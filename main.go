package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/manifoldco/promptui"
)

var runHelper = flag.Bool("runhelper", false, "Set to run the default debug loop without user input")

type MovingEntity struct {
	Location *LocationEntity
}

type Movable interface {
	GenerateMoveGoal() *LocationEntity
	PossibleMoveTargets() []*LocationEntity
	PromptMoveTargetSelection() *LocationEntity
	MoveTo(*LocationEntity)
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

type EnemyEntity struct {
	MovingEntity
	HealthPoolEntity
	ID     int
	Name   string
	Aloof  bool
	Hunter bool
	Damage int
}

type PlayerEntity struct {
	MovingEntity
	HealthPoolEntity
	ID                 int
	Name               string
	CardsInHand        int
	ResourcesAvailable int
}

func (pe *PlayerEntity) GenerateMoveGoal() *LocationEntity {
	return pe.PromptMoveTargetSelection()
}

type GameState struct {
	Locations []*LocationEntity
	Players   []*PlayerEntity
	Enemies   []*EnemyEntity
}

func presentPlayerSelect(gs *GameState) (*PlayerEntity, error) {
	activeArray := []string{}
	for _, v := range gs.Players {
		activeArray = append(activeArray, fmt.Sprintf("%s (currently at %s)", v.Name, v.Location.Name))
	}

	prompt := promptui.Select{
		Label: ">>> --- Choose a player to act --- <<<",
		Items: activeArray,
	}
	position, _, err := prompt.Run()

	if err != nil {
		return &PlayerEntity{}, err
	} else {
		return gs.Players[position], nil
	}
}

func presentActionSelect(player string) (string, error) {
	activeArray := []string{"Move", "Draw", "Resource", "Attack", "Evade"}
	label := fmt.Sprintf("<<< --- What will %s do? --- >>> ", player)

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

func main() {

	gs := &GameState{}
	setupBasicLevel(gs)
	running := true

	for running {

		playerToAct, err := presentPlayerSelect(gs)
		if err != nil {
			log.Println(err)
			running = false
			continue
		}
		fmt.Printf("%s selected to act!\n", playerToAct.Name)

		action, err := presentActionSelect(playerToAct.Name)
		if err != nil {
			log.Println(err)
			running = false
			continue
		}

		fmt.Printf("%s will perform a %s action.\n", playerToAct.Name, action)

		switch action {
		case "Move":
			moveeffectctx := &MoveEffectContext{playerToAct, 1}
			moveeff := MoveEffect{moveeffectctx}
			moveeff.Apply()
		default:
			fmt.Println("Targeted unimplemented action")
		}

	}
}

func setupBasicLevel(gs *GameState) {
	setupBasicLocations(gs)
	setupBasicPlayers(gs)
	setupBasicEnemy(gs)
}

func setupBasicLocations(gs *GameState) {

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

func setupBasicPlayers(gs *GameState) {
	log.Println("Setting up Jim...")
	Jim := &PlayerEntity{
		MovingEntity:       MovingEntity{Location: gs.Locations[0]},
		HealthPoolEntity:   HealthPoolEntity{CurrentHealth: 10, MaxHealth: 10},
		ID:                 1,
		Name:               "Jim Gordon",
		CardsInHand:        7,
		ResourcesAvailable: 5,
	}

	log.Println("Setting up Ivy...")
	Ivy := &PlayerEntity{
		MovingEntity:       MovingEntity{Location: gs.Locations[0]},
		HealthPoolEntity:   HealthPoolEntity{CurrentHealth: 8, MaxHealth: 8},
		ID:                 1,
		Name:               "Poison Ivy",
		CardsInHand:        7,
		ResourcesAvailable: 5,
	}

	gs.Players = append(gs.Players, Jim, Ivy)
}

func setupBasicEnemy(gs *GameState) {

	log.Println("Setting up Ghoul...")
	ghoul := &EnemyEntity{
		MovingEntity:     MovingEntity{Location: gs.Locations[len(gs.Locations)-1]},
		HealthPoolEntity: HealthPoolEntity{CurrentHealth: 5, MaxHealth: 5},
		ID:               1,
		Name:             "Noxious Ghoul",
		Aloof:            true,
		Hunter:           true,
		Damage:           1,
	}

	gs.Enemies = append(gs.Enemies, ghoul)
}
