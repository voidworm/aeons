import Phaser from 'phaser'
import { Card } from '../objects/Card'

export class MainScene extends Phaser.Scene {
  private cards: Card[] = []
  private lastActionText!: Phaser.GameObjects.Text

  constructor() {
    super('MainScene')
  }

  preload() {
  }

  create() {
    this.add.text(this.scale.width / 2 ,50, 'Along the mist-shrouded plains,', { color: '#000000'})
    this.add.text(this.scale.width / 2 ,70, 'Across the eternal plains', { color: '#000000'})
    this.add.text(this.scale.width / 2 ,90, 'And ever towards the ancient forest', { color: '#000000'})
    this.lastActionText = this.add.text(this.scale.width / 2 ,130, "...at long last, homebound.", { color: '#000000'})
    
    this.cards.push(new Card(this, (this.scale.width / 2)-400, 450, "Lightning Bolt", "Deal 3 damage to any target."))
    this.cards.push(new Card(this, (this.scale.width / 2), 450, "Cleanse", "Remove all debuffs from that target."))
    this.cards.push(new Card(this, (this.scale.width / 2)+400, 450, "Trap", "Summon a trap at target location"))
  }

  update(_timer: number, _delta: number) {
  }

  handleCardDrag(card: Card) {
    for (const other of this.cards) {
        if (other == card) continue
        const overlapping = Phaser.Geom.Intersects.RectangleToRectangle(card.getBounds(), other.getBounds())
        other.setHighlighted(overlapping)
    }
  }

  handleCardDrop(card: Card) {
    const highlighted = this.cards.find(other => other.isHighlighted) 
    this.lastActionText.text = `Cast ${card.name} on ${highlighted.name}`
    highlighted?.setHighlighted(false)
  }

}