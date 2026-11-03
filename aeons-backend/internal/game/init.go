package game

import (
	"aeons/internal/bonding"
	"aeons/internal/card"
	"aeons/internal/creature"
	"aeons/internal/distracting"
	"aeons/internal/hiding"
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
	gs.InitScenario()

	gs.KnownNodes = map[int]*collecting.Node{}
	gs.ReconcileVision()

	//init hands also sets playability of cards, so needs to be evoked after all creatures and nodes have been added to the board
	gs.initHandsForAllPlayers()
	gs.RefreshHands()
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
	basicHand = append(basicHand, gs.GenerateBasicMoveCard(), gs.GenerateBasicAttackCard(), gs.GenerateBasicDistractCard(), gs.GenerateBasicGatherCard(), gs.GenerateBasicBondCard(), gs.GenerateBasicHideCard())
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

func (gs *GameState) GenerateBasicHideCard() *card.PlayableCard {
	return &card.PlayableCard{
		ID:         gs.NextID(),
		Type:       card.PlayerHide,
		Name:       "Hiding",
		EffectText: "*Hide* at your location.",
		Flavor:     "Sometimes, just observing is wise.",
		Candidates: []card.Targetable{},
		TargetSpec: &card.TargetSpec{Type: card.TargetTypeLocation},
		Range:      1,
		Cost:       2,
	}
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
		case card.PlayerHide:
			c.Effect = &hiding.Effect {
				HidingPlayer: owner,
			}
		case card.PlayerAttack:
			target,ok := c.Target.(*creature.Unit)
			if !ok {
				//shouldnt happen
			}
			c.Effect = &confrontEffect {
				gs: gs,
				Player: owner,
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

// every kind of creature is in the grab bag three times, creaturesAtStart of them are drawn and placed one after another at random locations.
// A location that holds two creatures is skipped until every location holds one.
const creaturesAtStart = 8

func (gs *GameState) InitCreatures() {
	kinds := []func(*location.Unit) *creature.Unit{
		gs.NewWolfpack, gs.NewMoose, gs.NewStag, gs.NewRaccoon, gs.NewHare, gs.NewFox, gs.NewSuspiciousPlant, gs.NewWatchfulMycelium,
	}
	grabBag := slices.Repeat(kinds, 3)
	rand.Shuffle(len(grabBag), func(i, j int) {
		grabBag[i], grabBag[j] = grabBag[j], grabBag[i]
	})

	for _, newCreature := range grabBag[:creaturesAtStart] {
		candidates := gs.SpawnCandidates()
		gs.Creatures = append(gs.Creatures, newCreature(candidates[rand.IntN(len(candidates))]))
	}
}

func (gs *GameState) SpawnCandidates() []*location.Unit {
	everyLocationOccupied := !slices.ContainsFunc(gs.Locations, func(l *location.Unit) bool {
		return len(gs.EnemiesAtLocation(l)) == 0
	})
	if everyLocationOccupied {
		return gs.Locations
	}
	return slices.DeleteFunc(slices.Clone(gs.Locations), func(l *location.Unit) bool {
		return len(gs.EnemiesAtLocation(l)) >= 2
	})
}
