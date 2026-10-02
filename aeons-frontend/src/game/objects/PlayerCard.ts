import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { PlayerDTO } from './GameState';

export class PlayerCard extends Card {
  public cardsInHand: integer[];
  public activePlayer: boolean = false;

  constructor(
    scene: MainScene,
    x: number,
    y: number,
    id: number,
    cardname: string,
    cards: integer[],
  ) {
    super(scene, x, y, {
      id: id,
      name: cardname,
      flavor: '',
      faceColor: 0x8b5e34,
      artHeight: 256,
      shadow: 'hovers',
      showTextBox: false,
      titleBelowArt: true,
    });

    this.cardsInHand = cards;
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: PlayerDTO) {
    dto.hand;
    return new PlayerCard(scene, x, y, dto.id, dto.name, dto.hand);
  }

  setActivePlayer(active: boolean) {
    this.activePlayer = active;
  }
}
