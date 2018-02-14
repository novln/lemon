package lemon

import (
	"context"
	"runtime"
	"testing"
	"time"
)

type testHook struct {
	kill         chan struct{}
	stopCalled   bool
	startCalled  bool
	startError   error
	stopError    error
	startTimeout bool
	stopTimeout  bool
}

func (t *testHook) Start(ctx context.Context) error {
	for t.startTimeout {
		time.Sleep(100 * time.Millisecond)
	}
	if t.kill != nil {
		<-t.kill
	}
	t.startCalled = true
	return t.startError
}

func (t *testHook) Stop(ctx context.Context) error {
	for t.stopTimeout {
		time.Sleep(100 * time.Millisecond)
	}
	if t.kill != nil {
		t.kill <- struct{}{}
	}
	t.stopCalled = true
	return t.stopError
}

type panicHook struct {
	kill         chan struct{}
	stopCalled   bool
	startCalled  bool
	panicOnStart bool
	panicOnStop  bool
}

func (p *panicHook) Start(ctx context.Context) error {
	p.startCalled = true
	if p.panicOnStart {
		panic("Hook has a crashed: 0xDEADC0DE")
	}
	if p.kill != nil {
		<-p.kill
	}
	return nil
}

func (p *panicHook) Stop(ctx context.Context) error {
	p.stopCalled = true
	if p.panicOnStop {
		panic("Hook has a crashed: 0xDEADC0DE")
	}
	if p.kill != nil {
		p.kill <- struct{}{}
	}
	return nil
}

func inDelta(t *testing.T, actual, expected time.Duration, message string) {
	if actual > expected {
		t.Fatalf("%s: %s", message, actual)
	}
}

func inEpsilon(t *testing.T, actual, expected, epsilon time.Duration, message string) {

	if actual < (expected - epsilon) {
		t.Fatalf("%s: %s", message, actual)
	}

	if actual > (expected + epsilon) {
		t.Fatalf("%s: %s", message, actual)
	}
}

func hasACompleteLifecycle(t *testing.T, h *testHook, id string) {
	hasStarted(t, h, id)
	hasStopped(t, h, id)
}

func hasStarted(t *testing.T, h *testHook, id string) {
	if !h.startCalled {
		t.Fatalf("Hook %s should have been started.", id)
	}
}

func hasStopped(t *testing.T, h *testHook, id string) {
	if !h.stopCalled {
		t.Fatalf("Hook %s should have been stopped.", id)
	}
}

// A TestHandler is a test case.
type TestHandler func(*TestRuntime)

// TestRuntime exposes various components for a test case.
// It's a wrapper used to avoid deadlock and race conditions with go routine and a testing.T instance.
type TestRuntime struct {
	ctx  context.Context
	done chan struct{}
	test *testing.T
}

func (r *TestRuntime) Context() context.Context {
	return r.ctx
}

func (r *TestRuntime) Error(format string, args ...interface{}) {
	r.test.Errorf(format, args...)
	r.done <- struct{}{}
	runtime.Goexit()
}

func (r *TestRuntime) Log(format string, args ...interface{}) {
	r.test.Logf(format, args...)
}

func (r *TestRuntime) InEpsilon(actual, expected, epsilon time.Duration, message string) {
	if actual < (expected - epsilon) {
		r.Error("%s: %s", message, actual)
	}
	if actual > (expected + epsilon) {
		r.Error("%s: %s", message, actual)
	}
}

// Setup bootstrap a test case.
func Setup(callback func(*TestRuntime)) func(*testing.T) {
	return func(test *testing.T) {
		runtime := &TestRuntime{}
		runtime.ctx = context.Background()
		runtime.done = make(chan struct{}, 1)
		runtime.test = test

		// Execute test in a go routine...
		go func() {
			callback(runtime)
			runtime.done <- struct{}{}
		}()

		select {
		case <-runtime.done:
		case <-time.After(10 * time.Second):
			test.Fatal("Test has timeout.")
		}
	}
}
