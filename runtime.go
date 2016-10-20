package lemon

import (
	"context"
	"time"
)

// HookRuntime handles Hook lifecycle for the Engine.
//
// Under the hood it use flag and channel to communicate Hook's state. Using this with goroutine for either Start()
// or Stop() protect the Engine from a lock with a Hook blocking forever with either Start() and/or Stop().
//
// Unfortunately, this design increase the use of goroutine.
// However, it offer a great warranty of removing any blocking or deadlock issue while a shutdown is required.
type HookRuntime struct {
	// chan used by "stop" goroutine.
	c0 chan error
	// chan used by "start" goroutine.
	c1 chan error
	// wait flag for c0.
	w0 bool
	// wait flag for c1.
	w1 bool
}

func (hr *HookRuntime) start(h Hook) {
	go func() {
		hr.c1 <- h.Start()
	}()
}

func (hr *HookRuntime) stop(h Hook) {
	go func() {
		hr.c0 <- h.Stop()
	}()
}

func (hr *HookRuntime) init() {

	if hr.c0 == nil {
		hr.c0 = make(chan error, 1)
	}

	if hr.c1 == nil {
		hr.c1 = make(chan error, 1)
	}

	hr.w0 = true
	hr.w1 = true

}

// WaitForEvent will block until a shutdown of the given Hook is required.
// Also, if an error is returned, the Engine will shutdown every Hook.
func (hr *HookRuntime) WaitForEvent(ctx context.Context, h Hook) error {

	hr.init()
	hr.start(h)

	// Either context was cancelled, or an error has occurred during Hook startup.
	select {
	case <-ctx.Done():
		// Engine's context was cancelled.
		hr.stop(h)
		return nil
	case err := <-hr.c1:

		// Forward that c1 has stopped on shutdown.
		hr.c1 <- nil

		// If an error has occurred during Hook startup, we have to ignore Hook shutdown.
		if err != nil {
			hr.w0 = false
		}

		return err
	}
}

// Shutdown will gracefully shutdown the given hook, or kill it after timeout.
// It will also synchronise that Start() and Stop() have finished.
func (hr *HookRuntime) Shutdown(timeout time.Duration) []error {

	t := time.Now()
	failures := []error{}

	// Wait for previous hook to gracefully shutdown, or kill it after timeout.
	for {
		select {
		case err := <-hr.c1:
			if err != nil {
				failures = append(failures, err)
			}
			hr.w1 = false
		case err := <-hr.c0:
			if err != nil {
				failures = append(failures, err)
			}
			hr.w0 = false
		case <-time.After(timeout - time.Since(t)):
			return failures
		}

		if !hr.w1 && !hr.w0 {
			return failures
		}
	}

}
