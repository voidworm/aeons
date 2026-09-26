import Phaser from 'phaser'
import { PlayerCard } from '../objects/PlayerCard'
import { LocationCard } from '../objects/LocationCard'

export class MainScene extends Phaser.Scene {
  private playerCards: PlayerCard[] = []
  private locationCards: LocationCard[] = []
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
    
    var boltCard = new PlayerCard(this, (this.scale.width / 2)-400, this.scale.height, "Lightning Bolt", "Deal 3 damage to any target.")
    boltCard.on('cardDragged', (card: PlayerCard) => this.handleCardDrag(card))
    boltCard.on('cardDragStart', (card: PlayerCard) => this.handleCardDragStart(card))
    boltCard.on('cardDragEnd', (card: PlayerCard) => this.handleCardDrop(card))
    this.playerCards.push(boltCard)
    
    var cleanseCard = new PlayerCard(this, (this.scale.width / 2), this.scale.height, "Cleanse", "Remove all debuffs from that target.")
    cleanseCard.on('cardDragged', (card: PlayerCard) => this.handleCardDrag(card))
    cleanseCard.on('cardDragStart', (card: PlayerCard) => this.handleCardDragStart(card))
    cleanseCard.on('cardDragEnd', (card: PlayerCard) => this.handleCardDrop(card))
    this.playerCards.push(cleanseCard)

    var trapCard = new PlayerCard(this, (this.scale.width / 2)+400, this.scale.height, "Trap", "Summon a trap at target location")
    trapCard.on('cardDragged', (card: PlayerCard) => this.handleCardDrag(card))
    trapCard.on('cardDragStart', (card: PlayerCard) => this.handleCardDragStart(card))
    trapCard.on('cardDragEnd', (card: PlayerCard) => this.handleCardDrop(card))
    this.playerCards.push(trapCard)


    var secludedDen = new LocationCard(this, (this.scale.width/2)+400,400, "Secluded Den", "The secluded den is home to your tribe.")
    this.locationCards.push(secludedDen)

    var eternalPlains = new LocationCard(this, (this.scale.width/2),400, "Eternal Plains", "Across the eternal plains....")
    this.locationCards.push(eternalPlains)

    var windscarredCrag = new LocationCard(this, (this.scale.width/2)-400,400, "Windscarred Crag", "Only the mind of a child could not see the danger this place is home to")
    this.locationCards.push(windscarredCrag)

  }

  update(_timer: number, _delta: number) {
  }

  handleCardDragStart(card: PlayerCard) {
    this.children.bringToTop(card)
  }

  handleCardDrag(card: PlayerCard) {
    for (const other of this.playerCards) {
        if (other == card) continue
        const overlapping = Phaser.Geom.Intersects.RectangleToRectangle(card.getBounds(), other.getBounds())
        other.setHighlighted(overlapping)
    }
  }

  handleCardDrop(card: PlayerCard) {
    const highlighted = this.playerCards.find(other => other.isHighlighted) 
    
    if (!highlighted) {
        
        return
    }

    this.lastActionText.text = `Cast ${card.name} on ${highlighted.name}`
        highlighted?.setHighlighted(false)
  }

}