package server

import (
	"aeons/internal/game"
	"aeons/internal/harvesting"
)

type GameStateDTO struct {
	TurnCounter int `json:"turnCounter"`
	Locations []LocationDTO `json:"locations"`
	Players []PlayerDTO `json:"players"`
	Creatures []CreatureDTO `json:"creatures"`
	Harvestables []HarvestableDTO `json:"harvestables"`
}

type LocationDTO struct {
	ID	int `json:"id"`
	Name string `json:"name"`
	Connections []int `json:"connections"`
}

type PlayerDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"locationId"`
	Health     int    `json:"health"`
	MaxHealth  int    `json:"maxHealth"`
	Actions    int    `json:"actions"`
	Hand       int    `json:"cardsInHand"`
	Resources  int    `json:"resources"`	
}

type CreatureDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"locationId"`
	Health     int    `json:"health"`
	MaxHealth  int    `json:"maxHealth"`
}

type HarvestableDTO struct {
	ID         int          `json:"id"`
	Name       string       `json:"name"`
	LocationID int          `json:"locationId"`
	Healing    YieldDTO     `json:"healing"`
	Resources  YieldDTO     `json:"resources"`
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

func Snapshot(gs *game.GameState) GameStateDTO {
	var out GameStateDTO

	for _, l := range gs.Locations {
		conns := make([]int, 0, len(l.OutgoingConnections))
		for _, c := range l.OutgoingConnections {
			conns = append(conns, c.ID)
		}
		out.Locations = append(out.Locations, LocationDTO{ID: l.ID, Name: l.Name, Connections: conns})
	}

	for _, p := range gs.Players {
		out.Players = append(out.Players, PlayerDTO{
			ID: p.ID, Name: p.Name, LocationID: p.CurrentLocation.ID,
			Health: p.CurrentHealth, MaxHealth: p.MaxHealth,
			Actions: p.RemainingActions, Hand: p.CardsInHand, Resources: p.ResourcesAvailable,
		})
	}

	for _, c := range gs.Creatures {
		out.Creatures = append(out.Creatures, CreatureDTO{
			ID: c.ID, Name: c.Name, LocationID: c.CurrentLocation.ID,
			Health: c.CurrentHealth, MaxHealth: c.MaxHealth,
		})
	}

	for _, h := range gs.HarvestUnits {
		out.Harvestables = append(out.Harvestables, HarvestableDTO{
			ID: h.ID, Name: h.Name, LocationID: h.StaticLocation.ID,
			Healing:   yieldDTO(h.HealYieldConfig),
			Resources: yieldDTO(h.ResourceYieldConfig),
		})
	}

	return out
}