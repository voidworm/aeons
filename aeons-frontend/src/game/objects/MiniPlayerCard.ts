import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';

export class MiniPlayerCard extends Card {
  constructor(scene: MainScene, id: number, name: string) {
    super(scene, 0, 0, {
      id: id,
      name: name,
      flavor: '',
      faceColor: 0x8b5e34,
      artHeight: 320,
      shadow: 'flat',
      showTextBox: false,
      titleBelowArt: true,
    });

    this.hitbox.removeInteractive();
  }
}
