package lemon

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	// Signals is the default listening signals.
	Signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
)

// waitInterrupt will block until a shutdown notification is received.
func (e *Engine) waitInterrupt() {
	select {
	case <-e.interrupt:
	case <-e.parent.Done():
	}
}

// waitShutdownNotification will forward a shutdown notification on engine when a stop signal is received or
// when the parent context is terminated.
func (e *Engine) waitShutdownNotification() {

	signal.Notify(e.interrupt, e.signals...)

	e.waitInterrupt()

	if e.beforeShutdown != nil {
		e.beforeShutdown()
	}

	e.cancel()

}

// AddSignal will register the given signal has a trigger for a graceful shutdown.
func AddSignal(signal os.Signal) Option {
	return wrapOption(func(e *Engine) error {

		// Avoid repeated signal value.
		for i := range e.signals {
			if e.signals[i] == signal {
				return nil
			}
		}

		e.signals = append(e.signals, signal)
		return nil

	})
}
