import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { CURSOR_INTERACTIVE } from '../cursors';

export interface ButtonPalette {
  color: number;
  hover: number;
  outline: number;
}

//a game control button, same palette as the location connectors unless told otherwise; centred on its own origin
export class ControlButton extends Phaser.GameObjects.Container {
  static readonly WIDTH = 230; // the default, wider labels pass their own
  static readonly HEIGHT = 48;
  static readonly DEFAULT_PALETTE: ButtonPalette = { color: 0x5c3b1f, hover: 0x74492a, outline: 0x2a1809 };
  static readonly DEBUG_PALETTE: ButtonPalette = { color: 0x555555, hover: 0x8a8a8a, outline: 0x2b2b2b }; // marks playtest controls

  readonly label: Phaser.GameObjects.Text;
  private readonly background: Phaser.GameObjects.Rectangle;

  constructor(
    scene: MainScene,
    text: string,
    onClick: () => void,
    width = ControlButton.WIDTH,
    palette = ControlButton.DEFAULT_PALETTE,
  ) {
    super(scene, 0, 0);
    const { HEIGHT } = ControlButton;
    const { color, hover, outline } = palette;

    const background = new Phaser.GameObjects.Rectangle(scene, 0, 0, width, HEIGHT, color)
      .setStrokeStyle(4, outline)
      .setInteractive({ cursor: CURSOR_INTERACTIVE });
    this.label = new Phaser.GameObjects.Text(scene, 0, 0, text, {
      fontSize: '26px',
      fontStyle: 'bold',
      color: '#f3e3c3',
    }).setOrigin(0.5);

    background.on('pointerover', () => background.setFillStyle(hover));
    background.on('pointerout', () => background.setFillStyle(color));
    background.on('pointerup', onClick);

    this.background = background;
    this.add([background, this.label]);
  }

  setEnabled(enabled: boolean): void {
    this.background.input!.enabled = enabled;
    this.setAlpha(enabled ? 1 : 0.5);
  }
}
