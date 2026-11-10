import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { QuestDTO } from './GameState';

//a horizontal card like a location: the art on the left, the quest's goals on the right with its flavour at the bottom
export class QuestCard extends Card {
  static readonly QUEST_FACE_COLOR = 0x1d4f91;
  static readonly QUEST_WIDTH = 352;
  static readonly QUEST_HEIGHT = 252;
  private static readonly ART_WIDTH = 96;
  private static readonly PANEL_GAP = 12; // between the art and the text panel
  private static readonly PANEL_PAD = 8;
  private static readonly TEXT_GAP = 8;
  private static readonly MET_COLOR = '#1b5e20';
  static readonly HOVER_ZOOM_DELAY = Card.HOVER_ZOOM_DELAY / 2;
  private static readonly INTRO_ZOOM_FACTOR = 2;
  private static readonly INTRO_HOLD = 5000;
  private static readonly INTRO_DELAY = 300; // lets the previous quest leave first
  private static readonly SHRINK_DURATION = 300;
  private static readonly SCREEN_MARGIN = 12;

  public dto: QuestDTO;
  private readonly textWidth: number;
  private goalTexts: Phaser.GameObjects.Text[] = [];
  private titleText: Phaser.GameObjects.Text;
  private flavorText: Phaser.GameObjects.Text;
  private finishableOutline: Phaser.GameObjects.Graphics;
  private slot = { x: 0, y: 0, scale: 1 };
  private introPlaying = false;

  constructor(scene: MainScene, x: number, y: number, dto: QuestDTO) {
    const { QUEST_WIDTH: WIDTH, QUEST_HEIGHT: HEIGHT } = QuestCard;
    super(scene, x, y, {
      id: dto.id,
      name: dto.name,
      flavor: dto.flavor,
      faceColor: QuestCard.QUEST_FACE_COLOR,
      width: WIDTH,
      height: HEIGHT,
      artHeight: 0,
      shadow: 'flat',
      showTextBox: false,
      showTitle: false,
    });
    this.dto = dto;

    const panelX = Card.INSET + QuestCard.ART_WIDTH + QuestCard.PANEL_GAP;
    const panelHeight = HEIGHT - 2 * Card.INSET;
    const panelWidth = WIDTH - Card.INSET - panelX;
    this.textWidth = panelWidth - 2 * QuestCard.PANEL_PAD;
    const textX = panelX + QuestCard.PANEL_PAD;
    const textStyle = { fontSize: '11.2px', wordWrap: { width: this.textWidth } };

    this.titleText = this.label(textX, 0, dto.name, {
      ...textStyle,
      fontSize: '14.4px',
      fontStyle: 'bold',
    });
    this.flavorText = this.label(textX, Card.INSET + panelHeight - QuestCard.PANEL_PAD, dto.flavor, {
      ...textStyle,
      fontStyle: 'italic',
    }).setOrigin(0, 1);
    this.finishableOutline = new Phaser.GameObjects.Graphics(scene)
      .lineStyle(3, 0xf3e3c3, 1)
      .strokeRoundedRect(
        Card.BORDER,
        Card.BORDER,
        WIDTH - 2 * Card.BORDER,
        HEIGHT - 2 * Card.BORDER,
        Card.CORNER_RADIUS - Card.BORDER,
      );

    this.addFrameDetails(
      this.roundedRect(
        Card.INSET,
        Card.INSET,
        QuestCard.ART_WIDTH,
        HEIGHT - 2 * Card.INSET,
        Card.CORNER_RADIUS - Card.BORDER,
        0x2ecc71,
      ),
      this.rect(panelX, Card.INSET, panelWidth, panelHeight, 0xd8c9a3),
      this.titleText,
      this.flavorText,
      this.finishableOutline,
    );
    this.sync(dto);
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: QuestDTO) {
    return new QuestCard(scene, x, y, dto);
  }

  sync(dto: QuestDTO): void {
    this.dto = dto;
    this.finishableOutline.setVisible(dto.finishable);
    this.goalTexts.forEach((t) => t.destroy());
    this.goalTexts = dto.goals.map((goal) =>
      this.label(this.titleText.x, 0, `${goal.met ? '✓' : '•'} ${goal.text}`, {
        fontSize: '11.2px',
        wordWrap: { width: this.textWidth },
        color: goal.met ? QuestCard.MET_COLOR : '#000000',
      }),
    );
    this.addFrameDetails(...this.goalTexts);
    this.layoutPanelText();
  }

  private layoutPanelText(): void {
    let y = Card.INSET + QuestCard.PANEL_PAD;
    for (const text of [this.titleText, ...this.goalTexts]) {
      text.setY(y);
      y += text.height + QuestCard.TEXT_GAP;
    }
  }

  placeInSlot(x: number, y: number, scale: number): void {
    this.slot = { x, y, scale };
    if (this.introPlaying) return;
    this.setScale(scale);
    this.setPosition(x, y);
  }

  //the new quest grows out of its slot, holds still and settles back
  playIntro(): void {
    this.introPlaying = true;
    this.setScale(0);
    const { x, y } = this.zoomOffset(0);
    this.setPosition(x, y);
    this.scene.time.delayedCall(QuestCard.INTRO_DELAY, () => {
      if (!this.active) return;
      this.setScale(this.slot.scale); // the scale the intro settles back to
      this.startZoom(QuestCard.INTRO_ZOOM_FACTOR, 0);
      this.scene.time.delayedCall(QuestCard.INTRO_HOLD, () => {
        if (!this.active) return;
        this.introPlaying = false;
        super.cancelHoverZoom();
      });
    });
  }

  //a quest that is over shrinks away
  shrinkAway(): void {
    this.hitbox.disableInteractive();
    this.tweensManager.killTweensOf(this);
    this.tweensManager.add({
      targets: this,
      scale: 0,
      duration: QuestCard.SHRINK_DURATION,
      ease: 'Sine.easeIn',
      onComplete: () => this.destroy(),
    });
  }

  protected override canHoverZoom(): boolean {
    return !this.introPlaying;
  }

  protected override hoverZoomDelay(): number {
    return QuestCard.HOVER_ZOOM_DELAY;
  }

  //pointing away must not cut the intro short
  protected override cancelHoverZoom(): void {
    if (!this.introPlaying) super.cancelHoverZoom();
  }

  //a zoomed card keeps the slot's top right corner and is pushed up if it would leave the screen at the bottom
  protected override zoomOffset(scale: number): { x: number; y: number } {
    const { x, y, scale: slotScale } = this.slot;
    const { QUEST_WIDTH: WIDTH, QUEST_HEIGHT: HEIGHT } = QuestCard;
    const bottomLimit = this.scene.scale.height - QuestCard.SCREEN_MARGIN - (HEIGHT * scale) / 2;
    return {
      x: x + (WIDTH * (slotScale - scale)) / 2,
      y: Math.min(y + (HEIGHT * (scale - slotScale)) / 2, bottomLimit),
    };
  }

  //quests are only displayed, they are not moved around
  protected override canBeDragged(): boolean {
    return false;
  }
}
