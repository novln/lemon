package lemon

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	tests := map[string]TestHandler{
		"ErrOnStart": ErrOnStart,
		"ErrOnStop":  ErrOnStop,
	}

	for name, handler := range tests {
		t.Run(name, Setup(handler))
	}
}

func ErrOnStart(runtime *TestRuntime) {
	defer runtime.Done()

	failures := []error{}
	handler := func(err error) {
		failures = append(failures, err)
	}

	engine, err := New(context.Background(), Logger(handler))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
		return
	}

	hook := &testHook{}
	hook.startError = errors.New("An error has occurred: foobar")

	engine.Register(hook)

	err = engine.Start()
	if err == nil {
		runtime.Error("An error was expected")
		return
	}

	if err != hook.startError {
		runtime.Error("Unexpected failure: %+v", err)
		return
	}

	if len(failures) != 1 {
		runtime.Error("Unexpected failures: %+v", failures)
		return
	}

	if failures[0] != hook.startError {
		runtime.Error("Unexpected failure: %+v", failures[0])
		return
	}

	runtime.Log("Logger has received an error while engine was trying to manage a hook.")

}

func ErrOnStop(runtime *TestRuntime) {
	defer runtime.Done()

	failures := []error{}
	handler := func(err error) {
		failures = append(failures, err)
	}

	kill := 20 * time.Millisecond
	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, Logger(handler))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
		return
	}

	hook := &testHook{}
	hook.kill = make(chan struct{}, 1)
	hook.stopError = errors.New("An error has occurred: foobar")

	engine.Register(hook)

	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
		return
	}

	if len(failures) != 1 {
		runtime.Error("Unexpected failures: %+v", failures)
		return
	}

	if failures[0] != hook.stopError {
		runtime.Error("Unexpected failure: %+v", failures[0])
		return
	}

	runtime.Log("Logger has received an error while engine was trying to manage a hook.")

}

// TODO Move to v2
func TestLoggerErrOnStartAndStop(t *testing.T) {

	failures := []error{}
	handler := func(err error) {
		failures = append(failures, err)
	}

	kill := 20 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), kill)

	e, err := New(ctx, Logger(handler))
	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	// There is a trick with this testHook. Because h.kill is defined, h.startError will only be returned
	// when Stop() is called. So, this error will not be returned (which is the expected behavior)
	// by the engine's Start() method. Nonetheless, this error be forwarded on the error's logger,
	// along with the stop error.
	h := &testHook{}
	h.kill = make(chan struct{}, 1)
	h.startError = errors.New("An error has occurred: foobar")
	h.stopError = errors.New("Cannot stop service: foobar already closed")

	e.Register(h)

	d := make(chan struct{}, 1)

	go func() {

		if err = e.Start(); err != nil {
			t.Error("An error wasn't expected")
		}

		defer func() {
			d <- struct{}{}
		}()

		if len(failures) == 0 {
			t.Fatal("Failures were expected")
		}

		if len(failures) > 2 {
			t.Fatalf("Unexpected failures: %+v", failures[2:])
		}

		for _, err := range failures {
			if err != h.startError && err != h.stopError {
				t.Fatalf("Unexpected failure: %+v", err)
			}
		}

		t.Logf("Logger has received both errors while engine was trying to manage a hook.")

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(600 * time.Millisecond):
		t.Fatal("Engine should have stopped.")
	}

	cancel()
}
