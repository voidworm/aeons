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
  private static readonly CARD_WIDTH = 315; // rendered card width
  private static readonly CARD_STEP = 157.5; // center-to-center distance (overlap)
  private static readonly ROW_STEP = 220;
  private static readonly MARGIN = 50;
  private static readonly CARD_HEIGHT = 440;

  private slot = 0;
  private row = 0;

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

  handleGameStart(inputGameState: GameStateDTO) {
    this.slot = 0;
    this.row = 0;
    const factor = this.factor();

    console.log(inputGameState);

    for (const card of inputGameState.playables) {
      const c = PlayableCard.fromDto(this, 0, 0, card);
      c.on('cardDragEnd', this.onCardDrop, this);
      c.on('cardPointerOver', this.onCardPointerOver, this);
      c.on('cardDragStart', this.onCardDrag, this);
      this.playableCards.push(c);
    }

    var firstPlayerSet = false;
    for (const player of inputGameState.players) {
      const { x, y } = this.nextSlot(factor);

      const c = PlayerCard.fromDto(this, x, y, player);
      c.setScale(this.factor());
      c.on('cardPointerOver', this.onCardPointerOver, this);

      if (!firstPlayerSet) {
        firstPlayerSet = true;
        c.setActivePlayer(true);
      }

      this.players.push(c);
    }
    this.playerSwitcher.setPlayers(this.players);

    for (const player of this.players) {
      for (const card of this.playableCards.filter((c) => player.cardsInHand.includes(c.id))) {
        this.handContainer.addCard(player.id, card);
      }
    }

    for (const creature of inputGameState.creatures) {
      const { x, y } = this.nextSlot(factor);

      const c = CreatureCard.fromDto(this, x, y, creature);
      c.setScale(this.factor());
      this.creatureCards.push(c);
      c.on('cardDragEnd', this.onCardDrop, this);
      c.on('cardPointerOver', this.onCardPointerOver, this);
      c.on('cardDragStart', this.onCardDrag, this);
    }
    for (const harvestable of inputGameState.harvestables) {
      const { x, y } = this.nextSlot(factor);

      const c = HarvestableCard.fromDto(this, x, y, harvestable);
      c.setScale(this.factor());
      this.harvestableCards.push(c);
      c.on('cardDragEnd', this.onCardDrop, this);
      c.on('cardPointerOver', this.onCardPointerOver, this);
      c.on('cardDragStart', this.onCardDrag, this);
    }
    for (const location of inputGameState.locations) {
      const { x, y } = this.nextSlot(factor);

      const c = LocationCard.fromDto(this, x, y, location);
      c.on('cardDragEnd', this.onCardDrop, this);
      c.on('cardPointerOver', this.onCardPointerOver, this);
      c.on('cardDragStart', this.onCardDrag, this);
      c.setScale(this.factor());
      this.locationCards.push(c);
    }

    const active = this.players.find((p) => p.activePlayer);
    if (active) this.updateActivePCharacterTo(active.id);
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
  onCardDrag(_card: Card) {
    this.inDragging = true;
  }

  onCardDrop(card: Card) {
    this.inDragging = false;

    const locationHighlighted = this.locationCards.find((loc) => loc.isHovering);
    if (locationHighlighted) {
      this.send('poc_hello', {
        played_card: card.name,
        targeted_location: locationHighlighted.name,
      });
      locationHighlighted.isHovering = false;
    }
  }

  private nextSlot(factor: number): { x: number; y: number } {
    const margin = MainScene.MARGIN * factor;
    const step = MainScene.CARD_STEP * factor + 10 * factor;
    const halfWidth = (MainScene.CARD_WIDTH * factor) / 2;
    const halfHeight = (MainScene.CARD_HEIGHT * factor) / 2;

    let x = margin + halfWidth + step * this.slot;
    if (x + halfWidth > this.scale.width) {
      this.slot = 0;
      this.row++;
      x = margin + halfWidth;
    }

    const y = margin + halfHeight + MainScene.ROW_STEP * factor * this.row;
    this.slot++;
    return { x, y };
  }
}
