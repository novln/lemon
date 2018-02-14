package lemon

import (
	"context"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestSignal(t *testing.T) {
	tests := map[string]TestHandler{
		"Option":   SignalAddOption,
		"Shutdown": SignalShutdown,
	}

	for name, handler := range tests {
		t.Run(name, Setup(handler))
	}
}

func SignalAddOption(runtime *TestRuntime) {

	engine, err := New(runtime.Context(),
		AddSignal(syscall.SIGUSR1),
		AddSignal(syscall.SIGINT),
		AddSignal(syscall.SIGUSR2),
	)
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	hasUSR1 := false
	hasUSR2 := false

	for i := range engine.signals {
		if engine.signals[i] == syscall.SIGUSR1 {
			hasUSR1 = true
		}
		if engine.signals[i] == syscall.SIGUSR2 {
			hasUSR2 = true
		}
	}

	if !hasUSR1 {
		runtime.Error("Engine should listen on SIGUSR1 signal")
	}

	if !hasUSR2 {
		runtime.Error("Engine should listen on SIGUSR2 signal")
	}

	runtime.Log("Engine's configuration has a correct signal listener.")

}

func SignalShutdown(runtime *TestRuntime) {

	counter := int64(0)

	engine, err := New(context.Background(), BeforeShutdown(func() {
		runtime.Log("Engine has executed BeforeShutdown hook.")
		atomic.AddInt64(&counter, 1)
	}))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	hook := &testHook{}
	hook.kill = make(chan struct{}, 1)

	engine.Register(hook)

	engine.interrupt = make(chan os.Signal, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		engine.interrupt <- syscall.SIGINT
		time.Sleep(20 * time.Millisecond)
		engine.interrupt <- syscall.SIGINT
	}()

	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	shutdown := atomic.LoadInt64(&counter)

	if shutdown > 1 {
		runtime.Error("Engine shouldn't shutdown twice.")
	}

	if shutdown == 0 {
		runtime.Error("Engine should shutdown.")
	}

	runtime.Log("Engine has shutdown once.")

}
