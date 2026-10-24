import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import type { Card } from './Card';

//curved arrow in screen space, from the summoned card to the pointer
export class TargetingArrow extends Phaser.GameObjects.Graphics {
  static readonly COLOR = 0xf3e3c3;
  private static readonly LINE_WIDTH = 6;
  private static readonly HEAD_LENGTH = 26;
  private static readonly HEAD_WIDTH = 22;
  private static readonly HEAD_ROUNDING = 4; // corner radius of the head
  private static readonly BEND = 0.25; // how far the curve bows upwards, relative to the arrow length
  private static readonly DEPTH = 1100; // above the event log

  private readonly start: Phaser.GameObjects.Graphics; // only seen by the ui camera
  private behind: Card | null = null; // the card this arrow is currently a child of

  constructor(scene: MainScene) {
    super(scene);
    this.setDepth(TargetingArrow.DEPTH);
    scene.add.existing(this);
    this.start = scene.add.graphics().setDepth(TargetingArrow.DEPTH);
  }

  override destroy(fromScene?: boolean): void {
    this.start.destroy();
    super.destroy(fromScene);
  }

  //origin is the screen rectangle of the card the arrow starts from; it stays on top of that card even while the arrow runs behind another one
  pointFromTo(from: Phaser.Math.Vector2, to: Phaser.Math.Vector2, origin: Phaser.Geom.Rectangle): void {
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
    this.drawStart(shaft, origin);
  }

  //a copy of the shaft's first stretch, from the start until it leaves the origin card, drawn by the ui camera above everything
  private drawStart(shaft: Phaser.Curves.QuadraticBezier, origin: Phaser.Geom.Rectangle): void {
    const STEPS = 64;
    this.start.clear();
    if (!this.behind) return;
    const points: Phaser.Math.Vector2[] = [];
    for (let i = 0; i <= STEPS; i++) {
      const point = shaft.getPoint(i / STEPS);
      points.push(point);
      if (!origin.contains(point.x, point.y)) break;
    }
    this.start.lineStyle(TargetingArrow.LINE_WIDTH, TargetingArrow.COLOR, 1);
    this.start.strokePoints(points);
  }

  //the default: above everything, seen by the ui camera in screen space
  runAbove(main: Phaser.Cameras.Scene2D.Camera): void {
    this.cameraFilter = main.id;
    this.setScale(1).setRotation(0).setPosition(0, 0).setDepth(TargetingArrow.DEPTH);
    if (!this.behind) return;
    this.behind.remove(this);
    this.behind = null;
    this.scene.add.existing(this);
  }

  //slipped into the card between its shadow and its face, so only the face covers the arrow's end
  //the card's transform and the main camera's zoom and scroll are undone, so the arrow keeps drawing in screen coordinates
  runBehind(card: Card, main: Phaser.Cameras.Scene2D.Camera, ui: Phaser.Cameras.Scene2D.Camera): void {
    if (this.behind !== card) {
      this.behind?.remove(this);
      this.behind = card;
      card.addAt(this, 1);
    }
    this.cameraFilter = ui.id;
    this.start.cameraFilter = main.id;
    const scale = 1 / (main.zoom * card.scaleX);
    const half = new Phaser.Math.Vector2(main.width / 2, main.height / 2);
    const world = new Phaser.Math.Vector2(main.scrollX, main.scrollY).add(half).subtract(half.clone().scale(1 / main.zoom));
    const local = world.subtract(card).rotate(-card.rotation).scale(1 / card.scaleX);
    this.setScale(scale).setRotation(-card.rotation).setPosition(local.x, local.y);
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
