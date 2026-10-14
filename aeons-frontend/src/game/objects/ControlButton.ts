import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';

//a game control button, same palette as the location connectors; centred on its own origin
export class ControlButton extends Phaser.GameObjects.Container {
  static readonly WIDTH = 230;
  static readonly HEIGHT = 48;
  private static readonly COLOR = 0x5c3b1f;
  private static readonly COLOR_HOVER = 0x74492a;
  private static readonly OUTLINE_COLOR = 0x2a1809;

  readonly label: Phaser.GameObjects.Text;

  constructor(scene: MainScene, text: string, onClick: () => void) {
    super(scene, 0, 0);
    const { WIDTH, HEIGHT, COLOR, COLOR_HOVER, OUTLINE_COLOR } = ControlButton;

    const background = new Phaser.GameObjects.Rectangle(scene, 0, 0, WIDTH, HEIGHT, COLOR)
      .setStrokeStyle(4, OUTLINE_COLOR)
      .setInteractive({ useHandCursor: true });
    this.label = new Phaser.GameObjects.Text(scene, 0, 0, text, {
      fontSize: '26px',
      fontStyle: 'bold',
      color: '#f3e3c3',
    }).setOrigin(0.5);

    background.on('pointerover', () => background.setFillStyle(COLOR_HOVER));
    background.on('pointerout', () => background.setFillStyle(COLOR));
    background.on('pointerup', onClick);

    this.add([background, this.label]);
  }
}
