package lemon

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
)

type testHook struct {
	mutex        sync.Mutex
	kill         chan struct{}
	stopCalled   bool
	startCalled  bool
	stopDone     bool
	startDone    bool
	startError   error
	stopError    error
	startTimeout bool
	stopTimeout  bool
	panicOnStart bool
	panicOnStop  bool
}

func (t *testHook) Start(ctx context.Context) error {
	t.mutex.Lock()
	t.startCalled = true
	t.mutex.Unlock()
	if t.panicOnStart {
		panic("Hook has crashed: 0xDEADC0DE")
	}
	for t.startTimeout {
		time.Sleep(100 * time.Millisecond)
	}
	if t.kill != nil {
		<-t.kill
	}
	t.mutex.Lock()
	t.startDone = true
	t.mutex.Unlock()
	return t.startError
}

func (t *testHook) Stop(ctx context.Context) error {
	t.mutex.Lock()
	t.stopCalled = true
	t.mutex.Unlock()
	if t.panicOnStop {
		panic("Hook has crashed: 0xDEADC0DE")
	}
	for t.stopTimeout {
		time.Sleep(100 * time.Millisecond)
	}
	if t.kill != nil {
		t.kill <- struct{}{}
	}
	t.mutex.Lock()
	t.stopDone = true
	t.mutex.Unlock()
	return t.stopError
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

func (r *TestRuntime) InDelta(actual, expected time.Duration, message string) {
	if actual > expected {
		r.Error("%s: %s", message, actual)
	}
}

func (r *TestRuntime) InEpsilon(actual, expected, epsilon time.Duration, message string) {
	if actual < (expected - epsilon) {
		r.Error("%s: %s", message, actual)
	}
	if actual > (expected + epsilon) {
		r.Error("%s: %s", message, actual)
	}
}

func (r *TestRuntime) HasLifecycle(h *testHook, id string) {
	r.HasStarted(h, id)
	r.HasStopped(h, id)
}

func (r *TestRuntime) HasStarted(h *testHook, id string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if !h.startCalled || !h.startDone {
		r.Error("Hook %s should have been started.", id)
	}
}

func (r *TestRuntime) HasStopped(h *testHook, id string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if !h.stopCalled || !h.stopDone {
		r.Error("Hook %s should have been stopped.", id)
	}
}

func (r *TestRuntime) HasInvoked(h *testHook, id string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if !h.startCalled || h.startDone || h.stopCalled || h.stopDone {
		r.Error("Hook %s should have try to start.", id)
	}
}

func (r *TestRuntime) HasKill(h *testHook, id string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if !h.startCalled || !h.stopCalled {
		r.Error("Hook %s should have try to shutdown.", id)
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
