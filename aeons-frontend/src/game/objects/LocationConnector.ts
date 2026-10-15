import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import type { LocationDTO } from './GameState';
import type { LocationCard } from './LocationCard';

interface Link {
  from: LocationCard;
  to: LocationCard;
  bothWays: boolean;
}

interface Palette {
  outline: number;
  inline: number | null; // a veil only needs the outline pass, the inline would be veiled twice
  tip: number;
  tipOutline: number;
  alpha: number;
}

//draws a trail between the centres of connected locations, behind every card;
//flint arrowheads in the middle of the trail show which ways it can be walked
export class LocationConnector extends Phaser.GameObjects.Graphics {
  private static readonly OUTLINE_WIDTH = 12;
  private static readonly INLINE_WIDTH = 7;
  private static readonly TIP_OUTLINE_WIDTH = 4;
  private static readonly TIP_LENGTH = 44;
  private static readonly TIP_GAP = 28; // between the bases of a two way trail's tips
  private static readonly NORMAL: Palette = {
    outline: 0x2a1809,
    inline: 0x5c3b1f,
    tip: 0x9a9488,
    tipOutline: 0x2a1809,
    alpha: 1,
  };
  // laid over a trail that cannot be walked, the same veil a disabled card gets
  private static readonly VEIL: Palette = {
    outline: 0x111111,
    inline: null,
    tip: 0x111111,
    tipOutline: 0x111111,
    alpha: 0.5,
  };
  // arrowhead pointing along +x with its apex at the origin, so it can be rotated and placed along the trail
  private static readonly TIP_SHAPE = [
    [0, 0],
    [-30, -19],
    [-26, -11],
    [-44, -15],
    [-38, 0],
    [-44, 15],
    [-26, 11],
    [-30, 19],
  ];

  private links: Link[] = [];
  private focus: LocationCard | null = null; // while set, only the trails leading away from it stay lit

  constructor(scene: MainScene, dtos: LocationDTO[], cards: LocationCard[]) {
    super(scene);
    this.setDepth(-1);

    // a connection listed by one side only is one way, each pair is drawn once
    const listed = new Set(dtos.flatMap((dto) => dto.connections.map((id) => `${dto.id}>${id}`)));
    const seen = new Set<string>();
    for (const dto of dtos) {
      for (const otherId of dto.connections) {
        const key = [dto.id, otherId].sort((a, b) => a - b).join('-');
        const from = cards.find((c) => c.id === dto.id);
        const to = cards.find((c) => c.id === otherId);
        if (seen.has(key) || !from || !to) continue;
        seen.add(key);
        this.links.push({ from, to, bothWays: listed.has(`${otherId}>${dto.id}`) });
      }
    }

    scene.add.existing(this);
    this.redraw();
  }

  redraw(): void {
    const { NORMAL, VEIL } = LocationConnector;
    this.clear();
    // the veiled links go first, so a trail that can be walked is never veiled where they cross
    // all trails are drawn before any tip, so no trail runs over a tip
    const veiled = this.links.filter((link) => !this.isWalkable(link));
    const walkable = this.links.filter((link) => this.isWalkable(link));
    this.drawTrails(veiled, NORMAL);
    this.drawTrails(veiled, VEIL);
    this.drawTrails(walkable, NORMAL);
    this.drawAllTips(veiled, NORMAL);
    this.drawAllTips(veiled, VEIL);
    this.drawAllTips(walkable, NORMAL);
  }

  //veils every trail that cannot be walked away from the location
  focusOn(location: LocationCard): void {
    this.focus = location;
  }

  clearFocus(): void {
    this.focus = null;
  }

  private isWalkable(link: Link): boolean {
    return !this.focus || link.from === this.focus || (link.to === this.focus && link.bothWays);
  }

  private drawTrails(links: Link[], palette: Palette): void {
    const { OUTLINE_WIDTH, INLINE_WIDTH } = LocationConnector;
    // all outlines first, so crossing trails don't get an outline drawn over their inline
    const passes: [number, number | null][] = [
      [OUTLINE_WIDTH, palette.outline],
      [INLINE_WIDTH, palette.inline],
    ];
    for (const [width, color] of passes) {
      if (color === null) continue;
      this.lineStyle(width, color, palette.alpha);
      for (const { from, to } of links) this.lineBetween(from.x, from.y, to.x, to.y);
    }
  }

  private drawAllTips(links: Link[], palette: Palette): void {
    for (const link of links) this.drawTips(link, palette);
  }

  //the tips keep their size however far the locations are apart, only the trail stretches
  //a one way trail gets a single tip centred on the trail, a two way trail two tips back to back
  private drawTips({ from, to, bothWays }: Link, palette: Palette): void {
    const { TIP_LENGTH, TIP_GAP } = LocationConnector;
    const angle = Phaser.Math.Angle.Between(from.x, from.y, to.x, to.y);
    const middle = new Phaser.Math.Vector2((from.x + to.x) / 2, (from.y + to.y) / 2);
    const lengthwise = (distance: number) =>
      new Phaser.Math.Vector2(distance * from.scaleX, 0).rotate(angle);

    if (!bothWays) {
      this.drawTip(middle.clone().add(lengthwise(TIP_LENGTH / 2)), angle, from.scaleX, palette);
      return;
    }
    const apexOffset = lengthwise(TIP_LENGTH + TIP_GAP / 2);
    this.drawTip(middle.clone().add(apexOffset), angle, from.scaleX, palette);
    this.drawTip(middle.clone().subtract(apexOffset), angle + Math.PI, from.scaleX, palette);
  }

  private drawTip(apex: Phaser.Math.Vector2, angle: number, scale: number, palette: Palette): void {
    const { TIP_SHAPE, TIP_OUTLINE_WIDTH } = LocationConnector;
    const outline = TIP_SHAPE.map(([x, y]) =>
      new Phaser.Math.Vector2(x, y).rotate(angle).scale(scale).add(apex),
    );
    this.fillStyle(palette.tip, palette.alpha).fillPoints(outline, true);
    this.lineStyle(TIP_OUTLINE_WIDTH, palette.tipOutline, palette.alpha).strokePoints(
      outline,
      true,
    );
  }
}
