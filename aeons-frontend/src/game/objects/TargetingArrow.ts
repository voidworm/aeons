import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';

//curved arrow in screen space, from the summoned card to the pointer
export class TargetingArrow extends Phaser.GameObjects.Graphics {
  private static readonly COLOR = 0xf3e3c3;
  private static readonly LINE_WIDTH = 6;
  private static readonly HEAD_LENGTH = 26;
  private static readonly HEAD_WIDTH = 22;
  private static readonly HEAD_ROUNDING = 4; // corner radius of the head
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

    const shaft = new Phaser.Curves.QuadraticBezier(from, control, headBase);

    this.clear();
    this.lineStyle(LINE_WIDTH, COLOR, 1);
    this.strokePoints(shaft.getPoints(32));
    this.fillStyle(COLOR, 1);
    this.fillRoundedTriangle(to, headBase.clone().add(side), headBase.clone().subtract(side));
  }

  //the triangle is shrunk towards its incentre by the corner radius, then every corner is rounded by that radius
  private fillRoundedTriangle(
    a: Phaser.Math.Vector2,
    b: Phaser.Math.Vector2,
    c: Phaser.Math.Vector2,
  ): void {
    const radius = TargetingArrow.HEAD_ROUNDING;
    const [lenA, lenB, lenC] = [b.distance(c), a.distance(c), a.distance(b)];
    const perimeter = lenA + lenB + lenC;
    const incentre = a
      .clone()
      .scale(lenA)
      .add(b.clone().scale(lenB))
      .add(c.clone().scale(lenC))
      .scale(1 / perimeter);
    const s = perimeter / 2;
    const inradius = Math.sqrt(((s - lenA) * (s - lenB) * (s - lenC)) / s);
    const shrink = (inradius - radius) / inradius;
    const [pa, pb, pc] = [a, b, c].map((p) =>
      incentre.clone().add(p.clone().subtract(incentre).scale(shrink)),
    );

    this.fillTriangle(pa.x, pa.y, pb.x, pb.y, pc.x, pc.y);
    this.lineStyle(radius * 2, TargetingArrow.COLOR, 1);
    this.strokeTriangle(pa.x, pa.y, pb.x, pb.y, pc.x, pc.y);
    [pa, pb, pc].forEach((p) => this.fillCircle(p.x, p.y, radius));
  }
}
