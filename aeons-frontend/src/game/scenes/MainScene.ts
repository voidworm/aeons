import Phaser from 'phaser';
import { PlayableCard } from '../objects/PlayableCard';
import { LocationCard } from '../objects/LocationCard';
import { Card } from '../objects/Card';
import { connect } from '../../net/socket';
import { CreatureCard } from '../objects/CreatureCard';
import { HarvestableCard } from '../objects/HarvestableCard';
import type { CreatureEventDTO, GameStateDTO } from '../objects/GameState';
import { PlayerCard } from '../objects/PlayerCard';
import { HandContainer } from '../objects/HandContainer';
import { PlayerSwitcher } from '../objects/PlayerSwitcher';
import { LocationConnector } from '../objects/LocationConnector';
import { ControlButton } from '../objects/ControlButton';

export class MainScene extends Phaser.Scene {
  private static readonly CARD_WIDTH = 252; // rendered card width
  private static readonly MARGIN = 50;
  private static readonly CARD_HEIGHT = 352;

  private static readonly GRID_GAP = 75;
  private static readonly GRID_COLUMNS = 4;

  private handContainer!: HandContainer;
  private playerSwitcher!: PlayerSwitcher;
  private creaturePhaseButton!: ControlButton;
  private gameStarted = false;
  private creaturePhaseRunning = false;

  private playableCards: PlayableCard[] = [];
  private locationCards: LocationCard[] = [];
  private harvestableCards: HarvestableCard[] = [];
  private creatureCards: CreatureCard[] = [];
  private players: PlayerCard[] = [];

  private inDragging: boolean = false;
  private uiCamera!: Phaser.Cameras.Scene2D.Camera; // fixed camera for the hand and player switcher; the main camera pans over the board
  private boardBounds: Phaser.Geom.Rectangle | null = null;
  private static readonly ZOOM_MIN = 0.4;
  private static readonly ZOOM_MAX = 1.5;
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
    this.scale.on(Phaser.Scale.Events.RESIZE, this.onResize, this);

    this.handContainer = new HandContainer(this);
    this.playerSwitcher = new PlayerSwitcher(this, () => this.switchActivePlayer());
    this.creaturePhaseButton = new ControlButton(this, 'Trigger creature phase', () =>
      this.triggerCreaturePhase(),
    );
    this.uiCamera = this.cameras.add();
    this.initPanning();

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
        case 'poc_reply':
          this.send('poc_ack');
          break;
        case 'error':
          console.error(msg.payload);
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
    this.creaturePhaseButton.resize(gameSize.width, gameSize.height);
  }

  factor(): integer {
    return Math.min(1, this.scale.width / 1920, this.scale.height / 1080);
  }

  update(_timer: number, _delta: number) {
    this.connector?.redraw();
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
    const factor = this.factor();
    for (const [i, location] of inputGameState.locations.entries()) {
      const { x, y } = this.gridSlot(i, factor);

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
    this.hideFrom(this.cameras.main, this.creaturePhaseButton);
  }

  //TODO: sync health etc. onto the existing cards, once the creature phase animation is done
  private handleGameStateUpdate(_state: GameStateDTO) {}

  private triggerCreaturePhase() {
    if (!this.gameStarted || this.creaturePhaseRunning) return;
    this.creaturePhaseRunning = true;
    this.send('trigger_creature_phase');
  }

  //one creature after another, each waits for the previous one to arrive
  private async playCreaturePhase(events: CreatureEventDTO[]) {
    for (const event of events) {
      await this.moveCreature(event);
    }
    this.creaturePhaseRunning = false;
  }

  private moveCreature(event: CreatureEventDTO): Promise<void> {
    const card = this.creatureCards.find((c) => c.id === event.creatureId);
    const from = this.locationCards.find((l) => l.id === event.fromLocationId);
    const to = this.locationCards.find((l) => l.id === event.toLocationId);
    if (!card || !from || !to || from === to) return Promise.resolve();

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
      if (over.length > 0 || this.handContainer.containsPoint(pointer.x, pointer.y)) return;
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
    this.input.on('pointerup', () => (this.panLast = null));
    this.input.on('pointerupoutside', () => (this.panLast = null));
  }

  //zooms the main camera towards the pointer, keeping the board point under it in place
  private zoomAt(pointer: Phaser.Input.Pointer, deltaY: number) {
    if (this.handContainer.containsPoint(pointer.x, pointer.y)) return;
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
  }

  updateActivePCharacterTo(ident: number) {
    for (const player of this.players) {
      player.setActivePlayer(player.id === ident);
    }
    this.handContainer.setPlayerActive(ident);
    this.playerSwitcher.showActive(ident);
  }

  //once we want to highlight cards that are legal targets of a drag this needs to change
  onCardPointerOver(card: Card) {
    if (this.inDragging) return;
    this.children.bringToTop(card);
    this.children.queueDepthSort(); // bringToTop alone would put the card above higher-depth objects like the hand
    if (card instanceof LocationCard) card.raiseContents();
  }

  //disables the regular hover-response behaviour on other cards
  onCardDrag(card: Card) {
    this.inDragging = true;
    if (card instanceof PlayableCard) this.handContainer.startDrag(card);
  }

  //wiggles every enabled card under the pointer while a playable card is dragged
  onCardDragged(card: Card) {
    if (card.homeLocation) {
      card.homeLocation.constrainDrag(card);
      return;
    }
    if (card instanceof LocationCard) {
      card.syncContents();
      return;
    }
    if (!(card instanceof PlayableCard)) return;
    const { x, y } = this.boardPoint();
    for (const target of this.dropTargets()) {
      target.setDropTarget(target.getHitbox().contains(x, y));
    }
  }

  private dropTargets(): Card[] {
    return [...this.locationCards, ...this.creatureCards, ...this.harvestableCards];
  }

  onCardDrop(card: Card) {
    this.inDragging = false;
    this.dropTargets().forEach((t) => t.setDropTarget(false));
    if (!(card instanceof PlayableCard)) return;
    this.handContainer.returnCard(card);

    // a dragged hand card lives on the ui camera, so board cards never receive pointerover for it
    const { x, y } = this.boardPoint();
    const locationHighlighted = this.locationCards.find((loc) => loc.getHitbox().contains(x, y));
    if (locationHighlighted) {
      this.send('poc_hello', {
        played_card: card.name,
        targeted_location: locationHighlighted.name,
      });
    }
  }

  //temporary helper
  private stackAnchor(factor: number): { x: number; y: number } {
    const margin = MainScene.MARGIN * factor;
    const halfWidth = (MainScene.CARD_WIDTH * factor) / 2;
    const halfHeight = (MainScene.CARD_HEIGHT * factor) / 2;

    return { x: margin + halfWidth, y: this.scale.height - margin - halfHeight };
  }

  //grid for initial location placement
  private gridSlot(index: number, factor: number): { x: number; y: number } {
    const margin = MainScene.MARGIN * factor;
    const gap = MainScene.GRID_GAP * factor;
    const cardWidth = LocationCard.LOC_WIDTH * factor;
    const cardHeight = LocationCard.LOC_HEIGHT * factor;
    // contained cards hang below each location
    const hangHeight = (Card.HEIGHT - LocationCard.OVERLAP) * factor;

    const cols = MainScene.GRID_COLUMNS;
    const stepX = cardWidth + gap;
    const stepY = cardHeight + hangHeight + gap;
    const left = (this.scale.width - (cols * stepX - gap)) / 2;

    return {
      x: left + cardWidth / 2 + (index % cols) * stepX,
      y: margin + cardHeight / 2 + Math.floor(index / cols) * stepY,
    };
  }
}
