package game

import (
	"aeons/internal/enemy"
	"aeons/internal/healthpool"
	"aeons/internal/location"
	"aeons/internal/moving"
	"aeons/internal/player"
	"log"
)

func (gs *GameState) Init() {

	gs.InitLocations()
	gs.InitPlayers()
	gs.InitEnemies()
}

func (gs *GameState) InitLocations() {

	log.Println("Setting up Locations...")
	Porch := &location.LocationEntity{ID: 1, Name: "Porch"}
	DownstairsHallway := &location.LocationEntity{ID: 1, Name: "Downstairs Hallway"}
	Kitchen := &location.LocationEntity{ID: 1, Name: "Kitchen"}
	LivingRoom := &location.LocationEntity{ID: 1, Name: "Living Room"}
	UpstairsHallway := &location.LocationEntity{ID: 1, Name: "Upstairs Hallway"}
	SleepingRoom := &location.LocationEntity{ID: 1, Name: "Sleeping Room"}
	Attic := &location.LocationEntity{ID: 1, Name: "Attic"}

	log.Println("Setting up Locations linking...")
	Porch.OutgoingConnections = append(Porch.OutgoingConnections, DownstairsHallway)
	DownstairsHallway.OutgoingConnections = append(DownstairsHallway.OutgoingConnections, Porch, Kitchen, LivingRoom, UpstairsHallway)
	Kitchen.OutgoingConnections = append(Kitchen.OutgoingConnections, DownstairsHallway, LivingRoom)
	LivingRoom.OutgoingConnections = append(LivingRoom.OutgoingConnections, DownstairsHallway, Kitchen)
	UpstairsHallway.OutgoingConnections = append(UpstairsHallway.OutgoingConnections, DownstairsHallway, SleepingRoom, Attic)
	SleepingRoom.OutgoingConnections = append(SleepingRoom.OutgoingConnections, UpstairsHallway, Porch)
	Attic.OutgoingConnections = append(Attic.OutgoingConnections, UpstairsHallway)

	gs.Locations = append(gs.Locations, Porch, DownstairsHallway, Kitchen, LivingRoom, UpstairsHallway, SleepingRoom, Attic)
}

func (gs *GameState) InitPlayers() {
	log.Println("Setting up Jim...")
	Jim := &player.Player{
		MovingEntity:       moving.MovingEntity{Location: gs.Locations[0]},
		HealthPoolEntity:   healthpool.HealthPoolEntity{CurrentHealth: 10, MaxHealth: 10},
		ID:                 1,
		Name:               "Jim Gordon",
		CardsInHand:        7,
		ResourcesAvailable: 5,
		RemainingActions:   3,
	}

	log.Println("Setting up Ivy...")
	Ivy := &player.Player{
		MovingEntity:       moving.MovingEntity{Location: gs.Locations[0]},
		HealthPoolEntity:   healthpool.HealthPoolEntity{CurrentHealth: 8, MaxHealth: 8},
		ID:                 1,
		Name:               "Poison Ivy",
		CardsInHand:        7,
		ResourcesAvailable: 5,
		RemainingActions:   3,
	}

	gs.Players = append(gs.Players, Jim, Ivy)
}

func (gs *GameState) InitEnemies() {

	log.Println("Setting up Ghoul...")
	ghoul := &enemy.EnemyEntity{
		MovingEntity:     moving.MovingEntity{Location: gs.Locations[len(gs.Locations)-1]},
		HealthPoolEntity: healthpool.HealthPoolEntity{CurrentHealth: 5, MaxHealth: 5},
		ID:               1,
		Name:             "Noxious Ghoul",
		Aloof:            false,
		Hunter:           true,
		Damage:           2,
	}

	log.Println("Setting up Rat...")
	rat := &enemy.EnemyEntity{
		MovingEntity:     moving.MovingEntity{Location: gs.Locations[3]},
		HealthPoolEntity: healthpool.HealthPoolEntity{CurrentHealth: 2, MaxHealth: 2},
		ID:               1,
		Name:             "Chittering Rat",
		Aloof:            false,
		Hunter:           false,
		Damage:           1,
	}

	log.Println("Setting up Suspicious Plant...")
	plant := &enemy.EnemyEntity{
		MovingEntity:     moving.MovingEntity{Location: gs.Locations[3]},
		HealthPoolEntity: healthpool.HealthPoolEntity{CurrentHealth: 2, MaxHealth: 2},
		ID:               1,
		Name:             "Suspicious Plant",
		Aloof:            true,
		Hunter:           false,
		Damage:           1,
	}

	gs.Enemies = append(gs.Enemies, ghoul, rat, plant)
}
