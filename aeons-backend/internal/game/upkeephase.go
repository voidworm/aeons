package game

func (gs *GameState) ResolveUpkeepPhase() {
	gs.ResolveResourceRespawn()
	gs.ResolvePlayerRefresh()
}

func (gs *GameState) ResolveResourceRespawn() {
	for _, v := range gs.HarvestUnits {
		v.Tick()
	}
}

func (gs *GameState) ResolvePlayerRefresh() {
	for _, v := range gs.Players {
		v.RemainingActions = 3
	}
	gs.TurnCounter += 1
}
