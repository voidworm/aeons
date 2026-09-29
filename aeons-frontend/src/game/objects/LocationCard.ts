import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import Phaser from 'phaser';

export class LocationCard extends Card {
  constructor(scene: MainScene, x: number, y: number, cardname: string, cardtext: string) {
    super(scene, x, y, cardname, cardtext);
    this.frameInit(scene);
  }

  frameInit(scene: MainScene) {
    this.shadow = new Phaser.GameObjects.Rectangle(scene, 0, 0, 315, 440, 0x000000, 1);

    const border = new Phaser.GameObjects.Rectangle(scene, 0, 0, 315, 440, 0x000000);
    const face = new Phaser.GameObjects.Rectangle(scene, 0, 0, 295, 420, 0x105207);
    const art = new Phaser.GameObjects.Rectangle(scene, 0, -40, 255, 270, 0x2ecc71);
    const title = new Phaser.GameObjects.Text(scene, -127.5, -200, this.name, {
      color: '#000000',
      fontSize: '18px',
    });
    const textBox = new Phaser.GameObjects.Rectangle(scene, 0, 155, 255, 85, 0xd8c9a3);
    const effectText = new Phaser.GameObjects.Text(scene, -122.5, 140, this.effect, {
      color: '#000000',
      fontSize: '14px',
      wordWrap: { width: 235 },
    });

    this.content = scene.add.container(0, 0, [border, face, art, title, textBox, effectText]);
    this.add(this.shadow);
    this.add(this.content);

    scene.add.existing(this);
  }

  setFloating(input: boolean) {
    this.floating = input;
    this.tweensManager.add({
      targets: this.content,
      scale: input ? 1.05 : 1,
      y: input ? -10 : 0,
      duration: 150,
      ease: 'Sine.easeOut',
    });
    this.tweensManager.add({
      targets: this.shadow,
      scale: input ? 1.1 : 1,
      x: input ? 25 : 0,
      y: input ? -25 : 0,
      alpha: input ? 0.25 : 1,
      duration: 150,
      ease: 'Sine.easeOut',
    });
  }
}
