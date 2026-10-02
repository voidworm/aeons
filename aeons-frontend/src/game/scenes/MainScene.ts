import Phaser from 'phaser';
import { PlayableCard } from '../objects/PlayableCard';
import { LocationCard } from '../objects/LocationCard';
import { Card } from '../objects/Card';
import { connect } from '../../net/socket';
import { CreatureCard } from '../objects/CreatureCard';
import { HarvestableCard } from '../objects/HarvestableCard';
import type { GameStateDTO } from '../objects/GameState';
import { PlayerCard } from '../objects/PlayerCard';
import { HandContainer } from '../objects/HandContainer';
import { PlayerSwitcher } from '../objects/PlayerSwitcher';

export class MainScene extends Phaser.Scene {
  private static readonly CARD_WIDTH = 252; // rendered card width
  private static readonly MARGIN = 50;
  private static readonly CARD_HEIGHT = 352;

  private static readonly GRID_GAP = 75;
  private static readonly GRID_COLUMNS = 4;

  private handContainer!: HandContainer;
  private playerSwitcher!: PlayerSwitcher;

  private playableCards: PlayableCard[] = [];
  private locationCards: LocationCard[] = [];
  private harvestableCards: HarvestableCard[] = [];
  private creatureCards: CreatureCard[] = [];
  private players: PlayerCard[] = [];

  private inDragging: boolean = false;

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

    this.socket = connect((msg) => {
      switch (msg.type) {
        case 'game_state':
          this.handleGameStart(msg.payload);
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

    this.handContainer.resize(gameSize.width, gameSize.height);
    this.playerSwitcher.resize(gameSize.width, gameSize.height);
  }

  factor(): integer {
    return Math.min(1, this.scale.width / 1920, this.scale.height / 1080);
  }

  update(_timer: number, _delta: number) {}

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
      c.on('cardPointerOver', this.onCardPointerOver, this);

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
      this.locationCards.push(c);
    }
  }

  initHarvestNodes(inputGameState: GameStateDTO) {
    const factor = this.factor();
    for (const harvestable of inputGameState.harvestNodes) {
      const { x, y } = this.stackAnchor(factor);

      const c = HarvestableCard.fromDto(this, x, y, harvestable);
      c.setScale(this.factor());
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
    this.initHands(inputGameState);
    this.initPlayers(inputGameState);
    this.initLocations(inputGameState);
    this.initCreatures(inputGameState);
    this.initHarvestNodes(inputGameState);
    this.initFirstActivePlayer();
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
  }

  //disables the regular hover-response behaviour on other cards
  onCardDrag(card: Card) {
    this.inDragging = true;
    if (card instanceof PlayableCard) this.handContainer.startDrag(card);
  }

  //wiggles every enabled card under the pointer while a playable card is dragged
  onCardDragged(card: Card) {
    if (!(card instanceof PlayableCard)) return;
    const { worldX, worldY } = this.input.activePointer;
    for (const target of this.dropTargets()) {
      target.setDropTarget(target.getHitbox().contains(worldX, worldY));
    }
  }

  private dropTargets(): Card[] {
    return [...this.locationCards, ...this.creatureCards, ...this.harvestableCards];
  }

  onCardDrop(card: Card) {
    this.inDragging = false;
    this.dropTargets().forEach((t) => t.setDropTarget(false));
    if (card instanceof PlayableCard) this.handContainer.returnCard(card);

    const locationHighlighted = this.locationCards.find((loc) => loc.isHovering);
    if (locationHighlighted) {
      this.send('poc_hello', {
        played_card: card.name,
        targeted_location: locationHighlighted.name,
      });
      locationHighlighted.isHovering = false;
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
    const cardWidth = MainScene.CARD_WIDTH * factor;
    const cardHeight = MainScene.CARD_HEIGHT * factor;

    const cols = MainScene.GRID_COLUMNS;
    const stepX = cardWidth + gap;
    const stepY = cardHeight + gap;
    const left = (this.scale.width - (cols * stepX - gap)) / 2;

    return {
      x: left + cardWidth / 2 + (index % cols) * stepX,
      y: margin + cardHeight / 2 + Math.floor(index / cols) * stepY,
    };
  }
}
