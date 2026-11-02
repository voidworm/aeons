import type { MainScene } from '../scenes/MainScene';
import type { PlayableCardDTO } from './GameState';
import { PlayableCard } from './PlayableCard';

//the card that finishes a quest, in the blue of everything that belongs to quests
export class FinishQuestCard extends PlayableCard {
  static readonly QUEST_FACE_COLOR = 0x1d4f91;

  static fromDto(scene: MainScene, x: integer, y: integer, dto: PlayableCardDTO) {
    return new FinishQuestCard(
      scene,
      x,
      y,
      dto.id,
      dto.name,
      dto.effectText,
      dto.flavor,
      dto.cost,
      FinishQuestCard.QUEST_FACE_COLOR,
    ).applyDto(dto);
  }
}
