package lemon

import (
	"fmt"
	"testing"
)

type errOption struct {
	err error
}

func (o errOption) apply(e *Engine) error {
	return o.err
}

func TestOptionWithError(t *testing.T) {

	i := fmt.Errorf("Cannot update Engine with foo")
	o := errOption{i}

	_, err := New(o)

	if err != i {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err == nil {
		t.Fatal("An error was expected")
	}

	t.Logf("We received expected error from New(): %s.", err)

}
