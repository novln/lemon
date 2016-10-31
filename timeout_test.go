package lemon

import (
	"context"
	"testing"
	"time"
)

func TestTimeoutOption(t *testing.T) {

	timeout := 10 * time.Millisecond

	e, err := New(Timeout(timeout))
	if err != nil {
		t.Fatalf("An error wasn't expected: %s", err)
	}

	if e.Timeout() != timeout {
		t.Fatalf("Unexpected timeout: %s", e.Timeout())
	}

	if e.timeout != timeout {
		t.Fatalf("Unexpected timeout: %s", e.timeout)
	}

	t.Log("Engine's configuration has a correct timeout.")

}

type tBlockingStartHook struct {
	stop bool
}

func (t *tBlockingStartHook) Start() error {
	for !t.stop {
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (t *tBlockingStartHook) Stop() error {
	return nil
}

func TestTimeoutWithBlockingStart(t *testing.T) {

	timeout := 3 * time.Second
	kill := 500 * time.Millisecond
	epsilon := 30 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), kill)

	e, err := NewWithContext(ctx, Timeout(timeout))
	if err != nil {
		cancel()
		t.Fatalf("An error wasn't expected: %s", err)
	}

	e.Register(&tBlockingStartHook{})
	e.Register(&tBlockingStartHook{})
	e.Register(&tBlockingStartHook{})
	e.Register(&tBlockingStartHook{})
	e.Register(&tBlockingStartHook{})
	e.Register(&tBlockingStartHook{})
	e.Register(&tBlockingStartHook{})
	e.Register(&tBlockingStartHook{})

	d := make(chan struct{}, 1)

	go func() {

		t0 := time.Now()
		if err = e.Start(); err != nil {
			t.Errorf("An error wasn't expected: %s", err)
		}

		delta := time.Since(t0)

		defer func() {
			d <- struct{}{}
		}()

		inEpsilon(t, delta, (timeout + kill), epsilon, "Engine has shutdown with an unexpected amount of time...")
		t.Logf("Engine has shutdown with a correct timeout: %s.", delta)

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(5 * time.Second):
		t.Fatal("Engine should have stopped.")
	}

	cancel()
}

type tBlockingStopHook struct {
	kill chan struct{}
}

func (t *tBlockingStopHook) Start() error {
	<-t.kill
	return nil
}

func (t *tBlockingStopHook) Stop() error {
	// t.kill should be nil
	t.kill <- struct{}{}
	return nil
}

func TestTimeoutWithBlockingStop(t *testing.T) {

	timeout := 3 * time.Second
	kill := 500 * time.Millisecond
	epsilon := 60 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), kill)

	e, err := NewWithContext(ctx, Timeout(timeout))
	if err != nil {
		cancel()
		t.Fatalf("An error wasn't expected: %s", err)
	}

	e.Register(&tBlockingStopHook{})
	e.Register(&tBlockingStopHook{})
	e.Register(&tBlockingStopHook{})
	e.Register(&tBlockingStopHook{})
	e.Register(&tBlockingStopHook{})
	e.Register(&tBlockingStopHook{})
	e.Register(&tBlockingStopHook{})
	e.Register(&tBlockingStopHook{})

	d := make(chan struct{}, 1)

	go func() {

		t0 := time.Now()
		if err = e.Start(); err != nil {
			t.Errorf("An error wasn't expected: %s", err)
		}

		delta := time.Since(t0)

		defer func() {
			d <- struct{}{}
		}()

		inEpsilon(t, delta, (timeout + kill), epsilon, "Engine has shutdown with an unexpected amount of time...")
		t.Logf("Engine has shutdown with a correct timeout: %s.", delta)

	}()

	select {
	case <-d:
		t.Log("Engine has stopped.")
	case <-time.After(10 * time.Second):
		t.Fatal("Engine should have stopped.")
	}

	cancel()
}
