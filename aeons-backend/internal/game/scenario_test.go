package game

import (
	"aeons/internal/card"
	"aeons/internal/creature"
	"aeons/internal/quest"
	"testing"
)

func activeQuestNames(gs *GameState) []string {
	names := []string{}
	for _, q := range gs.ActiveQuests() {
		names = append(names, q.Name)
	}
	return names
}

func questNamed(t *testing.T, gs *GameState, name string) *quest.Quest {
	t.Helper()
	for _, q := range gs.ActiveQuests() {
		if q.Name == name {
			return q
		}
	}
	t.Fatalf("%s is not active, active quests: %v", name, activeQuestNames(gs))
	return nil
}

func finishCardOf(gs *GameState, playerIndex int) *card.PlayableCard {
	return cardOfType(gs.Players[playerIndex].CardsInHand, card.PlayerFinishQuest)
}

func playFinish(t *testing.T, gs *GameState, q *quest.Quest) {
	t.Helper()
	finish := finishCardOf(gs, 0)
	if finish == nil || !finish.CanBeCast {
		t.Fatalf("Finish Quest cannot be cast, finishable: %d", len(gs.FinishableQuests()))
	}
	if _, err := gs.PlayCard(finish.ID, q.ID, ""); err != nil {
		t.Fatal(err)
	}
}

func bondKindredToAllPlayers(gs *GameState) {
	for _, p := range gs.Players {
		c := gs.NewStag(p.CurrentLocation)
		c.Bonded, c.BondedTo = true, p
		gs.Creatures = append(gs.Creatures, c)
	}
	gs.RefreshHands()
}

func TestScenarioStartsWithBeyondTheBody(t *testing.T) {
	gs := newGame()
	if got := activeQuestNames(gs); len(got) != 1 || got[0] != "Beyond the Body" {
		t.Fatalf("active quests = %v", got)
	}
	if finishCardOf(gs, 0) != nil {
		t.Fatal("Finish Quest is in hand although nothing can be finished")
	}
}

func TestFinishingStartsTheNextQuestAndMandatoryBlocksEverythingElse(t *testing.T) {
	gs := newGame()
	bondKindredToAllPlayers(gs)

	first := questNamed(t, gs, "Beyond the Body")
	if !first.Finishable(gs.QuestWorld()) || finishCardOf(gs, 1) == nil {
		t.Fatal("Beyond the Body should be finishable and both players hold Finish Quest")
	}
	playFinish(t, gs, first)

	second := questNamed(t, gs, "Open the Eye")
	if finishCardOf(gs, 0) != nil {
		t.Fatal("Finish Quest stays in hand with nothing to finish")
	}

	hermit := gs.locationNamed("Hermit's Recluse")
	for _, p := range gs.Players {
		p.CurrentLocation = hermit
	}
	gs.Creatures = []*creature.Unit{}
	for _, p := range gs.Players {
		c := gs.NewStag(hermit)
		c.Bonded, c.BondedTo = true, p
		gs.Creatures = append(gs.Creatures, c)
	}
	gs.RefreshHands()

	unknown := gs.locationNamed("Redhorn Lake")
	unknown.Known = false
	if second.Finishable(gs.QuestWorld()) {
		t.Fatal("Open the Eye is finishable with an unknown location")
	}
	for _, l := range gs.Locations {
		l.Known = true
	}
	gs.RefreshHands()

	if !second.Finishable(gs.QuestWorld()) || !gs.MandatoryFinishPending() {
		t.Fatal("Open the Eye should be finishable and mandatory")
	}
	for _, c := range gs.Players[0].CardsInHand {
		if c.Type != card.PlayerFinishQuest && c.CanBeCast {
			t.Errorf("%s can be cast while a mandatory finish is pending", c.Name)
		}
	}

	playFinish(t, gs, second)
	if got := activeQuestNames(gs); len(got) != 3 {
		t.Fatalf("active quests = %v, want the three of Awaken the Elements", got)
	}
}

func TestFloodRequiresPaymentThatTheCallerSplits(t *testing.T) {
	gs := newGame()
	bondKindredToAllPlayers(gs)
	playFinish(t, gs, questNamed(t, gs, "Beyond the Body"))

	lake := gs.locationNamed("Redhorn Lake")
	gs.Players[0].CurrentLocation, gs.Players[1].CurrentLocation = lake, lake
	gs.Players[0].ResourcesAvailable, gs.Players[1].ResourcesAvailable = 6, 5

	open := questNamed(t, gs, "Open the Eye")
	open.Status = quest.StatusFinished
	gs.Scenario.Advance()
	gs.RefreshHands()

	flood := questNamed(t, gs, "A Flood is Coming")
	playFinish(t, gs, flood)
	if gs.PendingPayment != flood {
		t.Fatal("the flood quest should wait for a payment")
	}
	if _, err := gs.PlayCard(finishCardOf(gs, 0).ID, flood.ID, ""); err == nil {
		t.Error("a card was played while a payment is pending")
	}

	if err := gs.PayForPendingQuest(flood.ID, map[int]int{gs.Players[0].ID: 4, gs.Players[1].ID: 5}); err == nil {
		t.Error("a payment of 9 was accepted for a cost of 10")
	}
	if err := gs.PayForPendingQuest(flood.ID, map[int]int{gs.Players[0].ID: 0, gs.Players[1].ID: 10}); err == nil {
		t.Error("a player paid more than they hold")
	}
	if err := gs.PayForPendingQuest(flood.ID, map[int]int{gs.Players[0].ID: 6, gs.Players[1].ID: 4}); err != nil {
		t.Fatal(err)
	}
	if gs.Players[0].ResourcesAvailable != 0 || gs.Players[1].ResourcesAvailable != 1 {
		t.Errorf("resources left = %d and %d, want 0 and 1", gs.Players[0].ResourcesAvailable, gs.Players[1].ResourcesAvailable)
	}
	if flood.Status != quest.StatusFinished || gs.PendingPayment != nil {
		t.Error("the flood quest should be finished and no payment pending")
	}
}

func TestTwoOfThreeFinishThePlotlineAndTheOpenOneDecidesTheGoal(t *testing.T) {
	gs := newGame()
	gs.Scenario.Items[0].(*quest.Quest).Status = quest.StatusFinished
	gs.Scenario.Items[1].(*quest.Quest).Status = quest.StatusFinished
	gs.Scenario.Advance()

	flood := questNamed(t, gs, "A Flood is Coming")
	become := questNamed(t, gs, "Become One")
	plains := questNamed(t, gs, "Across the Eternal Plains")
	flood.Finish()
	plains.Finish()
	gs.Scenario.Advance()

	if become.Status != quest.StatusAbandoned {
		t.Fatalf("Become One status = %v, want abandoned", become.Status)
	}
	goal := questNamed(t, gs, "To understand eventually...").Goals[0].Text()
	if goal != "All players and bonded creatures are at Coniferous Grove" {
		t.Errorf("goal text = %q", goal)
	}
}

func TestBecomeOneSpawnsAPlantTwoMovesFromTheGrove(t *testing.T) {
	gs := newGame()
	gs.Creatures = []*creature.Unit{}
	gs.Scenario.Items[0].(*quest.Quest).Status = quest.StatusFinished
	gs.Scenario.Items[1].(*quest.Quest).Status = quest.StatusFinished
	gs.Scenario.Advance()

	grove := gs.locationNamed("Coniferous Grove")
	if len(gs.Creatures) != 1 || gs.Creatures[0].Name != plantName {
		t.Fatalf("creatures = %v, want one plant", gs.Creatures)
	}
	if moves := grove.DistanceTo(gs.Creatures[0].CurrentLocation) - 1; moves != 2 {
		t.Errorf("the plant spawned %d moves from the grove", moves)
	}
}

func TestEnragedSpiritSpawnsAtTheEndOfTheJourneyAndDefeatingItIsTheGoal(t *testing.T) {
	gs := newGame()
	gs.Creatures = []*creature.Unit{}
	gs.Scenario.Items[0].(*quest.Quest).Status = quest.StatusFinished
	gs.Scenario.Items[1].(*quest.Quest).Status = quest.StatusFinished
	gs.Scenario.Advance()
	flood := questNamed(t, gs, "A Flood is Coming")
	become := questNamed(t, gs, "Become One")
	flood.Finish()
	become.Finish()
	gs.Scenario.Advance()
	journeyEnd := questNamed(t, gs, "To understand eventually...")
	journeyEnd.Finish()
	gs.Scenario.Advance()

	last := questNamed(t, gs, "At long last, homebound.")
	var spirit *creature.Unit
	for _, c := range gs.Creatures {
		if c.Name == "Enraged Spirit" {
			spirit = c
		}
	}
	if spirit == nil || spirit.CurrentLocation != gs.locationNamed("Windswept Plains") {
		t.Fatalf("spirit = %v, want one at Windswept Plains", spirit)
	}
	if last.Finishable(gs.QuestWorld()) {
		t.Fatal("finishable before the spirit is defeated")
	}
	spirit.Distress = spirit.MaxDistress
	gs.ReconcileDefeats()
	if !last.Finishable(gs.QuestWorld()) {
		t.Fatal("not finishable after the spirit is defeated")
	}
}
