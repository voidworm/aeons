import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import { KindredCard } from './KindredCard';
import type { CreatureDTO } from './GameState';

export class CreatureCard extends Card {
  private static readonly FACE_COLOR = 0xad3d2b;
  private static readonly IDENT_LINE_HEIGHT = 16; // the identifier words take the first line of the text box
  private static readonly FACE_FADE_DURATION = 600; // ms, when becoming kindred
  private static readonly FOOTER_HEIGHT = 19.2;
  private static readonly SIZE_BADGE_RADIUS = 15;
  private static readonly SIZE_BADGE_BORDER = 3;
  private static readonly SIZE_BADGE_FILL = 0xd8c9a3; // the text box colour
  private static readonly DOT_RADIUS = 5;
  private static readonly DOT_GAP = 15;
  private static readonly DOT_EMPTY = 0x8a8a8a;
  private static readonly DOT_DISTRESSED = 0xc0392b;
  private static readonly DOT_OUTLINE = 0x2a1809;
  private exhausted = false;
  private shownFaceColor = CreatureCard.FACE_COLOR; // changes along the colour fade
  private faceColorGoal = CreatureCard.FACE_COLOR;
  private faceTween: Phaser.Tweens.Tween | null = null;
  private turnTween: Phaser.Tweens.Tween | null = null;
  private sizeBadge: Phaser.GameObjects.Arc | null = null;
  private distressDots: Phaser.GameObjects.Arc[] = [];
  private maxDistress: number;
  private size: number;
  public kindredOf: number | null = null; // id of the player this creature is kindred to
  private kindredLabel: Phaser.GameObjects.Text | null = null;

  constructor(
    scene: MainScene,
    x: number,
    y: number,
    id: number,
    name: string,
    identNames: string[],
    flavor: string,
    distress: number,
    maxDistress: number,
    size: number,
  ) {
    super(scene, x, y, {
      id: id,
      name: name,
      flavor: flavor,
      effectText: identNames.join(' · '),
      faceColor: CreatureCard.FACE_COLOR,
      artHeight: 136,
      shadow: 'flat',
      italicText: true,
      centerTitle: true,
      footerHeight: CreatureCard.FOOTER_HEIGHT,
    });

    this.maxDistress = maxDistress;
    this.size = size;
    this.initFrameDetails(distress);
  }

  // an exhausted creature lies on its side; the hitbox and shadow turn with it
  setExhausted(on: boolean, animate = true): Promise<void> {
    if (on === this.exhausted) return Promise.resolve();
    this.exhausted = on;
    const angle = on ? 90 : 0;
    this.turnTween?.stop();
    if (!animate) {
      [this.cardFrame, this.hitbox, this.shadow].forEach((part) => part.setAngle(angle));
      return Promise.resolve();
    }
    return new Promise((resolve) => {
      this.turnTween = this.tweensManager.add({
        targets: [this.cardFrame, this.hitbox, this.shadow],
        angle,
        duration: 400,
        ease: 'Sine.easeInOut',
        onComplete: () => resolve(),
        onStop: () => resolve(),
      });
    });
  }

  setFaceColor(color: number): void {
    super.setFaceColor(color);
    this.sizeBadge?.setStrokeStyle(CreatureCard.SIZE_BADGE_BORDER, color);
  }

  setKindredOf(ownerId: number | null, ownerName: string) {
    this.kindredOf = ownerId;
    this.fadeFaceTo(ownerId === null ? CreatureCard.FACE_COLOR : KindredCard.KINDRED_FACE_COLOR);
    this.kindredLabel?.destroy();
    this.kindredLabel = null;
    if (ownerId === null) return;

    const textBoxTop = Card.ART_Y + this.config.artHeight + Card.GAP;
    this.kindredLabel = this.label(
      this.cardWidth / 2,
      textBoxTop + Card.TEXT_PAD + CreatureCard.IDENT_LINE_HEIGHT,
      `Kindred to: ${ownerName}`,
      { fontSize: '11.2px', fontStyle: 'normal' },
    ).setOrigin(0.5, 0);
    this.addFrameDetails(this.kindredLabel);
  }

  private fadeFaceTo(goal: number) {
    if (goal === this.faceColorGoal) return;
    this.faceColorGoal = goal;
    this.faceTween?.stop();
    const from = Phaser.Display.Color.ValueToColor(this.shownFaceColor);
    const to = Phaser.Display.Color.ValueToColor(goal);
    const show = (color: number) => {
      this.shownFaceColor = color;
      this.setFaceColor(color);
    };
    this.faceTween = this.tweensManager.addCounter({
      from: 0,
      to: 100,
      duration: CreatureCard.FACE_FADE_DURATION,
      onUpdate: (tween) => {
        const { r, g, b } = Phaser.Display.Color.Interpolate.ColorWithColor(
          from,
          to,
          100,
          tween.getValue() ?? 0,
        );
        show(Phaser.Display.Color.GetColor(r, g, b));
      },
      onComplete: () => show(goal),
    });
  }

  initFrameDetails(distress: number) {
    this.addFrameDetails(...this.buildSizeBadge(), ...this.buildDistressDots());
    this.setDistress(distress);
  }

  // sits on the right edge, where the art meets the text box
  private buildSizeBadge(): Phaser.GameObjects.GameObject[] {
    const { SIZE_BADGE_RADIUS, SIZE_BADGE_BORDER, SIZE_BADGE_FILL } = CreatureCard;
    const x = this.cardWidth - Card.INSET;
    const y = Card.ART_Y + this.config.artHeight + Card.GAP / 2;
    this.sizeBadge = new Phaser.GameObjects.Arc(this.scene, x, y, SIZE_BADGE_RADIUS, 0, 360, false, SIZE_BADGE_FILL)
      .setStrokeStyle(SIZE_BADGE_BORDER, this.shownFaceColor);
    const number = this.label(x, y, String(this.size), { fontSize: '18px', fontStyle: 'bold' }).setOrigin(0.5);
    return [this.sizeBadge, number];
  }

  private buildDistressDots(): Phaser.GameObjects.Arc[] {
    const { FOOTER_HEIGHT, DOT_RADIUS, DOT_GAP, DOT_EMPTY, DOT_OUTLINE } = CreatureCard;
    const footerTop = this.cardHeight - Card.INSET - FOOTER_HEIGHT;
    const y = (footerTop + this.cardHeight - Card.BORDER) / 2;
    const firstX = this.cardWidth / 2 - ((this.maxDistress - 1) * DOT_GAP) / 2;
    this.distressDots = Array.from({ length: this.maxDistress }, (_, i) =>
      new Phaser.GameObjects.Arc(this.scene, firstX + i * DOT_GAP, y, DOT_RADIUS, 0, 360, false, DOT_EMPTY)
        .setStrokeStyle(1.5, DOT_OUTLINE),
    );
    return this.distressDots;
  }

  setDistress(distress: number) {
    const { DOT_DISTRESSED, DOT_EMPTY } = CreatureCard;
    this.distressDots.forEach((dot, i) => dot.setFillStyle(i < distress ? DOT_DISTRESSED : DOT_EMPTY));
  }

  // hidden creatures are faded out
  setHidden(hidden: boolean) {
    this.setAlpha(hidden ? 0.3 : 1);
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: CreatureDTO) {
    const card = new CreatureCard(
      scene,
      x,
      y,
      dto.id,
      dto.name,
      dto.identNames,
      dto.flavor,
      dto.distress,
      dto.maxDistress,
      dto.size,
    );
    void card.setExhausted(dto.exhausted, false);
    card.setHidden(dto.hidden);
    return card;
  }
}
