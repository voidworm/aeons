import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import Phaser from 'phaser';
import type { PlayableCardDTO } from './GameState';

export class PlayableCard extends Card {
  protected static readonly FACE_COLOR = 0x8b5e34;

  public effectType = -1;

  constructor(
    scene: MainScene,
    x: number,
    y: number,
    id: number,
    cardname: string,
    flavor: string,
    cost: number,
    faceColor: number = PlayableCard.FACE_COLOR,
  ) {
    super(scene, x, y, {
      id: id,
      name: cardname,
      flavor: flavor,
      faceColor,
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

  //disabled hand cards can still be re-sorted, they just can't be played
  protected canBeDragged(): boolean {
    return !this.isSuppressed();
  }

  protected applyDto(dto: PlayableCardDTO): this {
    this.effectType = dto.effectType;
    this.setDisabled(!dto.canBeCast);
    return this;
  }

  static fromDto(scene: MainScene, x: integer, y: integer, dto: PlayableCardDTO) {
    return new PlayableCard(scene, x, y, dto.id, dto.name, dto.effectText, dto.cost).applyDto(dto);
  }
}
