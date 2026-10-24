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
  private static readonly ROW_SPILL = 150; // how far the initial row may exceed the location's width before cards start to overlap

  //the box a location claims on the map: its own card plus the row of contained cards hanging below
  static readonly FOOTPRINT_WIDTH = LocationCard.LOC_WIDTH + LocationCard.ROW_SPILL;
  static readonly FOOTPRINT_HEIGHT = LocationCard.LOC_HEIGHT + Card.HEIGHT - LocationCard.OVERLAP;

  public connections: number[] = []; // ids of the locations reachable from here
  private contents: Card[] = [];
  private offsets = new Map<Card, Phaser.Math.Vector2>(); // card centre relative to the location centre, unscaled

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
    scene.events.on(Phaser.Scenes.Events.UPDATE, this.followWiggle, this);
    this.once(Phaser.GameObjects.Events.DESTROY, () =>
      scene.events.off(Phaser.Scenes.Events.UPDATE, this.followWiggle, this),
    );
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: LocationDTO) {
    const location = new LocationCard(scene, x, y, dto.id, dto.name, dto.flavor);
    location.connections = dto.connections;
    return location;
  }

  // lays the cards out in a row straddling the lower border; from then on they are free to move
  setContents(cards: Card[]): void {
    this.contents = [...cards];
    this.offsets.clear();

    const step = this.slotStep();
    const n = this.contents.length;
    this.contents.forEach((card, i) => {
      card.homeLocation = this;
      const x = (i - (n - 1) / 2) * step;
      const y = LocationCard.LOC_HEIGHT / 2 - LocationCard.OVERLAP + card.cardHeight / 2;
      this.offsets.set(card, new Phaser.Math.Vector2(x, y));
    });
    this.syncContents();
  }

  removeContent(card: Card): void {
    this.contents = this.contents.filter((c) => c !== card);
    this.offsets.delete(card);
    card.homeLocation = null;
  }

  // lays the row out again with the card included, which also snaps the card to its slot
  addContent(card: Card): void {
    this.setContents([...this.contents, card]);
  }

  // the contents ride along with the location, so they are raised with it and its shadow covers them
  override setFloating(lifted: boolean): void {
    super.setFloating(lifted);
    for (const card of this.contents) card.setLifted(lifted, false);
    if (lifted) {
      this.drawShadow();
      return;
    }
    this.settleContents();
    this.scene.time.delayedCall(Card.LIFT_DURATION, () => this.floating || this.drawShadow());
  }

  private settleContents(): void {
    this.syncContents();
    this.scene.tweens.add({
      targets: this.contents,
      angle: 0,
      duration: Card.LIFT_DURATION,
      ease: 'Sine.easeOut',
    });
  }

  // while floating, the shadow takes the shape of the location and its contents together
  protected override drawShadow(): void {
    super.drawShadow();
    if (!this.floating) return;
    for (const [card, offset] of this.offsets) {
      this.fillShadowRect(offset.x, offset.y, card.cardWidth, card.cardHeight);
    }
  }

  // keeps the contents on top of the location in the display list
  raiseContents(): void {
    for (const card of this.contents) this.scene.children.bringToTop(card);
    this.scene.children.queueDepthSort();
  }

  // moves the contents along with the location; while it is carried they turn around its centre with it
  syncContents(): void {
    const turn = this.floating ? this.rotation : 0;
    for (const [card, offset] of this.offsets) {
      const turned = offset.clone().scale(this.scaleX).rotate(turn);
      card.setPosition(this.x + turned.x, this.y + turned.y);
    }
  }

  // the location wiggles while it is carried; its contents wiggle with it instead of standing still
  private followWiggle(): void {
    if (!this.floating) return;
    this.syncContents();
    for (const card of this.contents) card.setAngle(this.angle);
  }

  // while a contained card is dragged: it may go anywhere as long as it still touches the location
  constrainDrag(card: Card): void {
    const s = this.scaleX;
    const reachX = ((LocationCard.LOC_WIDTH + card.cardWidth) / 2) * s;
    const reachY = ((LocationCard.LOC_HEIGHT + card.cardHeight) / 2) * s;
    const dx = Phaser.Math.Clamp(card.x - this.x, -reachX, reachX);
    const dy = Phaser.Math.Clamp(card.y - this.y, -reachY, reachY);
    card.setPosition(this.x + dx, this.y + dy);
    this.offsets.get(card)?.set(dx / s, dy / s);
  }

  private slotStep(): number {
    const n = this.contents.length;
    if (n < 2) return 0;
    const cardWidth = this.contents[0].cardWidth;
    const maxSpan = LocationCard.LOC_WIDTH + LocationCard.ROW_SPILL;
    return Math.min(cardWidth + LocationCard.SLOT_GAP, (maxSpan - cardWidth) / (n - 1));
  }
}
