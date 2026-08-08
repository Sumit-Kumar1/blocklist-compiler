package processor

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"blc/internal/models"
	"blc/internal/transformations"
)

const (
	outputPerm = 0o644
	// separator marks the boundary between two source lists in the output
	separator = "\n! ----\n\n"
)

// ErrTooFewRules reports a compile that fell below the configured sanity floor.
var ErrTooFewRules = errors.New("compiled list is below the minimum rule count")

type Metrics struct {
	transformationsTotal atomic.Int64
	transformationsTime  atomic.Int64
	blocklistsProcessed  atomic.Int64
}

// Options configures how a compiled list is published.
type Options struct {
	// OutputFile is replaced atomically once a compile passes the rule floor.
	OutputFile string
	// MinRules discards a compile producing fewer rules, keeping the previous list.
	MinRules int
}

type BlocklistProcessor struct {
	factory *transformations.TransformationFactory
	metrics *Metrics
	opts    Options
}

func NewBlocklistProcessor(opts Options) *BlocklistProcessor {
	return &BlocklistProcessor{
		factory: &transformations.TransformationFactory{},
		metrics: &Metrics{},
		opts:    opts,
	}
}

// Process applies a single blocklist's own transformations. Global
// transformations are deliberately left to ProcessAll, which runs them across
// the merged list so that cross-source duplicates and redundant subdomains are
// actually removed.
func (p *BlocklistProcessor) Process(ctx context.Context, blocklist *models.Blocklist) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, models.ErrCtxCancelled("blocklistProcessor", blocklist.Name)
	}

	// Read the blocklist data from temp file
	data, err := os.ReadFile(blocklist.TempName)
	if err != nil {
		return nil, models.BlockListError{Reason: err.Error(), BlocklistName: blocklist.Name}
	}

	pipeline, err := p.buildPipeline(blocklist.Transformations)
	if err != nil {
		return nil, err
	}

	return p.applyWithMetrics(ctx, pipeline, data)
}

func (p *BlocklistProcessor) buildPipeline(names []string) (*transformations.TransformationPipeline, error) {
	pipeline := &transformations.TransformationPipeline{}

	for _, name := range names {
		t, err := p.factory.Create(name)
		if err != nil {
			return nil, fmt.Errorf("creating transformation %q: %w", name, err)
		}

		pipeline.Add(t)
	}

	return pipeline, nil
}

func (p *BlocklistProcessor) applyWithMetrics(
	ctx context.Context, pipeline *transformations.TransformationPipeline, data []byte,
) ([]byte, error) {
	start := time.Now()

	data, err := pipeline.Apply(ctx, data)
	if err != nil {
		return nil, err
	}

	elapsed := time.Since(start)

	p.metrics.transformationsTotal.Add(int64(len(pipeline.Transformations)))
	p.metrics.transformationsTime.Add(elapsed.Microseconds())

	slog.LogAttrs(ctx, slog.LevelInfo, "transformations applied",
		slog.Int("count", len(pipeline.Transformations)),
		slog.Int64("duration_us", elapsed.Microseconds()))

	return data, nil
}

// ProcessAll transforms every blocklist, applies the config's exclusions, and
// publishes the result only when it clears the rule floor. A failed or
// undersized compile leaves the previously published list untouched.
func (p *BlocklistProcessor) ProcessAll(ctx context.Context, config *models.BlocklistConfig) error {
	exclusions, err := transformations.NewExclusionFilter(config.Exclusions)
	if err != nil {
		return err
	}

	var combined bytes.Buffer

	for i := range config.Blocklists {
		blocklist := &config.Blocklists[i]

		if ctx.Err() != nil {
			return models.ErrCtxCancelled("blocklist transformation processing", blocklist.Name)
		}

		processedData, err := p.Process(ctx, blocklist)
		if err != nil {
			slog.ErrorContext(ctx, "error processing blocklist", "name", blocklist.Name, "error", err.Error())
			return err
		}

		// Add separator between blocklists (except first)
		if i > 0 {
			combined.WriteString(separator)
		}

		combined.Write(processedData)

		p.metrics.blocklistsProcessed.Add(1)
		slog.InfoContext(ctx, "blocklist processed", "name", blocklist.Name, "index", i)
	}

	// Global transformations run across the merged list so that duplicates and
	// redundant subdomains spanning two sources are removed.
	globalPipeline, err := p.buildPipeline(config.Transformations)
	if err != nil {
		return err
	}

	merged, err := p.applyWithMetrics(ctx, globalPipeline, combined.Bytes())
	if err != nil {
		return err
	}

	published, err := exclusions.Apply(ctx, merged)
	if err != nil {
		return err
	}

	p.logMetrics(ctx)

	return p.publish(ctx, published)
}

// publish writes data to a temp file alongside the output and renames it into
// place, so a reader never observes a partially written list.
func (p *BlocklistProcessor) publish(ctx context.Context, data []byte) error {
	rules := countRules(data)
	if rules < p.opts.MinRules {
		slog.LogAttrs(ctx, slog.LevelError, "keeping previous list",
			slog.Int("rules", rules), slog.Int("minimum", p.opts.MinRules))

		return fmt.Errorf("%w: %d < %d", ErrTooFewRules, rules, p.opts.MinRules)
	}

	dir := filepath.Dir(p.opts.OutputFile)

	tmp, err := os.CreateTemp(dir, filepath.Base(p.opts.OutputFile)+".*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp output: %w", err)
	}

	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("writing temp output: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp output: %w", err)
	}

	if err := os.Chmod(tmpName, outputPerm); err != nil {
		return fmt.Errorf("setting output permissions: %w", err)
	}

	if err := os.Rename(tmpName, p.opts.OutputFile); err != nil {
		return fmt.Errorf("publishing output file: %w", err)
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "published list",
		slog.String("path", p.opts.OutputFile), slog.Int("rules", rules))

	return nil
}

// countRules counts lines that are neither blank nor comments.
func countRules(data []byte) int {
	var rules int

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), 4<<20)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "#") {
			continue
		}

		rules++
	}

	return rules
}

func (p *BlocklistProcessor) logMetrics(ctx context.Context) {
	slog.LogAttrs(ctx, slog.LevelInfo, "run metrics",
		slog.Int64("blocklists_processed", p.metrics.blocklistsProcessed.Load()),
		slog.Int64("transformations_total", p.metrics.transformationsTotal.Load()),
		slog.Int64("transformations_us", p.metrics.transformationsTime.Load()))
}
