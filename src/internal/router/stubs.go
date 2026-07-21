package router

// Stubs de estado vivo para EP-001. La fuente real de salud la inyecta el
// Health Monitor (EP-002) y la de cuota el Quota Manager (EP-003).

// StaticHealth marca todos los modelos como sanos.
type StaticHealth struct{}

func (StaticHealth) Healthy(_, _ string) bool { return true }

// StaticQuota devuelve una cuota fija no nula para todos los modelos.
type StaticQuota struct{}

func (StaticQuota) Remaining(_, _ string) int { return 1_000_000 }
