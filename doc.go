// Package lemon provides an engine to manage your components lifecycle.
//
// In order to hook your components with the engine lifecycle, you must define a startup (Start) and a
// shutdown mechanism (Stop).
//
// The engine will start every registered hook (or daemon, service, etc...) and block until it receives
// a signal (SIGINT, SIGTERM and SIGQUIT for example) or when the parent context is terminated.
//
// Start and Stop will be executed in separated goroutine, so be very carreful with any race conditions or deadlocks.
//
package lemon
