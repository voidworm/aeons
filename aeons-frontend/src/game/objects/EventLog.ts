import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { CURSOR_GRAB, CURSOR_INTERACTIVE } from '../cursors';

//upper left list of what happened, newest at the bottom; the wheel scrolls through older entries
export class EventLog extends Phaser.GameObjects.Container {
  private static readonly VISIBLE = 5;
  private static readonly WIDTH = 520;
  private static readonly ROW_HEIGHT = 30;
  private static readonly PADDING = 12;
  private static readonly MARGIN = 20;
  private static readonly HEIGHT = EventLog.VISIBLE * EventLog.ROW_HEIGHT + 2 * EventLog.PADDING;
  private static readonly OUTLINE_COLOR = 0x2a1809;

  private entries: string[] = [];
  private firstShown = 0; // index of the topmost visible entry
  private followsNewest = true;
  private rows: Phaser.GameObjects.Text[] = [];

  constructor(scene: MainScene) {
    super(scene, EventLog.MARGIN, EventLog.MARGIN);
    const { WIDTH, HEIGHT, ROW_HEIGHT, PADDING, VISIBLE, OUTLINE_COLOR } = EventLog;

    const background = new Phaser.GameObjects.Rectangle(scene, 0, 0, WIDTH, HEIGHT, 0x1b1209, 0.8)
      .setOrigin(0, 0)
      .setStrokeStyle(3, OUTLINE_COLOR)
      .setInteractive({ draggable: true, cursor: CURSOR_INTERACTIVE }); // also keeps hovers and clicks from reaching the board underneath
    this.enableDragging(background);
    background.on('wheel', (_pointer: Phaser.Input.Pointer, _dx: number, deltaY: number) =>
      this.scroll(Math.sign(deltaY)),
    );
    this.add(background);

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

  //moves by pointer deltas in screen space, the panel stays fully on screen
  private enableDragging(handle: Phaser.GameObjects.Rectangle): void {
    const last = new Phaser.Math.Vector2();
    handle.on('dragstart', (pointer: Phaser.Input.Pointer) => {
      last.set(pointer.x, pointer.y);
      this.scene.game.canvas.style.cursor = CURSOR_GRAB;
    });
    handle.on('drag', (pointer: Phaser.Input.Pointer) => {
      const { width, height } = this.scene.scale;
      this.setPosition(
        Phaser.Math.Clamp(this.x + pointer.x - last.x, 0, width - EventLog.WIDTH),
        Phaser.Math.Clamp(this.y + pointer.y - last.y, 0, height - EventLog.HEIGHT),
      );
      last.set(pointer.x, pointer.y);
    });
    handle.on('dragend', () => {
      this.scene.game.canvas.style.cursor = CURSOR_INTERACTIVE;
    });
  }

  append(entries: string[]): void {
    this.entries.push(...entries);
    if (this.followsNewest) this.firstShown = this.lastFirstShown();
    this.render();
  }

  containsPoint(x: number, y: number): boolean {
    const { WIDTH, HEIGHT } = EventLog;
    return new Phaser.Geom.Rectangle(this.x, this.y, WIDTH, HEIGHT).contains(x, y);
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
