import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { LocationDTO } from './GameState';

export class LocationCard extends Card {
  static readonly LOC_WIDTH = 352;
  static readonly LOC_HEIGHT = 252;
  // how far the contained cards reach up into the location from its lower border
  static readonly OVERLAP = 30;

  private static readonly SLOT_GAP = 12;
  private static readonly ROW_SPILL = 150; // how far the row may exceed the location's width before cards start to overlap

  private contents: Card[] = [];
  private slotTweens = new Map<Card, { tween: Phaser.Tweens.Tween; x: number; y: number }>();

  constructor(
    scene: MainScene,
    x: number,
    y: number,
    id: number,
    cardname: string,
    cardtext: string,
  ) {
    const { LOC_WIDTH: WIDTH, LOC_HEIGHT: HEIGHT } = LocationCard;
    super(scene, x, y, {
      id: id,
      name: cardname,
      flavor: cardtext,
      faceColor: 0x105207,
      width: WIDTH,
      height: HEIGHT,
      artHeight: HEIGHT - 36 - Card.INSET,
      shadow: 'flat',
      italicText: true,
      centerTitle: true,
      centerText: true,
      textBoxRect: { x: (WIDTH - 220) / 2, y: HEIGHT - 36 - 64, w: 220, h: 64 },
    });
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: LocationDTO) {
    return new LocationCard(scene, x, y, dto.id, dto.name, dto.flavor);
  }

  setContents(cards: Card[]): void {
    this.contents = [...cards];
    for (const card of this.contents) card.homeLocation = this;
  }

  // keeps the contents on top of the location in the display list
  raiseContents(): void {
    for (const card of this.contents) this.scene.children.bringToTop(card);
    this.scene.children.queueDepthSort();
  }

  // places the contents in a row straddling the lower border; `skip` is the card currently held by the pointer
  layoutContents(skip?: Card, animate = false): void {
    const step = this.slotStep();
    const n = this.contents.length;
    this.contents.forEach((card, i) => {
      if (card === skip) return;
      const offset = (i - (n - 1) / 2) * step;
      this.moveToSlot(card, this.x + offset * this.scaleX, this.slotY(card), animate);
    });
  }

  // while a contained card is dragged: lock it to the row and re-sort the others around it
  reorderAround(card: Card): void {
    const halfSpan = (this.rowSpan() / 2) * this.scaleX;
    card.x = Phaser.Math.Clamp(card.x, this.x - halfSpan, this.x + halfSpan);
    card.y = this.slotY(card);
    this.contents.sort((a, b) => a.x - b.x);
    this.layoutContents(card, true);
  }

  private slotY(card: Card): number {
    return (
      this.y +
      (LocationCard.LOC_HEIGHT / 2 - LocationCard.OVERLAP + card.cardHeight / 2) * this.scaleX
    );
  }

  private rowSpan(): number {
    const cardWidth = this.contents[0]?.cardWidth ?? Card.WIDTH;
    const step = this.slotStep();
    return cardWidth + (this.contents.length - 1) * step;
  }

  private slotStep(): number {
    const n = this.contents.length;
    if (n < 2) return 0;
    const cardWidth = this.contents[0].cardWidth;
    const maxSpan = LocationCard.LOC_WIDTH + LocationCard.ROW_SPILL;
    return Math.min(cardWidth + LocationCard.SLOT_GAP, (maxSpan - cardWidth) / (n - 1));
  }

  private moveToSlot(card: Card, x: number, y: number, animate: boolean): void {
    const running = this.slotTweens.get(card);
    if (running && running.x === x && running.y === y) return;
    running?.tween.stop();
    this.slotTweens.delete(card);

    if (!animate) {
      card.setPosition(x, y);
      return;
    }
    const tween = this.scene.tweens.add({
      targets: card,
      x,
      y,
      duration: 150,
      ease: 'Sine.easeOut',
      onComplete: () => this.slotTweens.delete(card),
    });
    this.slotTweens.set(card, { tween, x, y });
  }
}
