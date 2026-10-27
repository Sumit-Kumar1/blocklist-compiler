package main

import (
	"blc/internal"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	// LOG_LEVEL accepts debug, info, warn or error; anything else means info.
	var level slog.Level
	if err := level.UnmarshalText([]byte(os.Getenv("LOG_LEVEL"))); err != nil {
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	if err := internal.Run(ctx); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error in run", slog.String("error", err.Error()))

		return 1
	}

	return 0
}
