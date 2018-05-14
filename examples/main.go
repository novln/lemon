package main

import (
	"context"
	"fmt"
	"time"

	"github.com/novln/lemon"
)

type Ping struct {
}

func (p *Ping) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
			fmt.Println(ctx.Value("key"))
		}
	}
}

func (p *Ping) Stop(ctx context.Context) error {
	fmt.Println("Pong")
	return nil
}

func main() {

	ctx := context.Background()
	ctx = context.WithValue(ctx, "key", "Ping")
	ctx, cancel := context.WithTimeout(ctx, (10 * time.Second))
	defer cancel()

	engine, err := lemon.New(ctx)
	if err != nil {
		panic(err)
	}

	engine.Register(&Ping{})
	err = engine.Start()
	if err != nil {
		panic(err)
	}

}
