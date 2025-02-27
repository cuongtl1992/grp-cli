package plugin

import (
	"context"
)

// Plugin defines the interface for all job type plugins
type Plugin interface {
	// Name returns the unique identifier for this plugin
	Name() string
	
	// Description provides human-readable information
	Description() string
	
	// Version returns the semantic version of this plugin
	Version() string
	
	// ConfigSchema returns the JSON schema for validating plugin config
	ConfigSchema() *JSONSchema
	
	// Validate verifies configuration and prerequisites
	Validate(ctx context.Context, config map[string]interface{}) error
	
	// Execute runs the plugin with the given configuration
	Execute(ctx context.Context, config map[string]interface{}) (*Result, error)
	
	// Rollback reverts changes if execution fails
	Rollback(ctx context.Context, executionID string) error
}

// JSONSchema defines a simple JSON schema for config validation
type JSONSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]*JSONSchema `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
	Items      *JSONSchema            `json:"items,omitempty"`
}

// Result represents the outcome of a plugin execution
type Result struct {
	Success     bool                   `json:"success"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Message     string                 `json:"message,omitempty"`
	Artifacts   []Artifact             `json:"artifacts,omitempty"`
	ExecutionID string                 `json:"executionId"`
}

// Artifact represents a file or data produced by a plugin
type Artifact struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	ContentType string `json:"contentType"`
	Path        string `json:"path,omitempty"`
	Data        []byte `json:"data,omitempty"`
} 