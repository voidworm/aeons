package game

import (
	"aeons/internal/card"
	"aeons/internal/creature"
	"aeons/internal/collecting"
	"aeons/internal/location"
	"aeons/internal/player"
)

type GameState struct {
	nextID int
	eventLog []string
	Locations    []*location.Unit
	Players      []*player.Unit
	Creatures    []*creature.Unit
	Nodes []*collecting.Node
	PlayableCards 			[]*card.PlayableCard
}

func (gs *GameState) NextID() int {
	gs.nextID++
	return gs.nextID
}

func (gs *GameState) ReconcileDefeats() {
	aliveCreatures := gs.Creatures[:0]
	for _, v := range gs.Creatures {
		if v.CurrentHealth > 0 {
			aliveCreatures = append(aliveCreatures, v)
		}
	}
	gs.Creatures = aliveCreatures
}

func (gs *GameState) LocationHasEnemies(le *location.Unit) bool {
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation && !v.Bonded {
			return true
		}
	}
	return false
}

func (gs *GameState) LocationHasActivePlayers(le *location.Unit) bool {
	for _, v := range gs.Players {
		if le == v.CurrentLocation && !v.Fainted() {
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

func (gs *GameState) LocationHasBondableCreatures(le *location.Unit) bool {
	for _, v := range gs.Creatures {
		if v.CurrentLocation == le && v.IsBondable() {
			return true
		}
	}
	return false
}

func (gs *GameState) LocationHasBlockingCreature(le *location.Unit) bool {
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation && !v.Aloof && !v.Distracted && !v.Bonded {
			return true
		}
	}
	return false
}

func (gs *GameState) LocationCanBeLeft(le *location.Unit) bool {
	return !gs.LocationHasBlockingCreature(le)
}

func (gs *GameState) LocationHasGatherable(le *location.Unit) bool {
	for _, v := range gs.Nodes {
		if le == v.StaticLocation && len(v.Modes()) > 0 {
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

func (gs *GameState) NodesAtLocation(le *location.Unit) []*collecting.Node {

	found := []*collecting.Node{}
	for _, v := range gs.Nodes {
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
	case card.TargetTypeNode:
		return toTargetables(gs.Nodes)
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
		case card.TargetTypeNode:
			input.Candidates = gs.Candidates(card.TargetTypeNode)
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
		case card.TargetTypeNode:
			return getEntryWithID(gs.Nodes,id)
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
			case card.PlayerBond:
				v.CanBeCast = gs.KindredOf(p) == nil && gs.LocationHasBondableCreatures(p.CurrentLocation)
			case card.PlayerMoveKindred:
				kindred := gs.KindredOf(p)
				v.CanBeCast = kindred != nil && !v.OnCooldown && len(kindred.CurrentLocation.OutgoingConnections) > 0
			case card.PlayerGather:
				v.CanBeCast = gs.LocationHasGatherable(p.CurrentLocation) && !gs.LocationHasBlockingCreature(p.CurrentLocation)
			default:
				v.CanBeCast = true
		}
		if p.Energy < v.Cost {
			v.CanBeCast = false
		}
	}
}	