import Phaser from 'phaser';
import { PlayableCard } from '../objects/PlayableCard';
import { LocationCard } from '../objects/LocationCard';
import type { Card } from '../objects/Card';
import { connect } from '../../net/socket';

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

    this.add.text(this.scale.width / 2, 50, 'Along the mist-shrouded plains,', {
      color: '#000000',
    });
    this.add.text(this.scale.width / 2, 70, 'Across the eternal plains', { color: '#000000' });
    this.add.text(this.scale.width / 2, 90, 'And ever towards the ancient forest', {
      color: '#000000',
    });
    this.lastActionText = this.add.text(this.scale.width / 2, 130, '...at long last, homebound.', {
      color: '#000000',
    });

    this.socket = connect((msg) => {
      this.lastActionText.text = `server: ${msg.type}`;
      if (msg.type === 'poc_reply') this.send('poc_ack');
      if (msg.type === 'game_state') {
        if (msg.payload) {
          console.log('Received the new gamestate with a payload');
          this.handleGameStart(msg.payload);
        } else {
          console.log('Received the new gamestate but the payload is missing');
        }
      } else {
        console.log(msg.type);
      }
    });

    this.socket.addEventListener('open', () => this.send('new_game', { level: 'basic' }));

    //commenting out generics for now
    /*
    var boltCard = new PlayableCard(
      this,
      this.scale.width / 2 - 400,
      this.scale.height,
      'Fishing',
      'Attempt to fish at target location.',
      0,
    );
    boltCard.on('cardDragged', (card: PlayableCard) => this.handleCardDrag(card));
    boltCard.on('cardDragStart', (card: PlayableCard) => this.handleCardDragStart(card));
    boltCard.on('cardDragEnd', (card: PlayableCard) => this.handleCardDrop(card));
    this.playerCards.push(boltCard);

    var cleanseCard = new PlayableCard(
      this,
      this.scale.width / 2,
      this.scale.height,
      'Rite of Seeking',
      'Consult the spirits.\n\nIf you succeed, discover tracks at the location.',
      4,
    );
    cleanseCard.on('cardDragged', (card: PlayableCard) => this.handleCardDrag(card));
    cleanseCard.on('cardDragStart', (card: PlayableCard) => this.handleCardDragStart(card));
    cleanseCard.on('cardDragEnd', (card: PlayableCard) => this.handleCardDrop(card));

    cleanseCard.setDiabled(true);
    this.playerCards.push(cleanseCard);

    var trapCard = new PlayableCard(
      this,
      this.scale.width / 2 + 400,
      this.scale.height,
      'Trap',
      'Place a trap at target location',
      2,
    );
    trapCard.on('cardDragged', (card: PlayableCard) => this.handleCardDrag(card));
    trapCard.on('cardDragStart', (card: PlayableCard) => this.handleCardDragStart(card));
    trapCard.on('cardDragEnd', (card: PlayableCard) => this.handleCardDrop(card));
    this.playerCards.push(trapCard);

    /*
    var secludedDen = new LocationCard(
      this,
      this.scale.width / 2 + 400,
      400,
      'Secluded Den',
      'The secluded den is home to your tribe.',
    );
    secludedDen.on('cardDragStart', (card: Card) => this.handleCardDragStart(card));
    this.locationCards.push(secludedDen);

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

  update(_timer: number, _delta: number) {}

  handleCardDragStart(card: Card) {
    this.children.bringToTop(card);
  }

  handleGameStart(inputGameState: Record<string, unknown>) {
    console.log(inputGameState);
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
