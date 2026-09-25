import Phaser from 'phaser'

export class MainScene extends Phaser.Scene {
  private ellipse!: Phaser.GameObjects.Ellipse

  constructor() {
    super('MainScene')
  }

  preload() {
  }

  create() {
    this.add.text(20 ,20, 'The gateway has opened', { color: '#000000'})
    this.ellipse = this.add.ellipse(400,300,60,100,0x6c5ce7)
  }

  update(_timer: number, _delta: number) {
  }

}