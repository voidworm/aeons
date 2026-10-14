import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import { HandContainer } from './HandContainer';
import type { PlayerCard } from './PlayerCard';
import { MiniPlayerCard } from './MiniPlayerCard';
import { ControlButton } from './ControlButton';
import { CURSOR_INTERACTIVE } from '../cursors';
import { PlayerStatsPanel } from './PlayerStatsPanel';
import type { PlayerStats } from './PlayerStatsPanel';

export class PlayerSwitcher extends Phaser.GameObjects.Container {
  private static readonly PREVIEW_SCALE = 0.45;
  private static readonly PREVIEW_WIDTH = Card.WIDTH * PlayerSwitcher.PREVIEW_SCALE;
  private static readonly PREVIEW_HEIGHT = Card.HEIGHT * PlayerSwitcher.PREVIEW_SCALE;
  private static readonly PREVIEW_OFFSET_X = 16;
  static readonly STACK_WIDTH = PlayerSwitcher.PREVIEW_WIDTH + 2 * PlayerSwitcher.PREVIEW_OFFSET_X;
  private static readonly HAND_GAP = 250; // between the hand's outer cards and the panels / buttons
  private static readonly STATS_GAP = 16; // between the stack and the stats panel
  private static readonly NAME_GAP = 8; // between the stats panel and the name button above it
  private static readonly BUTTON_GAP = 20; // between the stacked buttons
  private static readonly BOTTOM_INSET = 20;
  private static readonly BUTTONS_LIFT = 40; // raises the phase buttons above the bottom edge
  private static readonly BACK_PREVIEW_OFFSET_Y = 24; // the back preview sits this much higher than the front one

  // the origin is the bottom right corner of the stats panel, the stack sits left of it, the name button above it
  // the phase buttons sit right of the hand, their positions are relative to this origin and set on resize
  // within the stack the active preview sits front-left and lower, the inactive one behind it, up and to the right
  private static readonly FRONT_SLOT = {
    x: -PlayerSwitcher.PREVIEW_OFFSET_X,
    y: -PlayerSwitcher.PREVIEW_HEIGHT / 2,
  };
  private static readonly BACK_SLOT = {
    x: PlayerSwitcher.PREVIEW_OFFSET_X,
    y: -PlayerSwitcher.PREVIEW_HEIGHT / 2 - PlayerSwitcher.BACK_PREVIEW_OFFSET_Y,
  };

  private previews = new Map<number, MiniPlayerCard>();
  private tweens: Phaser.Tweens.Tween[] = [];
  private nameButton: ControlButton; // shows the active player, click switches
  private labelTween: Phaser.Tweens.Tween | null = null;
  private statsPanel: PlayerStatsPanel;
  private stack: Phaser.GameObjects.Container; // the player previews, left of the stats panel
  private creaturePhaseButton: ControlButton;
  private upkeepButton: ControlButton;

  constructor(scene: MainScene, onSwitch: () => void, onCreaturePhase: () => void, onUpkeepPhase: () => void) {
    super(scene, 0, 0);
    const { PREVIEW_HEIGHT, BACK_PREVIEW_OFFSET_Y: OFFSET_Y, STATS_GAP, NAME_GAP, STACK_WIDTH } = PlayerSwitcher;
    const { HEIGHT } = ControlButton;

    const clickZone = new Phaser.GameObjects.Zone(
      scene,
      0,
      -PREVIEW_HEIGHT / 2 - OFFSET_Y / 2,
      STACK_WIDTH,
      PREVIEW_HEIGHT + OFFSET_Y,
    ).setInteractive({ cursor: CURSOR_INTERACTIVE });

    clickZone.on('pointerover', () => this.startWiggle());
    clickZone.on('pointerout', () => this.stopWiggle());
    clickZone.on('pointerup', onSwitch);

    this.stack = new Phaser.GameObjects.Container(scene, -PlayerStatsPanel.WIDTH - STATS_GAP - STACK_WIDTH / 2, 0);
    this.stack.add(clickZone);

    this.statsPanel = new PlayerStatsPanel(scene);
    this.statsPanel.setPosition(-PlayerStatsPanel.WIDTH / 2, -PlayerStatsPanel.HEIGHT / 2);

    this.nameButton = new ControlButton(scene, '', onSwitch);
    this.nameButton.setPosition(-PlayerStatsPanel.WIDTH / 2, -PlayerStatsPanel.HEIGHT - NAME_GAP - HEIGHT / 2);

    this.creaturePhaseButton = new ControlButton(scene, 'Creature phase', onCreaturePhase);
    this.upkeepButton = new ControlButton(scene, 'Upkeep phase', onUpkeepPhase);

    this.add([this.stack, this.statsPanel, this.nameButton, this.creaturePhaseButton, this.upkeepButton]);
    this.setDepth(1000); //if i dont do this the stupid things keeps moving in front of the cards
    this.resize(scene.scale.width, scene.scale.height);
    scene.add.existing(this);
  }

  resize(width: number, height: number): void {
    const { HAND_GAP, BOTTOM_INSET, BUTTON_GAP, BUTTONS_LIFT } = PlayerSwitcher;
    const handHalfWidth = HandContainer.expectedHalfWidth();
    const originX = width / 2 - handHalfWidth - HAND_GAP;
    const originY = height - BOTTOM_INSET;
    this.setPosition(originX, originY);

    // the phase buttons sit right of the hand, stacked upwards from the bottom
    const buttonX = width / 2 + handHalfWidth + HAND_GAP + ControlButton.WIDTH / 2 - originX;
    this.upkeepButton.setPosition(buttonX, -BUTTONS_LIFT - ControlButton.HEIGHT / 2);
    this.creaturePhaseButton.setPosition(buttonX, -BUTTONS_LIFT - ControlButton.HEIGHT * 1.5 - BUTTON_GAP);
  }

  setPlayers(players: PlayerCard[]): void {
    this.tweens.forEach((tween) => tween.stop());
    this.tweens = [];
    this.previews.forEach((preview) => preview.destroy());
    this.previews.clear();

    for (const player of players) {
      const preview = new MiniPlayerCard(this.scene as MainScene, player.id, player.name);
      preview.setScale(PlayerSwitcher.PREVIEW_SCALE);
      this.stack.add(preview);
      this.previews.set(player.id, preview);
      this.place(preview, player.activePlayer, false);
    }
  }

  setStats(stats: PlayerStats): void {
    this.statsPanel.setStats(stats);
  }

  showActive(id: number): void {
    this.previews.forEach((preview, previewId) => this.place(preview, previewId === id, true));
  }

  private setActiveName(name: string, animate: boolean): void {
    if (this.nameButton.label.text === name) return;
    this.labelTween?.stop();

    if (!animate) {
      this.nameButton.label.setText(name).setAlpha(1);
      return;
    }
    this.labelTween = this.scene.tweens.add({
      targets: this.nameButton.label,
      alpha: 0,
      duration: 125,
      onComplete: () => {
        this.nameButton.label.setText(name);
        this.labelTween = this.scene.tweens.add({
          targets: this.nameButton.label,
          alpha: 1,
          duration: 125,
        });
      },
    });
  }

  private place(preview: MiniPlayerCard, active: boolean, animate: boolean): void {
    const slot = active ? PlayerSwitcher.FRONT_SLOT : PlayerSwitcher.BACK_SLOT;
    preview.setDisabled(!active);
    if (active) {
      this.stack.bringToTop(preview);
      this.setActiveName(preview.name, animate);
    }

    if (!animate) {
      preview.setPosition(slot.x, slot.y);
      return;
    }
    this.scene.tweens.add({
      targets: preview,
      x: slot.x,
      y: slot.y,
      duration: 250,
      ease: 'Sine.easeInOut',
    });
  }

  private startWiggle(): void {
    this.tweens.forEach((tween) => tween.stop());
    this.tweens = [...this.previews.values()].map((preview) =>
      this.scene.tweens.add({
        targets: preview,
        angle: { from: -0.5, to: 0.5 },
        duration: 600,
        yoyo: true,
        repeat: -1,
        ease: 'Sine.easeInOut',
      }),
    );
  }

  private stopWiggle(): void {
    this.tweens.forEach((tween) => tween.stop());
    this.tweens = [];
    this.previews.forEach((preview) => {
      this.scene.tweens.add({ targets: preview, angle: 0, duration: 600, ease: 'Sine.easeOut' });
    });
  }
}
