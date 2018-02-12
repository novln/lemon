package lemon

import (
	"context"
)

// Hook defines a lifecycle mecanism for a component.
// If at least one Hook return an error with Start(), it will shutdown the engine.
// Either every Hook succeed to start, or none of them will...
type Hook interface {
	// Start is executed by runtime when a Hook should start.
	Start(context.Context) error
	// Stop is executed by runtime when a Hook should shutdown.
	// It is executed when given context is done, which is the same given from Start().
	// So, you may don't have anything to do if you handle ctx.Done() in Start().
	Stop(context.Context) error
}

// Register will attach the given hook on engine's lifecycle mechanism.
func (e *Engine) Register(hook Hook) {
	e.hooks = append(e.hooks, hook)
}

// BeforeShutdown will register a callback to execute when the engine will shutdown.
func BeforeShutdown(callback func()) Option {
	return wrapOption(func(e *Engine) error {
		e.beforeShutdown = callback
		return nil
	})
}

// AfterShutdown will register a callback to execute when the engine has shutdown.
func AfterShutdown(callback func()) Option {
	return wrapOption(func(e *Engine) error {
		e.afterShutdown = callback
		return nil
	})
}
