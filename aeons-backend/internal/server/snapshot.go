package server

import (
	"aeons/internal/card"
	"aeons/internal/game"
	"aeons/internal/collecting"
	"aeons/internal/player"
)

type GameStateDTO struct {
	TurnCounter  int               `json:"turnCounter"`
	Locations    []LocationDTO     `json:"locations"`
	Players      []PlayerDTO       `json:"players"`
	Creatures    []CreatureDTO     `json:"creatures"`
	Nodes        []NodeDTO         `json:"nodes"`
	CardsInHand  []PlayableCardDTO `json:"cardsInHand"`
}

type LocationDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Connections []int  `json:"connections"`
	Flavor      string `json:"flavor"`
}

type PlayerDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"locationId"`
	Wounds     int    `json:"wounds"`
	MaxWounds  int    `json:"maxWounds"`
	Fainted    bool   `json:"fainted"`
	Actions    int    `json:"actions"`
	Hand       []int  `json:"hand"`
	Food       int    `json:"food"`
	Resources  int    `json:"resources"`
	Energy     int    `json:"energy"`
}

type CreatureDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"locationId"`
	Health     int    `json:"health"`
	MaxHealth  int    `json:"maxHealth"`
	Flavor     string `json:"flavor"`
	Exhausted  bool   `json:"exhausted"`
	Bonded     bool   `json:"bonded"`
	KindredOf  *int   `json:"kindredOf"` // id of the player this creature is kindred to
}

type PlayableCardDTO struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Cost         int    `json:"cost"`
	EffectText   string `json:"effectText"`
	EffectType   int    `json:"effectType"`
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
		ID: c.ID, Name: c.Name, EffectText: c.EffectText,
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

func Snapshot(gs *game.GameState) GameStateDTO {
	var out GameStateDTO

	for _, l := range gs.Locations {
		conns := make([]int, 0, len(l.OutgoingConnections))
		for _, c := range l.OutgoingConnections {
			conns = append(conns, c.ID)
		}
		out.Locations = append(out.Locations, LocationDTO{ID: l.ID, Name: l.Name, Connections: conns, Flavor: l.Flavor})
	}

	for _, p := range gs.Players {
		hand := make([]int, 0, len(p.CardsInHand))
		for _, c := range p.CardsInHand {
			hand = append(hand, c.ID)
		}
		out.Players = append(out.Players, PlayerDTO{
			ID: p.ID, Name: p.Name, LocationID: p.CurrentLocation.ID,
			Wounds: p.Wounds, MaxWounds: player.MaxWounds, Fainted: p.Fainted(),
			Actions: p.RemainingActions, Hand: hand, Food: p.Food, Resources: p.ResourcesAvailable, Energy: p.Energy,
		})
	}

	for _, c := range gs.Creatures {
		var kindredOf *int
		if c.BondedTo != nil {
			kindredOf = &c.BondedTo.ID
		}
		out.Creatures = append(out.Creatures, CreatureDTO{KindredOf: kindredOf,
			ID: c.ID, Name: c.Name, Flavor: c.Flavor, LocationID: c.CurrentLocation.ID,
			Health: c.CurrentHealth, MaxHealth: c.MaxHealth, Exhausted: c.Distracted, Bonded: c.Bonded,
		})
	}

	for _, n := range gs.Nodes {
		out.Nodes = append(out.Nodes, NodeDTO{
			ID: n.ID, Name: n.Name, Flavor: n.Flavor, LocationID: n.StaticLocation.ID,
			Food: poolDTO(n.Food), Material: poolDTO(n.Material), Modes: modeStrings(n.Modes()),
		})
	}

	for _, c := range gs.PlayableCards {
		out.CardsInHand = append(out.CardsInHand, playableCardDTO(c))
	}

	return out
}
