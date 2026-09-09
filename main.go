package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"blc/internal"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGABRT, syscall.SIGKILL)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	if err := internal.Run(ctx); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error in run", slog.String("error", err.Error()))

		return 1
	}

	return 0
}
