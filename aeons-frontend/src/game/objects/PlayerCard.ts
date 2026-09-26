import Phaser from 'phaser'
import type { MainScene } from '../scenes/MainScene'
import { Card } from '../objects/Card'

export class PlayerCard extends Card {

    public border!: Phaser.GameObjects.Rectangle

    constructor(scene: MainScene, x: number, y: number, cardname: string, cardtext: string) {
        super(scene, x, y, cardname, cardtext)

        this.frameInit(scene)
    }

    frameInit(scene: MainScene) {
        this.border = scene.add.rectangle(0, 0, 315, 440, 0x000000)
        const face = scene.add.rectangle(0, 0, 295, 420, 0x8b5e34)
        const art = scene.add.rectangle(0, -90, 255, 170, 0x2ecc71)
        const title = scene.add.text(-127.5, -195, this.name, { color: '#000000', fontSize: '18px' })
        const textBox = scene.add.rectangle(0, 105, 255, 150, 0xd8c9a3)
        const effectText = scene.add.text(-122.5, 40, this.effect, {
            color: '#000000',
            fontSize: '14px',
            wordWrap: { width: 235 }
        })
        
        this.add([this.border, face, art, title, textBox, effectText])

        scene.add.existing(this)
    }

    onPointerOver() {
        this.border.fillColor = 0xaaaaaa
    }

    onPointerOut() {
           this.border.fillColor = 0x000000
    }
}