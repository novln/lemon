package lemon

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestAddSignal(t *testing.T) {

	e, err := New(AddSignal(syscall.SIGUSR1), AddSignal(syscall.SIGINT), AddSignal(syscall.SIGUSR2))
	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	hasUSR1 := false
	hasUSR2 := false

	for _, s := range e.signals {
		if s == syscall.SIGUSR1 {
			hasUSR1 = true
		}
		if s == syscall.SIGUSR2 {
			hasUSR2 = true
		}
	}

	if !hasUSR1 {
		t.Fatal("Engine should listen on SIGUSR1 signal")
	}

	if !hasUSR2 {
		t.Fatal("Engine should listen on SIGUSR2 signal")
	}

	t.Log("Engine's configuration has a correct signal listener.")

}

func TestShutdownWithMultipleSignal(t *testing.T) {

	i := 0

	e, err := New(BeforeShutdown(func() {
		t.Log("Engine has executed BeforeShutdown hook.")
		i++
	}))

	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	h := &testHook{}
	h.kill = make(chan struct{}, 1)

	e.interrupt = make(chan os.Signal, 1)

	e.Register(h)

	d := make(chan struct{}, 1)

	go func() {
		time.Sleep(200 * time.Millisecond)
		e.interrupt <- syscall.SIGINT
		time.Sleep(20 * time.Millisecond)
		e.interrupt <- syscall.SIGINT
	}()

	go func() {

		if err = e.Start(); err != nil {
			t.Errorf("An error wasn't expected: %s", err)
		}

		defer func() {
			d <- struct{}{}
		}()

		if i > 1 {
			t.Fatal("Engine shouldn't shutdown twice.")
		}

		if i == 0 {
			t.Fatal("Engine should shutdown.")
		}

		t.Log("Engine has shutdown once.")

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(600 * time.Millisecond):
		t.Fatal("Engine should have stopped.")
	}

}
