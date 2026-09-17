package healthpool

type HealthPoolEntity struct {
	CurrentHealth int
	MaxHealth     int
}

type Damageable interface {
	TakeDamage(int) int
}

type Healable interface {
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
