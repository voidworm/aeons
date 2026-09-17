package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/manifoldco/promptui"
)

var runHelper = flag.Bool("runhelper", false, "Set to run the default debug loop without user input")

type MovingEntity struct {
	Location *LocationEntity
}

type Movable interface {
	GenerateMoveGoal() *LocationEntity
	MovementPathTowards(*LocationEntity) ([]string, error)
	PossibleMoveTargets() []*LocationEntity
	PromptMoveTargetSelection() *LocationEntity
	MoveTo(*LocationEntity)
}

func (me *MovingEntity) MovementPathTowards(target *LocationEntity) ([]string, error) {
	path, err := me.Location.getShortestPathTo(target)
	if err != nil {
		return []string{}, err
	}
	return path, nil
}

func (me *MovingEntity) GenerateMoveGoal() *LocationEntity {
	return me.PossibleMoveTargets()[rand.IntN(len(me.PossibleMoveTargets()))]
}

func (me *MovingEntity) PossibleMoveTargets() []*LocationEntity {
	return me.Location.OutgoingConnections
}

func (me *MovingEntity) PromptMoveTargetSelection() *LocationEntity {
	optionsArray := []string{}

	for _, v := range me.PossibleMoveTargets() {
		optionsArray = append(optionsArray, v.Name)
	}

	prompt := promptui.Select{
		Label: ">>> --- Choose a location to move to --- <<<",
		Items: optionsArray,
	}
	position, _, err := prompt.Run()

	if err != nil {
		log.Printf("error when retrieving move goal for generic MovingEntitiy")
		return nil
	} else {
		return me.PossibleMoveTargets()[position]
	}
}

func (m *MovingEntity) MoveTo(target *LocationEntity) {
	m.Location = target
}

type Effect interface {
	Apply()
}

type MoveEffectContext struct {
	TargetEntity Movable
	MoveAmount   int
}

type MoveEffect struct {
	ctx *MoveEffectContext
}

func (me *MoveEffect) Apply() {
	for i := 1; i <= me.ctx.MoveAmount; i++ {
		goalLocation := me.ctx.TargetEntity.GenerateMoveGoal()
		me.ctx.TargetEntity.MoveTo(goalLocation)
	}
}

type HealthPoolEntity struct {
	CurrentHealth int
	MaxHealth     int
}

type Damageable interface {
	TakeDamage(int) int
	HealDamage(int) int
}

func (d *HealthPoolEntity) TakeDamage(amount int) int {
	d.CurrentHealth = max(d.CurrentHealth-amount, 0)
	return d.CurrentHealth
}

func (d *HealthPoolEntity) HealDamage(amount int) int {
	d.CurrentHealth = min(d.CurrentHealth+amount, d.MaxHealth)
	return d.CurrentHealth
}

type LocationEntity struct {
	ID                  int
	Name                string
	OutgoingConnections []*LocationEntity
}

func (le *LocationEntity) getShortestPathTo(target *LocationEntity) ([]string, error) {

	//stuff we need to check
	queue := []*LocationEntity{le}

	//stuff we have checked
	visited := []*LocationEntity{le}

	//the parents we took
	parents := make(map[string]string)

	found := false

	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if current == target {
			found = true
			break
		}
		for _, v := range current.OutgoingConnections {
			alreadyVisited := slices.ContainsFunc(visited, func(location *LocationEntity) bool {
				return (location == v)
			})
			if !alreadyVisited {
				visited = append(visited, v)
				queue = append(queue, v)
				parents[v.Name] = current.Name
			}
		}
	}

	if found {
		path := []string{target.Name}
		currentChild := target.Name
		for {
			parent := parents[currentChild]
			if parent == "" {
				break
			}
			path = append(path, parent)
			currentChild = parent
		}
		slices.Reverse(path)
		return path, nil
	}

	//since maps should be a graph this should never happen but you never know
	return []string{}, errors.New("Locations are not connected")
}

type EnemyEntity struct {
	MovingEntity
	HealthPoolEntity
	ID     int
	Name   string
	Aloof  bool
	Hunter bool
	Damage int
}

func (ee *EnemyEntity) GenerateMoveGoal() *LocationEntity {

	return nil
}

func (ee *EnemyEntity) determineHuntingTarget(players []*Player) (*Player, []string) {
	CurrentTargets := []*Player{}
	CurrentSteps := [][]string{}
	CurrentMinDistance := math.MaxInt

	for _, v := range players {
		steps, err := ee.Location.getShortestPathTo(v.Location)
		if err != nil {
			log.Println(err)
		}

		if len(steps) == CurrentMinDistance {
			CurrentTargets = append(CurrentTargets, v)
			CurrentSteps = append(CurrentSteps, steps)
		}

		if len(steps) < CurrentMinDistance {
			CurrentMinDistance = len(steps)
			CurrentTargets = []*Player{v}
			CurrentSteps = [][]string{steps}
		}
	}

	n := rand.IntN(len(CurrentTargets))
	finalTarget := CurrentTargets[n]
	finalSteps := CurrentSteps[n]
	return finalTarget, finalSteps
}

type Player struct {
	MovingEntity
	HealthPoolEntity
	ID                 int
	Name               string
	CardsInHand        int
	ResourcesAvailable int
	RemainingActions   int
}

func (pe *Player) GenerateMoveGoal() *LocationEntity {
	return pe.PromptMoveTargetSelection()
}

type GameState struct {
	Locations   []*LocationEntity
	Players     []*Player
	Enemies     []*EnemyEntity
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
	Porch := &LocationEntity{ID: 1, Name: "Porch"}
	DownstairsHallway := &LocationEntity{ID: 1, Name: "Downstairs Hallway"}
	Kitchen := &LocationEntity{ID: 1, Name: "Kitchen"}
	LivingRoom := &LocationEntity{ID: 1, Name: "Living Room"}
	UpstairsHallway := &LocationEntity{ID: 1, Name: "Upstairs Hallway"}
	SleepingRoom := &LocationEntity{ID: 1, Name: "Sleeping Room"}
	Attic := &LocationEntity{ID: 1, Name: "Attic"}

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
	Jim := &Player{
		MovingEntity:       MovingEntity{Location: gs.Locations[0]},
		HealthPoolEntity:   HealthPoolEntity{CurrentHealth: 10, MaxHealth: 10},
		ID:                 1,
		Name:               "Jim Gordon",
		CardsInHand:        7,
		ResourcesAvailable: 5,
		RemainingActions:   3,
	}

	log.Println("Setting up Ivy...")
	Ivy := &Player{
		MovingEntity:       MovingEntity{Location: gs.Locations[0]},
		HealthPoolEntity:   HealthPoolEntity{CurrentHealth: 8, MaxHealth: 8},
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
	ghoul := &EnemyEntity{
		MovingEntity:     MovingEntity{Location: gs.Locations[len(gs.Locations)-1]},
		HealthPoolEntity: HealthPoolEntity{CurrentHealth: 5, MaxHealth: 5},
		ID:               1,
		Name:             "Noxious Ghoul",
		Aloof:            false,
		Hunter:           true,
		Damage:           2,
	}

	log.Println("Setting up Rat...")
	rat := &EnemyEntity{
		MovingEntity:     MovingEntity{Location: gs.Locations[3]},
		HealthPoolEntity: HealthPoolEntity{CurrentHealth: 2, MaxHealth: 2},
		ID:               1,
		Name:             "Chittering Rat",
		Aloof:            false,
		Hunter:           false,
		Damage:           1,
	}

	log.Println("Setting up Suspicious Plant...")
	plant := &EnemyEntity{
		MovingEntity:     MovingEntity{Location: gs.Locations[3]},
		HealthPoolEntity: HealthPoolEntity{CurrentHealth: 2, MaxHealth: 2},
		ID:               1,
		Name:             "Suspicious Plant",
		Aloof:            true,
		Hunter:           false,
		Damage:           1,
	}

	gs.Enemies = append(gs.Enemies, ghoul, rat, plant)
}

func (gs *GameState) GetLocationByName(name string) *LocationEntity {
	//assumes locations are named uniquely!
	for _, v := range gs.Locations {
		if v.Name == name {
			return v
		}
	}

	return &LocationEntity{}
}

func (gs *GameState) PollNextTurn() (*Player, error) {

	selectable := []*Player{}
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
		return &Player{}, err
	} else {
		nextTurn := selectable[position]
		return nextTurn, nil
	}
}

func (gs *GameState) PromptNextAction(player *Player) (string, error) {
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

func (gs *GameState) playerHasEnemiesInRange(pce *Player) bool {
	for _, v := range gs.Enemies {
		if pce.Location == v.Location {
			return true
		}
	}

	return false
}

func (gs *GameState) enemiesAtSameLocationForPlayer(pce *Player) []*EnemyEntity {

	found := []*EnemyEntity{}
	for _, v := range gs.Enemies {
		if pce.Location == v.Location {
			found = append(found, v)
		}
	}
	return found
}

func (gs *GameState) enemyHasPlayersInRange(ee *EnemyEntity) bool {
	for _, v := range gs.Players {
		if ee.Location == v.Location {
			return true
		}
	}

	return false
}

func (gs *GameState) playersAtSameLocationForEnemy(ee *EnemyEntity) []*Player {

	found := []*Player{}
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
		moveEffect := MoveEffect{&MoveEffectContext{nextPlayer, 1}}
		moveEffect.Apply()
	case "Attack":
		gs.ResolveAttackForPlayer(nextPlayer)
	default:
		fmt.Println("Targeted unimplemented action")
	}
}

func (gs *GameState) ResolveAttackForPlayer(player *Player) {
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

func (gs *GameState) ResolveAttackForEnemy(enemy *EnemyEntity) {
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

func (gs *GameState) ResolveEnemyMovement(running *bool, enemy *EnemyEntity) {

	if enemy.Hunter {

		target, steps := enemy.determineHuntingTarget(gs.Players)

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

func (gs *GameState) ResolveEnemyAttacks(running *bool, enemy *EnemyEntity) {
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

func main() {

	gs := &GameState{TurnCounter: 1}
	gs.Init()
	running := true

	for running {
		log.Printf("Starting Turn %d", gs.TurnCounter)
		for gs.PlayerHaveActionsRemaining() {
			gs.ResolvePlayerPhaseStep(&running)
			gs.ReconcileDefeats()
		}

		gs.ResolveEnemyPhase(&running)
		gs.StartNewTurn(&running)
	}
}
