package lemon

import (
	"syscall"
	"testing"
)

func TestSignal(t *testing.T) {
	tests := map[string]TestHandler{
		"AddOption":     SignalAddOption,
		"DisableOption": SignalDisableOption,
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

func SignalDisableOption(runtime *TestRuntime) {

	engine, err := New(runtime.Context(),
		DisableSignal(),
	)
	if err != nil {
		runtime.Error("An error wasn't expected: %s", err)
	}
	if engine == nil {
		runtime.Error("Engine must be defined")
	}

	if len(engine.signals) != 0 {
		runtime.Error("Engine should not listen on any signals")
	}

	runtime.Log("Engine's configuration do not have signal listener.")

}
