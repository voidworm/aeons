import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import { HandContainer } from './HandContainer';
import type { PlayerCard } from './PlayerCard';
import { MiniPlayerCard } from './MiniPlayerCard';
import { ControlButton } from './ControlButton';

export class PlayerSwitcher extends Phaser.GameObjects.Container {
  private static readonly PREVIEW_SCALE = 0.45;
  private static readonly PREVIEW_WIDTH = Card.WIDTH * PlayerSwitcher.PREVIEW_SCALE;
  private static readonly PREVIEW_HEIGHT = Card.HEIGHT * PlayerSwitcher.PREVIEW_SCALE;
  private static readonly PREVIEW_OFFSET_X = 16;
  static readonly STACK_WIDTH = PlayerSwitcher.PREVIEW_WIDTH + 2 * PlayerSwitcher.PREVIEW_OFFSET_X;
  private static readonly EDGE_INSET = 240; // shifts the whole bundle left, away from the screen edge
  private static readonly LABEL_GAP = 6; // between the stack and the name button below it
  private static readonly BUTTON_GAP = 8; // between the stacked buttons
  // room reserved under the stack for the two buttons
  private static readonly LABEL_AREA =
    PlayerSwitcher.LABEL_GAP + 2 * ControlButton.HEIGHT + PlayerSwitcher.BUTTON_GAP;
  private static readonly BACK_PREVIEW_OFFSET_Y = 24; // the back preview sits this much higher than the front one

  // the origin is the bottom centre of the stack, the name button and the creature phase button hang below it
  // the active preview sits front-left and lower, the inactive one behind it, up and to the right
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

  constructor(scene: MainScene, onSwitch: () => void, onCreaturePhase: () => void) {
    super(scene, 0, 0);
    const { PREVIEW_HEIGHT, BACK_PREVIEW_OFFSET_Y: OFFSET_Y } = PlayerSwitcher;

    const clickZone = new Phaser.GameObjects.Zone(
      scene,
      0,
      -PREVIEW_HEIGHT / 2 - OFFSET_Y / 2,
      PlayerSwitcher.STACK_WIDTH,
      PREVIEW_HEIGHT + OFFSET_Y,
    ).setInteractive({ useHandCursor: true });

    clickZone.on('pointerover', () => this.startWiggle());
    clickZone.on('pointerout', () => this.stopWiggle());
    clickZone.on('pointerup', onSwitch);

    const { LABEL_GAP, BUTTON_GAP } = PlayerSwitcher;
    const { HEIGHT } = ControlButton;
    this.nameButton = new ControlButton(scene, '', onSwitch);
    this.nameButton.setY(LABEL_GAP + HEIGHT / 2);
    const phaseButton = new ControlButton(scene, 'Creature phase', onCreaturePhase);
    phaseButton.setY(LABEL_GAP + HEIGHT + BUTTON_GAP + HEIGHT / 2);

    this.add([clickZone, this.nameButton, phaseButton]);
    this.setDepth(1000); //if i dont do this the stupid things keeps moving in front of the cards
    this.resize(scene.scale.width, scene.scale.height);
    scene.add.existing(this);
  }

  resize(width: number, height: number): void {
    this.setPosition(
      width - HandContainer.PADDING - PlayerSwitcher.EDGE_INSET - PlayerSwitcher.STACK_WIDTH / 2,
      height - 20 - PlayerSwitcher.LABEL_AREA,
    );
  }

  setPlayers(players: PlayerCard[]): void {
    this.tweens.forEach((tween) => tween.stop());
    this.tweens = [];
    this.previews.forEach((preview) => preview.destroy());
    this.previews.clear();

    for (const player of players) {
      const preview = new MiniPlayerCard(this.scene as MainScene, player.id, player.name);
      preview.setScale(PlayerSwitcher.PREVIEW_SCALE);
      this.add(preview);
      this.previews.set(player.id, preview);
      this.place(preview, player.activePlayer, false);
    }
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
      this.bringToTop(preview);
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
