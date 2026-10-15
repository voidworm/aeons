import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';

export interface PlayerStats {
  energy: number;
  wounds: number;
  maxWounds: number;
  food: number;
  resources: number;
  location: string;
}

//small readout of the active player, origin is the panel's centre
export class PlayerStatsPanel extends Phaser.GameObjects.Container {
  static readonly WIDTH = 330;
  static readonly HEIGHT = 215;
  private static readonly COLOR = 0x5c3b1f;
  private static readonly OUTLINE_COLOR = 0x2a1809;
  private static readonly PADDING = 20;
  private static readonly ROW_HEIGHT = 37;
  private static readonly DOT_RADIUS = 11;
  private static readonly DOT_GAP = 36;
  private static readonly DOT_EMPTY = 0x8a8a8a;
  private static readonly DOT_WOUNDED = 0xc0392b;

  readonly dragHandle: Phaser.GameObjects.Rectangle;
  private readonly energy: Phaser.GameObjects.Text;
  private readonly food: Phaser.GameObjects.Text;
  private readonly resources: Phaser.GameObjects.Text;
  private readonly location: Phaser.GameObjects.Text;
  private readonly woundDots: Phaser.GameObjects.Arc[] = [];

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
    this.energy = this.rowText(scene, 0);
    const woundsLabel = this.rowText(scene, 1, 'Wounds:');
    this.food = this.rowText(scene, 2);
    this.resources = this.rowText(scene, 3);
    this.location = this.rowText(scene, 4);

    this.add([this.dragHandle, this.energy, woundsLabel, this.food, this.resources, this.location]);
    this.initWoundDots(scene, woundsLabel);
  }

  setStats(stats: PlayerStats): void {
    this.energy.setText(`Energy: ${stats.energy}`);
    this.food.setText(`Food: ${stats.food}`);
    this.resources.setText(`Resources: ${stats.resources}`);
    this.location.setText(`At: ${stats.location}`);
    this.woundDots.forEach((dot, i) =>
      dot.setFillStyle(i < stats.wounds ? PlayerStatsPanel.DOT_WOUNDED : PlayerStatsPanel.DOT_EMPTY),
    );
  }

  private rowY(row: number): number {
    const { HEIGHT, ROW_HEIGHT } = PlayerStatsPanel;
    return -HEIGHT / 2 + ROW_HEIGHT / 2 + 15 + row * ROW_HEIGHT;
  }

  private rowText(scene: MainScene, row: number, content = ''): Phaser.GameObjects.Text {
    return new Phaser.GameObjects.Text(scene, -PlayerStatsPanel.WIDTH / 2 + PlayerStatsPanel.PADDING, this.rowY(row), content, {
      fontSize: '24px',
      color: '#f3e3c3',
    }).setOrigin(0, 0.5);
  }

  private initWoundDots(scene: MainScene, label: Phaser.GameObjects.Text) {
    const { DOT_RADIUS, DOT_GAP, DOT_EMPTY, OUTLINE_COLOR } = PlayerStatsPanel;
    const startX = label.x + label.width + 24;
    for (let i = 0; i < 3; i++) {
      const dot = new Phaser.GameObjects.Arc(scene, startX + i * DOT_GAP, this.rowY(1), DOT_RADIUS, 0, 360, false, DOT_EMPTY)
        .setStrokeStyle(2, OUTLINE_COLOR);
      this.woundDots.push(dot);
      this.add(dot);
    }
  }
}
