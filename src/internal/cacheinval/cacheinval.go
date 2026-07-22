// Package cacheinval invalida la caché en RAM ante cambios en DB. En Fase 1 es
// no-op tras un feature flag; el patrón vigente es poll/retry ante miss. El
// worker completo de polling/webhook es Fase 2.
package cacheinval

// Invalidator gobierna la invalidación de caché. En MVP (Enabled=false) es no-op.
type Invalidator struct {
	// Enabled activa la invalidación (Fase 2). En Fase 1 queda en false.
	Enabled bool
	// OnInvalidate encola la hidratación asíncrona de una clave (Fase 2). Nil-safe.
	OnInvalidate func(key string)
	// MaxRetries acota los reintentos de OnMiss; 0 usa el default.
	MaxRetries int
}

// Invalidate invalida la caché de una clave. En Fase 1 (flag OFF) es no-op y
// devuelve false; en Fase 2 encola la hidratación y devuelve true.
func (i *Invalidator) Invalidate(key string) bool {
	if !i.Enabled {
		return false // Fase 1: no-op; el miss se resuelve por poll/retry
	}
	if i.OnInvalidate != nil {
		i.OnInvalidate(key)
	}
	return true
}

// OnMiss aplica el patrón poll/retry ante un miss de caché: reintenta la
// re-hidratación (fail-fast + retry) hasta MaxRetries, devolviendo el último
// error si se agotan. Es el mecanismo vigente en MVP para no quedar stale.
func (i *Invalidator) OnMiss(key string, rehydrate func() error) error {
	retries := i.MaxRetries
	if retries < 1 {
		retries = 3
	}
	var err error
	for k := 0; k < retries; k++ {
		if err = rehydrate(); err == nil {
			return nil
		}
	}
	return err
}
