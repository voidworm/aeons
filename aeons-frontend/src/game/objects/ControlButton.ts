import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { HandContainer } from './HandContainer';

//a game control button that lives in the hand area, bottom left
export class ControlButton extends Phaser.GameObjects.Container {
  private static readonly WIDTH = 220;
  private static readonly HEIGHT = 48;
  private static readonly MARGIN_BOTTOM = 20;

  constructor(scene: MainScene, label: string, onClick: () => void) {
    super(scene, 0, 0);
    const { WIDTH, HEIGHT } = ControlButton;

    const background = new Phaser.GameObjects.Rectangle(scene, 0, 0, WIDTH, HEIGHT, 0x8b5e34)
      .setStrokeStyle(2, 0xf3e3c3)
      .setInteractive({ useHandCursor: true });
    const text = new Phaser.GameObjects.Text(scene, 0, 0, label, {
      fontSize: '20px',
      fontStyle: 'bold',
      color: '#f3e3c3',
    }).setOrigin(0.5);

    background.on('pointerover', () => background.setFillStyle(0xa97544));
    background.on('pointerout', () => background.setFillStyle(0x8b5e34));
    background.on('pointerup', onClick);

    this.add([background, text]);
    this.setDepth(1000); // same layer as the player switcher, above the hand
    this.resize(scene.scale.width, scene.scale.height);
    scene.add.existing(this);
  }

  resize(_width: number, height: number): void {
    this.setPosition(
      HandContainer.PADDING + ControlButton.WIDTH / 2,
      height - ControlButton.MARGIN_BOTTOM - ControlButton.HEIGHT / 2,
    );
  }
}
