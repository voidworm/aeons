import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { LocationDTO } from './GameState';

export class LocationCard extends Card {
  constructor(scene: MainScene, x: number, y: number, cardname: string, cardtext: string) {
    super(scene, x, y, {
      name: cardname,
      flavor: cardtext,
      faceColor: 0x105207,
      artHeight: 270,
      shadow: 'flat',
      italicText: true,
    });
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: LocationDTO) {
    return new LocationCard(scene, x, y, dto.name, dto.flavor);
  }
}
