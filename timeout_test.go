package lemon

import (
	"context"
	"testing"
	"time"
)

func TestTimeoutOption(t *testing.T) {

	timeout := 10 * time.Millisecond

	e, err := New(context.Background(), Timeout(timeout))
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

func TestTimeoutOptionWithErr(t *testing.T) {

	timeout := -10 * time.Millisecond

	e, err := New(context.Background(), Timeout(timeout))
	if err == nil {
		t.Fatal("An error was expected")
	}

	if err != ErrTimeout {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e != nil {
		t.Fatal("Engine should be nil")
	}

	t.Log("Engine's configuration can't have a negative timeout.")

}

type tBlockingStartHook struct {
	stop bool
}

func (t *tBlockingStartHook) Start(ctx context.Context) error {
	for !t.stop {
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (t *tBlockingStartHook) Stop(ctx context.Context) error {
	return nil
}

func TestTimeoutWithBlockingStart(t *testing.T) {

	timeout := 3 * time.Second
	kill := 500 * time.Millisecond
	epsilon := 30 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), kill)

	e, err := New(ctx, Timeout(timeout))
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

func (t *tBlockingStopHook) Start(ctx context.Context) error {
	<-t.kill
	return nil
}

func (t *tBlockingStopHook) Stop(ctx context.Context) error {
	// t.kill should be nil
	t.kill <- struct{}{}
	return nil
}

func TestTimeoutWithBlockingStop(t *testing.T) {

	timeout := 3 * time.Second
	kill := 500 * time.Millisecond
	epsilon := 60 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), kill)

	e, err := New(ctx, Timeout(timeout))
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
