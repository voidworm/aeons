package game

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"

	"aeons/internal/effect"
	"aeons/internal/enemy"
	"aeons/internal/healthpool"
	"aeons/internal/location"
	"aeons/internal/moving"
	"aeons/internal/player"

	"github.com/manifoldco/promptui"
)

type GameState struct {
	Locations   []*location.LocationEntity
	Players     []*player.Player
	Enemies     []*enemy.EnemyEntity
	TurnCounter int
}

func (gs *GameState) ReconcileDefeats() {
	alivePlayers := gs.Players[:0]

	for _, v := range gs.Players {
		if v.CurrentHealth > 0 {
			alivePlayers = append(alivePlayers, v)
		}
	}

	aliveEnemies := gs.Enemies[:0]
	for _, v := range gs.Enemies {
		if v.CurrentHealth > 0 {
			aliveEnemies = append(aliveEnemies, v)
		}
	}

	gs.Enemies = aliveEnemies
	gs.Players = alivePlayers
}

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

func (gs *GameState) GetLocationByName(name string) *location.LocationEntity {
	//assumes locations are named uniquely!
	for _, v := range gs.Locations {
		if v.Name == name {
			return v
		}
	}

	return &location.LocationEntity{}
}

func (gs *GameState) PollNextTurn() (*player.Player, error) {

	selectable := []*player.Player{}
	for _, v := range gs.Players {
		if v.RemainingActions > 0 {
			selectable = append(selectable, v)
		}
	}

	activeArray := []string{}
	for _, v := range selectable {
		activeArray = append(activeArray, fmt.Sprintf("%s (currently at %s)", v.Name, v.Location.Name))
	}
	prompt := promptui.Select{
		Label: ">>> --- Choose a player to act --- <<<",
		Items: activeArray,
	}

	position, _, err := prompt.Run()
	if err != nil {
		return &player.Player{}, err
	} else {
		nextTurn := selectable[position]
		return nextTurn, nil
	}
}

func (gs *GameState) PromptNextAction(player *player.Player) (string, error) {
	activeArray := []string{"Move", "Draw", "Resource"}

	if gs.playerHasEnemiesInRange(player) {
		activeArray = append(activeArray, "Attack", "Evade")
	}

	label := fmt.Sprintf("<<< --- What will %s do? --- >>> ", player.Name)

	prompt := promptui.Select{
		Label: label,
		Items: activeArray,
	}
	_, action, err := prompt.Run()

	if err != nil {
		return "", err
	} else {
		return action, nil
	}
}

func (gs *GameState) playerHasEnemiesInRange(pce *player.Player) bool {
	for _, v := range gs.Enemies {
		if pce.Location == v.Location {
			return true
		}
	}

	return false
}

func (gs *GameState) enemiesAtSameLocationForPlayer(pce *player.Player) []*enemy.EnemyEntity {

	found := []*enemy.EnemyEntity{}
	for _, v := range gs.Enemies {
		if pce.Location == v.Location {
			found = append(found, v)
		}
	}
	return found
}

func (gs *GameState) enemyHasPlayersInRange(ee *enemy.EnemyEntity) bool {
	for _, v := range gs.Players {
		if ee.Location == v.Location {
			return true
		}
	}

	return false
}

func (gs *GameState) playersAtSameLocationForEnemy(ee *enemy.EnemyEntity) []*player.Player {

	found := []*player.Player{}
	for _, v := range gs.Players {
		if ee.Location == v.Location {
			found = append(found, v)
		}
	}

	return found
}

func (gs *GameState) PlayerHaveActionsRemaining() bool {
	for _, v := range gs.Players {
		if v.RemainingActions > 0 {
			return true
		}
	}
	return false
}

func (gs *GameState) ResolvePlayerPhaseStep(running *bool) {
	nextPlayer, err := gs.PollNextTurn()
	if err != nil {
		log.Println(err)
		*running = false
	}
	fmt.Printf("%s selected to act!\n", nextPlayer.Name)

	action, err := gs.PromptNextAction(nextPlayer)
	if err != nil {
		log.Println(err)
		*running = false
	}

	fmt.Printf("%s will perform a %s action.\n", nextPlayer.Name, action)
	nextPlayer.RemainingActions -= 1
	switch action {

	case "Move":
		moveEffect := effect.MoveEffect{&effect.MoveEffectContext{nextPlayer, 1}}
		moveEffect.Apply()
	case "Attack":
		gs.ResolveAttackForPlayer(nextPlayer)
	default:
		fmt.Println("Targeted unimplemented action")
	}
}

func (gs *GameState) ResolveAttackForPlayer(player *player.Player) {
	inRange := gs.enemiesAtSameLocationForPlayer(player)
	promptList := []string{}
	for _, v := range inRange {
		promptItem := fmt.Sprintf("%s (Remaining Health: %d/%d)", v.Name, v.CurrentHealth, v.MaxHealth)
		promptList = append(promptList, promptItem)
	}

	prompt := promptui.Select{
		Label: ">>> --- Choose a location to move to --- <<<",
		Items: promptList,
	}
	position, _, err := prompt.Run()
	if err != nil {
		log.Println("The attack failed because the input was not an enemy.")
	}
	target := inRange[position]

	target.TakeDamage(1)
	log.Printf("%s attacks %s down to %d/%d health.\n", player.Name, target.Name, target.CurrentHealth, target.MaxHealth)
	if target.CurrentHealth == 0 {
		log.Printf("%s has defeated %s!", player.Name, target.Name)
	}
}

func (gs *GameState) ResolveAttackForEnemy(enemy *enemy.EnemyEntity) {
	inRange := gs.playersAtSameLocationForEnemy(enemy)

	if len(inRange) == 0 {
		log.Printf("%s has no targets for an attack and doesn't attack.\n", enemy.Name)
		return
	}

	for _, target := range inRange {
		target.TakeDamage(enemy.Damage)
		log.Printf("%s attacks %s for %d damage down to %d/%d health.\n", enemy.Name, target.Name, enemy.Damage, target.CurrentHealth, target.MaxHealth)
		if target.CurrentHealth == 0 {
			log.Printf("%s has defeated %s!\n", target.Name, target.Name)
		}
	}
}

func (gs *GameState) ResolveEnemyPhase(running *bool) {
	for _, v := range gs.Enemies {
		gs.ResolveEnemyMovement(running, v)
		gs.ResolveEnemyAttacks(running, v)
	}
}

func (gs *GameState) ResolveEnemyMovement(running *bool, enemy *enemy.EnemyEntity) {

	if enemy.Hunter {

		target, steps := gs.DetermineHuntingTargetForEnemy(enemy)

		if len(steps) == 1 {
			log.Printf("%s is already at its prey location and does not need to move.\n", enemy.Name)
		} else {
			targetLocation := gs.GetLocationByName(steps[1])
			enemy.MoveTo(targetLocation)
			log.Printf("%s has %s as target and will move to %s to hunt its prey.\n", enemy.Name, target.Name, steps[1])
		}

	} else {
		log.Printf("%s is not a hunter and does not move.", enemy.Name)
	}
}

func (gs *GameState) ResolveEnemyAttacks(running *bool, enemy *enemy.EnemyEntity) {
	if !enemy.Aloof {
		gs.ResolveAttackForEnemy(enemy)
	} else {
		log.Printf("%s is aloof and doesn't attack.", enemy.Name)
	}
}

func (gs *GameState) StartNewTurn(running *bool) {
	for _, v := range gs.Players {
		v.RemainingActions = 3
	}
	gs.TurnCounter += 1
}

func (gs *GameState) DetermineHuntingTargetForEnemy(enemy *enemy.EnemyEntity) (*player.Player, []string) {
	CurrentTargets := []*player.Player{}
	CurrentSteps := [][]string{}
	CurrentMinDistance := math.MaxInt

	for _, v := range gs.Players {
		steps, err := enemy.Location.GetShortestPathTo(v.Location)
		if err != nil {
			log.Println(err)
		}

		if len(steps) == CurrentMinDistance {
			CurrentTargets = append(CurrentTargets, v)
			CurrentSteps = append(CurrentSteps, steps)
		}

		if len(steps) < CurrentMinDistance {
			CurrentMinDistance = len(steps)
			CurrentTargets = []*player.Player{v}
			CurrentSteps = [][]string{steps}
		}
	}

	n := rand.IntN(len(CurrentTargets))
	finalTarget := CurrentTargets[n]
	finalSteps := CurrentSteps[n]
	return finalTarget, finalSteps
}
