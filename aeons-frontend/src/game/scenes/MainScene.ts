import Phaser from 'phaser'
import { PlayerCard } from '../objects/PlayerCard'
import { LocationCard } from '../objects/LocationCard'
import type { Card } from '../objects/Card'

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
    
    var boltCard = new PlayerCard(this, (this.scale.width / 2)-400, this.scale.height, "Fishing", "Attempt to fish at target location.", 0)
    boltCard.on('cardDragged', (card: PlayerCard) => this.handleCardDrag(card))
    boltCard.on('cardDragStart', (card: PlayerCard) => this.handleCardDragStart(card))
    boltCard.on('cardDragEnd', (card: PlayerCard) => this.handleCardDrop(card))
    this.playerCards.push(boltCard)
    
    var cleanseCard = new PlayerCard(this, (this.scale.width / 2), this.scale.height, "Rite of Seeking", "Consult the spirits.\n\nIf you succeed, discover tracks at the location.", 4)
    cleanseCard.on('cardDragged', (card: PlayerCard) => this.handleCardDrag(card))
    cleanseCard.on('cardDragStart', (card: PlayerCard) => this.handleCardDragStart(card))
    cleanseCard.on('cardDragEnd', (card: PlayerCard) => this.handleCardDrop(card))
    this.playerCards.push(cleanseCard)

    var trapCard = new PlayerCard(this, (this.scale.width / 2)+400, this.scale.height, "Trap", "Place a trap at target location", 2)
    trapCard.on('cardDragged', (card: PlayerCard) => this.handleCardDrag(card))
    trapCard.on('cardDragStart', (card: PlayerCard) => this.handleCardDragStart(card))
    trapCard.on('cardDragEnd', (card: PlayerCard) => this.handleCardDrop(card))
    this.playerCards.push(trapCard)


    var secludedDen = new LocationCard(this, (this.scale.width/2)+400,400, "Secluded Den", "The secluded den is home to your tribe.")
    secludedDen.on('cardDragStart', (card: Card) => this.handleCardDragStart(card))
    this.locationCards.push(secludedDen)

    var eternalPlains = new LocationCard(this, (this.scale.width/2),400, "Eternal Plains", "Across the eternal plains....")
     eternalPlains.on('cardDragStart', (card: Card) => this.handleCardDragStart(card))
    this.locationCards.push(eternalPlains)

    var windscarredCrag = new LocationCard(this, (this.scale.width/2)-400,400, "Windscarred Crag", "Only the mind of a child could not see the danger this place is home to")
    windscarredCrag.on('cardDragStart', (card: Card) => this.handleCardDragStart(card))
    this.locationCards.push(windscarredCrag)

  }

  update(_timer: number, _delta: number) {
  }

  handleCardDragStart(card: Card) {
    this.children.bringToTop(card)
  }

  handleCardDrag(card: PlayerCard) {

    for (const location of this.locationCards) {
        const overlapping = Phaser.Geom.Intersects.RectangleToRectangle(card.getBounds(), location.getBounds())
        location.setHighlighted(overlapping)
    }
  }

  handleCardDrop(card: PlayerCard) {

    const locationHighlighted = this.locationCards.find(loc => loc.isHighlighted)
     if (!locationHighlighted) { 
    }else {
          this.lastActionText.text = `${card.name} played on ${locationHighlighted.name}`
        locationHighlighted?.setHighlighted(false)
    }
  }

}