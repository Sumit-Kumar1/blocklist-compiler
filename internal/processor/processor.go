package processor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"blc/internal/models"
	"blc/internal/transformations"
)

type blocklistCombined struct {
	mu       sync.Mutex
	combined strings.Builder
}

type Metrics struct {
	transformationsTotal int64
	transformationsTime  int64
	blocklistsProcessed  int64
}
type BlocklistProcessor struct {
	factory *transformations.TransformationFactory
	metrics *Metrics
}

func NewBlocklistProcessor() *BlocklistProcessor {
	return &BlocklistProcessor{
		factory: &transformations.TransformationFactory{},
		metrics: &Metrics{},
	}
}

func (p *BlocklistProcessor) Process(ctx context.Context, blocklist *models.Blocklist, globalTransforms []string) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, models.ErrCtxCancalled("blocklistProcessor", blocklist.Name)
	}

	// Read the blocklist data from temp file
	data, err := os.ReadFile(blocklist.TempName)
	if err != nil {
		return nil, models.BlockListError{Reason: err.Error(), BlocklistName: blocklist.Name}
	}

	// Create pipeline for blocklist-specific transformations
	specificPipeline := &transformations.TransformationPipeline{}
	for _, tName := range blocklist.Transformations {
		if ctx.Err() != nil {
			return nil, models.ErrCtxCancalled("blocklist transformation addition", tName)
		}

		t, err := p.factory.Create(tName)
		if err != nil {
			return nil, fmt.Errorf("creating transformation %q: %w", tName, err)
		}
		specificPipeline.Add(t)
	}

	// Apply blocklist-specific transformations with metrics
	data, err = p.applyTransformationsWithMetrics(ctx, specificPipeline, data)
	if err != nil {
		return nil, fmt.Errorf("error while applying specificPipeline: %w", err)
	}

	// Create pipeline for global transformations
	globalPipeline := &transformations.TransformationPipeline{}
	for _, tName := range globalTransforms {
		if ctx.Err() != nil {
			return nil, models.ErrCtxCancalled("processing global transformations", tName)
		}

		t, err := p.factory.Create(tName)
		if err != nil {
			return nil, fmt.Errorf("creating transformation %q: %w", tName, err)
		}
		globalPipeline.Add(t)
	}

	// Apply global transformations with metrics
	data, err = p.applyTransformationsWithMetrics(ctx, globalPipeline, data)
	if err != nil {
		return nil, fmt.Errorf("error while applyying globalPipeline: %w", err)
	}

	return data, nil
}

func (p *BlocklistProcessor) applyTransformationsWithMetrics(ctx context.Context, tp *transformations.TransformationPipeline, data []byte) ([]byte, error) {
	start := time.Now()
	var transformedData []byte
	var err error

	for _, t := range tp.Transformations {
		if ctx.Err() != nil {
			return nil, models.ErrCtxCancalled("transformation execution", t.Name())
		}

		// Apply transformation
		start := time.Now()
		transformedData, err = t.Apply(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("transformation %q failed: %w", t.Name(), err)
		}

		duration := time.Since(start)

		// Update metrics
		atomic.AddInt64(&p.metrics.transformationsTotal, 1)
		atomic.AddInt64(&p.metrics.transformationsTime, int64(duration.Microseconds()))

		data = transformedData
	}

	// Update block processing metrics
	atomic.AddInt64(&p.metrics.blocklistsProcessed, 1)
	slog.InfoContext(ctx, "transformations completed blocklist combined", slog.Int64("duration", time.Since(start).Microseconds()))

	return data, nil
}

func (p *BlocklistProcessor) ProcessAll(ctx context.Context, config *models.BlocklistConfig) error {
	combined := &blocklistCombined{
		combined: strings.Builder{},
	}

	for i, blocklist := range config.Blocklists {
		if ctx.Err() != nil {
			return models.ErrCtxCancalled("blocklist transformation processing", blocklist.Name)
		}

		processedData, err := p.Process(ctx, &blocklist, config.Transformations)
		if err != nil {
			slog.ErrorContext(ctx, "error processing blocklist", "name", blocklist.Name, "error", err.Error())
			return err
		}

		// Add separator between blocklists (except first)
		if i > 0 {
			combined.mu.Lock()
			combined.combined.WriteString("\n! ----\n\n")
			combined.mu.Unlock()
		}

		combined.mu.Lock()
		combined.combined.Write(processedData)
		combined.mu.Unlock()

		atomic.AddInt64(&p.metrics.blocklistsProcessed, 1)
		slog.InfoContext(ctx, "blocklist processed", "name", blocklist.Name, "index", i)
	}

	// Write final consolidated blocklist
	if combined.combined.Len() > 0 {
		if err := os.WriteFile("output.txt", []byte(combined.combined.String()), 0644); err != nil {
			return fmt.Errorf("error writing output file: %w", err)
		}
	}

	return nil
}
