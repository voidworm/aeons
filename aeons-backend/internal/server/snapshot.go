package server

import (
	"slices"
	"aeons/internal/card"
	"aeons/internal/game"
	"aeons/internal/collecting"
	"aeons/internal/location"
	"aeons/internal/player"
)

type GameStateDTO struct {
	Cycle        int               `json:"cycle"`
	Phase        string            `json:"phase"`
	Locations    []LocationDTO     `json:"locations"`
	Players      []PlayerDTO       `json:"players"`
	Creatures    []CreatureDTO     `json:"creatures"`
	Nodes        []NodeDTO         `json:"nodes"`
	CardsInHand  []PlayableCardDTO `json:"cardsInHand"`
	ScenarioName string            `json:"scenarioName"`
	Quests       []QuestDTO        `json:"quests"`
	PendingPayment *PaymentDTO     `json:"pendingPayment"`
	EndSpiritPhaseBlocked bool     `json:"endSpiritPhaseBlocked"`
}

type QuestDTO struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Flavor     string    `json:"flavor"`
	Goals      []GoalDTO `json:"goals"`
	Mandatory  bool      `json:"mandatory"`
	Finishable bool      `json:"finishable"`
}

type GoalDTO struct {
	Text string `json:"text"`
	Met  bool   `json:"met"`
}

// the players have to say who pays how much before the quest is finished
type PaymentDTO struct {
	QuestID int        `json:"questId"`
	Amount  int        `json:"amount"`
	Payers  []PayerDTO `json:"payers"`
}

type PayerDTO struct {
	PlayerID  int    `json:"playerId"`
	Name      string `json:"name"`
	Resources int    `json:"resources"`
}

type LocationDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Connections []int  `json:"connections"`
	Flavor      string `json:"flavor"`
	Visibility  string `json:"visibility"` // visible, memory (unveiled once, veiled now) or veiled
}

const (
	visibilityVisible = "visible"
	visibilityMemory  = "memory"
	visibilityVeiled  = "veiled"
)

func visibility(l *location.Unit, fogLifted bool) string {
	switch {
	case fogLifted:
		return visibilityVisible
	case !l.Veiled:
		return visibilityVisible
	case l.Known:
		return visibilityMemory
	default:
		return visibilityVeiled
	}
}

type PlayerDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"locationId"`
	Wounds     int    `json:"wounds"`
	MaxWounds  int    `json:"maxWounds"`
	Fainted    bool   `json:"fainted"`
	Hidden     bool   `json:"hidden"`
	Actions    int    `json:"actions"`
	Hand       []int  `json:"hand"`
	Food       int    `json:"food"`
	Resources  int    `json:"resources"`
	Energy     int    `json:"energy"`
	MaxEnergy  int    `json:"maxEnergy"`
}

type CreatureDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	IdentNames []string `json:"identNames"`
	LocationID int    `json:"locationId"`
	Distress    int    `json:"distress"`
	MaxDistress int    `json:"maxDistress"`
	Size        int    `json:"size"`
	Flavor     string `json:"flavor"`
	Exhausted  bool   `json:"exhausted"`
	Hidden     bool   `json:"hidden"`
	Bonded     bool   `json:"bonded"`
	KindredOf  *int   `json:"kindredOf"` // id of the player this creature is kindred to
}

type PlayableCardDTO struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Cost         int    `json:"cost"`
	EffectText   string `json:"effectText"`
	Flavor       string `json:"flavor"`
	EffectType  int    `json:"effectType"`
	CanBeCast    bool   `json:"canBeCast"`
	Range        int    `json:"range"`
	TargetType   *int   `json:"targetType"`
	CandidateIDs []int  `json:"candidateIds"`
}

type NodeDTO struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	LocationID int      `json:"locationId"`
	Food       PoolDTO  `json:"food"`
	Material   PoolDTO  `json:"material"`
	Modes      []string `json:"modes"` // what can be collected here right now
	Flavor     string   `json:"flavor"`
}

type PoolDTO struct {
	Remaining     int `json:"remaining"`
	Max           int `json:"max"`
	RegrowthTimer int `json:"regrowthTimer"`
	TurnsToRegrow int `json:"turnsToRegrow"`
}

func poolDTO(p *collecting.Pool) PoolDTO {
	return PoolDTO{
		Remaining:     p.Remaining,
		Max:           p.Max,
		RegrowthTimer: p.RegrowthTimer,
		TurnsToRegrow: p.TicksUntilRegrowth(),
	}
}

func modeStrings(modes []collecting.Mode) []string {
	out := make([]string, 0, len(modes))
	for _, m := range modes {
		out = append(out, string(m))
	}
	return out
}

func playableCardDTO(c *card.PlayableCard) PlayableCardDTO {
	out := PlayableCardDTO{
		ID: c.ID, Name: c.Name, EffectText: c.EffectText, Flavor: c.Flavor,
		Cost:       c.Cost,
		EffectType: int(c.Type), CanBeCast: c.CanBeCast, Range: c.Range,
		CandidateIDs: make([]int, 0, len(c.Candidates)),
	}
	if c.TargetSpec != nil {
		t := int(c.TargetSpec.Type)
		out.TargetType = &t
	}
	for _, t := range c.Candidates {
		out.CandidateIDs = append(out.CandidateIDs, t.GetID())
	}
	return out
}

func questDTOs(gs *game.GameState) []QuestDTO {
	finishable := gs.FinishableQuests()
	out := []QuestDTO{}
	for _, q := range gs.ActiveQuests() {
		goals := make([]GoalDTO, 0, len(q.Goals))
		for _, g := range q.Goals {
			goals = append(goals, GoalDTO{Text: g.Text(), Met: g.Met(gs.QuestWorld())})
		}
		out = append(out, QuestDTO{
			ID: q.ID, Name: q.Name, Flavor: q.Flavor, Goals: goals,
			Mandatory: q.Mandatory, Finishable: slices.Contains(finishable, q),
		})
	}
	return out
}

func paymentDTO(gs *game.GameState) *PaymentDTO {
	q := gs.PendingPayment
	if q == nil {
		return nil
	}
	payment := q.PaymentGoal().Payment
	out := &PaymentDTO{QuestID: q.ID, Amount: payment.Amount, Payers: []PayerDTO{}}
	for _, p := range payment.Payers(gs.QuestWorld()) {
		out.Payers = append(out.Payers, PayerDTO{PlayerID: p.ID, Name: p.Name, Resources: p.ResourcesAvailable})
	}
	return out
}

//fogLifted is a playtest aid: the players see everything, whatever they actually know
func Snapshot(gs *game.GameState, fogLifted bool) GameStateDTO {
	// veiled locations can leave these empty, which must reach the frontend as [] and not null
	out := GameStateDTO{Phase: string(gs.Phase), Cycle: gs.Cycle, Creatures: []CreatureDTO{}, Nodes: []NodeDTO{}}

	for _, l := range gs.Locations {
		conns := make([]int, 0, len(l.OutgoingConnections))
		for _, c := range l.OutgoingConnections {
			conns = append(conns, c.ID)
		}
		out.Locations = append(out.Locations, LocationDTO{ID: l.ID, Name: l.Name, Connections: conns, Flavor: l.Flavor, Visibility: visibility(l, fogLifted)})
	}

	for _, p := range gs.Players {
		hand := make([]int, 0, len(p.CardsInHand))
		for _, c := range p.CardsInHand {
			hand = append(hand, c.ID)
		}
		out.Players = append(out.Players, PlayerDTO{
			ID: p.ID, Name: p.Name, LocationID: p.CurrentLocation.ID,
			Wounds: p.Wounds, MaxWounds: player.MaxWounds, Fainted: p.Fainted(), Hidden: p.Hidden,
			Actions: p.RemainingActions, Hand: hand, Food: p.Food, Resources: p.ResourcesAvailable, Energy: p.Energy, MaxEnergy: game.AwakeningEnergy,
		})
	}

	for _, c := range gs.Creatures {
		if c.CurrentLocation.Veiled && !fogLifted {
			continue
		}
		if c.Hidden && !c.HiddenInSight && !fogLifted {
			continue
		}
		var kindredOf *int
		if c.BondedTo != nil {
			kindredOf = &c.BondedTo.ID
		}
		out.Creatures = append(out.Creatures, CreatureDTO{KindredOf: kindredOf,
			ID: c.ID, Name: c.Name, IdentNames: c.IdentNames, Flavor: c.Flavor, LocationID: c.CurrentLocation.ID,
			Distress: c.Distress, MaxDistress: c.MaxDistress, Size: c.Size,
			Exhausted: c.Exhausted, Hidden: c.Hidden, Bonded: c.Bonded,
		})
	}

	for _, live := range gs.Nodes {
		n, known := gs.KnownNodes[live.ID]
		if fogLifted {
			n, known = live, true
		}
		if !known {
			continue
		}
		out.Nodes = append(out.Nodes, NodeDTO{
			ID: n.ID, Name: n.Name, Flavor: n.Flavor, LocationID: n.StaticLocation.ID,
			Food: poolDTO(n.Food), Material: poolDTO(n.Material), Modes: modeStrings(n.Modes()),
		})
	}

	for _, c := range gs.PlayableCards {
		out.CardsInHand = append(out.CardsInHand, playableCardDTO(c))
	}

	if gs.Scenario != nil {
		out.ScenarioName = gs.Scenario.Name
	}
	out.Quests = questDTOs(gs)
	out.PendingPayment = paymentDTO(gs)
	out.EndSpiritPhaseBlocked = gs.MandatoryFinishPending()

	return out
}
