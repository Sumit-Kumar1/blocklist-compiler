package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"blc/internal/config"
	"blc/internal/models"
	"blc/internal/processor"
	"blc/internal/server"
)

// Run compiles the blocklist once, then keeps recompiling every cfg.Interval.
// When a list server is configured it serves the last published list throughout.
func Run(ctx context.Context) error {
	cfg := config.Load()

	agh := server.NewAdGuard(cfg.AGHAPI, cfg.AGHUser, cfg.AGHPass)

	if cfg.Interval <= 0 {
		// One-shot mode: a failed compile is the process's exit status.
		return compile(ctx, cfg, agh)
	}

	if cfg.ListenAddr == "" {
		return loop(ctx, cfg, agh)
	}

	srv := server.New(cfg.ListenAddr, cfg.OutputFile)
	errCh := make(chan error, 1)

	serveCtx, stopServing := context.WithCancel(ctx)
	defer stopServing()

	go func() { errCh <- srv.Start(serveCtx) }()

	err := loop(ctx, cfg, agh)

	stopServing()

	if serveErr := <-errCh; serveErr != nil {
		return serveErr
	}

	return err
}

// loop compiles immediately and then on every tick until ctx is done. A failed
// cycle is logged and the previously published list is left in place.
func loop(ctx context.Context, cfg *config.Config, agh *server.AdGuard) error {
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		if err := compile(ctx, cfg, agh); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "compile cycle failed, keeping previous list",
				slog.String("error", err.Error()))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// compile runs one full fetch-transform-publish cycle.
func compile(ctx context.Context, cfg *config.Config, agh *server.AdGuard) error {
	bc, err := loadBlocklistConfig(cfg)
	if err != nil {
		return err
	}

	defer cleanup(bc)

	if err := bc.Process(ctx); err != nil {
		return err
	}

	blocklistProcessor := processor.NewBlocklistProcessor(processor.Options{
		OutputFile: cfg.OutputFile,
		MinRules:   cfg.MinRules,
	})

	if err := blocklistProcessor.ProcessAll(ctx, bc); err != nil {
		return err
	}

	if err := agh.Refresh(ctx); err != nil {
		// AdGuard re-reads the list on its own schedule, so this is not fatal.
		slog.LogAttrs(ctx, slog.LevelWarn, "adguard refresh failed", slog.String("error", err.Error()))

		return nil
	}

	if !agh.Disabled {
		slog.InfoContext(ctx, "adguard refreshed")
	}

	return nil
}

func loadBlocklistConfig(cfg *config.Config) (*models.BlocklistConfig, error) {
	configFile := filepath.Clean(filepath.Join(cfg.ConfigPath, cfg.ConfigName))

	cfgData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var bc models.BlocklistConfig
	if err := json.Unmarshal(cfgData, &bc); err != nil {
		return nil, err
	}

	if err := bc.Validate(); err != nil {
		return nil, err
	}

	return &bc, nil
}

func cleanup(bc *models.BlocklistConfig) {
	for i := range bc.Blocklists {
		if bc.Blocklists[i].TempName == "" {
			continue
		}

		if err := os.Remove(bc.Blocklists[i].TempName); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Warn("removing temp file", "path", bc.Blocklists[i].TempName, "error", err.Error())
		}
	}
}
