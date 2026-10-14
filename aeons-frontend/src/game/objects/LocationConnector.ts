import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import type { LocationDTO } from './GameState';
import type { LocationCard } from './LocationCard';

//draws a line between the centres of connected locations, behind every card
export class LocationConnector extends Phaser.GameObjects.Graphics {
  private static readonly OUTLINE_WIDTH = 12;
  private static readonly OUTLINE_COLOR = 0x2a1809;
  private static readonly INLINE_WIDTH = 7;
  private static readonly INLINE_COLOR = 0x5c3b1f;

  private links: [LocationCard, LocationCard][] = [];

  constructor(scene: MainScene, dtos: LocationDTO[], cards: LocationCard[]) {
    super(scene);
    this.setDepth(-1);

    // a connection may be listed by one side only, each pair is drawn once
    const seen = new Set<string>();
    for (const dto of dtos) {
      for (const otherId of dto.connections) {
        const key = [dto.id, otherId].sort((a, b) => a - b).join('-');
        const from = cards.find((c) => c.id === dto.id);
        const to = cards.find((c) => c.id === otherId);
        if (seen.has(key) || !from || !to) continue;
        seen.add(key);
        this.links.push([from, to]);
      }
    }

    scene.add.existing(this);
    this.redraw();
  }

  redraw(): void {
    const { OUTLINE_WIDTH, OUTLINE_COLOR, INLINE_WIDTH, INLINE_COLOR } = LocationConnector;
    this.clear();
    // all outlines first, so crossing trails don't get an outline drawn over their inline
    for (const [width, color] of [
      [OUTLINE_WIDTH, OUTLINE_COLOR],
      [INLINE_WIDTH, INLINE_COLOR],
    ]) {
      this.lineStyle(width, color, 1);
      for (const [a, b] of this.links) this.lineBetween(a.x, a.y, b.x, b.y);
    }
  }
}
