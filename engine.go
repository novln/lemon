package lemon

import (
	"context"
	"os"
	"sync"
	"time"
)

// Engine is a components that handle your application lifecycle.
//
// Start
//
// It will start every registered hook (or daemon, service, etc...) and block until it
// receives a SIGINT, SIGTERM or SIGQUIT signal.
//
// For example:
//
//   e, err := New(ctx)
//   if err == nil {
//       e.Register(&MyService{})
//       e.Register(&MyWorker{})
//       e.Start()
//       // Wait until Ctrl-C
//   }
//
// Stop
//
// When your application has to stop, the engine will notify every hook to shutdown gracefully.
// However, if a hook fails to stop before timeout, the underlying goroutine will be destroyed...
//
type Engine struct {
	interrupt      chan os.Signal
	timeout        time.Duration
	hooks          []Hook
	wait           sync.WaitGroup
	parent         context.Context
	ctx            context.Context
	cancel         context.CancelFunc
	signals        []os.Signal
	beforeShutdown func()
	afterShutdown  func()
	logger         func(error)
	mutex          sync.Mutex
	cause          error
}

// New creates a new engine with given options.
//
// Options can change the timeout, register a signal, execute a pre-hook callback and many other behaviors.
func New(ctx context.Context, options ...Option) (*Engine, error) {

	e := &Engine{}
	e.parent = ctx
	e.init()

	for _, o := range options {
		if err := o.apply(e); err != nil {
			return nil, err
		}
	}

	return e, nil
}

// launch will start given hook.
func (e *Engine) launch(h Hook) {

	e.wait.Add(1)

	go func() {

		defer e.wait.Done()

		runtime := &HookRuntime{}

		// Wait for an event to notify this goroutine that a shutdown is required.
		// It could either be from engine's context or during Hook startup if an error has occurred.
		// NOTE: If HookRuntime returns an error, we have to shutdown every Hook...
		err := runtime.WaitForEvent(e.ctx, h)
		if err != nil {
			e.mutex.Lock()
			e.log(err)
			e.cancel()
			e.cause = err
			e.mutex.Unlock()
		}

		// Wait for hook to gracefully shutdown, or kill it after timeout.
		// This is handled by HookRuntime.
		for _, err := range runtime.Shutdown(e.timeout) {
			e.mutex.Lock()
			e.log(err)
			e.mutex.Unlock()
		}

	}()
}

// init configures default parameters for engine.
func (e *Engine) init() {

	if e.ctx == nil || e.cancel == nil {
		e.ctx, e.cancel = context.WithCancel(e.parent)
	}

	if e.timeout == 0 {
		e.timeout = DefaultTimeout
	}

	if len(e.signals) == 0 {
		e.signals = Signals
	}

	if e.interrupt == nil {
		e.interrupt = make(chan os.Signal, 1)
	}

}

// Start will launch the engine and start registered hooks.
// It will block until every hooks has shutdown, gracefully or with force...
func (e *Engine) Start() error {

	e.init()

	go e.waitShutdownNotification()

	for _, h := range e.hooks {
		e.launch(h)
	}

	e.wait.Wait()

	if e.afterShutdown != nil {
		e.afterShutdown()
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.cause

}
