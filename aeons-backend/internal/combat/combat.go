package combat

type HealthPool struct {
	CurrentHealth int
	MaxHealth     int
}

type Damageable interface {
	TakeDamage(int) int
}

type Healable interface {
	HealDamage(int) int
}

const BaseDamage = 1

func (d *HealthPool) TakeDamage(amount int) int {
	d.CurrentHealth = max(d.CurrentHealth-amount, 0)
	return d.CurrentHealth
}

func (d *HealthPool) HealDamage(amount int) int {
	d.CurrentHealth = min(d.CurrentHealth+amount, d.MaxHealth)
	return d.CurrentHealth
}

type DamageEffect struct {
	Target Damageable
}

func (de *DamageEffect) Apply() {
	de.Target.TakeDamage(BaseDamage)
}
