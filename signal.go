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

func (e *Engine) onSignalNotification() {

	if e.interrupt == nil {
		e.interrupt = make(chan os.Signal, 1)
	}

	signal.Notify(e.interrupt, e.signals...)

	for range e.interrupt {

		if e.interrupted {
			return
		}

		e.interrupted = true

		if e.beforeShutdown != nil {
			e.beforeShutdown()
		}

		e.cancel()

	}

}

// AddSignal will register the given signal has a trigger for a graceful shutdown.
func AddSignal(signal os.Signal) Option {
	return wrapOption(func(e *Engine) error {

		for _, s := range e.signals {
			if s == signal {
				return nil
			}
		}

		e.signals = append(e.signals, signal)
		return nil

	})
}
