import Phaser from 'phaser'
import type { MainScene } from '../scenes/MainScene'
import { Card } from '../objects/Card'

export class PlayerCard extends Card {

   

    constructor(scene: MainScene, x: number, y: number, cardname: string, cardtext: string, cost: integer) {
        super(scene, x, y, cardname, cardtext)
        this.frameInit(scene, cost)
    }

    frameInit(scene: MainScene, cost: integer) {

        var string = 
        this.border = scene.add.rectangle(0, 0, 315, 440, 0x000000)
        const face = scene.add.rectangle(0, 0, 295, 420, 0x8b5e34)
        const art = scene.add.rectangle(0, -90, 255, 170, 0x2ecc71)
        const title = scene.add.text(-127.5, -200, this.name, { color: '#000000', fontSize: '18px' })
        const textBox = scene.add.rectangle(0, 105, 255, 150, 0xd8c9a3)

        const costCircle = scene.add.circle(127.5, -192, 13, 0xc0c0c0)
        const costText = scene.add.text(127.5, -192, String(cost), { color: '#000000', fontSize: '18px' }).setOrigin(0.5)

        const effectText = scene.add.text(-122.5, 40, this.effect, {
            color: '#000000',
            fontSize: '14px',
            wordWrap: { width: 235 }
        })

        this.add([this.border, face, art, title, textBox, effectText, costCircle, costText])

        scene.add.existing(this)
    }

  
}