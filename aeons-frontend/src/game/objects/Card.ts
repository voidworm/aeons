import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import type { LocationCard } from './LocationCard';
import { CURSOR_INTERACTIVE } from '../cursors';
import { RichText } from './RichText';

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
  centerTitle?: boolean; // default false; centres the title horizontally at the top
  centerText?: boolean; // default false; centres the text box text horizontally and vertically
  footerHeight?: number; // default 0; shrinks the text box to make room for a footer line beneath it
  width?: number; // default Card.WIDTH
  height?: number; // default Card.HEIGHT
  textBoxRect?: { x: number; y: number; w: number; h: number }; // overrides the default text box placement, top-left card coordinates
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
  static readonly HOVER_ZOOM_DELAY = 3000; // ms of uninterrupted hovering before a board card zooms
  static readonly HOVER_ZOOM_FACTOR = 2.5;

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
  public readonly cardWidth: number;
  public readonly cardHeight: number;
  public floating = false; // currently lifted by a drag
  public flavor: string;
  public isHovering = false;
  public homeLocation: LocationCard | null = null; // set on cards that live at a location
  private isDropTarget = false;
  private zoomTimer: Phaser.Time.TimerEvent | null = null;
  private zoomTween: Phaser.Tweens.Tween | null = null;
  private restScale: number | null = null; // the card's scale before the hover zoom; non-null while zoomed

  protected readonly config: CardConfig;
  protected tween: Phaser.Tweens.Tween | null = null;
  protected tweensManager: Phaser.Tweens.TweenManager;
  protected isDisabled = false;
  private suppressed = false;
  private hoverTweens: Phaser.Tweens.Tween[] = []; // the lift tweens only, so stopping them leaves e.g. a creature's turn alone

  protected shadow: Phaser.GameObjects.Rectangle;
  protected cardFrame: Phaser.GameObjects.Container; // lift/scale target, sits at card centre
  protected surface: Phaser.GameObjects.Container; // top-left coordinate space
  protected hitbox: Phaser.GameObjects.Zone;
  protected disabledLayer!: Phaser.GameObjects.Rectangle;

  private dragLast = new Phaser.Math.Vector2();
  private dragTarget: Phaser.Math.Vector2 | null = null; // where the pointer wants the card, even if a constraint holds the card back

  constructor(scene: MainScene, x: number, y: number, config: CardConfig) {
    super(scene, x, y);
    this.config = config;
    this.cardWidth = config.width ?? Card.WIDTH;
    this.cardHeight = config.height ?? Card.HEIGHT;
    this.id = config.id;
    this.name = config.name;
    this.flavor = config.flavor;
    this.tweensManager = scene.tweens;

    const rest = SHADOW_REST[config.shadow];
    this.shadow = new Phaser.GameObjects.Rectangle(
      scene,
      rest.x,
      rest.y,
      this.cardWidth,
      this.cardHeight,
      0x000000,
    ).setAlpha(rest.alpha);

    this.surface = new Phaser.GameObjects.Container(
      scene,
      -this.cardWidth / 2,
      -this.cardHeight / 2,
    );
    this.cardFrame = new Phaser.GameObjects.Container(scene, 0, 0, [this.surface]);
    this.buildSurface();

    this.hitbox = new Phaser.GameObjects.Zone(scene, 0, 0, this.cardWidth, this.cardHeight);
    this.hitbox.setInteractive({ draggable: true, cursor: CURSOR_INTERACTIVE });
    this.bindInput();

    this.add([this.shadow, this.cardFrame, this.hitbox]);
    scene.add.existing(this);
  }

  private buildSurface(): void {
    const { BORDER, INSET, TITLE_Y, ART_Y, GAP, TEXT_PAD } = Card;
    const INNER_WIDTH = this.cardWidth - 2 * INSET;
    const BOTTOM = this.cardHeight - INSET;
    const {
      faceColor,
      artHeight,
      showTextBox = true,
      italicText = false,
      footerHeight = 0,
      titleBelowArt = false,
      centerTitle = false,
      centerText = false,
      textBoxRect,
    } = this.config;

    this.surface.add([
      this.rect(0, 0, this.cardWidth, this.cardHeight, 0x000000),
      this.rect(
        BORDER,
        BORDER,
        this.cardWidth - 2 * BORDER,
        this.cardHeight - 2 * BORDER,
        faceColor,
      ),
      this.rect(INSET, ART_Y, INNER_WIDTH, artHeight, 0x2ecc71),
      titleBelowArt
        ? this.label(INSET + INNER_WIDTH / 2, (ART_Y + artHeight + BOTTOM) / 2, this.name, {
            fontSize: '17.28px',
            fontStyle: 'bold',
            color: '#ffffff',
          }).setOrigin(0.5)
        : this.label(centerTitle ? this.cardWidth / 2 : INSET, TITLE_Y, this.name, {
            fontSize: '14.4px',
            fontStyle: 'bold',
            color: '#ffffff',
          }).setOrigin(centerTitle ? 0.5 : 0, 0),
    ]);

    if (showTextBox) {
      const top = ART_Y + artHeight + GAP;
      const box = textBoxRect ?? {
        x: INSET,
        y: top,
        w: INNER_WIDTH,
        h: BOTTOM - footerHeight - top,
      };
      const textX = box.x + box.w / 2;
      const textY = box.y + (centerText ? box.h / 2 : box.h - TEXT_PAD);
      const textStyle: Phaser.Types.GameObjects.Text.TextStyle = {
        ...TEXT_STYLE,
        fontSize: '11.2px',
        fontStyle: italicText ? 'italic' : 'normal',
        align: 'center',
        wordWrap: { width: box.w - 2 * TEXT_PAD },
      };
      this.surface.add([
        this.rect(box.x, box.y, box.w, box.h, 0xd8c9a3),
        this.flavor.includes('**')
          ? new RichText(
              this.scene,
              textX,
              textY,
              this.flavor,
              textStyle,
              box.w - 2 * TEXT_PAD,
              centerText ? 0.5 : 1,
            )
          : this.label(textX, textY, this.flavor, textStyle).setOrigin(0.5, centerText ? 0.5 : 1),
      ]);
    }

    this.disabledLayer = this.rect(0, 0, this.cardWidth, this.cardHeight, 0x111111, 0.5).setVisible(
      false,
    );
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
    const { BORDER, INSET } = Card;
    const INNER_WIDTH = this.cardWidth - 2 * INSET;
    const top = this.cardHeight - INSET - (this.config.footerHeight ?? 0);
    const y = (top + this.cardHeight - BORDER) / 2;
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
    if (!this.canBeDragged()) return;
    this.cancelHoverZoom();
    this.dragTarget = null;
    this.dragLast.set(pointer.x, pointer.y); // screen space: unaffected by which camera the pointer is currently over
    this.setFloating(true);
    this.emit('cardDragStart', this);
  }

  onDrag(pointer: Phaser.Input.Pointer) {
    if (!this.canBeDragged()) return;
    const zoom = this.dragZoom();
    this.dragTarget ??= new Phaser.Math.Vector2(this.x, this.y);
    this.dragTarget.x += (pointer.x - this.dragLast.x) / zoom;
    this.dragTarget.y += (pointer.y - this.dragLast.y) / zoom;
    this.setPosition(this.dragTarget.x, this.dragTarget.y);
    this.dragLast.set(pointer.x, pointer.y);
    this.emit('cardDragged', this);
  }

  // screen pixels per world pixel: board cards are seen through the zoomed main camera, hand cards through the ui camera
  private dragZoom(): number {
    const main = this.scene.cameras.main;
    return this.cameraFilter & main.id ? 1 : main.zoom;
  }

  onDragend() {
    if (!this.canBeDragged()) return;
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
    if (this.suppressed) return;
    this.emit('cardPointerOver', this);
    this.isHovering = true;
    this.setHoverLift(true);
    this.scheduleHoverZoom();
    if (this.tween?.isPlaying()) return;

    if (this.parentContainer) this.parentContainer.bringToTop(this);
    this.isHovering = true;
    this.startWiggle();
    this.emit('cardPointerOver', this);
  }

  // board cards only; hand cards are seen through the ui camera and the main camera ignores them
  private isBoardCard(): boolean {
    return !(this.cameraFilter & this.scene.cameras.main.id);
  }

  private scheduleHoverZoom(): void {
    if (this.floating || this.restScale !== null || this.zoomTimer || !this.isBoardCard()) return;
    this.zoomTimer = this.scene.time.delayedCall(Card.HOVER_ZOOM_DELAY, () => {
      this.zoomTimer = null;
      if (!this.active || !this.isHovering || this.floating) return;
      this.restScale = this.scaleX;
      this.scene.children.bringToTop(this);
      this.scene.children.queueDepthSort();
      this.tweenZoom(this.restScale * Card.HOVER_ZOOM_FACTOR);
    });
  }

  // ends the pending zoom or, if the card is already zoomed, shrinks it back to its rest scale
  private cancelHoverZoom(): void {
    this.zoomTimer?.remove();
    this.zoomTimer = null;
    if (this.restScale === null) return;
    const target = this.restScale;
    this.restScale = null;
    this.tweenZoom(target);
  }

  private tweenZoom(scale: number): void {
    this.zoomTween?.stop();
    this.zoomTween = this.tweensManager.add({
      targets: this,
      scale,
      duration: 250,
      ease: 'Sine.easeInOut',
    });
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
    on = on && !this.isInactive();
    if (on === this.isDropTarget) return;
    this.isDropTarget = on;
    if (on) this.startWiggle();
    else if (!this.isHovering) this.finishHover();
  }

  // on hover the card pops up (with its shadow) and settles back to rest, the wiggle signals the hover from there
  private setHoverLift(lifted: boolean): void {
    if (this.floating) return; //floating means it's beging dragged
    const restShadow = SHADOW_REST[this.config.shadow];
    this.hoverTweens.forEach((t) => t.stop());
    this.hoverTweens = [];

    const settle = () => {
      this.hoverTweens = [
        this.tweensManager.add({
          targets: this.cardFrame,
          scale: 1,
          y: 0,
          duration: 350,
          ease: 'Sine.easeInOut',
        }),
        this.tweensManager.add({
          targets: this.shadow,
          ...restShadow,
          duration: 350,
          ease: 'Sine.easeInOut',
        }),
      ];
    };

    if (!lifted) {
      settle();
      return;
    }

    this.hoverTweens = [
      this.tweensManager.add({
        targets: this.cardFrame,
        scale: 1.015,
        y: -4.8,
        duration: 260,
        ease: 'Sine.easeOut',
        onComplete: settle,
      }),
      this.tweensManager.add({
        targets: this.shadow,
        ...SHADOW_HOVER,
        duration: 260,
        ease: 'Sine.easeOut',
      }),
    ];
  }

  finishHover() {
    this.tween?.stop();
    this.tweensManager.add({ targets: this, angle: 0, duration: 600, ease: 'Sine.easeOut' });
  }

  onPointerOut() {
    this.isHovering = false;
    this.setHoverLift(false);
    this.cancelHoverZoom();
  }

  setDisabled(on: boolean) {
    this.isDisabled = on;
    this.refreshDisabledLayer();
  }

  // a temporary disable on top of the game-rule one, e.g. while a card is dragged that this one is no target of
  setSuppressed(on: boolean) {
    this.suppressed = on;
    this.refreshDisabledLayer();
  }

  protected isInactive(): boolean {
    return this.isDisabled || this.suppressed;
  }

  protected canBeDragged(): boolean {
    return !this.isInactive();
  }

  protected isSuppressed(): boolean {
    return this.suppressed;
  }

  isDisabledByRules(): boolean {
    return this.isDisabled;
  }

  private refreshDisabledLayer() {
    this.disabledLayer.setVisible(this.isInactive());
  }

  //a card that is only displayed (e.g. summoned) must not react to the pointer
  setPointerEnabled(on: boolean) {
    if (this.hitbox.input) this.hitbox.input.enabled = on;
  }

  getHitbox(): Phaser.Geom.Rectangle {
    return this.hitbox.getBounds();
  }
}
