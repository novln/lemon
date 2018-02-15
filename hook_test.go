package lemon

import (
	"context"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestHook(t *testing.T) {
	tests := map[string]TestHandler{
		"Lifecycle":              HookLifecycle,
		"BeforeShutdown/Context": HookBeforeShutdownWithContext,
		"AfterShutdown/Context":  HookAfterShutdownWithContext,
		"BeforeShutdown/Signal":  HookBeforeShutdownWithSignal,
		"AfterShutdown/Signal":   HookAfterShutdownWithSignal,
	}

	for name, handler := range tests {
		t.Run(name, Setup(handler))
	}
}

func HookLifecycle(runtime *TestRuntime) {

	kill := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx)
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	hook := &testHook{}
	hook.kill = make(chan struct{}, 1)

	engine.Register(hook)

	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	runtime.HasLifecycle(hook, "hook")

	runtime.Log("Engine has started then stopped given Hook")

}

func HookBeforeShutdownWithContext(runtime *TestRuntime) {

	before := int64(0)

	kill := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, BeforeShutdown(func() {
		runtime.Log("Engine has executed BeforeShutdown hook.")
		atomic.StoreInt64(&before, 1)
	}))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	hook1 := &testHook{}
	hook1.kill = make(chan struct{}, 1)

	hook2 := &testHook{}
	hook2.kill = make(chan struct{}, 1)

	engine.Register(hook1)
	engine.Register(hook2)

	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	if atomic.LoadInt64(&before) == 0 {
		runtime.Error("Engine should have executed BeforeShutdown hook.")
	}

	runtime.Log("Engine has executed BeforeShutdown hook.")

}

func HookAfterShutdownWithContext(runtime *TestRuntime) {

	after := int64(0)

	kill := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, AfterShutdown(func() {
		runtime.Log("Engine has executed AfterShutdown hook.")
		atomic.StoreInt64(&after, 1)
	}))
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	hook1 := &testHook{}
	hook1.kill = make(chan struct{}, 1)

	hook2 := &testHook{}
	hook2.kill = make(chan struct{}, 1)

	engine.Register(hook1)
	engine.Register(hook2)

	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	if atomic.LoadInt64(&after) == 0 {
		runtime.Error("Engine should have executed AfterShutdown hook.")
	}

	runtime.Log("Engine has executed AfterShutdown hook.")

}

func HookBeforeShutdownWithSignal(runtime *TestRuntime) {

	before := int64(0)

	engine, err := New(runtime.Context(), BeforeShutdown(func() {
		runtime.Log("Engine has executed BeforeShutdown hook.")
		atomic.StoreInt64(&before, 1)
	}))

	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	hook1 := &testHook{}
	hook1.kill = make(chan struct{}, 1)

	hook2 := &testHook{}
	hook2.kill = make(chan struct{}, 1)

	engine.Register(hook1)
	engine.Register(hook2)

	engine.interrupt = make(chan os.Signal, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		engine.interrupt <- syscall.SIGINT
	}()

	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	if atomic.LoadInt64(&before) == 0 {
		runtime.Error("Engine should have executed BeforeShutdown hook.")
	}

	runtime.Log("Engine has executed BeforeShutdown hook.")

}

func HookAfterShutdownWithSignal(runtime *TestRuntime) {

	after := int64(0)

	engine, err := New(runtime.Context(), AfterShutdown(func() {
		runtime.Log("Engine has executed AfterShutdown hook.")
		atomic.StoreInt64(&after, 1)
	}))

	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	hook1 := &testHook{}
	hook1.kill = make(chan struct{}, 1)

	hook2 := &testHook{}
	hook2.kill = make(chan struct{}, 1)

	engine.Register(hook1)
	engine.Register(hook2)

	engine.interrupt = make(chan os.Signal, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		engine.interrupt <- syscall.SIGINT
	}()

	err = engine.Start()
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}

	if atomic.LoadInt64(&after) == 0 {
		runtime.Error("Engine should have executed AfterShutdown hook.")
	}

	runtime.Log("Engine has executed AfterShutdown hook.")

}
