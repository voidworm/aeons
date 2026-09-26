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

        this.on('dragstart', ()=> {this.onDragStart()})
        this.on('drag', (pointer: Phaser.Input.Pointer, dragX: number, dragY: number) => {this.onDrag(pointer, dragX, dragY)})
        this.on('dragend', () => this.onDragend())

        this.on('pointerover', () => {this.onPointerOver()})
        this.on('pointerout', () => {this.onPointerOut()})
    }

    setHighlighted(input : boolean) {
        if (input) {
            this.alpha = 0.5
        }else {
            this.alpha = 1
        }
        this.isHighlighted = input
    }

    onDragStart() {
        this.setScale(1.05)
        this.emit('cardDragStart', this)
    }

    onDrag(pointer: Phaser.Input.Pointer, dragX: number, dragY: number) {
        this.x = dragX
        this.y = dragY
        this.emit('cardDragged', this)
    }

    onDragend() {
        this.setScale(1)
        this.emit('cardDragEnd', this)
    }

    onPointerOver() {
    }

    onPointerOut() {
    }

}