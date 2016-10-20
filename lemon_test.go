package lemon

import (
	"testing"
	"time"
)

type testHook struct {
	kill        chan struct{}
	stopCalled  bool
	startCalled bool
	startError  error
	stopError   error
}

func (t *testHook) Start() error {
	if t.kill != nil {
		<-t.kill
	}
	t.startCalled = true
	return t.startError
}

func (t *testHook) Stop() error {
	if t.kill != nil {
		t.kill <- struct{}{}
	}
	t.stopCalled = true
	return t.stopError
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
