package game

import (
	"aeons/internal/card"
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/evading"
	"aeons/internal/harvesting"
	"aeons/internal/location"
	"aeons/internal/moving"
	"aeons/internal/player"
	"log"
	"math/rand/v2"
	"slices"
)

func (gs *GameState) Init() {

	gs.InitLocations()
	gs.InitHarvestingUnits()
	gs.InitPlayers()
	gs.InitCreatures()
}

func (gs *GameState) InitLocations() {

	log.Println("[INIT] Setting up locations....")
	SecludedDen := &location.Unit{ID: gs.NextID(), Name: "Secluded Den"}
	WindsweptPlains := &location.Unit{ID: gs.NextID(), Name: "Windswept Plains"}
	RedhornLake := &location.Unit{ID: gs.NextID(), Name: "Redhorn Lake"}
	ConiferousGrove := &location.Unit{ID: gs.NextID(), Name: "Coniferous Grove"}
	AridPlateau := &location.Unit{ID: gs.NextID(), Name: "Arid Plateau"}
	SlumberingCrag := &location.Unit{ID: gs.NextID(), Name: "Slumbering Crag"}
	HermitsRecluse := &location.Unit{ID: gs.NextID(), Name: "Hermit's Recluse"}

	SecludedDen.OutgoingConnections = append(SecludedDen.OutgoingConnections, WindsweptPlains)
	WindsweptPlains.OutgoingConnections = append(WindsweptPlains.OutgoingConnections, SecludedDen, RedhornLake, ConiferousGrove, AridPlateau)
	RedhornLake.OutgoingConnections = append(RedhornLake.OutgoingConnections, WindsweptPlains, ConiferousGrove)
	ConiferousGrove.OutgoingConnections = append(ConiferousGrove.OutgoingConnections, WindsweptPlains, RedhornLake)
	AridPlateau.OutgoingConnections = append(AridPlateau.OutgoingConnections, WindsweptPlains, SlumberingCrag, HermitsRecluse)
	SlumberingCrag.OutgoingConnections = append(SlumberingCrag.OutgoingConnections, AridPlateau, SecludedDen)
	HermitsRecluse.OutgoingConnections = append(HermitsRecluse.OutgoingConnections, AridPlateau)

	gs.Locations = append(gs.Locations, SecludedDen, WindsweptPlains, RedhornLake, ConiferousGrove, AridPlateau, SlumberingCrag, HermitsRecluse)
}

func (gs *GameState) InitHarvestingUnits() {
	grabBag := append([]*location.Unit(nil), gs.Locations...)
	for range 5 {
		r := rand.IntN(len(grabBag))
		randomLocation := grabBag[r]
		grabBag = slices.Delete(grabBag, r, r+1)

		h := harvesting.GenerateRandomPredefinedHarvestingUnit(gs.NextID(), randomLocation)
		gs.HarvestUnits = append(gs.HarvestUnits, h)
		log.Printf("[INIT] Spawning harvestable %s at %s\n", h.Name, randomLocation.Name)
	}
}

func (gs *GameState) InitPlayers() {
	log.Println("[INIT] Setting up player Druid...")
	Druid := &player.Unit{
		MovingEntity:       moving.MovingEntity{CurrentLocation: gs.Locations[0]},
		HealthPool:         combat.HealthPool{CurrentHealth: 10, MaxHealth: 10},
		ID:                 gs.NextID(),
		Name:               "Druid",
		CardsInHand:        []*card.Card{},
		ResourcesAvailable: 5,
		RemainingActions:   3,
		Damage:             1,
	}

	Druid.CardsInHand = gs.GenerateBasicHand()

	log.Println("[INIT] Setting up player Scoundrel...")
	Scoundrel := &player.Unit{
		MovingEntity:       moving.MovingEntity{CurrentLocation: gs.Locations[0]},
		HealthPool:         combat.HealthPool{CurrentHealth: 8, MaxHealth: 8},
		ID:                 gs.NextID(),
		Name:               "Scoundrel",
		CardsInHand:        []*card.Card{},
		ResourcesAvailable: 5,
		RemainingActions:   3,
		Damage:             1,
	}

	Scoundrel.CardsInHand = gs.GenerateBasicHand()

	gs.Players = append(gs.Players, Druid, Scoundrel)
}

func (gs *GameState) getRandomLocation() *location.Unit {
	return gs.Locations[rand.IntN(len(gs.Locations))]
}

func (gs *GameState) GenerateBasicHand() []*card.Card {
	basicHand := []*card.Card{}
	basicHand = append(basicHand, gs.GenerateBasicMoveCard(), gs.GenerateBasicAttackCard(), gs.GenerateBasicDistractCard(), gs.GenerateBasicHarvestCard())
	gs.Cards = append(gs.Cards, basicHand...)
	return basicHand
}

func (gs *GameState) GenerateBasicMoveCard() *card.Card {
	BasicMoveCard := &card.Card {
		ID: gs.NextID(),
		Type: card.PlayerMove,
		Name: "Move",
		EffectText: "Moves your character to an adjacent location.",
		TargetSpec: nil,
		Candidates: []card.Targetable{},
		Effect: nil,
		Range: 1,
	}

	BasicMoveTargetLocationSpec := &card.TargetSpec {
		Type: card.TargetTypeLocation,
		Target: nil,
		Filter: nil,
	}

	BasicMoveCard.TargetSpec = BasicMoveTargetLocationSpec
	return BasicMoveCard
}

func (gs *GameState) GenerateBasicAttackCard() *card.Card {
	BasicAttackCard := &card.Card {
		ID: gs.NextID(),
		Name: "Attack",
		Type: card.PlayerAttack,
		EffectText: "Attacks a target creature at your location.",
		TargetSpec: nil,
		Candidates: []card.Targetable{},
		Effect: nil,
		Range: 1,
	}

	BasicAttackTargetCreatureSpec := &card.TargetSpec {
		Type: card.TargetTypeCreature,
		Target: nil,
		Filter: nil,
	}

	BasicAttackCard.TargetSpec = BasicAttackTargetCreatureSpec
	return BasicAttackCard
}

func (gs *GameState) GenerateBasicDistractCard() *card.Card {
	BasicDistractCard := &card.Card {
		ID: gs.NextID(),
		Name: "Distract",
		EffectText: "Distracts a target creature at your location.",
		Type:	card.PlayerDistract,
		TargetSpec: nil,
		Candidates: []card.Targetable{},
		Effect: nil,
		Range: 1,
	}

	BasicDistractTargetCreatureSpec := &card.TargetSpec {
		Type: card.TargetTypeCreature,
		Target: nil,
		Filter: nil,
	}

	BasicDistractCard.TargetSpec = BasicDistractTargetCreatureSpec
	return BasicDistractCard
}

func (gs *GameState) GenerateBasicHarvestCard() *card.Card {
	BasicHarvestCard := &card.Card {
		ID: gs.NextID(),
		Type: card.PlayerHarvest,
		Name: "Harvest",
		EffectText: "Harvests a target Harvestable at your location.",
		TargetSpec: nil,
		Candidates: []card.Targetable{},
		Effect: nil,
		Range: 1,
	}

	BasicHarvestTargetHarvestable := &card.TargetSpec {
		Type: card.TargetTypeHarvest,
		Target: nil,
		Filter: nil,
	}

	BasicHarvestCard.TargetSpec = BasicHarvestTargetHarvestable
	return BasicHarvestCard
}

func (gs *GameState) fillEffectForCard(c *card.Card, owner *player.Unit) {
		switch c.Type {
		case card.PlayerMove:
			location,ok := c.Target.(*location.Unit)
			if !ok {
				//shouldnt happen
			}
			c.Effect = &moving.MoveEffect {
				Entity: owner,
				Target: location,
			}
		case card.PlayerAttack:
			target,ok := c.Target.(*creature.Unit)
			if !ok {
				//shouldnt happen
			}
			c.Effect = &combat.DamageEffect {
				Source: owner,
				Target: target,
			}
		case card.PlayerDistract:
			target,ok := c.Target.(*creature.Unit)
			if !ok {
				//shouldnt happen
			}
			c.Effect = &evading.Effect {
				EvadingPlayer: owner,
				EvadedCreature: target,
			}
		case card.PlayerHarvest:
			target,ok := c.Target.(*harvesting.Unit)
			if !ok {
				//shouldnt happen
			}
			c.Effect = &harvesting.Effect {
				Player: owner,
				TargetHarvestable: target,
			}
	} 
}

func (gs *GameState) InitCreatures() {

	Stag := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: gs.getRandomLocation()},
		HealthPool:   combat.HealthPool{CurrentHealth: 5, MaxHealth: 5},
		ID:           gs.NextID(),
		Name:         "Stag",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(true),
			CreatureMovementConfig:  creature.GenerateCuriousConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
		},
		Damage: 2,
	}

	log.Printf("[INIT] Spawned creature %s at %s", Stag.Name, Stag.CurrentLocation.Name)

	Racoon := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: gs.getRandomLocation()},
		HealthPool:   combat.HealthPool{CurrentHealth: 2, MaxHealth: 2},
		ID:           gs.NextID(),
		Name:         "Raccoon",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(false),
			CreatureMovementConfig:  creature.GenerateCuriousConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
		},
		Damage: 1,
	}

	log.Printf("[INIT] Spawned creature %s at %s", Racoon.Name, Racoon.CurrentLocation.Name)

	Plant := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: gs.getRandomLocation()},
		HealthPool:   combat.HealthPool{CurrentHealth: 2, MaxHealth: 2},
		ID:           gs.NextID(),
		Name:         "Suspicious Plant",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(true),
			CreatureMovementConfig:  creature.GenerateShyConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
		},
		Damage: 1,
	}

	log.Printf("[INIT] Spawned creature %s at %s", Plant.Name, Plant.CurrentLocation.Name)

	gs.Creatures = append(gs.Creatures, Stag, Racoon, Plant)
}


