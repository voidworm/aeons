import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { ControlButton } from './ControlButton';

//modal notice that a player fainted, closes on its button
export class FaintedPopup extends Phaser.GameObjects.Container {
  private static readonly WIDTH = 480;
  private static readonly HEIGHT = 260;
  private static readonly COLOR = 0xd8c9a3; // same as the card text boxes
  private static readonly OUTLINE_COLOR = 0x2a1809;
  private static readonly DEPTH = 2000;

  constructor(scene: MainScene, playerName: string) {
    const { width, height } = scene.scale;
    super(scene, width / 2, height / 2);
    const { WIDTH, HEIGHT, COLOR, OUTLINE_COLOR } = FaintedPopup;

    // swallows clicks, so nothing behind the popup reacts
    const dim = new Phaser.GameObjects.Rectangle(scene, 0, 0, width, height, 0x000000, 0.5).setInteractive();
    const panel = new Phaser.GameObjects.Rectangle(scene, 0, 0, WIDTH, HEIGHT, COLOR).setStrokeStyle(4, OUTLINE_COLOR);
    const title = this.text(scene, `${playerName} fainted`, -HEIGHT / 2 + 40, '34px', true);
    const message = this.text(scene, 'The day has ended,\nyou must rest now.', 0, '30px');
    const close = new ControlButton(scene, 'OK', () => this.destroy());
    close.setY(HEIGHT / 2 - 55);

    this.add([dim, panel, title, message, close]);
    this.setDepth(FaintedPopup.DEPTH);
    scene.add.existing(this);
  }

  private text(scene: MainScene, content: string, y: number, fontSize: string, bold = false) {
    return new Phaser.GameObjects.Text(scene, 0, y, content, {
      fontSize,
      fontStyle: bold ? 'bold' : 'normal',
      color: '#000000',
      align: 'center',
      lineSpacing: 10,
    }).setOrigin(0.5);
  }
}
