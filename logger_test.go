package lemon

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLoggerErrOnStart(t *testing.T) {

	failures := []error{}
	handler := func(err error) {
		failures = append(failures, err)
	}

	e, err := New(context.Background(), Logger(handler))
	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h := &testHook{}
	h.startError = errors.New("An error has occurred: foobar")

	e.Register(h)

	d := make(chan struct{}, 1)

	go func() {

		if err = e.Start(); err == nil {
			t.Error("An error was expected")
		}

		defer func() {
			d <- struct{}{}
		}()

		if err != h.startError {
			t.Fatalf("Unexpected failure: %+v", err)
		}

		if len(failures) != 1 {
			t.Fatalf("Unexpected failures: %+v", failures)
		}

		if failures[0] != h.startError {
			t.Fatalf("Unexpected failure: %+v", failures[0])
		}

		t.Logf("Logger has received an error while engine was trying to manage a hook.")

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(600 * time.Millisecond):
		t.Fatal("Engine should have stopped.")
	}

}

func TestLoggerErrOnStop(t *testing.T) {

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

	h := &testHook{}
	h.kill = make(chan struct{}, 1)
	h.stopError = errors.New("An error has occurred: foobar")

	e.Register(h)

	d := make(chan struct{}, 1)

	go func() {

		if err = e.Start(); err != nil {
			t.Errorf("An error wasn't expected: %s", err)
		}

		defer func() {
			d <- struct{}{}
		}()

		if len(failures) != 1 {
			t.Fatalf("Unexpected failures: %+v", failures)
		}

		if failures[0] != h.stopError {
			t.Fatalf("Unexpected failure: %+v", failures[0])
		}

		t.Logf("Logger has received an error while engine was trying to manage a hook.")

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(600 * time.Millisecond):
		t.Fatal("Engine should have stopped.")
	}

	cancel()
}

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
