package game

import (
	"aeons/internal/card"
	"aeons/internal/creature"
	"aeons/internal/collecting"
	"aeons/internal/location"
	"aeons/internal/player"
	"aeons/internal/quest"
)

type GameState struct {
	nextID int
	Cycle int // the round counter, increases whenever a round ends
	ForwardStack []*creature.Unit // the creatures yet to act in the first pass of the round
	ReverseStack []*creature.Unit // the creatures that acted in the first pass, they act again in reverse
	Resting bool // the players rest, they and their bonded creatures are out of the hunters' sight
	creatureEvents []CreatureEvent
	eventLog []string
	Locations    []*location.Unit
	Players      []*player.Unit
	Creatures    []*creature.Unit
	Nodes []*collecting.Node
	KnownNodes map[int]*collecting.Node // the nodes as the players last saw them, by node id
	PlayableCards			[]*card.PlayableCard
	Scenario *quest.Group
	AdminMode bool // playtest, hands the players the cards that only admins have
	PendingPayment *quest.Quest // a quest whose goals are met, waiting for the players to say who pays
}

func (gs *GameState) NextID() int {
	gs.nextID++
	return gs.nextID
}

func (gs *GameState) ReconcileDefeats() {
	aliveCreatures := gs.Creatures[:0]
	for _, v := range gs.Creatures {
		if !v.IsDefeated() {
			aliveCreatures = append(aliveCreatures, v)
			continue
		}
		gs.Logf("%s is defeated", v.NameWithID())
		gs.Nodes = append(gs.Nodes, collecting.NewCorpse(gs.NextID(), v.Name, v.CurrentLocation))
	}
	gs.Creatures = aliveCreatures
}

func (gs *GameState) CreaturesAtLocationWhere(le *location.Unit, matches func(*creature.Unit) bool) []*creature.Unit {
	found := []*creature.Unit{}
	for _, v := range gs.Creatures {
		if le == v.CurrentLocation && matches(v) {
			found = append(found, v)
		}
	}
	return found
}

func (gs *GameState) LocationHasBlockingCreature(le *location.Unit) bool {
	for _, v := range gs.Creatures {
		blocks := v.Role.IsPredator() && v.Size >= player.Size
		if le == v.CurrentLocation && blocks && !v.Exhausted && !v.Bonded && !v.Hidden {
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
	case card.TargetTypeQuest:
		return toTargetables(gs.FinishableQuests())
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
		case card.TargetTypeQuest:
			return getEntryWithID(gs.FinishableQuests(),id)
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
				v.Candidates = toTargetables(gs.CreaturesAtLocationWhere(p.CurrentLocation, (*creature.Unit).IsConfrontable))
				v.CanBeCast = len(v.Candidates) > 0
			case card.PlayerDistract:
				v.Candidates = toTargetables(gs.CreaturesAtLocationWhere(p.CurrentLocation, (*creature.Unit).IsDistractable))
				v.CanBeCast = len(v.Candidates) > 0
			case card.PlayerBond:
				v.Candidates = toTargetables(gs.CreaturesAtLocationWhere(p.CurrentLocation, (*creature.Unit).IsBondable))
				v.CanBeCast = gs.KindredOf(p) == nil && len(v.Candidates) > 0
			case card.PlayerHide:
				v.CanBeCast = !p.Hidden
			case card.PlayerRest:
				v.CanBeCast = true
			case card.PlayerMoveKindred:
				kindred := gs.KindredOf(p)
				v.CanBeCast = kindred != nil && !v.OnCooldown && len(kindred.CurrentLocation.OutgoingConnections) > 0
			case card.PlayerGather:
				v.CanBeCast = gs.LocationHasGatherable(p.CurrentLocation) && !gs.LocationHasBlockingCreature(p.CurrentLocation)
			case card.PlayerFinishQuest:
				v.CanBeCast = len(v.Candidates) > 0
			case card.PlayerTeleport:
				v.CanBeCast = true
			default:
				v.CanBeCast = true
		}
		if p.Energy < v.Cost {
			v.CanBeCast = false
		}
		if v.Type != card.PlayerFinishQuest && v.Type != card.PlayerTeleport && gs.MandatoryFinishPending() {
			v.CanBeCast = false
		}
	}
}	