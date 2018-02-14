package lemon

import (
	"context"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	tests := map[string]TestHandler{
		"OkOption":  TimeoutOkOption,
		"ErrOption": TimeoutErrOption,
		"Start":     TimeoutStart,
		"Stop":      TimeoutStop,
	}

	for name, handler := range tests {
		t.Run(name, Setup(handler))
	}
}

func TimeoutOkOption(runtime *TestRuntime) {

	timeout := 10 * time.Millisecond

	engine, err := New(runtime.Context(), Timeout(timeout))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	if engine.Timeout() != timeout {
		runtime.Error("Unexpected timeout: %s", engine.Timeout())
	}

	if engine.timeout != timeout {
		runtime.Error("Unexpected timeout: %s", engine.timeout)
	}

	runtime.Log("Engine's configuration has a correct timeout.")

}

func TimeoutErrOption(runtime *TestRuntime) {

	timeout := -10 * time.Millisecond

	engine, err := New(context.Background(), Timeout(timeout))
	if err == nil {
		runtime.Error("An error was expected")
	}
	if engine != nil {
		runtime.Error("Engine should be undefined")
	}

	if err != ErrTimeout {
		runtime.Error("Unexpected error: %s", err)
	}

	runtime.Log("Engine's configuration can't have a negative timeout.")

}

func TimeoutStart(runtime *TestRuntime) {

	timeout := 3 * time.Second
	kill := 500 * time.Millisecond
	epsilon := 60 * time.Millisecond
	maximum := kill + timeout

	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, Timeout(timeout))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	create := func() *testHook {
		return &testHook{
			startTimeout: true,
		}
	}

	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())

	now := time.Now()
	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	delta := time.Since(now)

	runtime.InEpsilon(delta, maximum, epsilon, "Engine has shutdown with an unexpected amount of time...")

	runtime.Log("Engine has shutdown with a correct timeout: %s.", delta)

}

func TimeoutStop(runtime *TestRuntime) {

	timeout := 3 * time.Second
	kill := 500 * time.Millisecond
	epsilon := 60 * time.Millisecond
	maximum := kill + timeout

	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, Timeout(timeout))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	create := func() *testHook {
		return &testHook{
			kill:        make(chan struct{}),
			stopTimeout: true,
		}
	}

	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())
	engine.Register(create())

	now := time.Now()
	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	delta := time.Since(now)

	runtime.InEpsilon(delta, maximum, epsilon, "Engine has shutdown with an unexpected amount of time...")

	runtime.Log("Engine has shutdown with a correct timeout: %s.", delta)

}
