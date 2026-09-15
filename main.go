package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/manifoldco/promptui"
)

var runHelper = flag.Bool("runhelper", false, "Set to run the default debug loop without user input")

type Movable struct {
	Location *LocationEntity
}

func (m *Movable) PossibleMoveTargets() []*LocationEntity {
	return m.Location.OutgoingConnections
}

func (m *Movable) MoveTo(target *LocationEntity) {
	m.Location = target
}

type CanMove interface {
	PossibleMoveTargets() []*LocationEntity
	MoveTo(*LocationEntity)
}

type Damageable struct {
	CurrentHealth int
	MaxHealth     int
}

type CanBeDamaged interface {
	TakeDamage(int) int
	HealDamage(int) int
}

func (d *Damageable) TakeDamage(amount int) int {
	d.CurrentHealth = max(d.CurrentHealth-amount, 0)
	return d.CurrentHealth
}

func (d *Damageable) HealDamage(amount int) int {
	d.CurrentHealth = min(d.CurrentHealth+amount, d.MaxHealth)
	return d.CurrentHealth
}

type StaticObject struct {
	ID   int
	Name string
}

type MovingObject struct {
	Movable
	ID   int
	Name string
}

type LocationEntity struct {
	ID                  int
	Name                string
	OutgoingConnections []*LocationEntity
}

type EnemyEntity struct {
	Movable
	Damageable
	ID     int
	Name   string
	Aloof  bool
	Hunter bool
	Damage int
}

type PlayerEntity struct {
	Movable
	Damageable
	ID                 int
	Name               string
	CardsInHand        int
	ResourcesAvailable int
}

type GameState struct {
	Locations    []*LocationEntity
	Players      []*PlayerEntity
	Attached     []*AttachableEntity
	Enemies      []*EnemyEntity
	MovingObject []*MovingObject
	StaticObject []*StaticObject
}

type AttachableEntity struct {
	ID         int
	Name       string
	AttachedTo *PlayerEntity
}

type EffectContext struct {
	TargetLocation *LocationEntity
	TargetPlayer   *PlayerEntity
	TargetEnemy    *EnemyEntity
}

type Effect interface {
	Apply(ctx *EffectContext) error
}

type DealDamageToPlayerEffect struct {
	Damage int
}

func (e *DealDamageToPlayerEffect) Apply(ectx *EffectContext) error {

	if ectx.TargetPlayer == nil {
		return fmt.Errorf("TargetPlayer is required but was nil")
	}

	ectx.TargetPlayer.Damageable.TakeDamage(e.Damage)
	return nil
}

type HealDamageOnPlayerEffect struct {
	Amount int
}

func (e *HealDamageOnPlayerEffect) Apply(ectx *EffectContext) error {

	if ectx.TargetPlayer == nil {
		return fmt.Errorf("TargetPlayer is required but was nil")
	}

	ectx.TargetPlayer.Damageable.HealDamage(e.Amount)
	return nil
}

type DealDamageToEnemyEffect struct {
	Damage int
}

func (e *DealDamageToEnemyEffect) Apply(ectx *EffectContext) error {

	if ectx.TargetEnemy == nil {
		return fmt.Errorf("TargetEnemy is required but was nil")
	}

	ectx.TargetEnemy.Damageable.TakeDamage(e.Damage)
	return nil
}

type HealDamageOnEnemyEffect struct {
	Amount int
}

func (e *HealDamageOnEnemyEffect) Apply(ectx *EffectContext) error {

	if ectx.TargetEnemy == nil {
		return fmt.Errorf("TargetEnemy is required but was nil")
	}

	ectx.TargetEnemy.Damageable.HealDamage(e.Amount)
	return nil
}

type MovePlayerEffect struct {
	MoveAmount int
}

func (e *MovePlayerEffect) Apply(ectx *EffectContext) error {
	if ectx.TargetPlayer == nil {
		return fmt.Errorf("TargetPlayer is required but was nil")
	}
	for i := 0; i < e.MoveAmount; i++ {

		targets := ectx.TargetPlayer.PossibleMoveTargets()
		//simulate player input
		winner := AskPlayerForTargetSelection(targets)

		ectx.TargetPlayer.MoveTo(winner)
		log.Printf("%s has moved to %s \n", ectx.TargetPlayer.Name, winner.Name)

	}
	return nil
}

type MoveEnemyEffect struct {
	MoveAmount int
}

func (e *MoveEnemyEffect) Apply(ectx *EffectContext) error {
	if ectx.TargetEnemy == nil {
		return fmt.Errorf("TarGET ENEMY is required but was nil")
	}

	for i := 0; i < e.MoveAmount; i++ {

		targets := ectx.TargetEnemy.PossibleMoveTargets()
		//simulate hunting determination
		winner := AskPlayerForTargetSelection(targets)

		ectx.TargetEnemy.MoveTo(winner)
		log.Printf("%s has moved to %s \n", ectx.TargetEnemy.Name, winner.Name)

	}
	return nil
}

func AskPlayerForTargetSelection(in []*LocationEntity) *LocationEntity {
	winner := in[rand.IntN(len(in))]
	return winner
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
	flag.Parse()
	if *runHelper {
		helper()
		return
	}

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
		Movable:            Movable{Location: gs.Locations[0]},
		Damageable:         Damageable{CurrentHealth: 10, MaxHealth: 10},
		ID:                 1,
		Name:               "Jim Gordon",
		CardsInHand:        7,
		ResourcesAvailable: 5,
	}

	log.Println("Setting up Ivy...")
	Ivy := &PlayerEntity{
		Movable:            Movable{Location: gs.Locations[0]},
		Damageable:         Damageable{CurrentHealth: 8, MaxHealth: 8},
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
		Movable:    Movable{Location: gs.Locations[len(gs.Locations)-1]},
		Damageable: Damageable{CurrentHealth: 5, MaxHealth: 5},
		ID:         1,
		Name:       "Noxious Ghoul",
		Aloof:      true,
		Hunter:     true,
		Damage:     1,
	}

	gs.Enemies = append(gs.Enemies, ghoul)
}

func helper() {

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

	log.Println("Setting up Jim...")
	jim := PlayerEntity{
		Movable:            Movable{Location: Porch},
		Damageable:         Damageable{CurrentHealth: 10, MaxHealth: 10},
		ID:                 1,
		Name:               "Jim Gordon",
		CardsInHand:        7,
		ResourcesAvailable: 5,
	}

	log.Println("Setting up Ivy...")
	ivy := PlayerEntity{
		Movable:            Movable{Location: Porch},
		Damageable:         Damageable{CurrentHealth: 8, MaxHealth: 8},
		ID:                 1,
		Name:               "Poison Ivy",
		CardsInHand:        7,
		ResourcesAvailable: 5,
	}

	log.Println("Setting up Ghoul...")
	ghoul := EnemyEntity{
		Movable:    Movable{Location: Attic},
		Damageable: Damageable{CurrentHealth: 5, MaxHealth: 5},
		ID:         1,
		Name:       "Noxious Ghoul",
		Aloof:      true,
		Hunter:     true,
		Damage:     1,
	}

	ectx := &EffectContext{
		TargetPlayer: &jim,
		TargetEnemy:  &ghoul,
		// TargetLocation and TargetEnemy are automatically nil
	}

	mpe := &MovePlayerEffect{
		MoveAmount: 4,
	}

	mee := &MoveEnemyEffect{
		MoveAmount: 2,
	}

	err := mpe.Apply(ectx)
	if err != nil {
		log.Println("Jim laufen macht Krise")
	}

	ectx.TargetPlayer = &ivy
	err = mpe.Apply(ectx)
	if err != nil {
		log.Println("Ivy laufen macht Krise")
	}

	err = mee.Apply(ectx)
	if err != nil {
		log.Print("Ghoul laufen macht Krise")
	}

	if jim.Location.Name == ghoul.Location.Name {

		ectx.TargetPlayer = &jim

		log.Printf("%s and and %s both are in %s!\n", jim.Name, ghoul.Name, jim.Location.Name)
		log.Printf("%s will deal %v damage to Jim.\n", ghoul.Name, ghoul.Damage)
		ddtp := DealDamageToPlayerEffect{
			Damage: ghoul.Damage,
		}
		ddtp.Apply(ectx)
		log.Printf("%s's health dropped to %v\n", jim.Name, jim.CurrentHealth)
		log.Println("<< --------------------------------------- >>")
	}
	if ivy.Location.Name == ghoul.Location.Name {

		ectx.TargetPlayer = &ivy

		log.Printf("%s and and %s both are in %s!\n", ivy.Name, ghoul.Name, ivy.Location.Name)
		log.Printf("%s will deal %v damage to %s.\n", ghoul.Name, ghoul.Damage, ivy.Name)
		ddtp := DealDamageToPlayerEffect{
			Damage: ghoul.Damage,
		}
		ddtp.Apply(ectx)
		log.Printf("Ivy's health dropped to %v\n", ivy.CurrentHealth)
		log.Println("<< --------------------------------------- >>")
	}

	if jim.Location.Name == ivy.Location.Name {
		log.Printf("Hey, Jim and Ivy have met in %s!\n", jim.Location.Name)
		log.Println("They both heal a damage.")

		ectx.TargetPlayer = &jim
		ddtp := HealDamageOnPlayerEffect{
			Amount: 1,
		}
		ddtp.Apply(ectx)
		log.Printf("Jim's health healed up to %v\n", jim.CurrentHealth)

		ectx.TargetPlayer = &ivy
		ddtp.Apply(ectx)
		log.Printf("Ivy's health healed up to %v\n", ivy.CurrentHealth)
		log.Println("<< --------------------------------------- >>")
	}
}
