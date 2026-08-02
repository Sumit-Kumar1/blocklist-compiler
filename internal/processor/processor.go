package processor

import (
	"blc/internal/models"
	"blc/internal/transformations"
	"os"
	"strings"
)

// BlocklistProcessor handles processing blocklists using transformations
type BlocklistProcessor struct {
	factory *transformations.TransformationFactory
}

// NewBlocklistProcessor creates a new processor with default factory
func NewBlocklistProcessor() *BlocklistProcessor {
	return &BlocklistProcessor{
		factory: &transformations.TransformationFactory{},
	}
}

// Process applies transformations to a single blocklist
func (p *BlocklistProcessor) Process(blocklist *models.Blocklist, globalTransforms []string) ([]byte, error) {
	// Read the blocklist data from temp file
	data, err := os.ReadFile(blocklist.TempName)
	if err != nil {
		return nil, err
	}

	// Create pipeline for blocklist-specific transformations
	specificPipeline := &transformations.TransformationPipeline{}
	for _, tName := range blocklist.Transformations {
		t, err := p.factory.Create(tName)
		if err != nil {
			return nil, err
		}
		specificPipeline.Add(t)
	}

	// Apply blocklist-specific transformations
	data, err = specificPipeline.Apply(data)
	if err != nil {
		return nil, err
	}

	// Create pipeline for global transformations
	globalPipeline := &transformations.TransformationPipeline{}
	for _, tName := range globalTransforms {
		t, err := p.factory.Create(tName)
		if err != nil {
			return nil, err
		}
		globalPipeline.Add(t)
	}

	// Apply global transformations
	data, err = globalPipeline.Apply(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ProcessAll processes all blocklists in a config and writes consolidated output
func (p *BlocklistProcessor) ProcessAll(cfg *models.BlocklistConfig) error {
	var combined strings.Builder

	for i, blocklist := range cfg.Blocklists {
		// Process this blocklist
		processedData, err := p.Process(&blocklist, cfg.Transformations)
		if err != nil {
			return err
		}

		// Add separator between blocklists (except first)
		if i > 0 {
			combined.WriteString("\n! ----\n\n")
		}

		combined.Write(processedData)
	}

	// Write final consolidated blocklist
	return os.WriteFile("output.txt", []byte(combined.String()), 0644)
}
