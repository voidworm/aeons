package main

import (
	"fmt"
	"log"
	"math/rand/v2"
)

type LocationEntity struct {
	ID                  int
	Name                string
	OutgoingConnections []*LocationEntity
}

type EnemyEntity struct {
	ID         int
	Name       string
	AtLocation *LocationEntity
	Aloof      bool
	Hunter     bool
	Damage     int
	Horror     int
	Health     int
	Sanity     int
}

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

type StaticObject struct {
	ID   int
	Name string
}

type MovingObject struct {
	Movable
	ID   int
	Name string
}

type PlayerEntity struct {
	Movable
	ID                 int
	Name               string
	CardsInHand        int
	ResourcesAvailable int
	Health             int
	Sanity             int
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

	ectx.TargetPlayer.Health -= e.Damage
	return nil
}

type MovePlayerEffect struct {
	MoveAmount int
}

func (e *MovePlayerEffect) Apply(ectx *EffectContext) error {
	if ectx.TargetPlayer == nil {
		return fmt.Errorf("TargetPlayer is required but was nil")
	}

	log.Printf("Will move %s %v times, starting in %s", ectx.TargetPlayer.Name, e.MoveAmount, ectx.TargetPlayer.Location.Name)

	for i := 0; i < e.MoveAmount; i++ {

		targets := ectx.TargetPlayer.PossibleMoveTargets()
		//simulate player input
		winner := AskPlayerForTargetSelection(targets)

		ectx.TargetPlayer.MoveTo(winner)
		log.Printf("Have moved player %s to location %s \n", ectx.TargetPlayer.Name, winner.Name)

	}
	return nil
}

func AskPlayerForTargetSelection(in []*LocationEntity) *LocationEntity {
	winner := in[rand.IntN(len(in))]
	log.Printf("Will move to location %s", winner.Name)
	return winner
}

type DealDamageToEnemyEffect struct {
	Damage int
}

func (e *DealDamageToEnemyEffect) Apply(ectx *EffectContext) error {

	if ectx.TargetEnemy == nil {
		return fmt.Errorf("TargetPlayer is required but was nil")
	}

	ectx.TargetEnemy.Health -= e.Damage
	return nil
}

func main() {

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
		ID:                 1,
		Name:               "Jim Gordon",
		CardsInHand:        7,
		ResourcesAvailable: 5,
		Health:             10,
		Sanity:             7,
	}

	log.Println("Setting up Ivy...")
	ivy := PlayerEntity{
		Movable:            Movable{Location: Porch},
		ID:                 1,
		Name:               "Poison Ivy",
		CardsInHand:        7,
		ResourcesAvailable: 5,
		Health:             8,
		Sanity:             9,
	}

	ctx := &EffectContext{
		TargetPlayer: &jim,
		// TargetLocation and TargetEnemy are automatically nil
	}

	mpe := &MovePlayerEffect{
		MoveAmount: 4,
	}

	err := mpe.Apply(ctx)
	if err != nil {
		log.Println("wallah krise")
	}

	ctx.TargetPlayer = &ivy
	mpe.Apply(ctx)
	if err != nil {
		log.Println("wallah krise")
	}

	if jim.Location.Name == ivy.Location.Name {
		log.Printf("Hey, Jim and Ivy have met in %s!\n", jim.Location.Name)
	}
}
