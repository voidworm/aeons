package moving

import (
	"math/rand/v2"

	"aeons/internal/location"
)

type MovingEntity struct {
	Location *location.LocationEntity
}

type Movable interface {
	GenerateMoveGoal() *location.LocationEntity
	MovementPathTowards(*location.LocationEntity) ([]*location.LocationEntity, error)
	PossibleMoveTargets() []*location.LocationEntity
	MoveTo(*location.LocationEntity)
}

// this is a getter for the locations shortest path
// don't overwrite this
func (me *MovingEntity) GetShortestPathTo(target *location.LocationEntity) []*location.LocationEntity {
	return me.Location.GetShortestPathTo(target)
}

// this function returns a random connected location
// it should almost always be overwritten for your entity
func (me *MovingEntity) GenerateMoveGoal() *location.LocationEntity {
	linkedLocations := me.Location.OutgoingConnections
	return linkedLocations[rand.IntN(len(linkedLocations))]
}

// setter for location
// overwrite if you need to do stuff before or after moving
func (m *MovingEntity) MoveTo(target *location.LocationEntity) {
	m.Location = target
}
