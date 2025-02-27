package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/yourusername/grp-cli/internal/models"
)

// Loader handles loading and parsing configuration files
type Loader struct {
	resolver *Resolver
	cache    map[string]interface{}
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		resolver: NewResolver(),
		cache:    make(map[string]interface{}),
	}
}

// LoadPlan loads a release plan from a file
func (l *Loader) LoadPlan(filePath string) (*models.Plan, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// Verify file exists and is readable
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plan file does not exist: %s", filePath)
		}
		return nil, fmt.Errorf("error accessing plan file: %w", err)
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Parse as YAML
	var rawPlan map[string]interface{}
	if err := yaml.Unmarshal(data, &rawPlan); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the raw plan structure before processing
	if err := l.validateRawPlan(rawPlan); err != nil {
		return nil, fmt.Errorf("invalid plan structure: %w", err)
	}
	
	// Process includes
	baseDir := filepath.Dir(filePath)
	if includes, ok := rawPlan["includes"].([]interface{}); ok {
		for _, include := range includes {
			if includeMap, ok := include.(map[string]interface{}); ok {
				if path, ok := includeMap["path"].(string); ok {
					// Resolve path relative to the plan file
					includePath := filepath.Join(baseDir, path)
					if err := l.loadInclude(includePath); err != nil {
						return nil, fmt.Errorf("failed to load include %s: %w", path, err)
					}
				}
			}
		}
	}
	
	// Resolve variable references
	resolvedPlan, err := l.resolver.ResolveValues(rawPlan, l.cache)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve variables: %w", err)
	}
	
	// Convert to structured plan
	var plan models.Plan
	resolvedData, err := yaml.Marshal(resolvedPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to re-marshal plan: %w", err)
	}
	
	if err := yaml.Unmarshal(resolvedData, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan structure: %w", err)
	}
	
	return &plan, nil
}

// loadInclude loads an included configuration file
func (l *Loader) loadInclude(filePath string) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read include file: %w", err)
	}
	
	// Parse as YAML
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("failed to parse include YAML: %w", err)
	}
	
	// Get include key based on kind
	var key string
	if kind, ok := rawConfig["kind"].(string); ok {
		key = kind
	} else {
		// Use filename as fallback
		key = filepath.Base(filePath)
	}
	
	// Store in cache
	l.cache[key] = rawConfig
	
	return nil
}

func (l *Loader) validateRawPlan(raw map[string]interface{}) error {
	required := []string{"apiVersion", "kind", "metadata"}
	for _, field := range required {
		if _, ok := raw[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}