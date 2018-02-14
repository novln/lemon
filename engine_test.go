package lemon

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestShutdown(t *testing.T) {
	tests := map[string]TestHandler{
		"Once":            ShutdownOnce,
		"Signal":          ShutdownWithSignal,
		"Context":         ShutdownWithContext,
		"ErrHook/Start":   ShutdownWithHookErrorOnStart,
		"ErrHook/Stop":    ShutdownWithHookErrorOnStop,
		"PanicHook/Start": ShutdownWithHookPanicOnStart,
		"PanicHook/Stop":  ShutdownWithHookPanicOnStop,
	}

	for name, handler := range tests {
		t.Run(name, Setup(handler))
	}
}

func ShutdownOnce(runtime *TestRuntime) {

	counter := int64(0)

	engine, err := New(runtime.Context(), BeforeShutdown(func() {
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

func ShutdownWithSignal(runtime *TestRuntime) {

	kill := 200 * time.Millisecond
	epsilon := 20 * time.Millisecond
	maximum := 10 * time.Millisecond

	engine, err := New(runtime.Context())
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	create := func() *testHook {
		return &testHook{
			kill: make(chan struct{}, 1),
		}
	}

	hook1 := create()
	hook2 := create()
	hook3 := create()

	engine.Register(hook1)
	engine.Register(hook2)
	engine.Register(hook3)

	engine.interrupt = make(chan os.Signal, 1)
	go func() {
		time.Sleep(kill)
		engine.interrupt <- syscall.SIGINT
	}()

	now := time.Now()
	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	delta := time.Since(now)
	latency := delta - kill

	runtime.InDelta(latency, maximum, "Latency between signal and stop is too great")
	runtime.InEpsilon(delta, kill, epsilon, "Engine shouldn't stopped in this interval")

	runtime.HasLifecycle(hook1, "hook1")
	runtime.HasLifecycle(hook2, "hook2")
	runtime.HasLifecycle(hook3, "hook3")

	runtime.Log("Latency: %s", latency)

}

func ShutdownWithContext(runtime *TestRuntime) {

	kill := 200 * time.Millisecond
	epsilon := 20 * time.Millisecond
	maximum := 10 * time.Millisecond

	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx)
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	create := func() *testHook {
		return &testHook{
			kill: make(chan struct{}, 1),
		}
	}

	hook1 := create()
	hook2 := create()
	hook3 := create()

	engine.Register(hook1)
	engine.Register(hook2)
	engine.Register(hook3)

	now := time.Now()
	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	delta := time.Since(now)
	latency := delta - kill

	runtime.InDelta(latency, maximum, "Latency between signal and stop is too great")
	runtime.InEpsilon(delta, kill, epsilon, "Engine shouldn't stopped in this interval")

	runtime.HasLifecycle(hook1, "hook1")
	runtime.HasLifecycle(hook2, "hook2")
	runtime.HasLifecycle(hook3, "hook3")

	runtime.Log("Latency: %s", latency)

}

func ShutdownWithHookErrorOnStart(runtime *TestRuntime) {

	kill := 200 * time.Millisecond
	maximum := 10 * time.Millisecond

	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, Timeout(kill))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	createOk := func() *testHook {
		return &testHook{
			kill: make(chan struct{}, 1),
		}
	}

	createErr := func() *testHook {
		return &testHook{
			startError: errors.New("an error has occurred: foobar"),
		}
	}

	hook1 := createOk()
	hook2 := createErr()
	hook3 := createOk()

	engine.Register(hook1)
	engine.Register(hook2)
	engine.Register(hook3)

	now := time.Now()
	err = engine.Start()
	if err == nil {
		runtime.Error("An error was expected")
	}

	delta := time.Since(now)

	runtime.InDelta(delta, maximum, "Engine took way too long to shutdown")

	runtime.HasLifecycle(hook1, "hook1")
	runtime.HasStarted(hook2, "hook2")
	runtime.HasLifecycle(hook3, "hook3")

	runtime.Log("Shutdown was successful.")

}

func ShutdownWithHookErrorOnStop(runtime *TestRuntime) {

	kill := 200 * time.Millisecond
	epsilon := 20 * time.Millisecond
	maximum := kill + epsilon

	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, Timeout(kill))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	createOk := func() *testHook {
		return &testHook{
			kill: make(chan struct{}, 1),
		}
	}

	createErr := func() *testHook {
		return &testHook{
			kill:      make(chan struct{}, 1),
			stopError: errors.New("an error has occurred: foobar"),
		}
	}

	hook1 := createOk()
	hook2 := createErr()
	hook3 := createOk()

	engine.Register(hook1)
	engine.Register(hook2)
	engine.Register(hook3)

	now := time.Now()
	err = engine.Start()
	if err != nil {
		runtime.Error("Unexpected error: %s", err)
	}

	delta := time.Since(now)

	runtime.InDelta(delta, maximum, "Engine took way too long to shutdown")

	runtime.HasLifecycle(hook1, "hook1")
	runtime.HasKill(hook2, "hook2")
	runtime.HasLifecycle(hook3, "hook3")

	runtime.Log("Shutdown was successful.")

}

func ShutdownWithHookPanicOnStart(runtime *TestRuntime) {

	kill := 200 * time.Millisecond
	maximum := 10 * time.Millisecond

	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, Timeout(kill))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	createOk := func() *testHook {
		return &testHook{
			kill: make(chan struct{}, 1),
		}
	}

	createErr := func() *testHook {
		return &testHook{
			panicOnStart: true,
			kill:         make(chan struct{}, 1),
		}
	}

	hook1 := createOk()
	hook2 := createErr()
	hook3 := createOk()

	engine.Register(hook1)
	engine.Register(hook2)
	engine.Register(hook3)

	now := time.Now()
	err = engine.Start()
	if err == nil {
		runtime.Error("An error was expected")
	}

	if err.Error() != "lemon startup failed: Hook has crashed: 0xDEADC0DE" {
		runtime.Error("Unexpected error: %s", err)
	}

	delta := time.Since(now)

	runtime.InDelta(delta, maximum, "Engine took way too long to shutdown")

	runtime.HasLifecycle(hook1, "hook1")
	runtime.HasInvoked(hook2, "hook2")
	runtime.HasLifecycle(hook3, "hook3")

	runtime.Log("Shutdown was successful.")

}

func ShutdownWithHookPanicOnStop(runtime *TestRuntime) {

	kill := 200 * time.Millisecond
	epsilon := 20 * time.Millisecond
	maximum := 400 * time.Millisecond

	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, Timeout(kill))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	createOk := func() *testHook {
		return &testHook{
			kill: make(chan struct{}, 1),
		}
	}

	createErr := func() *testHook {
		return &testHook{
			panicOnStop: true,
			kill:        make(chan struct{}, 1),
		}
	}

	hook1 := createOk()
	hook2 := createErr()
	hook3 := createOk()

	engine.Register(hook1)
	engine.Register(hook2)
	engine.Register(hook3)

	now := time.Now()
	err = engine.Start()
	if err != nil {
		runtime.Error("Unexpected error: %s", err)
	}

	delta := time.Since(now)

	runtime.InEpsilon(delta, maximum, epsilon, "Engine took way too long to shutdown")

	runtime.HasLifecycle(hook1, "hook1")
	runtime.HasKill(hook2, "hook2")
	runtime.HasLifecycle(hook3, "hook3")

	runtime.Log("Shutdown was successful.")

}
