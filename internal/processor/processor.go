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
	"strconv"
	"strings"
	"time"

	"blc/internal/models"
	"blc/internal/transformations"
)

const (
	outputPerm = 0o644
	// separator marks the boundary between two source lists in the output
	separatorLine = "! ----"
	separator     = "\n" + separatorLine + "\n\n"
)

// ErrTooFewRules reports a compile that fell below the configured sanity floor.
var ErrTooFewRules = errors.New("compiled list is below the minimum rule count")

// Options configures how a compiled list is published.
type Options struct {
	// OutputFile is replaced atomically once a compile passes the rule floor.
	OutputFile string
	// MinRules discards a compile producing fewer rules, keeping the previous list.
	MinRules int
}

type BlocklistProcessor struct {
	opts Options
}

func NewBlocklistProcessor(opts Options) *BlocklistProcessor {
	return &BlocklistProcessor{opts: opts}
}

// Process applies a single blocklist's own transformations and filters. Global
// transformations are deliberately left to ProcessAll, which runs them across
// the merged list so that cross-source duplicates and redundant subdomains are
// actually removed.
func (p *BlocklistProcessor) Process(ctx context.Context, blocklist *models.Blocklist) ([]byte, error) {
	pipeline, err := buildPipeline(blocklist.Transformations)
	if err != nil {
		return nil, err
	}

	data, err := apply(ctx, pipeline, blocklist.Data)
	if err == nil {
		data, err = applyFilters(ctx, data, blocklist.Filters)
	}

	if err != nil {
		return nil, models.BlockListError{BlocklistName: blocklist.Name, Err: err}
	}

	return data, nil
}

// applyFilters drops excluded rules, then keeps only included ones.
func applyFilters(ctx context.Context, data []byte, filters models.Filters) ([]byte, error) {
	exclusions, err := transformations.NewExclusionFilter(filters.Exclusions)
	if err != nil {
		return nil, err
	}

	inclusions, err := transformations.NewInclusionFilter(filters.Inclusions)
	if err != nil {
		return nil, err
	}

	if data, err = exclusions.Apply(ctx, data); err != nil {
		return nil, err
	}

	return inclusions.Apply(ctx, data)
}

func buildPipeline(names []string) (*transformations.TransformationPipeline, error) {
	pipeline := &transformations.TransformationPipeline{}

	for _, name := range names {
		t, err := transformations.New(name)
		if err != nil {
			return nil, fmt.Errorf("creating transformation %q: %w", name, err)
		}

		pipeline.Add(t)
	}

	return pipeline, nil
}

func apply(ctx context.Context, pipeline *transformations.TransformationPipeline, data []byte) ([]byte, error) {
	start := time.Now()

	data, err := pipeline.Apply(ctx, data)
	if err != nil {
		return nil, err
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "transformations applied",
		slog.Int("count", len(pipeline.Transformations)),
		slog.Int64("duration_us", time.Since(start).Microseconds()))

	return data, nil
}

// ProcessAll transforms every blocklist, applies the config's filters, and
// publishes the result only when it clears the rule floor. A failed or
// undersized compile leaves the previously published list untouched.
func (p *BlocklistProcessor) ProcessAll(ctx context.Context, config *models.BlocklistConfig) error {
	start := time.Now()

	var combined bytes.Buffer

	for i := range config.Blocklists {
		blocklist := &config.Blocklists[i]

		if err := ctx.Err(); err != nil {
			return err
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

		// The raw download is no longer needed; let it be collected.
		blocklist.Data = nil

		slog.InfoContext(ctx, "blocklist processed", "name", blocklist.Name, "index", i)
	}

	// Global transformations run across the merged list so that duplicates and
	// redundant subdomains spanning two sources are removed.
	globalPipeline, err := buildPipeline(config.Transformations)
	if err != nil {
		return err
	}

	merged, err := apply(ctx, globalPipeline, combined.Bytes())
	if err != nil {
		return err
	}

	filtered, err := applyFilters(ctx, merged, config.Filters)
	if err != nil {
		return err
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "run metrics",
		slog.Int("blocklists_processed", len(config.Blocklists)),
		slog.Int64("duration_us", time.Since(start).Microseconds()))

	return p.publish(ctx, config, dropEmptySections(filtered))
}

// dropEmptySections removes source separators with no rule after them, as
// left behind when global deduplication empties a whole source.
func dropEmptySections(data []byte) []byte {
	var (
		out     bytes.Buffer
		pending [][]byte
		sep     bool
		sepAt   int
		emitted bool
	)

	out.Grow(len(data))

	for line := range bytes.Lines(data) {
		switch text := bytes.TrimRight(line, "\r\n"); {
		case len(text) == 0:
			pending = append(pending, line)
		case string(text) == separatorLine && sep:
			// Back-to-back separators surround an emptied source; keep only the first.
			pending = pending[:sepAt+1]
		case string(text) == separatorLine:
			sepAt, sep = len(pending), true
			pending = append(pending, line)
		default:
			// A separator with nothing before it is dropped too.
			if !sep || emitted {
				out.Write(bytes.Join(pending, nil))
			}

			pending, sep, emitted = pending[:0], false, true

			out.Write(line)
		}
	}

	if !sep {
		out.Write(bytes.Join(pending, nil))
	}

	return out.Bytes()
}

// header describes the list the way @adguard/hostlist-compiler does, which is
// what AdGuard Home reads its display title from.
func header(config *models.BlocklistConfig, rules int) []byte {
	var buf bytes.Buffer

	buf.WriteString("!\n")

	for _, field := range [][2]string{
		{"Title", config.Name},
		{"Description", config.Description},
		{"Homepage", config.Homepage},
		{"License", config.License},
		{"Version", config.Version},
		{"Last modified", time.Now().UTC().Format(time.RFC3339)},
		{"Rule count", strconv.Itoa(rules)},
	} {
		if field[1] != "" {
			fmt.Fprintf(&buf, "! %s: %s\n", field[0], field[1])
		}
	}

	buf.WriteString("!\n! Compiled by blocklist-compiler\n!\n")

	return buf.Bytes()
}

// publish writes the header and data to a temp file alongside the output and
// renames it into place, so a reader never observes a partially written list.
func (p *BlocklistProcessor) publish(ctx context.Context, config *models.BlocklistConfig, data []byte) error {
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

	if _, err := tmp.Write(append(header(config, rules), data...)); err != nil {
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
