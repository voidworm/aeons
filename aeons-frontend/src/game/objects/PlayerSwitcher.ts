import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { PlayerCard } from './PlayerCard';
import { MiniPlayerCard } from './MiniPlayerCard';
import { ControlButton } from './ControlButton';
import { CURSOR_INTERACTIVE } from '../cursors';
import { DragGroup } from './DragGroup';
import { uiScale } from '../layout/uiScale';
import { EventLog } from './EventLog';
import { PlayerStatsPanel } from './PlayerStatsPanel';
import type { PlayerStats } from './PlayerStatsPanel';
import type { Phase } from './GameState';

const PHASE_TITLES: Record<Phase, string> = {
  spirit: 'Spirit Phase',
  world: 'World Phase',
  awakening: 'Awakening Phase',
};
// what the phase button does next, in each phase
const PHASE_BUTTON_LABELS: Record<Phase, string> = {
  spirit: 'End Spirit Phase',
  world: 'Start Awakening Phase',
  awakening: 'Start Spirit Phase',
};

export class PlayerSwitcher extends Phaser.GameObjects.Container {
  private static readonly PREVIEW_SCALE = 0.45;
  private static readonly PREVIEW_WIDTH = Card.WIDTH * PlayerSwitcher.PREVIEW_SCALE;
  private static readonly PREVIEW_HEIGHT = Card.HEIGHT * PlayerSwitcher.PREVIEW_SCALE;
  private static readonly PREVIEW_OFFSET_X = 16;
  static readonly STACK_WIDTH = PlayerSwitcher.PREVIEW_WIDTH + 2 * PlayerSwitcher.PREVIEW_OFFSET_X;
  private static readonly STATS_GAP = 16; // between the stack and the stats panel
  private static readonly CLUSTER_WIDTH = PlayerStatsPanel.WIDTH + PlayerSwitcher.STATS_GAP + PlayerSwitcher.STACK_WIDTH; // previews and stats panel together
  private static readonly NAME_GAP = 8; // between the stats panel and the name button above it
  private static readonly LOG_GAP = 16; // between the event log and the phase button above it
  private static readonly PHASE_TITLE_GAP = 8; // between the phase button and the cycle counter above it
  private static readonly CYCLE_HEIGHT = 26; // the room the cycle counter takes between the title and the button
  private static readonly PHASE_BUTTON_WIDTH = 360;
  private static readonly BOTTOM_INSET = 20; // also the gap to the side edges

  private static readonly PANEL_LIFT = 15; // raises the whole switcher off the bottom edge
  private static readonly BUTTONS_LIFT = 40; // raises the event log, and the phase button above it, off the bottom edge
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
  private phaseButton: ControlButton;
  private phaseTitle: Phaser.GameObjects.Text;
  private cycleText: Phaser.GameObjects.Text;
  private phaseControl: Phaser.GameObjects.Container; // the phase title above the phase button, centred on the button

  constructor(scene: MainScene, onSwitch: () => void, onPhaseButton: () => void) {
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

    this.phaseButton = new ControlButton(scene, '', onPhaseButton, PlayerSwitcher.PHASE_BUTTON_WIDTH);
    const { PHASE_TITLE_GAP, CYCLE_HEIGHT } = PlayerSwitcher;
    this.cycleText = new Phaser.GameObjects.Text(scene, 0, -HEIGHT / 2 - PHASE_TITLE_GAP, '', {
      fontSize: '20px',
      color: '#f3e3c3',
    }).setOrigin(0.5, 1);
    this.phaseTitle = new Phaser.GameObjects.Text(scene, 0, -HEIGHT / 2 - PHASE_TITLE_GAP - CYCLE_HEIGHT, '', {
      fontSize: '26px',
      fontStyle: 'bold',
      color: '#f3e3c3',
    }).setOrigin(0.5, 1);
    this.phaseControl = new Phaser.GameObjects.Container(scene, 0, 0, [this.phaseButton, this.cycleText, this.phaseTitle]);

    this.add([this.stack, this.statsPanel, this.nameButton, this.phaseControl]);
    // the previews, name and stats are one unit that follows a drag on the stats panel
    new DragGroup(scene, [this.stack, this.statsPanel, this.nameButton]).enableOn(this.statsPanel.dragHandle);
    this.setDepth(1000); //if i dont do this the stupid things keeps moving in front of the cards
    this.resize(scene.scale.width, scene.scale.height);
    scene.add.existing(this);
  }

  resize(width: number, height: number): void {
    const { BOTTOM_INSET, BUTTONS_LIFT, CLUSTER_WIDTH } = PlayerSwitcher;
    const scale = uiScale(width, height);
    const inset = BOTTOM_INSET * scale;
    const originX = inset + CLUSTER_WIDTH * scale; // previews, name and stats sit in the bottom left corner
    const originY = height - (BOTTOM_INSET + PlayerSwitcher.PANEL_LIFT) * scale;
    this.setScale(scale);
    this.setPosition(originX, originY);

    // the event log sits in the bottom right corner, the phase button centred above it
    const logCentreX = (width - inset - (EventLog.WIDTH * scale) / 2 - originX) / scale;
    const logTop = -BUTTONS_LIFT - EventLog.HEIGHT;
    this.phaseControl.setPosition(logCentreX, logTop - PlayerSwitcher.LOG_GAP - ControlButton.HEIGHT / 2);
  }

  get phaseObjects(): Phaser.GameObjects.Container[] {
    return [this.phaseControl];
  }

  setPhase(phase: Phase, cycle: number): void {
    this.phaseTitle.setText(PHASE_TITLES[phase]);
    this.cycleText.setText(`Cycle ${cycle}`);
    this.phaseButton.label.setText(PHASE_BUTTON_LABELS[phase]);
  }

  setPhaseButtonEnabled(enabled: boolean): void {
    this.phaseButton.setEnabled(enabled);
  }

  //screen position of the bottom centre of the event log, below the phase button
  eventLogBottomCentre(): { x: number; y: number } {
    return {
      x: this.x + this.phaseControl.x * this.scaleX,
      y: this.y - PlayerSwitcher.BUTTONS_LIFT * this.scaleY,
    };
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
