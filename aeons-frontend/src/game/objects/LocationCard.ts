import Phaser from 'phaser'
import type { MainScene } from '../scenes/MainScene'
import { Card } from './Card'

export class LocationCard extends Card {

    public isHighlighted: boolean = false
    public border!: Phaser.GameObjects.Rectangle

    constructor(scene: MainScene, x: number, y: number, cardname: string, cardtext: string) {
        
        super(scene, x, y, cardname, cardtext)
        this.frameInit(scene)
    }

    frameInit(scene: MainScene) {
        this.border = scene.add.rectangle(0, 0, 315, 440, 0x000000)
        const face = scene.add.rectangle(0, 0, 295, 420, 0x105207)
        const art = scene.add.rectangle(0, -40, 255, 270, 0x2ecc71)
        const title = scene.add.text(-127.5, -195, this.name, { color: '#000000', fontSize: '18px' })
        const textBox = scene.add.rectangle(0, 155, 255, 100, 0xd8c9a3)
        const effectText = scene.add.text(-122.5, 140, this.effect, {
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