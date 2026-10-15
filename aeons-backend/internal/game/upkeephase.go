package game

const UpkeepEnergy = 4

// UpkeepEvent records one creature recovering from exhaustion in the upkeep phase.
type UpkeepEvent struct {
	CreatureID int
}

// ResolveUpkeepPhase runs the upkeep phase and returns what happened, in order.
func (gs *GameState) ResolveUpkeepPhase() []UpkeepEvent {
	gs.Logf("-- Upkeep phase --")
	gs.ResolveResourceRespawn()
	gs.ResolvePlayerRefresh()
	events := gs.ResolveCreatureRecovery()

	for _, p := range gs.Players {
		gs.EvaluatePlayerHandPlayability(p)
	}
	return events
}

func (gs *GameState) ResolveResourceRespawn() {
	for _, v := range gs.HarvestUnits {
		v.Tick()
	}
}

func (gs *GameState) ResolvePlayerRefresh() {
	for _, v := range gs.Players {
		if v.Fainted() {
			continue
		}
		v.RemainingActions = 3
		v.Energy = UpkeepEnergy
		gs.Logf("%s refreshes to %d energy", v.Name, UpkeepEnergy)
	}
}

// ResolveCreatureRecovery ends the turn state of all creatures, exhausted ones produce an event.
func (gs *GameState) ResolveCreatureRecovery() []UpkeepEvent {
	events := []UpkeepEvent{}
	for _, v := range gs.Creatures {
		if v.Distracted {
			events = append(events, UpkeepEvent{CreatureID: v.ID})
			gs.Logf("%s recovers from exhaustion", v.Name)
		}
		v.Tick()
	}
	return events
}
