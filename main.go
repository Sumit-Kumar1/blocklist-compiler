package main

import (
	"context"
	"os"
	"os/signal"

	"bcl/internal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := internal.Run(ctx); err != nil {
		return
	}
}
