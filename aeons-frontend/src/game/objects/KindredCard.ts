import type { MainScene } from '../scenes/MainScene';
import type { PlayableCardDTO } from './GameState';
import { PlayableCard } from './PlayableCard';

//cards that act on the player's kindred rather than the player, set apart by their colour
export class KindredCard extends PlayableCard {
  private static readonly KINDRED_FACE_COLOR = 0x4a2368;

  constructor(
    scene: MainScene,
    x: number,
    y: number,
    id: number,
    cardname: string,
    effectText: string,
    flavor: string,
    cost: number,
  ) {
    super(scene, x, y, id, cardname, effectText, flavor, cost, KindredCard.KINDRED_FACE_COLOR);
  }

  static fromDto(scene: MainScene, x: integer, y: integer, dto: PlayableCardDTO) {
    return new KindredCard(
      scene,
      x,
      y,
      dto.id,
      dto.name,
      dto.effectText,
      dto.flavor,
      dto.cost,
    ).applyDto(dto);
  }
}
