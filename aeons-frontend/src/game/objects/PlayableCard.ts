import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import Phaser from 'phaser';
import type { PlayableCardDTO } from './GameState';

export class PlayableCard extends Card {
  constructor(
    scene: MainScene,
    x: number,
    y: number,
    cardname: string,
    flavor: string,
    cost: number,
  ) {
    super(scene, x, y, {
      name: cardname,
      flavor: flavor,
      faceColor: 0x8b5e34,
      artHeight: 170,
      shadow: 'hovers',
    });

    const cx = Card.WIDTH - Card.INSET - 13;
    const cy = Card.TITLE_Y + 11;
    this.addFrameDetails(
      new Phaser.GameObjects.Arc(scene, cx, cy, 13, 0, 360, false, 0xc0c0c0),
      this.label(cx, cy, String(cost), { fontSize: '18px' }).setOrigin(0.5),
    );
  }

  static fromDto(scene: MainScene, x: integer, y: integer, dto: PlayableCardDTO) {
    const r = new PlayableCard(scene, x, y, dto.name, dto.effectText, dto.cost);
    r.setDisabled(dto.canBeCast);
    return r;
  }
}
