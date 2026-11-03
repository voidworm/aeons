package combat

type Damageable interface {
	TakeDamage(int) int
}

const BaseDamage = 1

type DamageEffect struct {
	Target Damageable
}

func (de *DamageEffect) Apply() {
	de.Target.TakeDamage(BaseDamage)
}
