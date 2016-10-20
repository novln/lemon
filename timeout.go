package lemon

import (
	"time"
)

const (
	// DefaultTimeout is the default (and maximum) amount of time the engine will wait for a hook to gracefully shutdown.
	DefaultTimeout = 5 * time.Second
)

// Timeout define the maximum amount of time the engine will wait for hooks to gracefully shut down.
func (e *Engine) Timeout() time.Duration {
	return e.timeout
}

// Timeout sets the maximum amount of time the engine will wait for hooks to gracefully shut down.
// After this timeout, hooks will be forcefully shut down by destroying underlying goroutines.
func Timeout(timeout time.Duration) Option {
	return wrapOption(func(e *Engine) error {
		e.timeout = timeout
		return nil
	})
}
