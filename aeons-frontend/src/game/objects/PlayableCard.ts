import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import Phaser from 'phaser';
import type { PlayableCardDTO } from './GameState';

export class PlayableCard extends Card {
  constructor(
    scene: MainScene,
    x: number,
    y: number,
    id: number,
    cardname: string,
    flavor: string,
    cost: number,
  ) {
    super(scene, x, y, {
      id: id,
      name: cardname,
      flavor: flavor,
      faceColor: 0x8b5e34,
      artHeight: 136,
      shadow: 'hovers',
    });

    const cx = Card.WIDTH - Card.INSET - 10.4;
    const cy = Card.TITLE_Y + 8.8;
    this.addFrameDetails(
      new Phaser.GameObjects.Arc(scene, cx, cy, 10.4, 0, 360, false, 0xc0c0c0),
      this.label(cx, cy, String(cost), { fontSize: '14.4px' }).setOrigin(0.5),
    );
  }

  static fromDto(scene: MainScene, x: integer, y: integer, dto: PlayableCardDTO) {
    const r = new PlayableCard(scene, x, y, dto.id, dto.name, dto.effectText, dto.cost);
    r.setDisabled(dto.canBeCast);
    return r;
  }
}
