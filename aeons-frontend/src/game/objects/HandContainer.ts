import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { PlayableCard } from './PlayableCard';

//one hand instance per player
//many hands per hand container
interface Hand {
  layer: Phaser.GameObjects.Container;
  cards: PlayableCard[];
  flight: Phaser.Tweens.Tween | null;
}

export class HandContainer extends Phaser.GameObjects.Container {
  static readonly HEIGHT = 220;
  static readonly PADDING = 40;
  private static readonly ARC_LIFT = 60; // shifts the circle up, so it is wider at the bottom corners
  private static readonly CARD_Y = 160;
  private static readonly STEP = Card.WIDTH / 1.5;
  private static readonly OFFSCREEN_Y = HandContainer.HEIGHT + Card.HEIGHT; //movement goal for outside the screen
  private static readonly MIN_SCALE = 0.3;
  private static readonly FLY_DURATION = 350;
  private static readonly ENTER_DELAY = 200; // lets the previous hand leave first

  private background: Phaser.GameObjects.Arc; // one big circle, only its top cap is on screen
  private hands = new Map<number, Hand>();
  private dragOrigins = new Map<PlayableCard, { x: number; y: number }>();
  private activeId: number | null = null;
  private containerWidth = 0;

  constructor(scene: MainScene) {
    super(scene, 0, 0);

    this.background = new Phaser.GameObjects.Arc(scene, 0, 0, 0, 0, 360, false, 0x000000, 0.1);
    this.add(this.background);

    this.resize(scene.scale.width, scene.scale.height);
    scene.add.existing(this);
  }

  resize(width: number, height: number): void {
    this.containerWidth = width;
    this.setPosition(0, height - HandContainer.HEIGHT);

    const halfChord = (width - 2 * HandContainer.PADDING) / 2;
    const radius = (halfChord ** 2 + HandContainer.HEIGHT ** 2) / (2 * HandContainer.HEIGHT);
    this.background.setRadius(radius).setIterations(2 / radius);
    this.background.setPosition(width / 2, radius - HandContainer.ARC_LIFT);
    this.hands.forEach((hand) => hand.layer.setX(width / 2));
  }

  addCard(playerId: number, card: PlayableCard): void {
    const hand = this.handOf(playerId);
    hand.cards.push(card);
    hand.layer.add(card);
    this.layout(hand);
  }

  removeCard(playerId: number, card: PlayableCard, destroy = false): void {
    const hand = this.handOf(playerId);
    hand.cards = hand.cards.filter((c) => c !== card);
    hand.layer.remove(card, destroy);
    this.layout(hand);
  }

  startDrag(card: PlayableCard): void {
    const hand = this.handHolding(card);
    if (!hand) return;

    this.dragOrigins.set(card, { x: card.x, y: card.y });
    const { tx, ty } = card.getWorldTransformMatrix();
    hand.layer.remove(card);
    this.scene.add.existing(card);
    card.setPosition(tx, ty);
  }

  //snap a card back to the hand after being dropped.
  returnCard(card: PlayableCard): void {
    const hand = this.handHolding(card);
    if (!hand || card.parentContainer === hand.layer) return;

    const pointer = this.scene.input.activePointer;
    const inHandArea = this.containsPoint(pointer.worldX, pointer.worldY);
    const local = hand.layer.getWorldTransformMatrix().applyInverse(card.x, card.y);
    hand.layer.add(card);
    card.setPosition(local.x, local.y);
    hand.cards.forEach((c) => hand.layer.bringToTop(c)); // keeps the original stacking order

    const origin = this.dragOrigins.get(card);
    this.dragOrigins.delete(card);
    if (inHandArea || !origin) return; // free placement inside the hand area

    this.scene.tweens.add({
      targets: card,
      x: origin.x,
      y: origin.y,
      duration: 250,
      ease: 'Sine.easeOut',
    });
  }

  //hand area = the visible cap of the background circle
  private containsPoint(x: number, y: number): boolean {
    const { tx: cx, ty: cy } = this.background.getWorldTransformMatrix();
    return Phaser.Math.Distance.Between(x, y, cx, cy) <= this.background.radius;
  }

  private handHolding(card: PlayableCard): Hand | undefined {
    return [...this.hands.values()].find((h) => h.cards.includes(card));
  }

  //includes animation for hand flying
  setPlayerActive(playerId: number): void {
    if (playerId === this.activeId) return;

    const leaving = this.activeId === null ? undefined : this.hands.get(this.activeId);
    this.activeId = playerId;

    if (leaving) this.flyOut(leaving);
    this.flyIn(this.handOf(playerId));
  }

  private flyIn(hand: Hand): void {
    if (!hand.layer.visible) {
      hand.layer.setY(HandContainer.OFFSCREEN_Y).setScale(HandContainer.MIN_SCALE);
    }
    hand.layer.setVisible(true);
    this.fly(hand, HandContainer.CARD_Y, 1, HandContainer.ENTER_DELAY, 'Back.easeOut');
  }

  private flyOut(hand: Hand): void {
    this.bringToTop(hand.layer); // leaving cards stay above the incoming hand
    this.fly(hand, HandContainer.OFFSCREEN_Y, HandContainer.MIN_SCALE, 0, 'Back.easeIn', () => {
      hand.layer.setVisible(false).setY(HandContainer.CARD_Y).setScale(1);
    });
  }

  private fly(
    hand: Hand,
    y: number,
    scale: number,
    delay: number,
    ease: string,
    onDone?: () => void,
  ): void {
    hand.flight?.stop();
    hand.flight = this.scene.tweens.add({
      targets: hand.layer,
      y,
      scale,
      duration: HandContainer.FLY_DURATION,
      delay,
      ease,
      onComplete: onDone,
    });
  }

  private handOf(playerId: number): Hand {
    let hand = this.hands.get(playerId);
    if (!hand) {
      const layer = new Phaser.GameObjects.Container(
        this.scene,
        this.containerWidth / 2,
        HandContainer.CARD_Y,
      ).setVisible(false);
      this.add(layer);
      hand = { layer, cards: [], flight: null };
      this.hands.set(playerId, hand);
    }
    return hand;
  }

  private layout(hand: Hand): void {
    hand.cards.forEach((card, i) => {
      card.setPosition(this.slotX(hand, i), 0);
    });
  }

  private slotX(hand: Hand, index: number): number {
    const span = (hand.cards.length - 1) * HandContainer.STEP;
    return index * HandContainer.STEP - span / 2;
  }
}
