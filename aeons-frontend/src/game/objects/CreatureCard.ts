import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { CreatureDTO } from './GameState';

export class CreatureCard extends Card {
  constructor(
    scene: MainScene,
    x: number,
    y: number,
    id: number,
    name: string,
    flavor: string,
    health: number,
    maxHealth: number,
    damage: number,
  ) {
    super(scene, x, y, {
      id: id,
      name: name,
      flavor: flavor,
      faceColor: 0xad3d2b,
      artHeight: 170,
      shadow: 'flat',
      italicText: true,
      footerHeight: 24,
    });

    this.initFrameDetails(maxHealth, health, damage);
  }

  initFrameDetails(maxHealth: integer, currentHealth: integer, damage: integer) {
    const hpPct = maxHealth > 0 ? Math.round((currentHealth / maxHealth) * 100) : 0;
    this.addFrameDetails(this.footer(`${hpPct}% HP    ·    ${damage} DMG`, 'center'));
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: CreatureDTO) {
    return new CreatureCard(
      scene,
      x,
      y,
      dto.id,
      dto.name,
      dto.flavor,
      dto.health,
      dto.maxHealth,
      dto.damage,
    );
  }
}
