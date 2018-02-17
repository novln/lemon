package lemon

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	tests := map[string]TestHandler{
		"ErrOnStart":   LoggerErrOnStart,
		"ErrOnStop":    LoggerErrOnStop,
		"ErrLifecycle": LoggerErrLifecycle,
	}

	for name, handler := range tests {
		t.Run(name, Setup(handler))
	}
}

func LoggerErrOnStart(runtime *TestRuntime) {

	failures := []error{}
	handler := func(err error) {
		failures = append(failures, err)
	}

	engine, err := New(runtime.Context(), Logger(handler))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	hook := &testHook{}
	hook.startError = errors.New("an error has occurred: foobar")

	engine.Register(hook)

	err = engine.Start()
	if err == nil {
		runtime.Error("An error was expected")
	}

	if err != hook.startError {
		runtime.Error("Unexpected failure: %+v", err)
	}

	if len(failures) != 1 {
		runtime.Error("Unexpected failures: %+v", failures)
	}

	if failures[0] != hook.startError {
		runtime.Error("Unexpected failure: %+v", failures[0])
	}

	runtime.Log("Logger has received an error while engine was trying to start a hook.")

}

func LoggerErrOnStop(runtime *TestRuntime) {

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
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	hook := &testHook{}
	hook.kill = make(chan struct{}, 1)
	hook.stopError = errors.New("an error has occurred: foobar")

	engine.Register(hook)

	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	if len(failures) == 0 {
		runtime.Error("Failures were expected")
	}

	if len(failures) != 1 {
		runtime.Error("Unexpected failures: %+v", failures)
	}

	if failures[0] != hook.stopError {
		runtime.Error("Unexpected failure: %+v", failures[0])
	}

	runtime.Log("Logger has received an error while engine was trying to stop a hook.")

}

func LoggerErrLifecycle(runtime *TestRuntime) {

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
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	// There is a trick with this testHook. Because h.kill is defined, h.startError will only be returned
	// when Stop() is called. So, this error will not be returned (which is the expected behavior)
	// by the engine's Start() method. Nonetheless, this error be forwarded on the error's logger,
	// along with the stop error.
	hook := &testHook{}
	hook.kill = make(chan struct{}, 1)
	hook.startError = errors.New("an error has occurred: foobar")
	hook.stopError = errors.New("cannot stop service: foobar already closed")

	engine.Register(hook)

	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	if len(failures) == 0 {
		runtime.Error("Failures were expected")
	}

	if len(failures) > 2 {
		runtime.Error("Unexpected failures: %+v", failures)
	}

	for _, err := range failures {
		if err != hook.startError && err != hook.stopError {
			runtime.Error("Unexpected failure: %+v", err)
		}
	}

	runtime.Log("Logger has received both errors while engine was trying to manage a hook lifecycle.")

}
