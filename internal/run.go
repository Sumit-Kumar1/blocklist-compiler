package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"blc/internal/config"
	"blc/internal/models"
	"blc/internal/processor"
)

func Run(ctx context.Context) error {
	var bc models.BlocklistConfig
	var processor = processor.NewBlocklistProcessor()

	cfg := config.LoadConfig()

	ff := filepath.Join(cfg.Config.Path, cfg.Config.Name)

	if !filepath.IsLocal(ff) {
		return fmt.Errorf("filepath %s is not local path", ff)
	}

	cfgData, err := os.ReadFile(ff)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(cfgData, &bc); err != nil {
		return err
	}

	if err := bc.Validate(); err != nil {
		return err
	}

	if err := bc.Process(); err != nil {
		return err
	}

	return processor.ProcessAll(&bc)
}
