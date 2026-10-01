import Phaser from 'phaser';
import { PlayableCard } from '../objects/PlayableCard';
import { LocationCard } from '../objects/LocationCard';
import type { Card } from '../objects/Card';
import { connect } from '../../net/socket';
import { CreatureCard } from '../objects/CreatureCard';
import { HarvestableCard } from '../objects/HarvestableCard';
import type { GameStateDTO } from '../objects/GameState';

export class MainScene extends Phaser.Scene {
  private playerCards: PlayableCard[] = [];
  private locationCards: LocationCard[] = [];
  private lastActionText!: Phaser.GameObjects.Text;
  private socket!: WebSocket;
  constructor() {
    super('MainScene');
  }

  preload() {}

  async create() {
    this.events.once(Phaser.Scenes.Events.DESTROY, () => this.socket.close());
    this.scale.on(Phaser.Scale.Events.RESIZE, this.onResize, this);

    this.lastActionText = this.add.text(this.scale.width / 2, 130, '...at long last, homebound.', {
      color: '#000000',
    });

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

    //commenting out generics for now
    /*
    var raccoonCard = new CreatureCard(
      this,
      this.scale.width / 2 + 200,
      this.scale.height - 400,
      'Raccoon',
      'A raccoon wandering around. What might he do next, steal your berries?',
    );
    raccoonCard.on('cardDragged', (card: CreatureCard) => this.handleCardDrag(card));
    raccoonCard.on('cardDragStart', (card: CreatureCard) => this.handleCardDragStart(card));
    raccoonCard.on('cardDragEnd', (card: CreatureCard) => this.handleCardDrop(card));
    this.playerCards.push(raccoonCard);

    var riteOfSeekingCard = new PlayableCard(
      this,
      this.scale.width / 2,
      this.scale.height,
      'Rite of Seeking',
      'Consult the spirits.\n\nIf you succeed, discover tracks at the location.',
      4,
    );
    riteOfSeekingCard.on('cardDragged', (card: PlayableCard) => this.handleCardDrag(card));
    riteOfSeekingCard.on('cardDragStart', (card: PlayableCard) => this.handleCardDragStart(card));
    riteOfSeekingCard.on('cardDragEnd', (card: PlayableCard) => this.handleCardDrop(card));

    this.playerCards.push(riteOfSeekingCard);

    var cleanseCard = new PlayableCard(
      this,
      this.scale.width / 2 + 200,
      this.scale.height,
      'Cleanse',
      'Cleanse all debuffs from target player.',
      1,
    );
    cleanseCard.on('cardDragged', (card: PlayableCard) => this.handleCardDrag(card));
    cleanseCard.on('cardDragStart', (card: PlayableCard) => this.handleCardDragStart(card));
    cleanseCard.on('cardDragEnd', (card: PlayableCard) => this.handleCardDrop(card));

    this.playerCards.push(cleanseCard);

    var berryBush = new HarvestableCard(
      this,
      this.scale.width / 2 + 400,
      this.scale.height,
      'Berry Bush',
      'Berries around here are known to be very tasty. Especially raccoons love them.',
    );
    berryBush.on('cardDragged', (card: HarvestableCard) => this.handleCardDrag(card));
    berryBush.on('cardDragStart', (card: HarvestableCard) => this.handleCardDragStart(card));
    berryBush.on('cardDragEnd', (card: HarvestableCard) => this.handleCardDrop(card));
    this.playerCards.push(berryBush);

    var secludedDen = new LocationCard(
      this,
      this.scale.width / 2 + 400,
      400,
      'Secluded Den',
      'The secluded den is home to your tribe.',
    );
    secludedDen.on('cardDragStart', (card: Card) => this.handleCardDragStart(card));
    this.locationCards.push(secludedDen);

    /*
    var eternalPlains = new LocationCard(
      this,
      this.scale.width / 2,
      400,
      'Eternal Plains',
      'Across the eternal plains....',
    );
    eternalPlains.on('cardDragStart', (card: Card) => this.handleCardDragStart(card));
    this.locationCards.push(eternalPlains);

    var windscarredCrag = new LocationCard(
      this,
      this.scale.width / 2 - 400,
      400,
      'Windscarred Crag',
      'Only the mind of a child could not see the danger this place is home to',
    );
    windscarredCrag.on('cardDragStart', (card: Card) => this.handleCardDragStart(card));
    this.locationCards.push(windscarredCrag);
    */
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
    const factor = Math.min(1, gameSize.width / 1920, gameSize.height / 1080);

    for (const card of [...this.playerCards, ...this.locationCards]) {
      card.setScale(factor);
      if (card.floating) continue;
      card.setPosition(card.x * ratioX, card.y * ratioY);
    }
  }

  update(_timer: number, _delta: number) {}

  handleCardDragStart(card: Card) {
    this.children.bringToTop(card);
  }

  handleGameStart(inputGameState: GameStateDTO) {
    for (const card of inputGameState.cards) {
      console.log(`Found card ${card.name}`);
    }
    for (const creature of inputGameState.creatures) {
      console.log(`Found card ${creature.name}`);
    }
    for (const harvestable of inputGameState.harvestables) {
      console.log(`Found card ${harvestable.name}`);
    }
    for (const location of inputGameState.locations) {
      console.log(`Found card ${location.name}`);
    }
    for (const player of inputGameState.players) {
      console.log(`Found card ${player.name}`);
    }
  }

  handleCardDrag(card: PlayableCard) {
    for (const location of this.locationCards) {
      const overlapping = Phaser.Geom.Intersects.RectangleToRectangle(
        card.getOverlapBounds(),
        location.getOverlapBounds(),
      );
      if (overlapping) location.onPointerOver();
      else location.onPointerOut();
    }
  }

  handleCardDrop(card: PlayableCard) {
    const locationHighlighted = this.locationCards.find((loc) => loc.isHovering);
    if (locationHighlighted) {
      this.send('poc_hello', {
        played_card: card.name,
        targeted_location: locationHighlighted.name,
      });
      this.lastActionText.text = `${card.name} played on ${locationHighlighted.name}`;
      locationHighlighted.isHovering = false;
    }
  }
}
