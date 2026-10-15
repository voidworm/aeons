import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';

export interface PlayerStats {
  energy: number;
  health: number;
  maxHealth: number;
  resources: number;
  location: string;
}

//small readout of the active player, origin is the panel's centre
export class PlayerStatsPanel extends Phaser.GameObjects.Container {
  static readonly WIDTH = 330;
  static readonly HEIGHT = 150;
  private static readonly COLOR = 0x5c3b1f;
  private static readonly OUTLINE_COLOR = 0x2a1809;

  readonly dragHandle: Phaser.GameObjects.Rectangle;
  private readonly text: Phaser.GameObjects.Text;

  constructor(scene: MainScene) {
    super(scene, 0, 0);
    const { WIDTH, HEIGHT, COLOR, OUTLINE_COLOR } = PlayerStatsPanel;

    this.dragHandle = new Phaser.GameObjects.Rectangle(
      scene,
      0,
      0,
      WIDTH,
      HEIGHT,
      COLOR,
    ).setStrokeStyle(4, OUTLINE_COLOR);
    this.text = new Phaser.GameObjects.Text(scene, 0, 0, '', {
      fontSize: '24px',
      color: '#f3e3c3',
      lineSpacing: 8,
      align: 'left',
    }).setOrigin(0.5);

    this.add([this.dragHandle, this.text]);
  }

  setStats(stats: PlayerStats): void {
    this.text.setText([
      `Energy:    ${stats.energy}`,
      `Health:    ${stats.health}/${stats.maxHealth}`,
      `Resources: ${stats.resources}`,
      `At: ${stats.location}`,
    ]);
  }
}
