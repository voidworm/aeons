import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { PlayerSwitcher } from './PlayerSwitcher';

//list of what happened above the phase buttons, newest at the bottom; the wheel scrolls through older entries
export class EventLog extends Phaser.GameObjects.Container {
  private static readonly VISIBLE = 5;
  private static readonly WIDTH = PlayerSwitcher.PHASE_BUTTONS_WIDTH;
  private static readonly ROW_HEIGHT = 30;
  private static readonly PADDING = 12;
  private static readonly HEIGHT = EventLog.VISIBLE * EventLog.ROW_HEIGHT + 2 * EventLog.PADDING;
  private static readonly OUTLINE_COLOR = 0x2a1809;

  readonly dragHandle: Phaser.GameObjects.Rectangle;
  private entries: string[] = [];
  private firstShown = 0; // index of the topmost visible entry
  private followsNewest = true;
  private rows: Phaser.GameObjects.Text[] = [];

  constructor(scene: MainScene) {
    super(scene, 0, 0);
    const { WIDTH, HEIGHT, ROW_HEIGHT, PADDING, VISIBLE, OUTLINE_COLOR } = EventLog;

    this.dragHandle = new Phaser.GameObjects.Rectangle(scene, 0, 0, WIDTH, HEIGHT, 0x1b1209, 0.8)
      .setOrigin(0, 0)
      .setStrokeStyle(3, OUTLINE_COLOR);
    this.dragHandle.on('wheel', (_pointer: Phaser.Input.Pointer, _dx: number, deltaY: number) =>
      this.scroll(Math.sign(deltaY)),
    );
    this.add(this.dragHandle);

    for (let i = 0; i < VISIBLE; i++) {
      const row = new Phaser.GameObjects.Text(scene, PADDING, PADDING + i * ROW_HEIGHT, '', {
        fontSize: '20px',
        color: '#f3e3c3',
      });
      this.rows.push(row);
      this.add(row);
    }

    this.setDepth(1000);
    scene.add.existing(this);
  }

  setBottomCentre(x: number, y: number): void {
    const width = EventLog.WIDTH * this.scaleX;
    const height = EventLog.HEIGHT * this.scaleY;
    this.setPosition(
      Phaser.Math.Clamp(x - width / 2, 0, this.scene.scale.width - width),
      Math.max(0, y - height),
    );
  }

  append(entries: string[]): void {
    this.entries.push(...entries);
    if (this.followsNewest) this.firstShown = this.lastFirstShown();
    this.render();
  }

  containsPoint(x: number, y: number): boolean {
    const { WIDTH, HEIGHT } = EventLog;
    return new Phaser.Geom.Rectangle(
      this.x,
      this.y,
      WIDTH * this.scaleX,
      HEIGHT * this.scaleY,
    ).contains(x, y);
  }

  private scroll(direction: number): void {
    this.firstShown = Phaser.Math.Clamp(this.firstShown + direction, 0, this.lastFirstShown());
    this.followsNewest = this.firstShown === this.lastFirstShown();
    this.render();
  }

  private lastFirstShown(): number {
    return Math.max(0, this.entries.length - EventLog.VISIBLE);
  }

  private render(): void {
    this.rows.forEach((row, i) => row.setText(this.entries[this.firstShown + i] ?? ''));
  }
}
