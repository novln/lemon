package lemon

import (
	"context"
	"errors"
	"testing"
)

func TestOption(t *testing.T) {
	tests := map[string]TestHandler{
		"WithError": OptionWithError,
	}

	for name, handler := range tests {
		t.Run(name, Setup(handler))
	}
}

type errOption struct {
	err error
}

func (o errOption) apply(e *Engine) error {
	return o.err
}

func OptionWithError(runtime *TestRuntime) {

	expected := errors.New("cannot update engine with foobar")
	option := errOption{
		err: expected,
	}

	engine, err := New(context.Background(), option)

	if err == nil {
		runtime.Error("An error was expected")
	}

	if err != expected {
		runtime.Error("Unexpected error: %s", err)
	}

	if engine != nil {
		runtime.Error("Engine should be undefined")
	}

	runtime.Log("We received expected error: %s.", err)

}
