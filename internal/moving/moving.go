package moving

import (
	"log"
	"math/rand/v2"

	"github.com/manifoldco/promptui"

	"aeons/internal/location"
)

type MovingEntity struct {
	Location *location.LocationEntity
}

type Movable interface {
	GenerateMoveGoal() *location.LocationEntity
	MovementPathTowards(*location.LocationEntity) ([]string, error)
	PossibleMoveTargets() []*location.LocationEntity
	PromptMoveTargetSelection() *location.LocationEntity
	MoveTo(*location.LocationEntity)
}

func (me *MovingEntity) MovementPathTowards(target *location.LocationEntity) ([]string, error) {
	path, err := me.Location.GetShortestPathTo(target)
	if err != nil {
		return []string{}, err
	}
	return path, nil
}

func (me *MovingEntity) GenerateMoveGoal() *location.LocationEntity {
	return me.PossibleMoveTargets()[rand.IntN(len(me.PossibleMoveTargets()))]
}

func (me *MovingEntity) PossibleMoveTargets() []*location.LocationEntity {
	return me.Location.OutgoingConnections
}

func (me *MovingEntity) PromptMoveTargetSelection() *location.LocationEntity {
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

func (m *MovingEntity) MoveTo(target *location.LocationEntity) {
	m.Location = target
}
