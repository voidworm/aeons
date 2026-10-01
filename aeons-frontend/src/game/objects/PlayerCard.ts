import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { PlayerDTO } from './GameState';

export class PlayerCard extends Card {
  constructor(scene: MainScene, x: number, y: number, cardname: string) {
    super(scene, x, y, {
      name: cardname,
      flavor: '',
      faceColor: 0x8b5e34,
      artHeight: 320,
      shadow: 'hovers',
      showTextBox: false,
      titleBelowArt: true,
    });
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: PlayerDTO) {
    return new PlayerCard(scene, x, y, dto.name);
  }
}
