package server

import (
	"aeons/internal/card"
	"aeons/internal/game"
	"aeons/internal/harvesting"
)

type GameStateDTO struct {
	TurnCounter int `json:"turnCounter"`
	Locations []LocationDTO `json:"locations"`
	Players []PlayerDTO `json:"players"`
	Creatures []CreatureDTO `json:"creatures"`
	HarvestNodes []HarvestableDTO `json:"harvestNodes"`
	CardsInHand []PlayableCardDTO `json:"cardsInHand"`
}

type LocationDTO struct {
	ID	int `json:"id"`
	Name string `json:"name"`
	Connections []int `json:"connections"`
	Flavor string `json:"flavor"`
}

type PlayerDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"locationId"`
	Health     int    `json:"health"`
	MaxHealth  int    `json:"maxHealth"`
	Actions    int    `json:"actions"`
	Hand       []int  `json:"hand"`
	Resources  int    `json:"resources"`
	Damage     int    `json:"damage"`
}

type CreatureDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"locationId"`
	Health     int    `json:"health"`
	MaxHealth  int    `json:"maxHealth"`
	Damage     int    `json:"damage"`
	Flavor		string `json:"flavor"`
}

type PlayableCardDTO struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Cost int `json:"cost"`
	EffectText   string `json:"effectText"`
	EffectType   int    `json:"effectType"`
	CanBeCast    bool   `json:"canBeCast"`
	Range        int    `json:"range"`
	TargetType   *int   `json:"targetType"`
	CandidateIDs []int  `json:"candidateIds"`
}

type HarvestableDTO struct {
	ID         int          `json:"id"`
	Name       string       `json:"name"`
	LocationID int          `json:"locationId"`
	Healing    YieldDTO     `json:"healing"`
	Resources  YieldDTO     `json:"resources"`
	Flavor	string	`json:"flavor"`
}

type YieldDTO struct {
	Capacity     int `json:"capacity"`
	MaxCapacity  int `json:"maxCapacity"`
	YieldAmount  int `json:"yieldAmount"`
	RespawnTicks int `json:"respawnTicks"`
	CurrentTicks int `json:"currentTicks"`
}

func yieldDTO(yc *harvesting.YieldConfig) YieldDTO {
	return YieldDTO{
		Capacity:     yc.CurrentCapacity,
		MaxCapacity:  yc.MaxCapacity,
		YieldAmount:  yc.YieldAmount,
		RespawnTicks: yc.RespawnTicks,
		CurrentTicks: yc.CurrentTicks,
	}
}

func playableCardDTO(c *card.PlayableCard) PlayableCardDTO {
	out := PlayableCardDTO{
		ID: c.ID, Name: c.Name, EffectText: c.EffectText,
		Cost: c.Cost,
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
			Health: p.CurrentHealth, MaxHealth: p.MaxHealth,
			Actions: p.RemainingActions, Hand: hand, Resources: p.ResourcesAvailable,
			Damage: p.Damage,
		})
	}

	for _, c := range gs.Creatures {
		out.Creatures = append(out.Creatures, CreatureDTO{
			ID: c.ID, Name: c.Name, Flavor: c.Flavor, LocationID: c.CurrentLocation.ID,
			Health: c.CurrentHealth, MaxHealth: c.MaxHealth, Damage: c.Damage,
		})
	}

	for _, h := range gs.HarvestUnits {
		out.HarvestNodes = append(out.HarvestNodes, HarvestableDTO{
			ID: h.ID, Name: h.Name, Flavor: h.Flavor, LocationID: h.StaticLocation.ID,
			Healing:   yieldDTO(h.HealYieldConfig),
			Resources: yieldDTO(h.ResourceYieldConfig),
		})
	}

	for _, c := range gs.PlayableCards {
		out.CardsInHand = append(out.CardsInHand, playableCardDTO(c))
	}

	return out
}