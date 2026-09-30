import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import Phaser from 'phaser';

export class HarvestableCard extends Card {
  constructor(scene: MainScene, x: number, y: number, cardname: string, cardtext: string) {
    super(scene, x, y, cardname, cardtext);
    this.frameInit(scene);
  }

  frameInit(scene: MainScene) {
    this.shadow = new Phaser.GameObjects.Rectangle(scene, 14, -14, 315, 440, 0x000000, 0.5);
    const border = new Phaser.GameObjects.Rectangle(scene, 0, 0, 315, 440, 0x000000);
    const face = new Phaser.GameObjects.Rectangle(scene, 0, 0, 295, 420, 0x333333);
    const art = new Phaser.GameObjects.Rectangle(scene, 0, -90, 255, 170, 0x2ecc71);
    const title = new Phaser.GameObjects.Text(scene, -127.5, -200, this.name, {
      color: '#000000',
      fontSize: '18px',
    });
    const textBox = new Phaser.GameObjects.Rectangle(scene, 0, 105, 255, 150, 0xd8c9a3);

    const effectText = new Phaser.GameObjects.Text(scene, -122.5, 40, this.effect, {
      color: '#000000',
      fontSize: '14px',
      wordWrap: { width: 235 },
    });

    this.disabledLayer = new Phaser.GameObjects.Rectangle(scene, 0, 0, 315, 440, 0x111111);
    this.disabledLayer.setVisible(false);

    this.content = scene.add.container(0, 0, [
      border,
      face,
      art,
      title,
      textBox,
      effectText,
      this.disabledLayer,
    ]);
    this.add(this.shadow);
    this.add(this.content);

    scene.add.existing(this);
  }
}
