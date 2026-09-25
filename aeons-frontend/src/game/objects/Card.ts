import Phaser from 'phaser'
import type { MainScene } from '../scenes/MainScene'

export class Card extends Phaser.GameObjects.Container {

    public isHighlighted: boolean = false
    public name: string = "Generic card"
    public effect: string = "Generic effect"

    constructor(scene: MainScene, x: number, y: number, cardname: string, cardtext: string) {
        super(scene, x, y)

        this.setSize(315,440)
        this.setInteractive({ draggable: true })
        
        this.name = cardname
        this.effect = cardtext

        const border = scene.add.rectangle(0, 0, 315, 440, 0x000000)
        const face = scene.add.rectangle(0, 0, 295, 420, 0x8b5e34)
        const art = scene.add.rectangle(0, -90, 255, 170, 0x2ecc71)
        const title = scene.add.text(-127.5, -195, cardname, { color: '#000000', fontSize: '18px' })
        const textBox = scene.add.rectangle(0, 105, 255, 150, 0xd8c9a3)
        const effectText = scene.add.text(-122.5, 40, cardtext, {
            color: '#000000',
            fontSize: '14px',
            wordWrap: { width: 235 }
        })

        this.add([border, face, art, title, textBox, effectText])
        scene.add.existing(this)

        this.on('dragstart', ()=> {
            this.setScale(1.05)
            scene.children.bringToTop(this)
        })

        this.on('drag', (pointer: Phaser.Input.Pointer, dragX: number, dragY: number) => {
            this.x = dragX
            this.y = dragY
            scene.handleCardDrag(this)
        })

        this.on('dragend', () => {
            this.setScale(1)
            scene.handleCardDrop(this)
        })
    }

    setHighlighted(input : boolean) {
        if (input) {
            this.alpha = 0.5
        }else {
            this.alpha = 1
        }
        this.isHighlighted = input
    }

  
}