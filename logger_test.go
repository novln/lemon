package lemon

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestLoggerErrOnStart(t *testing.T) {

	failures := []error{}
	sync := &sync.Mutex{}
	handler := func(err error) {
		sync.Lock()
		defer sync.Unlock()
		failures = append(failures, err)
	}

	e, err := New(Logger(handler))
	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h := &testHook{}
	h.startError = errors.New("An error has occurred: foobar")

	e.Register(h)

	d := make(chan struct{}, 1)

	go func() {

		e.Start()
		defer func() {
			d <- struct{}{}
		}()

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
	sync := &sync.Mutex{}
	handler := func(err error) {
		sync.Lock()
		defer sync.Unlock()
		failures = append(failures, err)
	}

	kill := 20 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), kill)

	e, err := NewWithContext(ctx, Logger(handler))
	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h := &testHook{}
	h.kill = make(chan struct{}, 1)
	h.stopError = errors.New("An error has occurred: foobar")

	e.Register(h)

	d := make(chan struct{}, 1)

	go func() {

		e.Start()
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
	sync := &sync.Mutex{}
	handler := func(err error) {
		sync.Lock()
		defer sync.Unlock()
		failures = append(failures, err)
	}

	kill := 20 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), kill)

	e, err := NewWithContext(ctx, Logger(handler))
	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h := &testHook{}
	h.kill = make(chan struct{}, 1)
	h.startError = errors.New("An error has occurred: foobar")
	h.stopError = errors.New("Cannot stop service: foobar already closed")

	e.Register(h)

	d := make(chan struct{}, 1)

	go func() {

		e.Start()
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
