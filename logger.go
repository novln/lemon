package lemon

// Logger sets an optional error handler.
// Use this Option if you want to receives errors that occurs during startup or shutdown.
// Also, an engine's internal mutex avoid race conditions.
func Logger(handler func(err error)) Option {
	return wrapOption(func(e *Engine) error {
		e.logger = handler
		return nil
	})
}

// log will forward errors that occurs during startup or shutdown, if a logger is defined.
func (e *Engine) log(err error) {
	if err != nil && e.logger != nil {
		e.logger(err)
	}
}
