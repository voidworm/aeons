package game

import (
	"aeons/internal/bonding"
	"aeons/internal/card"
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/distracting"
	"aeons/internal/collecting"
	"aeons/internal/location"
	"aeons/internal/moving"
	"aeons/internal/player"
	"log"
	"math/rand/v2"
	"slices"
)

func (gs *GameState) Init() {
	gs.Phase = PhaseSpirit
	gs.Cycle = 1

	gs.InitLocations()
	gs.InitPlayers()
	gs.InitNodes()
	gs.InitCreatures()

	gs.KnownNodes = map[int]*collecting.Node{}
	gs.ReconcileVision()

	//init hands also sets playability of cards, so needs to be evoked after all creatures and nodes have been added to the board
	gs.initHandsForAllPlayers()
}

func (gs *GameState) InitLocations() {

	log.Println("[INIT] Setting up locations....")
	SecludedDen := &location.Unit{ID: gs.NextID(), Name: "Secluded Den", Flavor: "The den has been your home for many years."}
	WindsweptPlains := &location.Unit{ID: gs.NextID(), Name: "Windswept Plains", Flavor: "The plains lie calm, the grasses and weeds wail in the wind."}
	RedhornLake := &location.Unit{ID: gs.NextID(), Name: "Redhorn Lake", Flavor: "When the moon rises, the lake glimmers in an ominous red."}
	ConiferousGrove := &location.Unit{ID: gs.NextID(), Name: "Coniferous Grove", Flavor: "The smell of resin and needles swirls through the shadowy grove."}
	AridPlateau := &location.Unit{ID: gs.NextID(), Name: "Arid Plateau", Flavor: "Climb the plateau to gain an excellent view over what you call home."}
	SlumberingCrag := &location.Unit{ID: gs.NextID(), Name: "Slumbering Crag", Flavor: "Those who commune with the earth element find peace here."}
	HermitsRecluse := &location.Unit{ID: gs.NextID(), Name: "Hermit's Recluse", Flavor: "The blind woman known as the Hermit calls this little cave her home."}

	SecludedDen.OutgoingConnections = append(SecludedDen.OutgoingConnections, WindsweptPlains)
	WindsweptPlains.OutgoingConnections = append(WindsweptPlains.OutgoingConnections, SecludedDen, RedhornLake, ConiferousGrove, AridPlateau)
	RedhornLake.OutgoingConnections = append(RedhornLake.OutgoingConnections, WindsweptPlains, ConiferousGrove)
	ConiferousGrove.OutgoingConnections = append(ConiferousGrove.OutgoingConnections, WindsweptPlains, RedhornLake)
	AridPlateau.OutgoingConnections = append(AridPlateau.OutgoingConnections, WindsweptPlains, SlumberingCrag, HermitsRecluse)
	SlumberingCrag.OutgoingConnections = append(SlumberingCrag.OutgoingConnections, AridPlateau, SecludedDen)
	HermitsRecluse.OutgoingConnections = append(HermitsRecluse.OutgoingConnections, AridPlateau)

	gs.Locations = append(gs.Locations, SecludedDen, WindsweptPlains, RedhornLake, ConiferousGrove, AridPlateau, SlumberingCrag, HermitsRecluse)
}

func (gs *GameState) InitNodes() {
	grabBag := append([]*location.Unit(nil), gs.Locations...)
	for range 5 {
		r := rand.IntN(len(grabBag))
		randomLocation := grabBag[r]
		grabBag = slices.Delete(grabBag, r, r+1)

		n := collecting.GenerateRandomNode(gs.NextID(), randomLocation)
		gs.Nodes = append(gs.Nodes, n)
		log.Printf("[INIT] Spawning node %s at %s\n", n.Name, randomLocation.Name)
	}
}

func (gs *GameState) InitPlayers() {
	log.Println("[INIT] Setting up player Druid...")
	Druid := &player.Unit{
		MovingEntity:       moving.MovingEntity{CurrentLocation: gs.Locations[0]},
		ID:                 gs.NextID(),
		Name:               "Druid",
		CardsInHand:        []*card.PlayableCard{},
		ResourcesAvailable: 0,
		RemainingActions:   3,
		Energy:             AwakeningEnergy,
	}

	log.Println("[INIT] Setting up player Scoundrel...")
	Scoundrel := &player.Unit{
		MovingEntity:       moving.MovingEntity{CurrentLocation: gs.Locations[0]},
		ID:                 gs.NextID(),
		Name:               "Scoundrel",
		CardsInHand:        []*card.PlayableCard{},
		ResourcesAvailable: 0,
		RemainingActions:   3,
		Energy:             AwakeningEnergy,
	}
	gs.Players = append(gs.Players, Druid, Scoundrel)
}

func (gs *GameState)initHandsForAllPlayers() {
	for _,p := range gs.Players {
		p.CardsInHand = gs.GenerateBasicHand()
		gs.EvaluatePlayerHandPlayability(p)
	}
}

func (gs *GameState) GenerateBasicHand() []*card.PlayableCard {
	basicHand := []*card.PlayableCard{}
	basicHand = append(basicHand, gs.GenerateBasicMoveCard(), gs.GenerateBasicAttackCard(), gs.GenerateBasicDistractCard(), gs.GenerateBasicGatherCard(), gs.GenerateBasicBondCard())
	gs.PlayableCards = append(gs.PlayableCards, basicHand...)
	return basicHand
}

func (gs *GameState) GenerateBasicMoveCard() *card.PlayableCard {
	BasicMoveCard := &card.PlayableCard {
		ID: gs.NextID(),
		Type: card.PlayerMove,
		Name: "Move",
		EffectText: "*Move* your character to target adjacent location.",
		Flavor: "Mundo goes where he pleases.",
		TargetSpec: nil,
		Candidates: []card.Targetable{},
		Effect: nil,
		Range: 1,
		Cost: 1,
	}

	BasicMoveTargetLocationSpec := &card.TargetSpec {
		Type: card.TargetTypeLocation,
		Target: nil,
		Filter: nil,
	}

	BasicMoveCard.TargetSpec = BasicMoveTargetLocationSpec
	return BasicMoveCard
}

func (gs *GameState) GenerateBasicAttackCard() *card.PlayableCard {
	BasicAttackCard := &card.PlayableCard {
		ID: gs.NextID(),
		Name: "Confront",
		Type: card.PlayerAttack,
		EffectText: "*Confront* a creature at your location.",
		Flavor: "We do what we must.",
		TargetSpec: nil,
		Candidates: []card.Targetable{},
		Effect: nil,
		Range: 1,
		Cost: 1,
	}

	BasicAttackTargetCreatureSpec := &card.TargetSpec {
		Type: card.TargetTypeCreature,
		Target: nil,
		Filter: nil,
	}

	BasicAttackCard.TargetSpec = BasicAttackTargetCreatureSpec
	return BasicAttackCard
}

func (gs *GameState) GenerateBasicDistractCard() *card.PlayableCard {
	BasicDistractCard := &card.PlayableCard {
		ID: gs.NextID(),
		Name: "Distract",
		EffectText: "*Distract* a target creature at your location.",
		Flavor: "Look, over there!",
		Type:	card.PlayerDistract,
		TargetSpec: nil,
		Candidates: []card.Targetable{},
		Effect: nil,
		Range: 1,
		Cost: 1,
	}

	BasicDistractTargetCreatureSpec := &card.TargetSpec {
		Type: card.TargetTypeCreature,
		Target: nil,
		Filter: nil,
	}

	BasicDistractCard.TargetSpec = BasicDistractTargetCreatureSpec
	return BasicDistractCard
}

func (gs *GameState) GenerateBasicGatherCard() *card.PlayableCard {
	BasicGatherCard := &card.PlayableCard {
		ID: gs.NextID(),
		Type: card.PlayerGather,
		Name: "Gather",
		EffectText: "*Harvest* or *Forage* at your location",
		Flavor: "Work work!",
		TargetSpec: nil,
		Candidates: []card.Targetable{},
		Effect: nil,
		Range: 1,
		Cost: 1,
	}

	BasicGatherTargetNode := &card.TargetSpec {
		Type: card.TargetTypeNode,
		Target: nil,
		Filter: nil,
	}

	BasicGatherCard.TargetSpec = BasicGatherTargetNode
	return BasicGatherCard
}

func (gs *GameState) GenerateBasicBondCard() *card.PlayableCard {
	BasicBondCard := &card.PlayableCard {
		ID: gs.NextID(),
		Type: card.PlayerBond,
		Name: "Bond",
		EffectText: "*Bond* with target bondable creature at your location.",
		Flavor: "Well, hello there!",
		TargetSpec: nil,
		Candidates: []card.Targetable{},
		Effect: nil,
		Range: 1,
		Cost: 1,
	}

	BasicBondTargetCreatureSpec := &card.TargetSpec {
		Type: card.TargetTypeCreature,
		Target: nil,
		Filter: nil,
	}

	BasicBondCard.TargetSpec = BasicBondTargetCreatureSpec
	return BasicBondCard
}

func (gs *GameState) GenerateKindredMoveCard() *card.PlayableCard {
	return &card.PlayableCard{
		ID:         gs.NextID(),
		Type:       card.PlayerMoveKindred,
		Name:       "Kindred Move",
		EffectText: "Cooldown\n*Move* your Kindred to target adjacent location.",
		Flavor:     "It's time to go, my friend.",
		TargetSpec: &card.TargetSpec{Type: card.TargetTypeLocation},
		Candidates: []card.Targetable{},
		Range:      1,
		Cost:       0,
	}
}

func (gs *GameState) fillEffectForCard(c *card.PlayableCard, owner *player.Unit, mode collecting.Mode) {
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
		case card.PlayerMoveKindred:
			c.Effect = &moving.MoveEffect {
				Entity: gs.KindredOf(owner),
				Target: c.Target.(*location.Unit),
			}
		case card.PlayerAttack:
			target,ok := c.Target.(*creature.Unit)
			if !ok {
				//shouldnt happen
			}
			c.Effect = &creature.PlayerAttackEffect {
				Target: target,
			}
		case card.PlayerDistract:
			target,ok := c.Target.(*creature.Unit)
			if !ok {
				//shouldnt happen
			}
			c.Effect = &distracting.Effect {
				DistractingPlayer: owner,
				DistractedCreature: target,
			}
		case card.PlayerBond:
			target,ok := c.Target.(*creature.Unit)
			if !ok {
				//shouldnt happen
			}
			c.Effect = &bonding.Effect {
				BondingPlayer: owner,
				BondedCreature: target,
			}
		case card.PlayerGather:
			target,ok := c.Target.(*collecting.Node)
			if !ok {
				//shouldnt happen
			}
			c.Effect = &collecting.Effect {
				Player: owner,
				Node: target,
				Mode: mode,
			}
	} 
}

func (gs *GameState) InitCreatures() {
	freeLocations := slices.Clone(gs.Locations)
	rand.Shuffle(len(freeLocations), func(i, j int) {
		freeLocations[i], freeLocations[j] = freeLocations[j], freeLocations[i]
	})

	gs.Creatures = append(gs.Creatures, gs.NewWolf(freeLocations[0]), gs.NewStag(freeLocations[1]))
	for _, l := range freeLocations[2:5] {
		gs.Creatures = append(gs.Creatures, gs.NewRacoonOrPlant(l))
	}

	rand.Shuffle(len(gs.Creatures), func(i, j int) {
		gs.Creatures[i], gs.Creatures[j] = gs.Creatures[j], gs.Creatures[i]
	})
}

func (gs *GameState) NewRacoonOrPlant(at *location.Unit) *creature.Unit {
	if rand.IntN(2) == 0 {
		return gs.NewRacoon(at)
	}
	return gs.NewPlant(at)
}

func (gs *GameState) NewStag(at *location.Unit) *creature.Unit {
	stag := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: at},
		HealthPool:   combat.HealthPool{CurrentHealth: 5, MaxHealth: 5},
		ID:           gs.NextID(),
		Name:         "Stag",
		Flavor:       "A large stag that wanders the area aimlessly.",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(true),
			CreatureMovementConfig:  creature.GenerateCuriousShyConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
			CreatureBondConfig:      creature.GenerateBondConfig(3),
			CreatureHuntingConfig:   creature.GenerateHuntingConfig(false, true),
		},
	}

	log.Printf("[INIT] Spawned creature %s at %s", stag.NameWithID(), stag.CurrentLocation.Name)
	return stag
}

func (gs *GameState) NewRacoon(at *location.Unit) *creature.Unit {
	racoon := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: at},
		HealthPool:   combat.HealthPool{CurrentHealth: 2, MaxHealth: 2},
		ID:           gs.NextID(),
		Name:         "Raccoon",
		Flavor:       "A nifty little raccoon, known to steal berries. Especially yours.",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(true),
			CreatureMovementConfig:  creature.GenerateCuriousConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
			CreatureBondConfig:      creature.GenerateBondConfig(2),
			CreatureHuntingConfig:   creature.GenerateHuntingConfig(false, true),
		},
	}

	log.Printf("[INIT] Spawned creature %s at %s", racoon.NameWithID(), racoon.CurrentLocation.Name)
	return racoon
}

func (gs *GameState) NewPlant(at *location.Unit) *creature.Unit {
	plant := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: at},
		HealthPool:   combat.HealthPool{CurrentHealth: 2, MaxHealth: 2},
		ID:           gs.NextID(),
		Name:         "Suspicious Plant",
		Flavor:       "Hmmmmm...",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(true),
			CreatureMovementConfig:  creature.GenerateShyConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
			CreatureBondConfig:      creature.GenerateBondConfig(2),
			CreatureHuntingConfig:   creature.GenerateHuntingConfig(false, true),
		},
	}

	log.Printf("[INIT] Spawned creature %s at %s", plant.NameWithID(), plant.CurrentLocation.Name)
	return plant
}

func (gs *GameState) NewWolf(at *location.Unit) *creature.Unit {
	wolf := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: at},
		HealthPool:   combat.HealthPool{CurrentHealth: 4, MaxHealth: 4},
		ID:           gs.NextID(),
		Name:         "Wolf",
		Flavor:       "Prowling, stalking, these woods belong to the pack...",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(false),
			CreatureMovementConfig:  creature.GenerateCuriousConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
			CreatureBondConfig:      creature.GenerateBondConfig(3),
			CreatureHuntingConfig:   creature.GenerateHuntingConfig(true, false),
		},
	}

	log.Printf("[INIT] Spawned creature %s at %s", wolf.NameWithID(), wolf.CurrentLocation.Name)
	return wolf
}


