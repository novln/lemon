package lemon

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestHook(t *testing.T) {
	tests := map[string]TestHandler{
		"Lifecycle":              Lifecycle,
		"BeforeShutdown/Context": BeforeShutdownHookWithContext,
		"AfterShutdown/Context":  AfterShutdownHookWithContext,
		"BeforeShutdown/Signal":  BeforeShutdownHookWithSignal,
		"AfterShutdown/Signal":   AfterShutdownHookWithContext,
	}

	for name, handler := range tests {
		t.Run(name, Setup(handler))
	}
}

func Lifecycle(runtime *TestRuntime) {

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

	if !hook.startCalled {
		runtime.Error("Engine should have start given Hook")
	}

	if !hook.stopCalled {
		runtime.Error("Engine should have stop given Hook")
	}

	runtime.Log("Engine has started then stopped given Hook")

}

func BeforeShutdownHookWithContext(runtime *TestRuntime) {

	before := false

	kill := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, BeforeShutdown(func() {
		runtime.Log("Engine has executed BeforeShutdown hook.")
		before = true
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

	if !before {
		runtime.Error("Engine should have executed BeforeShutdown hook.")
	}

	runtime.Log("Engine has executed BeforeShutdown hook.")

}

func AfterShutdownHookWithContext(runtime *TestRuntime) {

	after := false

	kill := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(runtime.Context(), kill)
	defer cancel()

	engine, err := New(ctx, AfterShutdown(func() {
		runtime.Log("Engine has executed AfterShutdown hook.")
		after = true
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

	if !after {
		runtime.Error("Engine should have executed AfterShutdown hook.")
	}

	runtime.Log("Engine has executed AfterShutdown hook.")

}

func BeforeShutdownHookWithSignal(runtime *TestRuntime) {

	before := false

	engine, err := New(runtime.Context(), BeforeShutdown(func() {
		runtime.Log("Engine has executed BeforeShutdown hook.")
		before = true
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

	if !before {
		runtime.Error("Engine should have executed BeforeShutdown hook.")
	}

	runtime.Log("Engine has executed BeforeShutdown hook.")

}

func AfterShutdownHookWithSignal(runtime *TestRuntime) {

	after := false

	engine, err := New(runtime.Context(), AfterShutdown(func() {
		runtime.Log("Engine has executed AfterShutdown hook.")
		after = true
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

	if !after {
		runtime.Error("Engine should have executed AfterShutdown hook.")
	}

	runtime.Log("Engine has executed AfterShutdown hook.")

}
