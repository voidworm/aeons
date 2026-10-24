import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';

//curved arrow in screen space, from the summoned card to the pointer
export class TargetingArrow extends Phaser.GameObjects.Graphics {
  private static readonly COLOR = 0xf3e3c3;
  private static readonly LINE_WIDTH = 6;
  private static readonly HEAD_LENGTH = 26;
  private static readonly HEAD_WIDTH = 22;
  private static readonly BEND = 0.25; // how far the curve bows upwards, relative to the arrow length
  private static readonly DEPTH = 1100; // above the event log

  constructor(scene: MainScene) {
    super(scene);
    this.setDepth(TargetingArrow.DEPTH);
    scene.add.existing(this);
  }

  pointFromTo(from: Phaser.Math.Vector2, to: Phaser.Math.Vector2): void {
    const { COLOR, LINE_WIDTH, HEAD_LENGTH, HEAD_WIDTH, BEND } = TargetingArrow;
    const control = new Phaser.Math.Vector2(
      (from.x + to.x) / 2,
      Math.min(from.y, to.y) - from.distance(to) * BEND,
    );
    const curve = new Phaser.Curves.QuadraticBezier(from, control, to);
    const direction = curve.getTangent(1);
    const headBase = to.clone().subtract(direction.clone().scale(HEAD_LENGTH));
    const side = new Phaser.Math.Vector2(-direction.y, direction.x).scale(HEAD_WIDTH / 2);

    this.clear();
    this.lineStyle(LINE_WIDTH, COLOR, 1);
    this.strokePoints(curve.getPoints(32));
    this.fillStyle(COLOR, 1);
    this.fillTriangle(
      to.x,
      to.y,
      headBase.x + side.x,
      headBase.y + side.y,
      headBase.x - side.x,
      headBase.y - side.y,
    );
  }
}
