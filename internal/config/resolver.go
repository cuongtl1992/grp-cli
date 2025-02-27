package config

import (
	"fmt"
	"regexp"
	"strings"
)

// Resolver handles variable and reference resolution
type Resolver struct {
	// Regular expression for variable references ${...}
	refRegex *regexp.Regexp
}

// NewResolver creates a new resolver
func NewResolver() *Resolver {
	return &Resolver{
		refRegex: regexp.MustCompile(`\${([^}]+)}`),
	}
}

// ResolveValues processes all variable references in a configuration
func (r *Resolver) ResolveValues(config map[string]interface{}, context map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// Process each key-value pair
	for key, value := range config {
		resolvedValue, err := r.resolveValue(value, context)
		if err != nil {
			return nil, fmt.Errorf("error resolving value for key %s: %w", key, err)
		}
		result[key] = resolvedValue
	}
	
	return result, nil
}

// resolveValue handles a single value which might be a string, map, slice, or primitive
func (r *Resolver) resolveValue(value interface{}, context map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return r.resolveString(v, context)
	case map[string]interface{}:
		return r.ResolveValues(v, context)
	case []interface{}:
		return r.resolveSlice(v, context)
	default:
		// For primitives (int, bool, etc.), return as is
		return v, nil
	}
}

// resolveString handles variable substitution in strings
func (r *Resolver) resolveString(value string, context map[string]interface{}) (interface{}, error) {
	// Check if the entire string is a reference
	if r.refRegex.MatchString(value) && strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		// Extract reference path
		path := value[2 : len(value)-1]
		
		// Resolve the reference
		resolvedValue, err := r.resolvePath(path, context)
		if err != nil {
			return nil, err
		}
		
		// Return the resolved value with its original type
		return resolvedValue, nil
	}
	
	// Handle partial substitutions
	result := r.refRegex.ReplaceAllStringFunc(value, func(match string) string {
		// Extract reference path
		path := match[2 : len(match)-1]
		
		// Resolve the reference
		resolvedValue, err := r.resolvePath(path, context)
		if err != nil {
			// Just return the original reference if resolution fails
			return match
		}
		
		// Convert to string for interpolation
		return fmt.Sprintf("%v", resolvedValue)
	})
	
	return result, nil
}

// resolveSlice handles variable substitution in slices
func (r *Resolver) resolveSlice(slice []interface{}, context map[string]interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(slice))
	
	for i, item := range slice {
		resolvedItem, err := r.resolveValue(item, context)
		if err != nil {
			return nil, fmt.Errorf("error resolving array item %d: %w", i, err)
		}
		result[i] = resolvedItem
	}
	
	return result, nil
}

// resolvePath handles dot-notation path resolution (e.g., "variables.service.port")
func (r *Resolver) resolvePath(path string, context map[string]interface{}) (interface{}, error) {
	parts := strings.Split(path, ".")
	
	// Start with the top-level context
	var current interface{} = context
	
	// Navigate through the path
	for _, part := range parts {
		// Handle map access
		if currentMap, ok := current.(map[string]interface{}); ok {
			var exists bool
			current, exists = currentMap[part]
			if !exists {
				return nil, fmt.Errorf("reference path not found: %s", path)
			}
			continue
		}
		
		// Can't navigate further
		return nil, fmt.Errorf("invalid reference path: %s", path)
	}
	
	return current, nil
} 