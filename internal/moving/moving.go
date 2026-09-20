package moving

import (
	"aeons/internal/location"
)

type MovingEntity struct {
	CurrentLocation *location.LocationEntity
}

type Movable interface {
	MoveTowards(*location.LocationEntity)
	MoveTo(*location.LocationEntity)
}

// generic move towards function
// will calc the path to the target and then move once towards the target
// can be overwritten if your move towards is not a bfs
func (m *MovingEntity) MoveTowards(target *location.LocationEntity) {
	steps := m.CurrentLocation.GetShortestPathTo(target)
	switch len(steps) {
	case 1:
		//the entity is already at the target we don't need to do anything
	default:
		m.MoveTo(steps[1])
	}
}

// setter for location
// overwrite if you need to do stuff before or after moving
func (m *MovingEntity) MoveTo(target *location.LocationEntity) {
	m.CurrentLocation = target
}

type MoveEffect struct {
	Entity Movable
	Target *location.LocationEntity
}

func (hme *MoveEffect) Apply() {
	hme.Entity.MoveTowards(hme.Target)
}
