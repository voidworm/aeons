import Phaser from 'phaser';
import { PlayableCard } from '../objects/PlayableCard';
import { LocationCard } from '../objects/LocationCard';
import { Card } from '../objects/Card';
import { connect } from '../../net/socket';
import { CreatureCard } from '../objects/CreatureCard';
import { HarvestableCard } from '../objects/HarvestableCard';
import {
  EFFECT_PLAYER_ATTACK,
  EFFECT_PLAYER_BOND,
  EFFECT_PLAYER_DISTRACT,
  EFFECT_PLAYER_HARVEST,
  EFFECT_PLAYER_MOVE,
} from '../objects/GameState';
import type {
  CreatureEventDTO,
  GameStateDTO,
  LocationDTO,
  HarvestResultDTO,
  PlayerDTO,
  UpkeepEventDTO,
} from '../objects/GameState';
import { PlayerCard } from '../objects/PlayerCard';
import { HandContainer } from '../objects/HandContainer';
import { PlayerSwitcher } from '../objects/PlayerSwitcher';
import { EventLog } from '../objects/EventLog';
import { HarvestResultPopup } from '../objects/HarvestResultPopup';
import { FaintedPopup } from '../objects/FaintedPopup';
import { LocationConnector } from '../objects/LocationConnector';
import { DragGroup } from '../objects/DragGroup';
import { layoutLocations } from '../layout/locationLayout';
import { TargetingArrow } from '../objects/TargetingArrow';
import { CURSOR_DEFAULT, CURSOR_GRAB, CURSOR_INTERACTIVE } from '../cursors';

export class MainScene extends Phaser.Scene {
  private static readonly CARD_WIDTH = 252; // rendered card width
  private static readonly MARGIN = 50;
  private static readonly CARD_HEIGHT = 352;

  private handContainer!: HandContainer;
  private playerSwitcher!: PlayerSwitcher;
  private eventLog!: EventLog;
  private gameStarted = false;
  private phaseRunning = false; // a creature or upkeep phase is being animated
  private dragSuppressed: Card[] = []; // cards disabled for the duration of a drag
  private pendingState: GameStateDTO | null = null; // an update that arrived mid creature phase

  private playableCards: PlayableCard[] = [];
  private locationCards: LocationCard[] = [];
  private harvestableCards: HarvestableCard[] = [];
  private creatureCards: CreatureCard[] = [];
  private players: PlayerCard[] = [];
  private playerStates: PlayerDTO[] = []; // latest server state, feeds the stats panel

  private static readonly SUMMON_DURATION = 300;
  private static readonly SUMMON_HOLD = 600; // an untargeted card rests at the summon position before it resolves

  private inDragging: boolean = false;
  private targeting: {
    card: PlayableCard;
    legal: Card[];
    arrow: TargetingArrow;
    pressed: boolean; // a click started while targeting, so the release of the card drag doesn't count
  } | null = null;
  private uiCamera!: Phaser.Cameras.Scene2D.Camera; // fixed camera for the hand and player switcher; the main camera pans over the board
  private boardBounds: Phaser.Geom.Rectangle | null = null;
  private static readonly ZOOM_MIN = 0.4;
  private static readonly ZOOM_MAX = 1.5;
  private static readonly ZOOM_INITIAL = 0.75;
  private static readonly ZOOM_SPEED = 0.001; // per wheel delta unit

  private connector: LocationConnector | null = null;
  private panLast: Phaser.Math.Vector2 | null = null;

  private socket!: WebSocket;

  constructor() {
    super('MainScene');
  }

  preload() {}

  async create() {
    this.events.once(Phaser.Scenes.Events.DESTROY, () => this.socket.close());
    this.input.setDefaultCursor(CURSOR_DEFAULT);
    this.scale.on(Phaser.Scale.Events.RESIZE, this.onResize, this);

    this.handContainer = new HandContainer(this);
    this.playerSwitcher = new PlayerSwitcher(
      this,
      () => this.switchActivePlayer(),
      () => this.triggerCreaturePhase(),
      () => this.triggerUpkeepPhase(),
    );
    this.eventLog = new EventLog(this);
    this.placeEventLog();
    new DragGroup(this, [this.eventLog, ...this.playerSwitcher.phaseButtons]).enableOn(
      this.eventLog.dragHandle,
    );
    this.uiCamera = this.cameras.add();
    this.cameras.main.setZoom(MainScene.ZOOM_INITIAL);
    this.initPanning();
    this.initTargetingInput();

    this.socket = connect((msg) => {
      switch (msg.type) {
        case 'initial_game_state':
          this.handleGameStart(msg.payload);
          break;
        case 'updated_game_state':
          this.handleGameStateUpdate(msg.payload);
          break;
        case 'creature_phase':
          void this.playCreaturePhase(msg.payload.events);
          break;
        case 'upkeep_phase':
          void this.playUpkeepPhase(msg.payload.events);
          break;
        case 'log':
          this.eventLog.append(msg.payload.entries);
          break;
        case 'harvest_result':
          this.showHarvestResult(msg.payload);
          break;
        case 'poc_reply':
          this.send('poc_ack');
          break;
        case 'error':
          console.error(msg.payload);
          this.eventLog.append([`Error: ${msg.payload}`]);
          break;
      }
    });

    this.socket.addEventListener('open', () => this.send('new_game', { level: 'basic' }));
  }

  send(type: string, payload?: unknown): void {
    this.socket.send(JSON.stringify({ type, payload }));
  }

  private onResize(
    gameSize: Phaser.Structs.Size,
    _baseSize: Phaser.Structs.Size,
    _displaySize: Phaser.Structs.Size,
    previousWidth: number,
    previousHeight: number,
  ): void {
    const ratioX = gameSize.width / previousWidth;
    const ratioY = gameSize.height / previousHeight;
    const factor = this.factor();

    for (const card of [
      ...this.playableCards,
      ...this.locationCards,
      ...this.players,
      ...this.harvestableCards,
      ...this.creatureCards,
    ]) {
      card.setScale(factor);
      if (card.floating) continue;
      card.setPosition(card.x * ratioX, card.y * ratioY);
    }
    for (const location of this.locationCards) location.syncContents();
    this.updateBoardBounds();
    this.clampScroll();

    this.handContainer.resize(gameSize.width, gameSize.height);
    this.playerSwitcher.resize(gameSize.width, gameSize.height);
    this.placeEventLog();
  }

  private placeEventLog() {
    const { x, y } = this.playerSwitcher.abovePhaseButtons();
    this.eventLog.setBottomCentre(x, y);
  }

  factor(): integer {
    return Math.min(1, this.scale.width / 1920, this.scale.height / 1080);
  }

  update(_timer: number, _delta: number) {
    this.connector?.redraw();
    this.updateTargeting();
    if (this.inDragging || this.panLast) this.game.canvas.style.cursor = CURSOR_GRAB; // phaser swaps it on every hover change
  }

  handleCardDragStart(_card: Card) {}

  initHands(inputGameState: GameStateDTO) {
    for (const card of inputGameState.cardsInHand) {
      const c = PlayableCard.fromDto(this, 0, 0, card);
      c.on('cardDragEnd', this.onCardDrop, this);
      c.on('cardPointerOver', this.onCardPointerOver, this);
      c.on('cardDragStart', this.onCardDrag, this);
      c.on('cardDragged', this.onCardDragged, this);
      this.playableCards.push(c);
    }
  }

  initPlayers(inputGameState: GameStateDTO) {
    const factor = this.factor();

    for (const player of inputGameState.players) {
      const { x, y } = this.stackAnchor(factor);

      const c = PlayerCard.fromDto(this, x, y, player);
      c.setScale(this.factor());
      c.on('cardDragEnd', this.onCardDrop, this);
      c.on('cardPointerOver', this.onCardPointerOver, this);
      c.on('cardDragStart', this.onCardDrag, this);
      c.on('cardDragged', this.onCardDragged, this);
      this.hideFromUiCamera(c);

      this.players.push(c);
    }
    this.playerSwitcher.setPlayers(this.players);

    for (const player of this.players) {
      for (const card of this.playableCards.filter((c) => player.cardsInHand.includes(c.id))) {
        this.handContainer.addCard(player.id, card);
      }
    }
  }

  initLocations(inputGameState: GameStateDTO) {
    const slots = this.locationSlots(inputGameState.locations);
    for (const location of inputGameState.locations) {
      const { x, y } = slots.get(location.id)!;

      const c = LocationCard.fromDto(this, x, y, location);
      c.on('cardDragEnd', this.onCardDrop, this);
      c.on('cardPointerOver', this.onCardPointerOver, this);
      c.on('cardDragStart', this.onCardDrag, this);
      c.on('cardDragged', this.onCardDragged, this);
      c.setScale(this.factor());
      this.hideFromUiCamera(c);
      this.locationCards.push(c);
    }
    this.connector = new LocationConnector(this, inputGameState.locations, this.locationCards);
    this.hideFrom(this.uiCamera, this.connector);
  }

  initHarvestNodes(inputGameState: GameStateDTO) {
    const factor = this.factor();
    for (const harvestable of inputGameState.harvestNodes) {
      const { x, y } = this.stackAnchor(factor);

      const c = HarvestableCard.fromDto(this, x, y, harvestable);
      c.setScale(this.factor());
      this.hideFromUiCamera(c);
      this.harvestableCards.push(c);
      c.on('cardDragEnd', this.onCardDrop, this);
      c.on('cardPointerOver', this.onCardPointerOver, this);
      c.on('cardDragStart', this.onCardDrag, this);
      c.on('cardDragged', this.onCardDragged, this);
    }
  }

  initCreatures(inputGameState: GameStateDTO) {
    const factor = this.factor();
    for (const creature of inputGameState.creatures) {
      const { x, y } = this.stackAnchor(factor);

      const c = CreatureCard.fromDto(this, x, y, creature);
      c.setScale(this.factor());
      this.hideFromUiCamera(c);
      this.creatureCards.push(c);
      c.on('cardDragEnd', this.onCardDrop, this);
      c.on('cardPointerOver', this.onCardPointerOver, this);
      c.on('cardDragStart', this.onCardDrag, this);
      c.on('cardDragged', this.onCardDragged, this);
    }
  }

  initFirstActivePlayer() {
    this.players[0].setActivePlayer(true);
    this.updateActivePCharacterTo(this.players[0].id);
  }

  handleGameStart(inputGameState: GameStateDTO) {
    this.gameStarted = true;
    this.playerStates = inputGameState.players;
    this.initHands(inputGameState);
    this.initPlayers(inputGameState);
    this.initLocations(inputGameState);
    this.initCreatures(inputGameState);
    this.initHarvestNodes(inputGameState);
    this.assignToLocations(inputGameState);
    this.initFirstActivePlayer();
    this.hideUiFromMainCamera();
  }

  //camera filters are per object, so everything the hand/switcher own has to be flagged once it exists
  private hideUiFromMainCamera() {
    this.hideFrom(this.cameras.main, this.handContainer);
    this.hideFrom(this.cameras.main, this.playerSwitcher);
    this.hideFrom(this.cameras.main, this.eventLog);
  }

  //TODO: sync health etc. onto the existing cards too
  //the update arrives right behind the creature phase events, so it waits until the animation is done
  private handleGameStateUpdate(state: GameStateDTO) {
    if (this.phaseRunning) {
      this.pendingState = state;
      return;
    }
    this.syncState(state);
  }

  private syncState(state: GameStateDTO) {
    this.playerStates = state.players;
    this.refreshStatsPanel();
    for (const dto of state.cardsInHand) {
      this.playableCards.find((c) => c.id === dto.id)?.setDisabled(!dto.canBeCast);
    }
    for (const dto of state.creatures) {
      const card = this.creatureCards.find((c) => c.id === dto.id);
      card?.setHealth(dto.health);
      void card?.setExhausted(dto.exhausted);
      card?.setBonded(dto.bonded);
    }
    for (const dto of state.harvestNodes) {
      this.harvestableCards.find((h) => h.id === dto.id)?.setYields(dto.nutrition, dto.resources);
    }
    for (const dto of state.players) {
      const card = this.players.find((p) => p.id === dto.id);
      card?.setDisabled(dto.fainted);
      const to = this.locationCards.find((l) => l.id === dto.locationId);
      if (card && to) void this.moveCardToLocation(card, to);
    }
  }

  private showHarvestResult(result: HarvestResultDTO) {
    const harvestable = this.harvestableCards.find((h) => h.id === result.harvestableId);
    const popup = new HarvestResultPopup(
      this,
      harvestable?.name ?? '',
      result.nutrition,
      result.resources,
    );
    this.hideFrom(this.cameras.main, popup);
  }

  private showFainted(playerId: number) {
    const player = this.players.find((p) => p.id === playerId);
    this.hideFrom(this.cameras.main, new FaintedPopup(this, player?.name ?? ''));
  }

  private triggerCreaturePhase() {
    if (!this.gameStarted || this.phaseRunning) return;
    this.phaseRunning = true;
    this.send('trigger_creature_phase');
  }

  private triggerUpkeepPhase() {
    if (!this.gameStarted || this.phaseRunning) return;
    this.phaseRunning = true;
    this.send('trigger_upkeep_phase');
  }

  //one creature after another, each waits for the previous one to arrive and, if it exhausted, to turn
  private async playCreaturePhase(events: CreatureEventDTO[]) {
    for (const event of events) {
      await this.moveCreature(event);
      if (event.exhausted)
        await this.creatureCards.find((c) => c.id === event.creatureId)?.setExhausted(true);
    }
    this.finishPhase();
    for (const attack of events.flatMap((e) => e.attacks).filter((a) => a.fainted)) {
      this.showFainted(attack.targetPlayerId);
    }
  }

  //exhausted creatures turn upright again, one after another
  private async playUpkeepPhase(events: UpkeepEventDTO[]) {
    for (const event of events) {
      await this.creatureCards.find((c) => c.id === event.creatureId)?.setExhausted(false);
    }
    this.finishPhase();
  }

  private finishPhase() {
    this.phaseRunning = false;
    if (this.pendingState) {
      this.syncState(this.pendingState);
      this.pendingState = null;
    }
  }

  private moveCreature(event: CreatureEventDTO): Promise<void> {
    const card = this.creatureCards.find((c) => c.id === event.creatureId);
    const to = this.locationCards.find((l) => l.id === event.toLocationId);
    if (!card || !to) return Promise.resolve();
    return this.moveCardToLocation(card, to);
  }

  //re-homes the card to the location and glides it over from where it stands now
  private moveCardToLocation(card: Card, to: LocationCard): Promise<void> {
    const from = card.homeLocation;
    if (!from || from === to) return Promise.resolve();

    const start = { x: card.x, y: card.y };
    from.removeContent(card);
    to.addContent(card); // snaps the card to its new slot, so we can read the destination
    const end = { x: card.x, y: card.y };
    card.setPosition(start.x, start.y);
    to.raiseContents();
    this.children.bringToTop(card); // travels above everything

    return new Promise((resolve) => {
      this.tweens.add({
        targets: card,
        x: end.x,
        y: end.y,
        duration: 500,
        ease: 'Sine.easeInOut',
        onComplete: () => resolve(),
      });
    });
  }

  private hideFromUiCamera(card: Card) {
    this.hideFrom(this.uiCamera, card);
  }

  private hideFrom(camera: Phaser.Cameras.Scene2D.Camera, object: Phaser.GameObjects.GameObject) {
    object.cameraFilter |= camera.id;
    if (object instanceof Phaser.GameObjects.Container) {
      for (const child of object.list) this.hideFrom(camera, child);
    }
  }

  //dragging empty space pans the board; the hand area and anything interactive don't start a pan
  private initPanning() {
    this.input.on('pointerdown', (pointer: Phaser.Input.Pointer, over: unknown[]) => {
      if (
        this.targeting ||
        over.length > 0 ||
        this.handContainer.containsPoint(pointer.x, pointer.y)
      )
        return;
      this.updateBoardBounds();
      this.panLast = new Phaser.Math.Vector2(pointer.x, pointer.y);
    });
    this.input.on('pointermove', (pointer: Phaser.Input.Pointer) => {
      if (!this.panLast) return;
      const cam = this.cameras.main;
      cam.scrollX -= (pointer.x - this.panLast.x) / cam.zoom;
      cam.scrollY -= (pointer.y - this.panLast.y) / cam.zoom;
      this.panLast.set(pointer.x, pointer.y);
      this.clampScroll();
    });
    this.input.on(
      'wheel',
      (pointer: Phaser.Input.Pointer, _over: unknown[], _dx: number, deltaY: number) =>
        this.zoomAt(pointer, deltaY),
    );
    this.input.on('pointerup', () => this.endPan());
    this.input.on('pointerupoutside', () => this.endPan());
  }

  private endPan() {
    if (!this.panLast) return;
    this.panLast = null;
    this.restoreCursor();
  }

  //zooms the main camera towards the pointer, keeping the board point under it in place
  private zoomAt(pointer: Phaser.Input.Pointer, deltaY: number) {
    if (this.handContainer.containsPoint(pointer.x, pointer.y)) return;
    if (this.eventLog.containsPoint(pointer.x, pointer.y)) return;
    const cam = this.cameras.main;
    const oldZoom = cam.zoom;
    const newZoom = Phaser.Math.Clamp(
      oldZoom * Math.exp(-deltaY * MainScene.ZOOM_SPEED),
      MainScene.ZOOM_MIN,
      MainScene.ZOOM_MAX,
    );
    // the camera zooms around its centre: world = (screen - centre) / zoom + scroll + centre
    const halfW = cam.width / 2;
    const halfH = cam.height / 2;
    const worldX = (pointer.x - halfW) / oldZoom + cam.scrollX + halfW;
    const worldY = (pointer.y - halfH) / oldZoom + cam.scrollY + halfH;

    cam.setZoom(newZoom);
    cam.scrollX = worldX - (pointer.x - halfW) / newZoom - halfW;
    cam.scrollY = worldY - (pointer.y - halfH) / newZoom - halfH;
    this.updateBoardBounds();
    this.clampScroll();
  }

  private updateBoardBounds() {
    const cards = [
      ...this.locationCards,
      ...this.players,
      ...this.creatureCards,
      ...this.harvestableCards,
    ];
    if (cards.length === 0) return;
    this.boardBounds = cards
      .map((c) => c.getHitbox())
      .reduce((union, r) => Phaser.Geom.Rectangle.Union(union, r, union), cards[0].getHitbox());
  }

  //the board is centred and usually fits the screen, so allow dragging its edges up to the screen centre
  private clampScroll() {
    if (!this.boardBounds) return;
    const cam = this.cameras.main;
    const { left, right, top, bottom } = this.boardBounds;
    cam.scrollX = Phaser.Math.Clamp(
      cam.scrollX,
      Math.min(0, left - cam.width / 2),
      Math.max(0, right - cam.width / 2),
    );
    cam.scrollY = Phaser.Math.Clamp(
      cam.scrollY,
      Math.min(0, top - cam.height / 2),
      Math.max(0, bottom - cam.height / 2),
    );
  }

  //pointer position in board coordinates, whichever camera the pointer is currently over
  private boardPoint(): Phaser.Math.Vector2 {
    const pointer = this.input.activePointer;
    return this.cameras.main.getWorldPoint(pointer.x, pointer.y);
  }

  //links every player/creature/harvestable to the location it holds the id of
  private assignToLocations(inputGameState: GameStateDTO) {
    for (const location of this.locationCards) {
      const here = (dto: { locationId: number }) => dto.locationId === location.id;
      const cardsOf = <T extends Card>(cards: T[], dtos: { id: number; locationId: number }[]) =>
        dtos.filter(here).flatMap((dto) => cards.filter((c) => c.id === dto.id));

      location.setContents([
        ...cardsOf(this.players, inputGameState.players),
        ...cardsOf(this.creatureCards, inputGameState.creatures),
        ...cardsOf(this.harvestableCards, inputGameState.harvestNodes),
      ]);
      location.raiseContents();
    }
  }

  switchActivePlayer() {
    const inactivePlayer = this.players.find((p) => !p.activePlayer)!;
    this.updateActivePCharacterTo(inactivePlayer.id);
    this.eventLog.append([`Switching to ${inactivePlayer.name}`]);
  }

  updateActivePCharacterTo(ident: number) {
    for (const player of this.players) {
      player.setActivePlayer(player.id === ident);
    }
    this.handContainer.setPlayerActive(ident);
    this.playerSwitcher.showActive(ident);
    this.refreshStatsPanel();
  }

  private refreshStatsPanel() {
    const active = this.players.find((p) => p.activePlayer);
    const state = this.playerStates.find((p) => p.id === active?.id);
    if (!state) return;
    const location = this.locationCards.find((l) => l.id === state.locationId);
    this.playerSwitcher.setStats({
      energy: state.energy,
      wounds: state.wounds,
      maxWounds: state.maxWounds,
      food: state.food,
      resources: state.resources,
      location: location?.name ?? '',
    });
  }

  //once we want to highlight cards that are legal targets of a drag this needs to change
  onCardPointerOver(card: Card) {
    if (this.inDragging || this.targeting) return;
    this.children.bringToTop(card);
    this.children.queueDepthSort(); // bringToTop alone would put the card above higher-depth objects like the hand
    if (card instanceof LocationCard) card.raiseContents();
  }

  private restoreCursor() {
    const hovering = this.input.hitTestPointer(this.input.activePointer).length > 0;
    this.game.canvas.style.cursor = hovering ? CURSOR_INTERACTIVE : CURSOR_DEFAULT;
  }

  //disables the regular hover-response behaviour on other cards
  onCardDrag(card: Card) {
    this.inDragging = true;
    if (!(card instanceof PlayableCard)) return;
    this.handContainer.startDrag(card);
    if (card.isDisabledByRules()) this.suppressBoard();
    else if (this.hasTargetRestriction(card)) this.suppressAllButLegalTargets(card);
  }

  //an unplayable card has no valid target, so the whole board is veiled while it is dragged
  private suppressBoard() {
    this.dragSuppressed = [
      ...this.locationCards,
      ...this.players,
      ...this.creatureCards,
      ...this.harvestableCards,
    ];
    this.dragSuppressed.forEach((c) => c.setSuppressed(true));
  }

  private hasTargetRestriction(card: PlayableCard): boolean {
    return [
      EFFECT_PLAYER_MOVE,
      EFFECT_PLAYER_DISTRACT,
      EFFECT_PLAYER_ATTACK,
      EFFECT_PLAYER_HARVEST,
      EFFECT_PLAYER_BOND,
    ].includes(card.effectType);
  }

  //while a move, distract, attack or harvest card is dragged only its legal targets stay enabled
  private suppressAllButLegalTargets(dragged: PlayableCard) {
    const legal = this.legalTargets(dragged);
    this.dragSuppressed = [
      ...this.playableCards,
      ...this.locationCards,
      ...this.players,
      ...this.creatureCards,
      ...this.harvestableCards,
    ].filter((c) => c !== dragged && !legal.includes(c));
    this.dragSuppressed.forEach((c) => c.setSuppressed(true));
    const here = this.players.find((p) => p.activePlayer)?.homeLocation;
    if (dragged.effectType === EFFECT_PLAYER_MOVE && here) this.connector?.focusOn(here);
  }

  private legalTargets(card: PlayableCard): Card[] {
    const here = this.players.find((p) => p.activePlayer)?.homeLocation;
    if (card.effectType === EFFECT_PLAYER_MOVE) {
      return this.locationCards.filter((l) => here?.connections.includes(l.id));
    }
    if (card.effectType === EFFECT_PLAYER_HARVEST) {
      return this.harvestableCards.filter((h) => h.homeLocation === here);
    }
    return this.creatureCards.filter((c) => c.homeLocation === here);
  }

  onCardDragged(card: Card) {
    if (card.homeLocation) {
      card.homeLocation.constrainDrag(card);
      return;
    }
    if (card instanceof LocationCard) {
      card.syncContents();
      return;
    }
  }

  private releaseSuppressed() {
    this.dragSuppressed.forEach((c) => c.setSuppressed(false));
    this.dragSuppressed = [];
    this.connector?.clearFocus();
  }

  private dropTargets(): Card[] {
    return [...this.locationCards, ...this.creatureCards, ...this.harvestableCards];
  }

  onCardDrop(card: Card) {
    this.inDragging = false;
    this.restoreCursor();
    if (!(card instanceof PlayableCard)) return;
    const pointer = this.input.activePointer;
    if (card.isDisabledByRules() || this.handContainer.containsPoint(pointer.x, pointer.y)) {
      this.handContainer.returnCard(card);
      this.releaseSuppressed();
      return;
    }
    void this.summon(card);
  }

  //the card glides above the event log; targeted cards then wait for a click, the others resolve right away
  private async summon(card: PlayableCard) {
    card.setPointerEnabled(false);
    const { x, y } = this.eventLog.summonPoint(card.cardHeight * card.scaleY);
    await new Promise<void>((resolve) => {
      this.tweens.add({
        targets: card,
        x,
        y,
        duration: MainScene.SUMMON_DURATION,
        ease: 'Sine.easeOut',
        onComplete: () => resolve(),
      });
    });

    if (this.hasTargetRestriction(card)) {
      this.startTargeting(card);
      return;
    }
    await new Promise((resolve) => this.time.delayedCall(MainScene.SUMMON_HOLD, resolve));
    this.send('poc_hello', { played_card: card.name });
    this.finishSummon(card);
  }

  private startTargeting(card: PlayableCard) {
    const arrow = new TargetingArrow(this);
    this.hideFrom(this.cameras.main, arrow);
    this.targeting = { card, legal: this.legalTargets(card), arrow, pressed: false };
  }

  //escape and clicks on empty board cancel, a click on a legal target confirms, anything else is ignored
  private initTargetingInput() {
    this.input.keyboard?.on('keydown-ESC', () => this.cancelTargeting());
    this.input.on('pointerdown', () => {
      if (this.targeting) this.targeting.pressed = true;
    });
    this.input.on('pointerup', (pointer: Phaser.Input.Pointer) => {
      if (!this.targeting?.pressed) return;
      this.targeting.pressed = false;
      const { x, y } = this.boardPoint();
      const target = this.targeting.legal.find((t) => t.getHitbox().contains(x, y));
      if (target) this.confirmTarget(target);
      else if (this.input.hitTestPointer(pointer).length === 0) this.cancelTargeting();
    });
  }

  private updateTargeting() {
    if (!this.targeting) return;
    const { card, legal, arrow } = this.targeting;
    const pointer = this.input.activePointer;
    arrow.pointFromTo(
      new Phaser.Math.Vector2(card.x, card.y - (card.cardHeight * card.scaleY) / 2),
      new Phaser.Math.Vector2(pointer.x, pointer.y),
    );
    const { x, y } = this.boardPoint();
    for (const target of this.dropTargets()) {
      target.setDropTarget(legal.includes(target) && target.getHitbox().contains(x, y));
    }
  }

  private confirmTarget(target: Card) {
    const card = this.targeting!.card;
    this.send('play_card', { cardId: card.id, targetId: target.id });
    this.finishSummon(card);
  }

  private cancelTargeting() {
    if (!this.targeting) return;
    this.finishSummon(this.targeting.card);
  }

  //ends the summon and sends the card back to its hand slot
  private finishSummon(card: PlayableCard) {
    this.targeting?.arrow.destroy();
    this.targeting = null;
    this.dropTargets().forEach((t) => t.setDropTarget(false));
    this.releaseSuppressed();
    card.setPointerEnabled(true);
    this.handContainer.returnCard(card, true);
  }

  //temporary helper
  private stackAnchor(factor: number): { x: number; y: number } {
    const margin = MainScene.MARGIN * factor;
    const halfWidth = (MainScene.CARD_WIDTH * factor) / 2;
    const halfHeight = (MainScene.CARD_HEIGHT * factor) / 2;

    return { x: margin + halfWidth, y: this.scale.height - margin - halfHeight };
  }

  //connected locations end up close together, and the whole map is centred on the screen
  private locationSlots(locations: LocationDTO[]): Map<number, { x: number; y: number }> {
    const factor = this.factor();
    const footprint = {
      width: LocationCard.FOOTPRINT_WIDTH,
      height: LocationCard.FOOTPRINT_HEIGHT,
      anchorY: (LocationCard.LOC_HEIGHT - LocationCard.FOOTPRINT_HEIGHT) / 2, // the location sits in the upper part of its footprint
    };
    const layout = layoutLocations(locations, footprint, this.scale.width / this.scale.height);

    return new Map(
      [...layout].map(([id, p]) => [
        id,
        { x: this.scale.width / 2 + p.x * factor, y: this.scale.height / 2 + p.y * factor },
      ]),
    );
  }
}
