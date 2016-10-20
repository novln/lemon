package lemon

// Option is used to set options for the engine.
type Option interface {
	apply(*Engine) error
}

type option struct {
	callback func(*Engine) error
}

func (o option) apply(e *Engine) error {
	return o.callback(e)
}

func wrapOption(f func(*Engine) error) Option {
	return option{f}
}
