import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';

export class Card extends Phaser.GameObjects.Container {
  public floating: boolean = false;
  public name: string = 'Generic card';
  public effect: string = 'Generic effect';

  public pointed: boolean = false;

  public isHovering = false;
  protected tween: Phaser.Tweens.Tween | null = null;
  protected tweensManager: Phaser.Tweens.TweenManager;

  protected shadow!: Phaser.GameObjects.Rectangle;
  protected disabledLayer!: Phaser.GameObjects.Rectangle;
  protected isDisabled: boolean;
  protected content!: Phaser.GameObjects.Container;

  constructor(scene: MainScene, x: number, y: number, cardname: string, cardtext: string) {
    super(scene, x, y);
    this.isDisabled = false;
    this.setSize(315, 440);
    this.setInteractive({ draggable: true });
    this.name = cardname;
    this.effect = cardtext;
    this.tweensManager = scene.tweens;

    this.on('dragstart', () => {
      this.onDragStart();
    });
    this.on('drag', (pointer: Phaser.Input.Pointer, dragX: number, dragY: number) => {
      this.onDrag(pointer, dragX, dragY);
    });
    this.on('dragend', () => this.onDragend());

    this.on('pointerover', () => {
      this.onPointerOver();
    });
    this.on('pointerout', () => {
      this.onPointerOut();
    });
  }

  onDragStart() {
    if (this.isDisabled) return;

    this.setFloating(true);
    this.emit('cardDragStart', this);
  }

  setFloating(input: boolean) {
    this.floating = input;
    this.tweensManager.add({
      targets: this.content,
      scale: input ? 1.05 : 1,
      y: input ? -10 : 0,
      duration: 150,
      ease: 'Sine.easeOut',
    });
    this.tweensManager.add({
      targets: this.shadow,
      scale: input ? 1.1 : 1,
      x: input ? 20 : 14,
      y: input ? -20 : -14,
      alpha: input ? 0.5 : 1,
      duration: 150,
      ease: 'Sine.easeOut',
    });
  }

  onDrag(_pointer: Phaser.Input.Pointer, dragX: number, dragY: number) {
    if (this.isDisabled) return;

    this.x = dragX;
    this.y = dragY;
    this.emit('cardDragged', this);
  }

  onDragend() {
    if (this.isDisabled) return;

    this.setFloating(false);
    this.emit('cardDragEnd', this);
  }

  onPointerOver() {
    this.isHovering = true;
    if (this.tween?.isPlaying()) return;

    this.tween = this.tweensManager.add({
      targets: this,
      angle: { from: -0.5, to: 0.5 },
      duration: 600,
      yoyo: true,
      repeat: -1,
      ease: 'Sine.easeInOut',
      onRepeat: () => {
        if (!this.isHovering) this.finishHover();
      },
    });
  }

  finishHover() {
    this.tween?.stop();
    this.tweensManager.add({
      targets: this,
      angle: 0,
      duration: 600,
      ease: 'Sine.easeOut',
    });
  }

  onPointerOut() {
    this.isHovering = false;
  }

  setDiabled(on: boolean) {
    this.isDisabled = on;
    this.disabledLayer.setVisible(on);
  }

  getOverlapBounds(): Phaser.Geom.Rectangle {
    return this.content.getBounds();
  }
}
