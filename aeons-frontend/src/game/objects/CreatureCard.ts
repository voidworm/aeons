import type Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { CreatureDTO } from './GameState';

export class CreatureCard extends Card {
  private exhausted = false;
  private turnTween: Phaser.Tweens.Tween | null = null;
  private stats!: Phaser.GameObjects.Text;
  private maxHealth: number;

  constructor(
    scene: MainScene,
    x: number,
    y: number,
    id: number,
    name: string,
    flavor: string,
    health: number,
    maxHealth: number,
  ) {
    super(scene, x, y, {
      id: id,
      name: name,
      flavor: flavor,
      faceColor: 0xad3d2b,
      artHeight: 136,
      shadow: 'flat',
      italicText: true,
      footerHeight: 19.2,
    });

    this.maxHealth = maxHealth;
    this.initFrameDetails(health);
  }

  // an exhausted creature lies on its side; the hitbox turns with it
  setExhausted(on: boolean, animate = true): Promise<void> {
    if (on === this.exhausted) return Promise.resolve();
    this.exhausted = on;
    const angle = on ? 90 : 0;
    this.turnTween?.stop();
    if (!animate) {
      this.cardFrame.setAngle(angle);
      this.hitbox.setAngle(angle);
      return Promise.resolve();
    }
    return new Promise((resolve) => {
      this.turnTween = this.tweensManager.add({
        targets: [this.cardFrame, this.hitbox],
        angle,
        duration: 400,
        ease: 'Sine.easeInOut',
        onComplete: () => resolve(),
        onStop: () => resolve(),
      });
    });
  }

  initFrameDetails(currentHealth: integer) {
    this.stats = this.footer(this.statsText(currentHealth), 'center');
    this.addFrameDetails(this.stats);
  }

  setHealth(health: number) {
    this.stats.setText(this.statsText(health));
  }

  private statsText(health: number): string {
    const hpPct = this.maxHealth > 0 ? Math.round((health / this.maxHealth) * 100) : 0;
    return `${hpPct}% HP`;
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: CreatureDTO) {
    const card = new CreatureCard(
      scene,
      x,
      y,
      dto.id,
      dto.name,
      dto.flavor,
      dto.health,
      dto.maxHealth,
    );
    void card.setExhausted(dto.exhausted, false);
    return card;
  }
}
