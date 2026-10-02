import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';

export type ShadowMode = 'hovers' | 'flat';

export interface CardConfig {
  id: number;
  name: string;
  flavor: string;
  faceColor: number;
  artHeight: number;
  shadow: ShadowMode;
  showTextBox?: boolean; // default true; the text box fills the space below the art
  italicText?: boolean; // default false
  titleBelowArt?: boolean; // default false; centred, bold, 20% larger, beneath the art instead of above it
  footerHeight?: number; // default 0; shrinks the text box to make room for a footer line beneath it
}

interface ShadowPose {
  x: number;
  y: number;
  scale: number;
  alpha: number;
}

const SHADOW_REST: Record<ShadowMode, ShadowPose> = {
  hovers: { x: 11.2, y: -11.2, scale: 1, alpha: 0.5 },
  flat: { x: 0, y: 0, scale: 1, alpha: 0 },
};
const SHADOW_LIFTED: ShadowPose = { x: 20, y: -20, scale: 1.1, alpha: 0.25 };
const SHADOW_HOVER: ShadowPose = { x: 16, y: -16, scale: 1.025, alpha: 0.35 };

const TEXT_STYLE: Phaser.Types.GameObjects.Text.TextStyle = {
  color: '#000000',
  fontFamily: '"Times New Roman", Times, serif',
};

export class Card extends Phaser.GameObjects.Container {
  static readonly WIDTH = 252;
  static readonly HEIGHT = 352;

  // layout, all in top-left card coordinates
  protected static readonly BORDER = 8;
  protected static readonly INSET = 24; // border + padding; left/right/bottom margin of art and text box
  protected static readonly INNER_WIDTH = Card.WIDTH - 2 * Card.INSET; // 204
  protected static readonly TITLE_Y = 16;
  protected static readonly ART_Y = 36;
  protected static readonly GAP = 12;
  protected static readonly BOTTOM = Card.HEIGHT - Card.INSET; // 328
  protected static readonly TEXT_PAD = 6.4;

  public readonly id: number;
  public floating = false; // currently lifted by a drag
  public flavor: string;
  public isHovering = false;
  private isDropTarget = false;

  protected readonly config: CardConfig;
  protected tween: Phaser.Tweens.Tween | null = null;
  protected tweensManager: Phaser.Tweens.TweenManager;
  protected isDisabled = false;

  protected shadow: Phaser.GameObjects.Rectangle;
  protected cardFrame: Phaser.GameObjects.Container; // lift/scale target, sits at card centre
  protected surface: Phaser.GameObjects.Container; // top-left coordinate space
  protected hitbox: Phaser.GameObjects.Zone;
  protected disabledLayer!: Phaser.GameObjects.Rectangle;

  private dragLast = new Phaser.Math.Vector2();

  constructor(scene: MainScene, x: number, y: number, config: CardConfig) {
    super(scene, x, y);
    this.config = config;
    this.id = config.id;
    this.name = config.name;
    this.flavor = config.flavor;
    this.tweensManager = scene.tweens;

    const rest = SHADOW_REST[config.shadow];
    this.shadow = new Phaser.GameObjects.Rectangle(
      scene,
      rest.x,
      rest.y,
      Card.WIDTH,
      Card.HEIGHT,
      0x000000,
    ).setAlpha(rest.alpha);

    this.surface = new Phaser.GameObjects.Container(scene, -Card.WIDTH / 2, -Card.HEIGHT / 2);
    this.cardFrame = new Phaser.GameObjects.Container(scene, 0, 0, [this.surface]);
    this.buildSurface();

    this.hitbox = new Phaser.GameObjects.Zone(scene, 0, 0, Card.WIDTH, Card.HEIGHT);
    this.hitbox.setInteractive({ draggable: true });
    this.bindInput();

    this.add([this.shadow, this.cardFrame, this.hitbox]);
    scene.add.existing(this);
  }

  private buildSurface(): void {
    const { BORDER, INSET, INNER_WIDTH, TITLE_Y, ART_Y, GAP, BOTTOM, TEXT_PAD } = Card;
    const {
      faceColor,
      artHeight,
      showTextBox = true,
      italicText = false,
      footerHeight = 0,
      titleBelowArt = false,
    } = this.config;

    this.surface.add([
      this.rect(0, 0, Card.WIDTH, Card.HEIGHT, 0x000000),
      this.rect(BORDER, BORDER, Card.WIDTH - 2 * BORDER, Card.HEIGHT - 2 * BORDER, faceColor),
      this.rect(INSET, ART_Y, INNER_WIDTH, artHeight, 0x2ecc71),
      titleBelowArt
        ? this.label(INSET + INNER_WIDTH / 2, (ART_Y + artHeight + BOTTOM) / 2, this.name, {
            fontSize: '17.28px',
            fontStyle: 'bold',
            color: '#ffffff',
          }).setOrigin(0.5)
        : this.label(INSET, TITLE_Y, this.name, {
            fontSize: '14.4px',
            fontStyle: 'bold',
            color: '#ffffff',
          }),
    ]);

    if (showTextBox) {
      const top = ART_Y + artHeight + GAP;
      const boxHeight = BOTTOM - footerHeight - top;
      this.surface.add([
        this.rect(INSET, top, INNER_WIDTH, boxHeight, 0xd8c9a3),
        this.label(INSET + INNER_WIDTH / 2, top + boxHeight - TEXT_PAD, this.flavor, {
          fontSize: '11.2px',
          fontStyle: italicText ? 'italic' : 'normal',
          align: 'center',
          wordWrap: { width: INNER_WIDTH - 2 * TEXT_PAD },
        }).setOrigin(0.5, 1),
      ]);
    }

    this.disabledLayer = this.rect(0, 0, Card.WIDTH, Card.HEIGHT, 0x111111, 0.5).setVisible(false);
    this.surface.add(this.disabledLayer);
  }

  protected rect(x: number, y: number, w: number, h: number, color: number, alpha = 1) {
    return new Phaser.GameObjects.Rectangle(this.scene, x, y, w, h, color, alpha).setOrigin(0, 0);
  }

  protected label(
    x: number,
    y: number,
    content: string,
    style: Phaser.Types.GameObjects.Text.TextStyle = {},
  ) {
    return new Phaser.GameObjects.Text(this.scene, x, y, content, { ...TEXT_STYLE, ...style });
  }

  // footer line between the text box and the card edge; align 'right' sits flush with the art's right edge
  protected footer(content: string, align: 'center' | 'right'): Phaser.GameObjects.Text {
    const { BORDER, INSET, BOTTOM, INNER_WIDTH } = Card;
    const top = BOTTOM - (this.config.footerHeight ?? 0);
    const y = (top + Card.HEIGHT - BORDER) / 2;
    const x = align === 'center' ? INSET + INNER_WIDTH / 2 : INSET + INNER_WIDTH;
    return this.label(x, y, content, { fontSize: '12.8px', color: '#ffffff' }).setOrigin(
      align === 'center' ? 0.5 : 1,
      0.5,
    );
  }

  //inserts below the disabled layer
  protected addFrameDetails(...objects: Phaser.GameObjects.GameObject[]): void {
    for (const o of objects) this.surface.addAt(o, this.surface.length - 1);
  }

  private bindInput(): void {
    this.hitbox.on('dragstart', (pointer: Phaser.Input.Pointer) => this.onDragStart(pointer));
    this.hitbox.on('drag', (pointer: Phaser.Input.Pointer) => this.onDrag(pointer));
    this.hitbox.on('dragend', () => this.onDragend());
    this.hitbox.on('pointerover', () => this.onPointerOver());
    this.hitbox.on('pointerout', () => this.onPointerOut());
  }

  onDragStart(pointer: Phaser.Input.Pointer) {
    if (this.isDisabled) return;
    this.dragLast.set(pointer.worldX, pointer.worldY);
    this.setFloating(true);
    this.emit('cardDragStart', this);
  }

  onDrag(pointer: Phaser.Input.Pointer) {
    if (this.isDisabled) return;
    this.x += pointer.worldX - this.dragLast.x;
    this.y += pointer.worldY - this.dragLast.y;
    this.dragLast.set(pointer.worldX, pointer.worldY);
    this.emit('cardDragged', this);
  }

  onDragend() {
    if (this.isDisabled) return;
    this.setFloating(false);
    this.emit('cardDragEnd', this);
  }

  setFloating(lifted: boolean) {
    this.floating = lifted;
    const pose = lifted ? SHADOW_LIFTED : SHADOW_REST[this.config.shadow];
    this.tweensManager.add({
      targets: this.cardFrame,
      scale: lifted ? 1.05 : 1,
      y: lifted ? -8 : 0,
      duration: 150,
      ease: 'Sine.easeOut',
    });
    this.tweensManager.add({
      targets: this.shadow,
      ...pose,
      duration: 150,
      ease: 'Sine.easeOut',
    });
  }

  onPointerOver() {
    this.emit('cardPointerOver', this);
    this.isHovering = true;
    this.setHoverLift(true);
    if (this.tween?.isPlaying()) return;

    if (this.parentContainer) this.parentContainer.bringToTop(this);
    this.isHovering = true;
    this.startWiggle();
    this.emit('cardPointerOver', this);
  }

  private startWiggle(): void {
    if (this.tween?.isPlaying()) return;
    this.tween = this.tweensManager.add({
      targets: this,
      angle: { from: -0.5, to: 0.5 },
      duration: 600,
      yoyo: true,
      repeat: -1,
      ease: 'Sine.easeInOut',
      onRepeat: () => {
        if (!this.isHovering && !this.isDropTarget) this.finishHover();
      },
    });
  }

  // wiggles while a playable card is dragged over it; disabled cards are no valid target
  setDropTarget(on: boolean): void {
    on = on && !this.isDisabled;
    if (on === this.isDropTarget) return;
    this.isDropTarget = on;
    if (on) this.startWiggle();
    else if (!this.isHovering) this.finishHover();
  }

  // on hover the card pops up (with its shadow) and settles back to rest, the wiggle signals the hover from there
  private setHoverLift(lifted: boolean): void {
    if (this.floating) return; //floating means it's beging dragged
    const restShadow = SHADOW_REST[this.config.shadow];
    this.tweensManager.killTweensOf(this.cardFrame);
    this.tweensManager.killTweensOf(this.shadow);

    const settle = () => {
      this.tweensManager.add({
        targets: this.cardFrame,
        scale: 1,
        y: 0,
        duration: 350,
        ease: 'Sine.easeInOut',
      });
      this.tweensManager.add({
        targets: this.shadow,
        ...restShadow,
        duration: 350,
        ease: 'Sine.easeInOut',
      });
    };

    if (!lifted) {
      settle();
      return;
    }

    this.tweensManager.add({
      targets: this.cardFrame,
      scale: 1.015,
      y: -4.8,
      duration: 260,
      ease: 'Sine.easeOut',
      onComplete: settle,
    });
    this.tweensManager.add({
      targets: this.shadow,
      ...SHADOW_HOVER,
      duration: 260,
      ease: 'Sine.easeOut',
    });
  }

  finishHover() {
    this.tween?.stop();
    this.tweensManager.add({ targets: this, angle: 0, duration: 600, ease: 'Sine.easeOut' });
  }

  onPointerOut() {
    this.isHovering = false;
    this.setHoverLift(false);
  }

  setDisabled(on: boolean) {
    this.isDisabled = on;
    this.disabledLayer.setVisible(on);
  }

  getHitbox(): Phaser.Geom.Rectangle {
    return this.hitbox.getBounds();
  }
}
