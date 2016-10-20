package lemon

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestHookLifecycle(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), (500 * time.Millisecond))

	e, err := NewWithContext(ctx)
	if err != nil {
		cancel()
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h := &testHook{}
	h.kill = make(chan struct{}, 1)

	d := make(chan struct{}, 1)

	e.Register(h)

	go func() {

		e.Start()
		defer func() {
			d <- struct{}{}
		}()

		if !h.startCalled {
			t.Fatal("Engine should have start given Hook")
		}

		if !h.stopCalled {
			t.Fatal("Engine should have stop given Hook")
		}

		t.Log("Engine has started then stopped given Hook")

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(600 * time.Millisecond):
		t.Fatal("Engine should have stopped.")
	}

	cancel()

}

func TestBeforeShutdownHookWithCancelContext(t *testing.T) {

	var before bool

	ctx, cancel := context.WithTimeout(context.Background(), (500 * time.Millisecond))

	e, err := NewWithContext(ctx, BeforeShutdown(func() {
		t.Log("Engine has executed BeforeShutdown hook.")
		before = true
	}))

	if err != nil {
		cancel()
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h1 := &testHook{}
	h1.kill = make(chan struct{}, 1)

	h2 := &testHook{}
	h2.kill = make(chan struct{}, 1)

	e.Register(h1)
	e.Register(h2)

	d := make(chan struct{}, 1)

	go func() {

		e.Start()
		defer func() {
			d <- struct{}{}
		}()

		if !before {
			t.Fatal("Engine should have executed BeforeShutdown hook.")
		}

		t.Log("Engine has executed BeforeShutdown hook.")

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(600 * time.Millisecond):
		t.Fatal("Engine should have stopped.")
	}

	cancel()

}

func TestAfterShutdownHookWithCancelContext(t *testing.T) {

	var after bool

	ctx, cancel := context.WithTimeout(context.Background(), (500 * time.Millisecond))

	e, err := NewWithContext(ctx, AfterShutdown(func() {
		t.Log("Engine has executed AfterShutdown hook.")
		after = true
	}))

	if err != nil {
		cancel()
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h1 := &testHook{}
	h1.kill = make(chan struct{}, 1)

	h2 := &testHook{}
	h2.kill = make(chan struct{}, 1)

	e.Register(h1)
	e.Register(h2)

	d := make(chan struct{}, 1)

	go func() {

		e.Start()
		defer func() {
			d <- struct{}{}
		}()

		if !after {
			t.Fatal("Engine should have executed AfterShutdown hook.")
		}

		t.Log("Engine has executed AfterShutdown hook.")

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(600 * time.Millisecond):
		t.Fatal("Engine should have stopped.")
	}

	cancel()

}

func TestBeforeShutdownHookWithSignal(t *testing.T) {

	var before bool

	e, err := New(BeforeShutdown(func() {
		t.Log("Engine has executed BeforeShutdown hook.")
		before = true
	}))

	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h1 := &testHook{}
	h1.kill = make(chan struct{}, 1)

	h2 := &testHook{}
	h2.kill = make(chan struct{}, 1)

	e.interrupt = make(chan os.Signal, 1)
	e.Register(h1)
	e.Register(h2)

	go func() {
		time.Sleep(200 * time.Millisecond)
		e.interrupt <- syscall.SIGINT
	}()

	d := make(chan struct{}, 1)

	go func() {

		e.Start()
		defer func() {
			d <- struct{}{}
		}()

		if !before {
			t.Fatal("Engine should have executed BeforeShutdown hook.")
		}

		t.Log("Engine has executed BeforeShutdown hook.")

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Engine should have stopped.")
	}

}

func TestAfterShutdownHookWithSignal(t *testing.T) {

	var after bool

	e, err := New(AfterShutdown(func() {
		t.Log("Engine has executed AfterShutdown hook.")
		after = true
	}))

	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h1 := &testHook{}
	h1.kill = make(chan struct{}, 1)

	h2 := &testHook{}
	h2.kill = make(chan struct{}, 1)

	e.interrupt = make(chan os.Signal, 1)
	e.Register(h1)
	e.Register(h2)

	go func() {
		time.Sleep(200 * time.Millisecond)
		e.interrupt <- syscall.SIGINT
	}()

	d := make(chan struct{}, 1)

	go func() {

		e.Start()
		defer func() {
			d <- struct{}{}
		}()

		if !after {
			t.Fatal("Engine should have executed AfterShutdown hook.")
		}

		t.Log("Engine has executed AfterShutdown hook.")

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Engine should have stopped.")
	}

}
