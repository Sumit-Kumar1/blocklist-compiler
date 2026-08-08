package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"blc/internal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	if err := internal.Run(ctx); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error in run", slog.String("error", err.Error()))
		return
	}
}
