package game

import (
	"aeons/internal/card"
	"aeons/internal/creature"
	"aeons/internal/harvesting"
	"aeons/internal/location"
	"aeons/internal/player"
)

type GameState struct {
	nextID int
	Locations    []*location.Unit
	Players      []*player.Unit
	Creatures    []*creature.Unit
	HarvestUnits []*harvesting.Unit
	PlayableCards 			[]*card.PlayableCard
}

func (gs *GameState) NextID() int {
	gs.nextID++
	return gs.nextID
}

func (gs *GameState) ReconcileDefeats() {
	alivePlayers := gs.Players[:0]

	for _, v := range gs.Players {
		if v.CurrentHealth > 0 {
			alivePlayers = append(alivePlayers, v)
		}
	}

	aliveCreatures := gs.Creatures[:0]
	for _, v := range gs.Creatures {
		if v.CurrentHealth > 0 {
			aliveCreatures = append(aliveCreatures, v)
		}
	}
	gs.Creatures = aliveCreatures
	gs.Players = alivePlayers
}

func (gs *GameState) LocationHasEnemies(le *location.Unit) bool {
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation {
			return true
		}
	}
	return false
}

func (gs *GameState) LocationHasPlayers(le *location.Unit) bool {
	for _, v := range gs.Players {
		if le == v.CurrentLocation {
			return true
		}
	}
	return false
}

func (gs *GameState) LocationHaEvadableCreatures(le *location.Unit) bool {

	if !gs.LocationHasEnemies(le) {
		return false
	}

	for _, v := range gs.Creatures {
		if v.IsEvadable() {
			if v.CurrentLocation == le {
				return true
			}
		}
	}
	return false
}

func (gs *GameState) LocationCanBeLeft(le *location.Unit) bool {

	if !gs.LocationHasEnemies(le) {
		return true
	}

	for _, v := range gs.Creatures {
		if le == v.CurrentLocation {
			creatureBlocksMovement := !v.Aloof && !v.Exhausted
			if creatureBlocksMovement {
				return false
			}
		}
	}

	return true
}

func (gs *GameState) LocationHasHarvest(le *location.Unit) bool {
	for _, v := range gs.HarvestUnits {
		if le == v.StaticLocation {
			return true
		}
	}

	return false
}

func (gs *GameState) EnemiesAtLocation(le *location.Unit) []*creature.Unit {

	found := []*creature.Unit{}
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation {
			found = append(found, v)
		}
	}
	return found
}
func (gs *GameState) EvadableEnemiesAtLocation(le *location.Unit) []*creature.Unit {

	found := []*creature.Unit{}
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation && v.IsEvadable() {
			found = append(found, v)
		}
	}
	return found
}

func (gs *GameState) HarvestUnitsAtLocation(le *location.Unit) []*harvesting.Unit {

	found := []*harvesting.Unit{}
	for _, v := range gs.HarvestUnits {
		if le == v.StaticLocation {
			found = append(found, v)
		}
	}
	return found
}

func (gs *GameState) PlayersAtLocation(ee *creature.Unit) []*player.Unit {

	found := []*player.Unit{}
	for _, v := range gs.Players {
		if ee.CurrentLocation == v.CurrentLocation {
			found = append(found, v)
		}
	}

	return found
}

func (gs *GameState) PlayerHaveActionsRemaining() bool {
	for _, v := range gs.Players {
		if v.RemainingActions > 0 {
			return true
		}
	}
	return false
}

func (gs *GameState) Candidates(kind card.TargetType) []card.Targetable{

	switch kind {
	case card.TargetTypeCreature:
		return toTargetables(gs.Creatures)
	case card.TargetTypeHarvest:
		return toTargetables(gs.HarvestUnits)
	case card.TargetTypePlayer:
		return toTargetables(gs.Players)
	case card.TargetTypeLocation:
		return toTargetables(gs.Locations)
	default:
		return []card.Targetable{}
	}
}

func toTargetables[T card.Targetable](in []T) []card.Targetable {
	out := make([]card.Targetable,0,len(in))
	for _,v := range in {
		out = append(out, v)
	}
	return out
}

func getEntryWithID[T card.Targetable](in []T, id int) card.Targetable {
	for _,v := range in {
		if v.GetID() == id {
			return  v
		}
	}
	return nil
}


func (gs *GameState) FillCardWithCandidates(input card.PlayableCard) {
		switch input.TargetSpec.Type {
		case card.TargetTypeCreature:
			input.Candidates = gs.Candidates(card.TargetTypeCreature)
		case card.TargetTypePlayer:
			input.Candidates = gs.Candidates(card.TargetTypePlayer)
		case card.TargetTypeHarvest:
			input.Candidates = gs.Candidates(card.TargetTypeHarvest)
		case card.TargetTypeLocation:
			input.Candidates = gs.Candidates(card.TargetTypeLocation)
		default:
			input.Candidates = nil	
		}
	}

func (gs *GameState) GetTargetWithIdAndType(id int, ttype card.TargetType) card.Targetable {
		switch ttype {
		case card.TargetTypeCreature:
			return getEntryWithID(gs.Creatures,id)
		case card.TargetTypePlayer:
			return getEntryWithID(gs.Players,id)
		case card.TargetTypeHarvest:
			return getEntryWithID(gs.HarvestUnits,id)
		case card.TargetTypeLocation:
			return getEntryWithID(gs.Locations,id)
		default:
			return nil	
		}
}


func (gs *GameState) EvaluatePlayerHandPlayability(p *player.Unit) {
	for _,v := range p.CardsInHand {
		switch v.Type {
			case card.PlayerMove: 
				if gs.LocationCanBeLeft(p.CurrentLocation){
					v.CanBeCast = true
				}else {
					v.CanBeCast = false
				}
			case card.PlayerAttack: 
				if gs.LocationHasEnemies(p.CurrentLocation){
					v.CanBeCast = true
				}else {
					v.CanBeCast = false
				}
			case card.PlayerDistract:
				if gs.LocationHaEvadableCreatures(p.CurrentLocation) {
					v.CanBeCast = true
				}else {
					v.CanBeCast = false
				}
			case card.PlayerHarvest:
				if gs.LocationHasHarvest(p.CurrentLocation) {
					v.CanBeCast = true
				}else {
					v.CanBeCast = false
				}
			default:
				v.CanBeCast = true
		}
	}
}	