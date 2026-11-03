package game

import (
	"aeons/internal/card"
	"aeons/internal/collecting"
	"aeons/internal/creature"
	"aeons/internal/location"
	"slices"
	"testing"
)

// an empty board, the players stay at the Secluded Den
func newBoard() (*GameState, *location.Unit) {
	gs := newGame()
	gs.Creatures = []*creature.Unit{}
	return gs, gs.locationNamed("Windswept Plains")
}

func (gs *GameState) place(c *creature.Unit) *creature.Unit {
	gs.Creatures = append(gs.Creatures, c)
	return c
}

func TestHuntingFollowsSize(t *testing.T) {
	gs, plains := newBoard()
	wolfpack := gs.NewWolfpack(plains)
	fox := gs.NewFox(plains)
	moose := gs.NewMoose(plains)
	stag := gs.NewStag(plains)
	hare := gs.NewHare(plains)
	plant := gs.NewSuspiciousPlant(plains)
	fox4 := gs.NewHare(plains)
	fox4.Size = 4
	player := gs.Players[0]

	cases := []struct {
		name   string
		hunter *creature.Unit
		target creature.Quarry
		want   bool
	}{
		{"predator hunts smaller prey", wolfpack, hare, true},
		{"predator hunts prey of the same size", wolfpack, stag, true},
		{"predator does not hunt larger prey", wolfpack, moose, false},
		{"predator hunts players, who count as size 6", wolfpack, player, true},
		{"predator that is prey hunts smaller prey", fox, hare, true},
		{"predator that is prey does not hunt prey of the same size", fox, fox4, false},
		{"predator that is prey does not hunt larger prey", fox, stag, false},
		{"predator that is prey does not hunt larger players", fox, player, false},
		{"nobody hunts plants", wolfpack, plant, false},
		{"prey does not hunt", stag, hare, false},
	}
	for _, c := range cases {
		if got := c.hunter.CanHunt(c.target); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}

	hare.Hidden = true
	if wolfpack.CanHunt(hare) {
		t.Error("a hidden prey is hunted")
	}
}

func TestPickTargetPrefersLowestRemainingDistressThenSmallerSize(t *testing.T) {
	gs, plains := newBoard()
	hare, raccoon, stag := gs.NewHare(plains), gs.NewRaccoon(plains), gs.NewStag(plains)
	stag.Distress = 4 // 1 remaining, the hare has 2, the raccoon 3

	if got := pickTarget([]creature.Quarry{hare, raccoon, stag}); got != stag {
		t.Errorf("picked %v, want the stag with the least distress left", got)
	}
	hare.Distress = 1 // now tied with the stag at 1, the hare is smaller
	if got := pickTarget([]creature.Quarry{stag, raccoon, hare}); got != hare {
		t.Errorf("picked %v, want the smaller hare", got)
	}
}

func TestPlayersCountThreeDistressForTheHunt(t *testing.T) {
	gs, plains := newBoard()
	druid := gs.Players[0]
	druid.Wounds = 2 // 1 remaining
	hare := gs.NewHare(plains)
	hare.Distress = 0 // 2 remaining

	if got := pickTarget([]creature.Quarry{hare, druid}); got != druid {
		t.Errorf("picked %v, want the player", got)
	}
}

func TestPreyAtItsBurrowHidesWhenAHunterArrives(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(plains))
	wolfpack := gs.place(gs.NewWolfpack(plains))

	events := gs.ReactToArrival(wolfpack)
	if !hare.Hidden || hare.Exhausted || len(events) != 0 {
		t.Errorf("hidden = %v, exhausted = %v, events = %d, want hidden and ready", hare.Hidden, hare.Exhausted, len(events))
	}
}

func TestPreyAwayFromItsBurrowRunsTowardsItAndExhausts(t *testing.T) {
	gs, plains := newBoard()
	lake := gs.locationNamed("Redhorn Lake")
	raccoon := gs.place(gs.NewRaccoon(lake))
	raccoon.CurrentLocation = plains
	wolfpack := gs.place(gs.NewWolfpack(plains))

	gs.ReactToArrival(wolfpack)
	if raccoon.CurrentLocation != lake || !raccoon.Exhausted || raccoon.Distress != 0 {
		t.Errorf("at %s, exhausted = %v, distress = %d", raccoon.CurrentLocation.Name, raccoon.Exhausted, raccoon.Distress)
	}
}

func TestPreyWithoutBurrowHopsToANeighbourWithoutHunters(t *testing.T) {
	gs, plains := newBoard()
	stag := gs.place(gs.NewStag(plains))
	wolfpack := gs.place(gs.NewWolfpack(plains))
	grove := gs.locationNamed("Coniferous Grove")
	gs.place(gs.NewWolfpack(grove))

	gs.ReactToArrival(wolfpack)
	if stag.CurrentLocation == plains || stag.CurrentLocation == grove || !stag.Exhausted {
		t.Errorf("stag at %s, exhausted = %v", stag.CurrentLocation.Name, stag.Exhausted)
	}
}

func TestPreyDoesNotRunFromLargerPreyOrReadyHuntersItIsTooBigFor(t *testing.T) {
	gs, plains := newBoard()
	moose := gs.place(gs.NewMoose(plains))
	wolfpack := gs.place(gs.NewWolfpack(plains))

	gs.ReactToArrival(wolfpack)
	if moose.CurrentLocation != plains || moose.Exhausted {
		t.Error("a moose runs from a wolfpack that cannot hunt it")
	}
}

func TestExhaustedPreyDoesNotRun(t *testing.T) {
	gs, plains := newBoard()
	stag := gs.place(gs.NewStag(plains))
	stag.Exhausted = true
	wolfpack := gs.place(gs.NewWolfpack(plains))

	gs.ReactToArrival(wolfpack)
	if stag.CurrentLocation != plains {
		t.Error("an exhausted prey ran away")
	}
}

func TestPredatorAttacksTwiceAndStaysReady(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(plains))
	hare.Hidden = false
	hare.CanHide = false
	wolfpack := gs.place(gs.NewWolfpack(plains))

	gs.ActivatePredator(wolfpack)
	gs.ReconcileDefeats()

	if slices.Contains(gs.Creatures, hare) {
		t.Error("the hare survived two attacks")
	}
	if wolfpack.Exhausted {
		t.Error("the wolfpack exhausted itself")
	}
	corpse := slices.IndexFunc(gs.Nodes, func(n *collecting.Node) bool { return n.Name == "Hare corpse" })
	if corpse == -1 || gs.Nodes[corpse].StaticLocation != plains {
		t.Error("no corpse was left at the hare's location")
	}
}

func TestExhaustedPredatorOnlyReadies(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(plains))
	hare.CanHide = false
	wolfpack := gs.place(gs.NewWolfpack(plains))
	wolfpack.Exhausted = true

	gs.ActivatePredator(wolfpack)
	if wolfpack.Exhausted || hare.Distress != 0 {
		t.Errorf("exhausted = %v, hare distress = %d, want it ready and nothing attacked", wolfpack.Exhausted, hare.Distress)
	}
}

func TestPredatorWithoutTargetMovesAwayFromWhereItWasAndLocksOnToPrey(t *testing.T) {
	gs, plains := newBoard()
	lake := gs.locationNamed("Redhorn Lake")
	grove := gs.locationNamed("Coniferous Grove")
	hare := gs.place(gs.NewHare(grove))
	hare.CanHide = false
	wolfpack := gs.place(gs.NewWolfpack(lake))
	wolfpack.LastLocation = plains

	gs.huntMove(wolfpack)
	if wolfpack.CurrentLocation != grove || wolfpack.Chasing != hare {
		t.Errorf("wolfpack at %s, chasing %v, want it at the grove locked on to the hare", wolfpack.CurrentLocation.Name, wolfpack.Chasing)
	}
}

func TestChaseEndsWhenThePreyHidesOrGetsTwoMovesAway(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(gs.locationNamed("Redhorn Lake")))
	wolfpack := gs.place(gs.NewWolfpack(plains))
	wolfpack.Chasing = hare

	if !stillChasing(wolfpack) {
		t.Error("the chase ended with the prey one move away")
	}
	hare.CurrentLocation = gs.locationNamed("Slumbering Crag")
	if stillChasing(wolfpack) {
		t.Error("the chase goes on with the prey two moves away")
	}
	hare.CurrentLocation = gs.locationNamed("Redhorn Lake")
	hare.Hidden = true
	if stillChasing(wolfpack) {
		t.Error("the chase goes on with the prey hidden")
	}
}

func TestHiddenPreyRelievesDistressAndComesOutOnceNoPredatorIsAround(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(plains))
	hare.Hidden = true
	hare.Distress = 1
	wolfpack := gs.place(gs.NewWolfpack(plains))

	gs.ActivatePrey(hare)
	gs.ReconcileHiding()
	if !hare.Hidden || hare.Distress != 0 {
		t.Errorf("hidden = %v, distress = %d, want hidden with the distress relieved while a predator is there", hare.Hidden, hare.Distress)
	}

	wolfpack.CurrentLocation = gs.locationNamed("Redhorn Lake")
	gs.ReconcileHiding()
	if hare.Hidden {
		t.Error("the hare stays hidden without distress and without a predator")
	}
}

func TestHiddenPreyWithDistressStaysHiddenWithoutPredator(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(plains))
	hare.Hidden = true
	hare.Distress = 1

	gs.ReconcileHiding()
	if !hare.Hidden {
		t.Error("the hare came out while it is still distressed")
	}
}

func TestDistressedPreyHeadsHomeAndHides(t *testing.T) {
	gs, plains := newBoard()
	lake := gs.locationNamed("Redhorn Lake")
	hare := gs.place(gs.NewHare(lake))
	hare.CurrentLocation = gs.locationNamed("Slumbering Crag") // Crag - Plateau - Plains - Lake
	hare.Distress = 1

	gs.ActivatePrey(hare)
	if hare.CurrentLocation != gs.locationNamed("Arid Plateau") && hare.CurrentLocation != plains {
		t.Errorf("hare at %s", hare.CurrentLocation.Name)
	}
	gs.ActivatePrey(hare)
	gs.ActivatePrey(hare)
	if hare.CurrentLocation != lake || !hare.Hidden {
		t.Errorf("hare at %s, hidden = %v, want hidden at its burrow", hare.CurrentLocation.Name, hare.Hidden)
	}
}

func TestExhaustedPreyReadiesThenMoves(t *testing.T) {
	gs, plains := newBoard()
	stag := gs.place(gs.NewStag(plains))
	stag.Exhausted = true

	events := gs.ActivatePrey(stag)
	if stag.Exhausted || len(events) != 2 {
		t.Errorf("exhausted = %v, events = %d, want ready and one move", stag.Exhausted, len(events))
	}
}

func TestCalmPreyStaysWithinItsRadius(t *testing.T) {
	for range 50 {
		gs, plains := newBoard()
		raccoon := gs.place(gs.NewRaccoon(plains))
		for range 10 {
			gs.ActivatePrey(raccoon)
			if plains.StepsTo(raccoon.CurrentLocation) > raccoon.Radius {
				t.Fatalf("raccoon wandered to %s", raccoon.CurrentLocation.Name)
			}
		}
	}
}

func TestConfrontDistressesAndSendsPreyFleeingWithoutExhausting(t *testing.T) {
	gs, plains := newBoard()
	stag := gs.place(gs.NewStag(plains))
	confront := &confrontEffect{gs: gs, Player: gs.Players[0], Target: stag}

	confront.Apply()
	if stag.Distress != 1 || stag.CurrentLocation == plains || stag.Exhausted {
		t.Errorf("distress = %d, at %s, exhausted = %v", stag.Distress, stag.CurrentLocation.Name, stag.Exhausted)
	}
}

func TestConfrontOnlyTargetsPreyAndDistractOnlyPredators(t *testing.T) {
	gs, plains := newBoard()
	wolfpack, stag, fox := gs.place(gs.NewWolfpack(plains)), gs.place(gs.NewStag(plains)), gs.place(gs.NewFox(plains))
	plant := gs.place(gs.NewSuspiciousPlant(plains))
	hare := gs.place(gs.NewHare(plains))
	hare.Hidden = true
	gs.Players[0].CurrentLocation = plains
	gs.RefreshHands()

	candidatesOf := func(kind card.EffectType) []card.Targetable {
		return cardOfType(gs.Players[0].CardsInHand, kind).Candidates
	}
	cases := []struct {
		name string
		kind card.EffectType
		want []*creature.Unit
	}{
		{"confront", card.PlayerAttack, []*creature.Unit{stag, fox}},
		{"distract", card.PlayerDistract, []*creature.Unit{wolfpack, fox}},
		{"bond", card.PlayerBond, []*creature.Unit{stag, plant}},
	}
	for _, c := range cases {
		got := candidatesOf(c.kind)
		if len(got) != len(c.want) {
			t.Errorf("%s targets %d creatures, want %d", c.name, len(got), len(c.want))
			continue
		}
		for _, want := range c.want {
			if !slices.Contains(got, card.Targetable(want)) {
				t.Errorf("%s cannot target %s", c.name, want.Name)
			}
		}
	}
}

func TestPlayCardRejectsHiddenAndWrongKindOfTarget(t *testing.T) {
	gs, plains := newBoard()
	wolfpack := gs.place(gs.NewWolfpack(plains))
	gs.place(gs.NewStag(plains))
	gs.Players[0].CurrentLocation = plains
	gs.Phase = PhaseSpirit
	gs.RefreshHands()

	confront := cardOfType(gs.Players[0].CardsInHand, card.PlayerAttack)
	if _, err := gs.PlayCard(confront.ID, wolfpack.ID, ""); err == nil {
		t.Error("a predator was confronted")
	}
}

func TestCalmPreyMovesOnceAndDistressedPreyUpToTwice(t *testing.T) {
	gs, plains := newBoard()
	calm := gs.place(gs.NewStag(plains))
	if events := gs.ActivatePrey(calm); len(events) != 1 {
		t.Errorf("calm prey moved %d times, want 1", len(events))
	}
	distressed := gs.place(gs.NewStag(plains))
	distressed.Distress = 1
	if events := gs.ActivatePrey(distressed); len(events) != 2 {
		t.Errorf("distressed prey moved %d times, want 2", len(events))
	}
}

func TestPreyThatIsNoPredatorWandersTowardsPlants(t *testing.T) {
	for range 30 {
		gs, plains := newBoard()
		grove := gs.locationNamed("Coniferous Grove")
		plant := gs.place(gs.NewWatchfulMycelium(grove))
		plant.Bonded = true
		stag := gs.place(gs.NewStag(plains))

		gs.ActivatePrey(stag)
		if stag.CurrentLocation != grove {
			t.Fatalf("stag went to %s, want the grove with the plant", stag.CurrentLocation.Name)
		}
	}
}

func TestBondedCreaturesAndPlantsDoNotActInTheWorldPhase(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(plains))
	hare.Bonded, hare.BondedTo = true, gs.Players[0]
	plant := gs.place(gs.NewSuspiciousPlant(plains))

	for range 20 {
		gs.ResolveWorldPhase()
	}
	if hare.CurrentLocation != plains || plant.CurrentLocation != plains {
		t.Error("a bonded creature or a plant moved")
	}
}

func TestBondedHunterScaresPreyButBondedPreyIsStillHunted(t *testing.T) {
	gs, plains := newBoard()
	stag := gs.place(gs.NewStag(plains))
	kindredWolfpack := gs.place(gs.NewWolfpack(plains))
	kindredWolfpack.Bonded, kindredWolfpack.BondedTo = true, gs.Players[0]
	gs.ReactToArrival(kindredWolfpack)
	if stag.CurrentLocation == plains {
		t.Error("prey stayed with a bonded hunter")
	}

	bondedStag := gs.place(gs.NewStag(plains))
	bondedStag.Bonded, bondedStag.BondedTo = true, gs.Players[1]
	gs.ReactToArrival(kindredWolfpack)
	if bondedStag.CurrentLocation != plains || !kindredWolfpack.CanHunt(bondedStag) {
		t.Error("a bonded prey ran, or is not hunted")
	}
}

func TestSpawningDrawsEightCreaturesAndNeverCrowdsALocationWhileOthersAreEmpty(t *testing.T) {
	for range 50 {
		gs := newGame()
		if len(gs.Creatures) != 8 {
			t.Fatalf("%d creatures spawned, want 8", len(gs.Creatures))
		}
		for _, l := range gs.Locations {
			if len(gs.EnemiesAtLocation(l)) > 2 {
				t.Fatalf("%s holds %d creatures", l.Name, len(gs.EnemiesAtLocation(l)))
			}
		}
	}
}

func TestIdentNamesFindEveryHare(t *testing.T) {
	gs, plains := newBoard()
	tim := gs.NewHare(plains)
	tim.Name = "Tim"
	if !tim.HasIdentName("hare") || tim.HasIdentName("Fox") {
		t.Error("ident names do not identify Tim as a hare")
	}
}

func TestPreyInADeadEndCanLeaveEvenToWhereItCameFrom(t *testing.T) {
	gs, _ := newBoard()
	den := gs.locationNamed("Secluded Den")
	plains := gs.locationNamed("Windswept Plains")
	stag := gs.place(gs.NewStag(den))
	stag.LastLocation = plains

	gs.ActivatePrey(stag)
	if stag.CurrentLocation != plains {
		t.Errorf("stag at %s, want it back at the plains", stag.CurrentLocation.Name)
	}
}

func TestPreyOnAPlantLocationStillLeaves(t *testing.T) {
	gs, plains := newBoard()
	grove := gs.locationNamed("Coniferous Grove")
	gs.place(gs.NewWatchfulMycelium(grove))
	stag := gs.place(gs.NewStag(plains))

	gs.ActivatePrey(stag)
	if stag.CurrentLocation != grove {
		t.Fatalf("stag at %s, want the plant location", stag.CurrentLocation.Name)
	}
	for range 5 {
		before := stag.CurrentLocation
		gs.ActivatePrey(stag)
		if stag.CurrentLocation == before || stag.CurrentLocation == stag.LastLocation {
			t.Fatalf("stag stayed or went straight back: %s", stag.CurrentLocation.Name)
		}
	}
}

func TestPreyHeldInHidingByAPredatorMovesOnceAfterThePredatorLeft(t *testing.T) {
	for range 30 {
		gs, plains := newBoard()
		hare := gs.place(gs.NewHare(plains))
		hare.Hide(true)
		wolfpack := gs.place(gs.NewWolfpack(plains))
		wolfpack.LastLocation = gs.locationNamed("Secluded Den")

		gs.ResolveWorldPhase()
		if hare.Hidden || hare.CurrentLocation == plains {
			t.Fatalf("hare hidden = %v, at %s, want it out and moved on", hare.Hidden, hare.CurrentLocation.Name)
		}
	}
}

func TestPreyStaysHiddenWhileThePredatorStays(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(plains))
	hare.Hide(true)
	wolfpack := gs.place(gs.NewWolfpack(plains))
	wolfpack.Exhausted = true

	gs.ResolveWorldPhase()
	if !hare.Hidden || hare.CurrentLocation != plains {
		t.Errorf("hare hidden = %v, at %s, want it hidden at the plains", hare.Hidden, hare.CurrentLocation.Name)
	}
}

func TestOnlyPredatorsOfSizeSixOrMoreBlockMovement(t *testing.T) {
	gs, plains := newBoard()
	fox := gs.place(gs.NewFox(plains))
	if gs.LocationHasBlockingCreature(plains) {
		t.Error("a fox blocks movement")
	}
	wolfpack := gs.place(gs.NewWolfpack(plains))
	if !gs.LocationHasBlockingCreature(plains) {
		t.Error("a wolfpack does not block movement")
	}
	wolfpack.Exhausted = true
	fox.Size = 6
	if !gs.LocationHasBlockingCreature(plains) {
		t.Error("a predator of size 6 does not block movement")
	}
}

func TestCreaturesUnexhaustWhenTheBondIsComplete(t *testing.T) {
	gs, plains := newBoard()
	wolfpack := gs.place(gs.NewWolfpack(plains))
	wolfpack.Exhausted = true

	for range wolfpack.BondThreshold - 1 {
		wolfpack.Bond(gs.Players[0])
		if !wolfpack.Exhausted {
			t.Fatal("the wolfpack unexhausted before the bond was complete")
		}
	}
	wolfpack.Bond(gs.Players[0])
	if !wolfpack.Bonded || wolfpack.Exhausted {
		t.Errorf("bonded = %v, exhausted = %v, want bonded and ready", wolfpack.Bonded, wolfpack.Exhausted)
	}
}

func TestConfrontedPreyAtItsBurrowHidesBecauseItIsDistressed(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(plains))
	(&confrontEffect{gs: gs, Player: gs.Players[0], Target: hare}).Apply()
	if !hare.Hidden || hare.Distress != 1 || hare.CurrentLocation != plains || hare.Exhausted {
		t.Errorf("hidden = %v, distress = %d, at %s, exhausted = %v, want it hidden and ready at its burrow",
			hare.Hidden, hare.Distress, hare.CurrentLocation.Name, hare.Exhausted)
	}
}

func TestCalmPreyDoesNotHideWithAPlayerAround(t *testing.T) {
	gs, plains := newBoard()
	hare := gs.place(gs.NewHare(plains))
	gs.Players[0].CurrentLocation = plains

	gs.ActivatePrey(hare)
	gs.ReconcileHiding()
	if hare.Hidden {
		t.Error("a calm hare hid because a player is there")
	}
}

func TestPreyOnlyHidesOnItsOwnTurnWhileDistressed(t *testing.T) {
	gs, plains := newBoard()
	calm := gs.place(gs.NewHare(plains))
	gs.ActivatePrey(calm)
	if calm.Hidden {
		t.Error("a calm hare hid on its own turn")
	}

	distressed := gs.place(gs.NewHare(plains))
	distressed.Distress = 1
	gs.ActivatePrey(distressed)
	if !distressed.Hidden {
		t.Error("a distressed hare at its burrow did not hide")
	}
}

func TestExhaustedPreyDoesNotStepAwayWhenConfronted(t *testing.T) {
	gs, plains := newBoard()
	stag := gs.place(gs.NewStag(plains))
	stag.Exhausted = true
	(&confrontEffect{gs: gs, Player: gs.Players[0], Target: stag}).Apply()
	if stag.CurrentLocation != plains || stag.Distress != 1 {
		t.Errorf("stag at %s with distress %d, want it to stay and take the distress", stag.CurrentLocation.Name, stag.Distress)
	}
}

func TestExhaustedPreyReadiesInAnEventBeforeItMoves(t *testing.T) {
	gs, plains := newBoard()
	stag := gs.place(gs.NewStag(plains))
	stag.Exhausted = true

	events := gs.ActivatePrey(stag)
	if len(events) != 2 || !events[0].Ready || events[1].FromLocationID == events[1].ToLocationID {
		t.Errorf("events = %+v, want a ready event and then a move", events)
	}
}

func TestCreaturesDoNotMoveAgainInThePhaseTheyExhaustedIn(t *testing.T) {
	for range 200 {
		gs, plains := newBoard()
		for _, newCreature := range []func(*location.Unit) *creature.Unit{gs.NewWolfpack, gs.NewWolfpack, gs.NewStag, gs.NewRaccoon, gs.NewHare, gs.NewFox, gs.NewFox} {
			gs.place(newCreature(plains))
		}
		exhausted := map[int]bool{}
		for _, event := range gs.ResolveWorldPhase() {
			if exhausted[event.CreatureID] && event.FromLocationID != event.ToLocationID {
				t.Fatalf("creature %d moved after it exhausted", event.CreatureID)
			}
			exhausted[event.CreatureID] = exhausted[event.CreatureID] || event.Exhausted
		}
	}
}
