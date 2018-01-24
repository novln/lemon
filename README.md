# Lemon

[![Documentation][godoc-img]][godoc-url]
![License][license-img]
[![Build Status][travis-img]][travis-url]
[![Coverage Status][coverage-img]][coverage-url]
[![Report Status][goreport-img]][goreport-url]

An engine to manage your components lifecycle.

[![Lemon][lemon-img]][lemon-url]

## Introduction

Lemon is an engine that manage your components lifecycle using a startup and shutdown mechanism.

It will start every registered hook _(or daemon, service, etc...)_ and block until it receives a signal
(**SIGINT**, **SIGTERM** and **SIGQUIT** for example) or when a parent context _(if provided)_ is terminated...

> **NOTE:** startup and shutdown procedure will be executed in separated goroutine: so be very careful with
any race conditions or deadlocks.

## Example

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/november-eleven/lemon"
)

type Ping struct {
	kill chan struct{}
}

func (p *Ping) Start() error {

	if p.kill == nil {
		p.kill = make(chan struct{}, 1)
	}

	for {
		select {
		case <-p.kill:
			return nil
		case <-time.After(2 * time.Second):
			fmt.Println("Ping")
		}
	}
}

func (p *Ping) Stop() error {
	if p.kill != nil {
		p.kill <- struct{}{}
	}
	return nil
}

func main() {

	t := 5 * time.Second
	e, err := lemon.New(lemon.Timeout(t), lemon.Logger(func(err error) {
		fmt.Fprintln(os.Stderr, err)
	}))

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(255)
	}

	e.Register(&Ping{})
	e.Start()

}

```

## Versioning or Vendoring

Expect compatibility break from `master` branch.
Copy and/or Fork are highly recommended.

> **NOTE:** semver tags or branches could be provided, if needed.

## License

This is Free Software, released under the [`MIT License`](LICENSE).

[lemon-url]: https://github.com/november-eleven/lemon
[lemon-img]: https://raw.githubusercontent.com/november-eleven/lemon/master/lemon.png
[godoc-url]: https://godoc.org/github.com/november-eleven/lemon
[godoc-img]: https://godoc.org/github.com/november-eleven/lemon?status.svg
[license-img]: https://img.shields.io/badge/license-MIT-blue.svg
[travis-url]: https://travis-ci.org/november-eleven/lemon
[travis-img]: https://travis-ci.org/november-eleven/lemon.svg?branch=master
[coverage-url]: https://coveralls.io/github/november-eleven/lemon?branch=master
[coverage-img]: https://coveralls.io/repos/github/november-eleven/lemon/badge.svg?branch=master
[goreport-url]: https://goreportcard.com/report/november-eleven/lemon
[goreport-img]: https://goreportcard.com/badge/november-eleven/lemon
