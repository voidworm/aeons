import Phaser from 'phaser';
import { PlayableCard } from '../objects/PlayableCard';
import { KindredCard } from '../objects/KindredCard';
import { LocationCard } from '../objects/LocationCard';
import { Card } from '../objects/Card';
import { connect } from '../../net/socket';
import { CreatureCard } from '../objects/CreatureCard';
import { NodeCard } from '../objects/NodeCard';
import { GatherOptionCard } from '../objects/GatherOptionCard';
import { ControlButton } from '../objects/ControlButton';
import {
  EFFECT_PLAYER_ATTACK,
  EFFECT_PLAYER_BOND,
  EFFECT_KINDRED_MOVE,
  EFFECT_PLAYER_DISTRACT,
  EFFECT_PLAYER_GATHER,
  EFFECT_PLAYER_MOVE,
} from '../objects/GameState';
import type {
  AwakeningEventDTO,
  CreatureDTO,
  CreatureEventDTO,
  GameStateDTO,
  LocationDTO,
  NodeDTO,
  GatherResultDTO,
  PlayableCardDTO,
  Phase,
  PlayerDTO,
} from '../objects/GameState';
import { PlayerCard } from '../objects/PlayerCard';
import { HandContainer } from '../objects/HandContainer';
import { PlayerSwitcher } from '../objects/PlayerSwitcher';
import { EventLog } from '../objects/EventLog';
import { GatherResultPopup } from '../objects/GatherResultPopup';
import { FaintedPopup } from '../objects/FaintedPopup';
import { LocationConnector } from '../objects/LocationConnector';
import { DragGroup } from '../objects/DragGroup';
import { layoutLocations } from '../layout/locationLayout';
import { uiScale } from '../layout/uiScale';
import { TargetingArrow } from '../objects/TargetingArrow';
import { CURSOR_DEFAULT, CURSOR_GRAB, CURSOR_INTERACTIVE } from '../cursors';

export class MainScene extends Phaser.Scene {
  private static readonly CARD_WIDTH = 252; // rendered card width
  private static readonly MARGIN = 50;
  private static readonly DEBUG_BUTTON_WIDTH = 300;
  private static readonly CARD_HEIGHT = 352;

  private handContainer!: HandContainer;
  private playerSwitcher!: PlayerSwitcher;
  private eventLog!: EventLog;
  private gameStarted = false;
  private phase: Phase = 'spirit';
  private phaseRunning = false; // a world or awakening phase is being animated
  private splits: GatherOptionCard[] = []; // stand in for nodes that offer both harvesting and foraging, from drag start until the cast ends
  private dragSuppressed: Card[] = []; // cards disabled for the duration of a drag
  private pendingState: GameStateDTO | null = null; // an update that arrived mid world phase

  private playableCards: PlayableCard[] = [];
  private locationCards: LocationCard[] = [];
  private nodeCards: NodeCard[] = [];
  private creatureCards: CreatureCard[] = [];
  private players: PlayerCard[] = [];
  private playerStates: PlayerDTO[] = []; // latest server state, feeds the stats panel

  private static readonly SPLIT_START_SCALE = 0.3; // relative to the node, the halves start this small
  private static readonly SPLIT_DURATION = 400; // ms the halves take to grow out of the node
  private static readonly SUMMON_HOLD = 600; // an untargeted card rests at the summon position before it resolves

  private inDragging: boolean = false;
  private targeting: {
    card: PlayableCard;
    legal: Card[];
    arrow: TargetingArrow;
    pressed: boolean; // a click started while targeting, so the release of the card drag doesn't count
    hovered: Card | null; // the legal target under the pointer, raised above its neighbours
  } | null = null;
  private uiCamera!: Phaser.Cameras.Scene2D.Camera; // fixed camera for the hand and player switcher; the main camera pans over the board
  private boardBounds: Phaser.Geom.Rectangle | null = null;
  private static readonly ZOOM_MIN = 0.4;
  private static readonly ZOOM_MAX = 1.5;
  private static readonly ZOOM_DEFAULT = 0.65;
  private static readonly INTRO_ZOOM_START = 0.4;
  private static readonly INTRO_ZOOM_DURATION = 1500; // ms
  private static readonly LAYOUT_SIZE = 1500; // the frame the locations settle in; the smaller, the closer they sit
  private static readonly WORLD_SIZE = 2000; // the map is laid out as if the screen were this big, whatever it really is
  private static readonly ZOOM_SPEED = 0.001; // per wheel delta unit

  private zoomDebug!: Phaser.GameObjects.Text;
  private debugButtons: ControlButton[] = []; // playtest controls under the zoom readout
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
      () => this.onPhaseButton(),
    );
    this.eventLog = new EventLog(this);
    this.placeEventLog();
    new DragGroup(this, [this.eventLog, ...this.playerSwitcher.phaseObjects]).enableOn(
      this.eventLog.dragHandle,
    );
    this.uiCamera = this.cameras.add();
    this.initPanning();
    this.initZoomDebug();
    this.initDebugButtons();
    this.initTargetingInput();

    this.socket = connect((msg) => {
      switch (msg.type) {
        case 'initial_game_state':
          this.handleGameStart(msg.payload);
          break;
        case 'updated_game_state':
          this.handleGameStateUpdate(msg.payload);
          break;
        case 'phase_changed':
          this.setPhase(msg.payload.phase, msg.payload.cycle);
          break;
        case 'world_phase':
          void this.playWorldPhase(msg.payload.events);
          break;
        case 'awakening_phase':
          void this.playAwakeningPhase(msg.payload.events);
          break;
        case 'log':
          this.eventLog.append(msg.payload.entries);
          break;
        case 'gather_result':
          this.showGatherResult(msg.payload);
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
      ...this.playableCards.filter((c) => !c.parentContainer), // cards in a hand are scaled with the hand
      ...this.locationCards,
      ...this.players,
      ...this.nodeCards,
      ...this.creatureCards,
    ]) {
      card.setScale(factor);
      if (card.floating || this.locationCards.includes(card as LocationCard)) continue; // locations sit at fixed world positions
      card.setPosition(card.x * ratioX, card.y * ratioY);
    }
    for (const location of this.locationCards) location.syncContents();
    this.updateBoardBounds();
    this.clampScroll();

    this.handContainer.resize(gameSize.width, gameSize.height);
    this.playerSwitcher.resize(gameSize.width, gameSize.height);
    this.placeEventLog();
    this.zoomDebug.setX(gameSize.width - 12);
    this.placeDebugButtons();
  }

  private initZoomDebug() {
    this.zoomDebug = this.add
      .text(this.scale.width - 12, 12, '', {
        fontFamily: 'monospace',
        fontSize: '18px',
        color: '#ffffff',
      })
      .setOrigin(1, 0)
      .setDepth(1000);
    this.hideFrom(this.cameras.main, this.zoomDebug);
  }

  private initDebugButtons() {
    this.debugButtons = [
      this.createDebugButton('World phase', () => this.triggerDebugPhase('trigger_world_phase')),
      this.createDebugButton('Awakening phase', () => this.triggerDebugPhase('trigger_awakening_phase')),
    ];
    for (const button of this.debugButtons) {
      button.setDepth(1000);
      this.add.existing(button);
      this.hideFrom(this.cameras.main, button);
    }
    this.placeDebugButtons();
  }

  private createDebugButton(text: string, onClick: () => void): ControlButton {
    return new ControlButton(this, text, onClick, MainScene.DEBUG_BUTTON_WIDTH, ControlButton.DEBUG_PALETTE);
  }

  private placeDebugButtons() {
    const { HEIGHT } = ControlButton;
    const { DEBUG_BUTTON_WIDTH } = MainScene;
    const scale = this.factor();
    const top = 12 + this.zoomDebug.height + 8; // right below the zoom readout
    this.debugButtons.forEach((button, i) => {
      button.setScale(scale);
      button.setPosition(
        this.scale.width - 12 - (DEBUG_BUTTON_WIDTH * scale) / 2,
        top + (HEIGHT * scale) / 2 + i * (HEIGHT * scale + 8),
      );
    });
  }

  private placeEventLog() {
    this.eventLog.setScale(this.factor());
    const { x, y } = this.playerSwitcher.eventLogBottomCentre();
    this.eventLog.setBottomCentre(x, y);
  }

  factor(): integer {
    return uiScale(this.scale.width, this.scale.height);
  }

  update(_timer: number, _delta: number) {
    this.zoomDebug.setText(
      `zoom ${Math.round(this.cameras.main.zoom * 100)}% · ${Math.round(this.game.loop.actualFps)} fps`,
    );
    this.connector?.redraw();
    this.updateTargeting();
    if (this.inDragging || this.panLast) this.game.canvas.style.cursor = CURSOR_GRAB; // phaser swaps it on every hover change
  }

  handleCardDragStart(_card: Card) {}

  initHands(inputGameState: GameStateDTO) {
    for (const card of inputGameState.cardsInHand) this.createPlayableCard(card);
  }

  private createPlayableCard(dto: PlayableCardDTO): PlayableCard {
    const create =
      dto.effectType === EFFECT_KINDRED_MOVE ? KindredCard.fromDto : PlayableCard.fromDto;
    const c = create(this, 0, 0, dto);
    c.on('cardDragEnd', this.onCardDrop, this);
    c.on('cardPointerOver', this.onCardPointerOver, this);
    c.on('cardDragStart', this.onCardDrag, this);
    c.on('cardDragged', this.onCardDragged, this);
    this.playableCards.push(c);
    return c;
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

  initNodes(inputGameState: GameStateDTO) {
    for (const node of inputGameState.nodes) this.spawnNode(node);
  }

  initCreatures(inputGameState: GameStateDTO) {
    for (const creature of inputGameState.creatures) this.spawnCreature(creature);
  }

  private spawnNode(dto: NodeDTO): NodeCard {
    const { x, y } = this.stackAnchor(this.factor());
    const c = NodeCard.fromDto(this, x, y, dto);
    this.prepareBoardCard(c);
    this.nodeCards.push(c);
    return c;
  }

  private spawnCreature(dto: CreatureDTO): CreatureCard {
    const { x, y } = this.stackAnchor(this.factor());
    const c = CreatureCard.fromDto(this, x, y, dto);
    this.prepareBoardCard(c);
    this.creatureCards.push(c);
    return c;
  }

  private prepareBoardCard(c: Card) {
    c.setScale(this.factor());
    this.hideFromUiCamera(c);
    c.on('cardDragEnd', this.onCardDrop, this);
    c.on('cardPointerOver', this.onCardPointerOver, this);
    c.on('cardDragStart', this.onCardDrag, this);
    c.on('cardDragged', this.onCardDragged, this);
  }

  //a card at a location the players can't see any more is gone; one at a newly visible location appears there
  private placeAtLocation(card: Card, locationId: number) {
    const to = this.locationCards.find((l) => l.id === locationId);
    if (!to) return;
    to.addContent(card);
    to.raiseContents();
  }

  private removeCreatureCards(state: GameStateDTO) {
    const shown = new Set(state.creatures.map((c) => c.id));
    for (const card of this.creatureCards.filter((c) => !shown.has(c.id))) {
      card.homeLocation?.removeContent(card);
      card.destroy();
    }
    this.creatureCards = this.creatureCards.filter((c) => shown.has(c.id));
  }

  initFirstActivePlayer() {
    this.players[0].setActivePlayer(true);
    this.updateActivePCharacterTo(this.players[0].id);
  }

  handleGameStart(inputGameState: GameStateDTO) {
    this.gameStarted = true;
    this.setPhase(inputGameState.phase, inputGameState.cycle);
    this.playerStates = inputGameState.players;
    this.initHands(inputGameState);
    this.initPlayers(inputGameState);
    this.initLocations(inputGameState);
    this.initCreatures(inputGameState);
    this.initNodes(inputGameState);
    this.assignToLocations(inputGameState);
    this.initFirstActivePlayer();
    this.hideUiFromMainCamera();
    this.updateBoardBounds();
    this.centreCameraOnActivePlayer();
  }

  //the camera starts on the active player
  private centreCameraOnActivePlayer() {
    const cam = this.cameras.main;
    const centre = MainScene.WORLD_SIZE / 2;
    const target = this.activePlayerCard() ?? { x: centre, y: centre };
    cam.setZoom(MainScene.INTRO_ZOOM_START);
    cam.scrollX = target.x - cam.width / 2;
    cam.scrollY = target.y - cam.height / 2;
    this.tweens.add({
      targets: cam,
      zoom: MainScene.ZOOM_DEFAULT,
      duration: MainScene.INTRO_ZOOM_DURATION,
      ease: 'Linear',
    });
  }

  //camera filters are per object, so everything the hand/switcher own has to be flagged once it exists
  private hideUiFromMainCamera() {
    this.hideFrom(this.cameras.main, this.handContainer);
    this.hideFrom(this.cameras.main, this.playerSwitcher);
    this.hideFrom(this.cameras.main, this.eventLog);
  }

  //TODO: sync health etc. onto the existing cards too
  //the update arrives right behind the world phase events, so it waits until the animation is done
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
    this.syncHands(state);
    for (const dto of state.cardsInHand) {
      this.playableCards.find((c) => c.id === dto.id)?.setDisabled(!dto.canBeCast);
    }
    for (const dto of state.locations) {
      this.locationCards.find((l) => l.id === dto.id)?.setVisibility(dto.visibility);
    }
    this.removeCreatureCards(state);
    for (const dto of state.creatures) {
      let card = this.creatureCards.find((c) => c.id === dto.id);
      if (!card) {
        card = this.spawnCreature(dto);
        this.placeAtLocation(card, dto.locationId);
      }
      card.setHealth(dto.health);
      void card?.setExhausted(dto.exhausted);
      const owner = state.players.find((p) => p.id === dto.kindredOf);
      card?.setKindredOf(dto.kindredOf, owner?.name ?? '');
      const to = this.locationCards.find((l) => l.id === dto.locationId);
      if (card && to) void this.moveCardToLocation(card, to);
    }
    for (const dto of state.nodes) {
      let card = this.nodeCards.find((n) => n.id === dto.id);
      if (card) {
        card.sync(dto);
      } else {
        card = this.spawnNode(dto);
        this.placeAtLocation(card, dto.locationId);
      }
    }
    for (const dto of state.players) {
      const card = this.players.find((p) => p.id === dto.id);
      card?.setDisabled(dto.fainted);
      const to = this.locationCards.find((l) => l.id === dto.locationId);
      if (card && to) void this.moveCardToLocation(card, to);
    }
  }

  //cards come and go with the server's hands, e.g. Bond is swapped for Move Kindred
  private syncHands(state: GameStateDTO) {
    const handCardIds = new Set(state.players.flatMap((p) => p.hand));
    for (const card of this.playableCards.filter((c) => !handCardIds.has(c.id))) {
      const owner = this.players.find((p) => p.cardsInHand.includes(card.id));
      this.tweens.killTweensOf(card);
      if (owner) this.handContainer.removeCard(owner.id, card, true);
      this.playableCards = this.playableCards.filter((c) => c !== card);
    }
    for (const dto of state.cardsInHand) {
      if (this.playableCards.some((c) => c.id === dto.id)) continue;
      const owner = state.players.find((p) => p.hand.includes(dto.id));
      if (!owner) continue;
      const card = this.createPlayableCard(dto);
      this.handContainer.addCard(owner.id, card);
      this.hideFrom(this.cameras.main, card);
    }
    for (const dto of state.players) {
      const player = this.players.find((p) => p.id === dto.id);
      if (player) player.cardsInHand = dto.hand;
    }
  }

  private showGatherResult(result: GatherResultDTO) {
    const node = this.nodeCards.find((n) => n.id === result.nodeId);
    const popup = new GatherResultPopup(this, node?.name ?? '', result.mode);
    this.hideFrom(this.cameras.main, popup);
  }

  private showFainted(playerId: number) {
    const player = this.players.find((p) => p.id === playerId);
    this.hideFrom(this.cameras.main, new FaintedPopup(this, player?.name ?? ''));
  }

  //the phase button advances the game by one step, the board stays usable while it waits for the next click
  private onPhaseButton() {
    if (!this.gameStarted || this.phaseRunning) return;
    switch (this.phase) {
      case 'spirit':
        this.setPhaseRunning(true);
        this.send('end_spirit_phase');
        break;
      case 'world':
        this.setPhaseRunning(true);
        this.send('start_awakening_phase');
        break;
      case 'awakening':
        this.send('start_spirit_phase'); // nothing to animate, the update arrives right away
        break;
    }
  }

  private triggerDebugPhase(type: 'trigger_world_phase' | 'trigger_awakening_phase') {
    if (!this.gameStarted || this.phaseRunning) return;
    this.setPhaseRunning(true);
    this.send(type);
  }

  private setPhase(phase: Phase, cycle: number) {
    this.phase = phase;
    this.playerSwitcher.setPhase(phase, cycle);
  }

  private setPhaseRunning(running: boolean) {
    this.phaseRunning = running;
    this.playerSwitcher.setPhaseButtonEnabled(!running);
  }

  //one creature after another, each waits for the previous one to arrive and, if it exhausted, to turn
  private async playWorldPhase(events: CreatureEventDTO[]) {
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
  private async playAwakeningPhase(events: AwakeningEventDTO[]) {
    for (const event of events) {
      await this.creatureCards.find((c) => c.id === event.creatureId)?.setExhausted(false);
    }
    this.finishPhase();
  }

  private finishPhase() {
    this.setPhaseRunning(false);
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
      ...this.nodeCards,
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

  //links every player/creature/node to the location it holds the id of
  private assignToLocations(inputGameState: GameStateDTO) {
    for (const location of this.locationCards) {
      const here = (dto: { locationId: number }) => dto.locationId === location.id;
      const cardsOf = <T extends Card>(cards: T[], dtos: { id: number; locationId: number }[]) =>
        dtos.filter(here).flatMap((dto) => cards.filter((c) => c.id === dto.id));

      location.setContents([
        ...cardsOf(this.players, inputGameState.players),
        ...cardsOf(this.creatureCards, inputGameState.creatures),
        ...cardsOf(this.nodeCards, inputGameState.nodes),
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
    else if (this.hasTargetRestriction(card)) {
      this.suppressAllButLegalTargets(card);
      if (card.effectType === EFFECT_PLAYER_GATHER)
        this.splits = this.splitNodes(this.legalTargets(card));
    }
  }

  //an unplayable card has no valid target, so the whole board is veiled while it is dragged
  private suppressBoard() {
    this.dragSuppressed = [
      ...this.locationCards,
      ...this.players,
      ...this.creatureCards,
      ...this.nodeCards,
    ];
    this.dragSuppressed.forEach((c) => c.setSuppressed(true));
  }

  private hasTargetRestriction(card: PlayableCard): boolean {
    return [
      EFFECT_PLAYER_MOVE,
      EFFECT_PLAYER_DISTRACT,
      EFFECT_PLAYER_ATTACK,
      EFFECT_PLAYER_GATHER,
      EFFECT_PLAYER_BOND,
      EFFECT_KINDRED_MOVE,
    ].includes(card.effectType);
  }

  //while a move, distract, attack, bond or gather card is dragged only its legal targets stay enabled
  private suppressAllButLegalTargets(dragged: PlayableCard) {
    const legal = this.legalTargets(dragged);
    this.dragSuppressed = [
      ...this.playableCards,
      ...this.locationCards,
      ...this.players,
      ...this.creatureCards,
      ...this.nodeCards,
    ].filter((c) => c !== dragged && !legal.includes(c));
    this.dragSuppressed.forEach((c) => c.setSuppressed(true));
    const mover =
      dragged.effectType === EFFECT_KINDRED_MOVE ? this.activeKindred() : this.activePlayerCard();
    if (dragged.effectType === EFFECT_PLAYER_MOVE || dragged.effectType === EFFECT_KINDRED_MOVE) {
      if (mover?.homeLocation) this.connector?.focusOn(mover.homeLocation);
    }
  }

  private activePlayerCard(): PlayerCard | undefined {
    return this.players.find((p) => p.activePlayer);
  }

  //the creature that is kindred to the active player
  private activeKindred(): CreatureCard | undefined {
    const id = this.activePlayerCard()?.id;
    return this.creatureCards.find((c) => c.kindredOf === id);
  }

  private legalTargets(card: PlayableCard): Card[] {
    const here = this.players.find((p) => p.activePlayer)?.homeLocation;
    if (card.effectType === EFFECT_PLAYER_MOVE) {
      return this.locationCards.filter((l) => here?.connections.includes(l.id));
    }
    if (card.effectType === EFFECT_KINDRED_MOVE) {
      const kindredAt = this.activeKindred()?.homeLocation;
      return this.locationCards.filter((l) => kindredAt?.connections.includes(l.id));
    }
    if (card.effectType === EFFECT_PLAYER_GATHER) {
      return this.nodeCards.filter((n) => n.homeLocation === here && n.modes.length > 0);
    }
    return this.creatureCards.filter((c) => c.homeLocation === here && c.kindredOf === null);
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
    this.splits.forEach((s) => s.destroy());
    this.splits = [];
    this.nodeCards.forEach((n) => n.setVisible(true));
    this.dragSuppressed.forEach((c) => c.setSuppressed(false));
    this.dragSuppressed = [];
    this.connector?.clearFocus();
  }

  private dropTargets(): Card[] {
    return [...this.locationCards, ...this.creatureCards, ...this.nodeCards, ...this.splits];
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

  //the card is put down where it was dropped; targeted cards then wait for a click, the others resolve right away
  private async summon(card: PlayableCard) {
    card.setPointerEnabled(false);
    const placed = card.playAnimation();

    if (this.hasTargetRestriction(card)) {
      this.startTargeting(card);
      return;
    }
    await placed;
    await new Promise((resolve) => this.time.delayedCall(MainScene.SUMMON_HOLD, resolve));
    this.send('poc_hello', { played_card: card.name });
    this.finishSummon(card);
  }

  private startTargeting(card: PlayableCard) {
    const arrow = new TargetingArrow(this);
    this.hideFrom(this.cameras.main, arrow);
    arrow.setVisible(false); // shown once the pointer moves
    const legal = this.legalTargets(card);
    const splitNodes = new Set(this.splits.map((s) => s.id));
    this.targeting = {
      card,
      legal: [...legal.filter((t) => !splitNodes.has(t.id)), ...this.splits],
      arrow,
      pressed: false,
      hovered: null,
    };
  }

  //a node that offers harvesting and foraging splits into one card each, side by side where the node stands;
  //the halves grow out of the node
  private splitNodes(legal: Card[]): GatherOptionCard[] {
    const nodes = legal.filter((t): t is NodeCard => t instanceof NodeCard && t.modes.length > 1);
    return nodes.flatMap((node) => {
      const gap = (Card.WIDTH * node.scaleX) / 2 + 8;
      return node.modes.map((mode, i) => {
        const half = new GatherOptionCard(this, node.x, node.y, node.dto, mode);
        half.setScale(MainScene.SPLIT_START_SCALE * node.scaleX);
        half.setDepth(1500);
        this.hideFromUiCamera(half);
        node.setVisible(false);
        this.tweens.add({
          targets: half,
          x: node.x + (i === 0 ? -gap : gap),
          scale: node.scaleX,
          duration: MainScene.SPLIT_DURATION,
          ease: 'Back.easeOut',
        });
        return half;
      });
    });
  }

  //escape and clicks on empty board cancel, a click on a legal target confirms, anything else is ignored
  private initTargetingInput() {
    this.input.keyboard?.on('keydown-ESC', () => this.cancelTargeting());
    this.input.on('pointermove', () => this.targeting?.arrow.setVisible(true));
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
    const { x, y } = this.boardPoint();
    const under = legal.filter((t) => t.getHitbox().contains(x, y));
    const hovered = under.reduce<Card | null>(
      (top, t) => (top && this.children.getIndex(top) > this.children.getIndex(t) ? top : t),
      null,
    ); // of overlapping cards the one on top is the one the pointer is on
    for (const target of this.dropTargets()) target.setDropTarget(target === hovered);
    this.raiseHovered(hovered);
    if (hovered) arrow.runBehind(hovered, this.cameras.main, this.uiCamera);
    else arrow.runAbove(this.cameras.main);
    arrow.pointFromTo(
      card.textBoxCentre(),
      new Phaser.Math.Vector2(pointer.x, pointer.y),
      card.getShadowedBounds(),
    );
  }

  //the target under the pointer comes to the front, so neighbours overlapping it don't cover its outline
  private raiseHovered(hovered: Card | null) {
    if (!this.targeting || hovered === this.targeting.hovered) return;
    this.targeting.hovered = hovered;
    if (!hovered) return;
    this.children.bringToTop(hovered);
    this.children.queueDepthSort();
    if (hovered instanceof LocationCard) hovered.raiseContents();
  }

  private confirmTarget(target: Card) {
    const card = this.targeting!.card;
    const mode =
      target instanceof GatherOptionCard
        ? target.mode
        : target instanceof NodeCard
          ? target.modes[0]
          : undefined;
    this.send('play_card', { cardId: card.id, targetId: target.id, mode });
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
    card.cancelPlayAnimation();
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

  //connected locations end up close together, and the whole map is centred in the world
  private locationSlots(locations: LocationDTO[]): Map<number, { x: number; y: number }> {
    const footprint = {
      width: LocationCard.FOOTPRINT_WIDTH,
      height: LocationCard.FOOTPRINT_HEIGHT,
      anchorY: (LocationCard.LOC_HEIGHT - LocationCard.FOOTPRINT_HEIGHT) / 2, // the location sits in the upper part of its footprint
    };
    const layout = layoutLocations(locations, footprint, {
      width: MainScene.LAYOUT_SIZE,
      height: MainScene.LAYOUT_SIZE,
    });
    const centre = MainScene.WORLD_SIZE / 2;

    return new Map([...layout].map(([id, p]) => [id, { x: centre + p.x, y: centre + p.y }]));
  }
}
