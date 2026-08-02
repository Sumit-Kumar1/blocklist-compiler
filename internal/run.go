package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Run(ctx context.Context) error {
	var bc blocklistConfig

	cfg := loadConfig()

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

	if err := bc.validate(); err != nil {
		return err
	}

	if err := bc.process(); err != nil {
		return err
	}

	return nil
}
