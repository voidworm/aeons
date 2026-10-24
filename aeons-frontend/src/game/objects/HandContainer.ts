import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { PlayableCard } from './PlayableCard';
import { uiScale } from '../layout/uiScale';

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
  private static readonly EXPECTED_HAND_SIZE = 7; // hands rarely get bigger than this for now
  private static readonly MIN_SCALE = 0.3;
  private static readonly FLY_DURATION = 350;
  private static readonly RETURN_MIN_SCALE = 0.05; // how small a returning card gets
  private static readonly SHRINK_DURATION = 500;
  private static readonly GROW_DURATION = 750;
  private static readonly ENTER_DELAY = 200; // lets the previous hand leave first
  private static readonly DEPTH = 900; // above the board, below the player switcher
  private static readonly DRAG_DEPTH = 950; // a card pulled out of the hand stays above the hand area

  private background: Phaser.GameObjects.Arc; // one big circle, only its top cap is on screen
  private hands = new Map<number, Hand>();
  private returning = new Set<PlayableCard>(); // cards currently shrinking on their way back
  private dragOrigins = new Map<PlayableCard, { x: number; y: number }>();
  private activeId: number | null = null;
  private containerWidth = 0;

  constructor(scene: MainScene) {
    super(scene, 0, 0);

    this.background = new Phaser.GameObjects.Arc(scene, 0, 0, 0, 0, 360, false, 0x04140a, 0.9);
    this.background.setInteractive({
      hitArea: {},
      hitAreaCallback: (_area: unknown, x: number, y: number) => this.insideBackground(x, y),
    }); // only there so board cards underneath don't register hovers through the hand
    this.add(this.background);

    this.setDepth(HandContainer.DEPTH);
    this.resize(scene.scale.width, scene.scale.height);
    scene.add.existing(this);
  }

  //distance from the screen centre to the outer edge of the outermost card of an expected-size hand
  static expectedHalfWidth(): number {
    const span = (HandContainer.EXPECTED_HAND_SIZE - 1) * HandContainer.STEP;
    return span / 2 + Card.WIDTH / 2;
  }

  //the container is scaled as a whole, so its content is laid out in units of the scaled-up screen width
  resize(width: number, height: number): void {
    const scale = uiScale(width, height);
    const localWidth = width / scale;
    this.containerWidth = localWidth;
    this.setScale(scale);
    this.setPosition(0, height - HandContainer.HEIGHT * scale);

    const halfChord = (localWidth - 2 * HandContainer.PADDING) / 2;
    const radius = (halfChord ** 2 + HandContainer.HEIGHT ** 2) / (2 * HandContainer.HEIGHT);
    this.background.setRadius(radius).setIterations(2 / radius);
    this.background.setPosition(localWidth / 2, radius - HandContainer.ARC_LIFT);
    this.hands.forEach((hand) => hand.layer.setX(localWidth / 2));
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
    hand.layer.remove(card);
    this.dragOrigins.delete(card);
    this.returning.delete(card);
    if (destroy) card.destroy(); // the card may be outside the layer, e.g. still shrinking on its way back
    this.layout(hand);
  }

  startDrag(card: PlayableCard): void {
    const hand = this.handHolding(card);
    if (!hand) return;

    this.dragOrigins.set(card, { x: card.x, y: card.y });
    const { tx, ty } = card.getWorldTransformMatrix();
    hand.layer.remove(card);
    this.scene.add.existing(card);
    card.setDepth(HandContainer.DRAG_DEPTH);
    card.setScale(this.scaleX * hand.layer.scaleX);
    card.setPosition(tx, ty);
  }

  //a card dropped over the hand area stays where it is; otherwise it shrinks away where it is,
  //reappears tiny in its hand slot and grows back; snapBack forces the second even over the hand area
  returnCard(card: PlayableCard, snapBack = false): void {
    const hand = this.handHolding(card);
    if (!hand || card.parentContainer === hand.layer || this.returning.has(card)) return;

    const pointer = this.scene.input.activePointer;
    const inHandArea = this.containsPoint(pointer.x, pointer.y);
    const origin = this.dragOrigins.get(card);
    this.dragOrigins.delete(card);

    if ((inHandArea && !snapBack) || !origin) {
      const local = hand.layer.getWorldTransformMatrix().applyInverse(card.x, card.y);
      this.putInHand(hand, card, local.x, local.y, 1);
      return;
    }
    this.returning.add(card);
    this.scene.tweens.add({
      targets: card,
      scale: HandContainer.RETURN_MIN_SCALE,
      duration: HandContainer.SHRINK_DURATION,
      ease: 'Sine.easeIn',
      onComplete: () => {
        this.returning.delete(card);
        this.putInHand(hand, card, origin.x, origin.y, HandContainer.RETURN_MIN_SCALE);
        this.scene.tweens.add({
          targets: card,
          scale: 1,
          duration: HandContainer.GROW_DURATION,
          ease: 'Back.easeOut',
        });
      },
    });
  }

  private putInHand(hand: Hand, card: PlayableCard, x: number, y: number, scale: number): void {
    hand.layer.add(card);
    card.setDepth(0);
    card.setScale(scale);
    card.setPosition(x, y);
    hand.cards.forEach((c) => hand.layer.bringToTop(c)); // keeps the original stacking order
  }

  //x and y are local to the circle's bounding box
  private insideBackground(x: number, y: number): boolean {
    const { radius } = this.background;
    return Phaser.Math.Distance.Between(x, y, radius, radius) <= radius;
  }

  //hand area = the visible cap of the background circle
  containsPoint(x: number, y: number): boolean {
    const { tx: cx, ty: cy } = this.background.getWorldTransformMatrix();
    return Phaser.Math.Distance.Between(x, y, cx, cy) <= this.background.radius * this.scaleX;
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
