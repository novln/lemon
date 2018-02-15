# Lemon

[![Documentation][godoc-img]][godoc-url]
![License][license-img]
[![Build Status][travis-img]][travis-url]
[![Report Status][goreport-img]][goreport-url]

An engine to manage your components lifecycle.

[![Lemon][lemon-img]][lemon-url]

## Introduction

Lemon is an engine that manage your components lifecycle using a startup and shutdown mechanism.

It will start every registered hook _(or daemon, service, etc...)_ and block until it receives a signal
(**SIGINT**, **SIGTERM** and **SIGQUIT** for example) or when a parent context is terminated...

> **NOTE:** startup and shutdown procedure will be executed in separated goroutine: so be very careful with
any race conditions or deadlocks.

## Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/novln/lemon"
)

// Let's define a simple Ping hook...
type Ping struct {}

// Start will be executed when lemon's engine will try to start this Hook.
// Your hook can perfectly use the given context (see example), or any blocking operation...
func (p *Ping) Start(ctx context.Context) error {

	fmt.Println("[ping] Start")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
			fmt.Println("[ping] Echo Request")
		}
	}
}

// However, if you don't use <-ctx.Done(), you must cancel your blocking operation in Stop.
func (p *Ping) Stop(ctx context.Context) error {
	fmt.Println("[ping] Stop")
	return nil
}

func main() {

	timeout := 5 * time.Second
	ctx := context.Background()

	engine, err := lemon.New(ctx, lemon.Timeout(timeout), lemon.Logger(func(err error) {
		fmt.Fprintln(os.Stderr, err)
	}))

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(255)
	}

	engine.Register(&Ping{})
	engine.Start()

}

```

## Versioning or Vendoring

Expect compatibility break from `master` branch.

Using [Go dependency management tool](https://github.com/golang/dep) is **highly recommended**.

> **NOTE:** semver tags or branches could be provided, if needed.

## License

This is Free Software, released under the [`MIT License`](LICENSE).

[lemon-url]: https://github.com/novln/lemon
[lemon-img]: https://raw.githubusercontent.com/novln/lemon/master/lemon.png
[godoc-url]: https://godoc.org/github.com/novln/lemon
[godoc-img]: https://godoc.org/github.com/novln/lemon?status.svg
[license-img]: https://img.shields.io/badge/license-MIT-blue.svg
[travis-url]: https://travis-ci.org/novln/lemon
[travis-img]: https://travis-ci.org/novln/lemon.svg?branch=master
[goreport-url]: https://goreportcard.com/report/novln/lemon
[goreport-img]: https://goreportcard.com/badge/novln/lemon
