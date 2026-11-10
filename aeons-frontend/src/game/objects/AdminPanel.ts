import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { CARD_FONT_FAMILY } from '../fonts';
import { ControlButton } from './ControlButton';

//playtest controls in the upper left corner, the origin is the panel's top left corner
//only the veil toggle is there at first, the rest appears once the veil is lifted
export class AdminPanel extends Phaser.GameObjects.Container {
  static readonly WIDTH = 384;
  private static readonly PADDING = 14;
  private static readonly ROW_GAP = 10;
  private static readonly COLOR = 0x2a1809;
  private static readonly OUTLINE = 0x5c3b1f;
  private static readonly TEXT_COLOR = '#c9b48f';

  private readonly background: Phaser.GameObjects.Rectangle;
  private readonly fogButton: ControlButton;
  private readonly readout: Phaser.GameObjects.Text;
  private readonly worldwakeButton: ControlButton;
  private readonly activationOrder: Phaser.GameObjects.Text;
  private lifted = false;

  constructor(scene: MainScene, onFogToggle: (lifted: boolean) => void, onWorldwake: () => void) {
    super(scene, 0, 0);
    const { WIDTH, PADDING, COLOR, OUTLINE } = AdminPanel;
    const innerWidth = WIDTH - 2 * PADDING;

    this.background = new Phaser.GameObjects.Rectangle(scene, 0, 0, WIDTH, 0, COLOR, 0.85)
      .setOrigin(0, 0)
      .setStrokeStyle(4, OUTLINE)
      .setInteractive(); // keeps hovers and clicks from reaching the board underneath

    this.fogButton = new ControlButton(scene, 'Lift the veil', () => {
      this.lifted = !this.lifted;
      this.fogButton.label.setText(this.lifted ? 'The mist comes...' : 'Lift the veil');
      this.layout();
      onFogToggle(this.lifted);
    }, innerWidth);
    this.readout = this.createText(scene, 20);
    this.worldwakeButton = new ControlButton(scene, 'Worldwake', onWorldwake, innerWidth);
    this.activationOrder = this.createText(scene, 22);

    this.add([this.background, this.fogButton, this.readout, this.worldwakeButton, this.activationOrder]);
    this.layout();
    this.setDepth(1000);
    scene.add.existing(this);
  }

  setReadout(zoomPercent: number, fps: number): void {
    this.readout.setText(`zoom ${zoomPercent}% · ${fps} fps`);
  }

  setWorldwakeEnabled(enabled: boolean): void {
    this.worldwakeButton.setEnabled(enabled);
  }

  setActivationOrder(names: string[]): void {
    this.activationOrder.setText(names.map((name, i) => `${i + 1}. ${name}`).join('\n'));
    this.layout();
  }

  //stacks the visible rows from the top and fits the background around them
  private layout(): void {
    const { WIDTH, PADDING, ROW_GAP } = AdminPanel;
    const { HEIGHT } = ControlButton;
    const centreX = WIDTH / 2;
    let y = PADDING;

    this.fogButton.setPosition(centreX, y + HEIGHT / 2);
    y += HEIGHT;

    const showOrder = this.lifted && this.activationOrder.text !== '';
    this.readout.setVisible(this.lifted);
    this.worldwakeButton.setVisible(this.lifted);
    this.activationOrder.setVisible(showOrder);
    if (this.lifted) {
      y += ROW_GAP;
      this.readout.setPosition(PADDING, y);
      y += this.readout.height + ROW_GAP;
      this.worldwakeButton.setPosition(centreX, y + HEIGHT / 2);
      y += HEIGHT;
    }
    if (showOrder) {
      y += ROW_GAP;
      this.activationOrder.setPosition(PADDING, y);
      y += this.activationOrder.height;
    }
    y += PADDING;
    this.background.setSize(WIDTH, y);
    this.setSize(WIDTH, y);
  }

  private createText(scene: MainScene, size: number): Phaser.GameObjects.Text {
    return new Phaser.GameObjects.Text(scene, 0, 0, '', {
      fontFamily: CARD_FONT_FAMILY,
      fontSize: `${size}px`,
      color: AdminPanel.TEXT_COLOR,
    });
  }
}
