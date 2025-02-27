package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	goplugin "plugin"

	"github.com/yourusername/grp-cli/pkg/plugin"
)

// Manager handles plugin discovery, loading and execution
type Manager struct {
	registry       map[string]plugin.Plugin
	pluginDir      string
	mutex          sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager(pluginDir string) *Manager {
	return &Manager{
		registry:       make(map[string]plugin.Plugin),
		pluginDir:      pluginDir,
	}
}

// LoadPlugins discovers and loads all plugins from the plugin directory
func (pm *Manager) LoadPlugins() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	// Ensure plugin directory exists
	if _, err := os.Stat(pm.pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin directory does not exist: %s", pm.pluginDir)
	}
	
	// Find all .so files in the plugin directory
	files, err := filepath.Glob(filepath.Join(pm.pluginDir, "*.so"))
	if err != nil {
		return fmt.Errorf("failed to search plugin directory: %w", err)
	}
	
	// Load each plugin
	for _, file := range files {
		if err := pm.loadPlugin(file); err != nil {
			fmt.Printf("Warning: Failed to load plugin %s: %v\n", file, err)
			continue
		}
	}
	
	return nil
}

// loadPlugin loads a single plugin from a .so file
func (pm *Manager) loadPlugin(path string) error {
	// Open the plugin
	plug, err := goplugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}
	
	// Look up the exported Plugin symbol
	symPlugin, err := plug.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin does not export 'Plugin' symbol: %w", err)
	}
	
	// Assert that the symbol is a Plugin
	plg, ok := symPlugin.(plugin.Plugin)
	if !ok {
		return fmt.Errorf("plugin does not implement the Plugin interface")
	}
	
	// Register the plugin
	pm.registry[plg.Name()] = plg
	
	return nil
}

// RegisterPlugin adds a plugin to the registry
func (pm *Manager) RegisterPlugin(plg plugin.Plugin) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	name := plg.Name()
	if _, exists := pm.registry[name]; exists {
		return fmt.Errorf("plugin %s is already registered", name)
	}
	
	pm.registry[name] = plg
	return nil
}

// GetPlugin retrieves a plugin by name
func (pm *Manager) GetPlugin(name string) (plugin.Plugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	plg, exists := pm.registry[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	
	return plg, nil
}

// ExecutePlugin runs a specific plugin with provided configuration
func (pm *Manager) ExecutePlugin(ctx context.Context, jobType string, config map[string]interface{}) (*plugin.Result, error) {
	// Get the plugin
	plg, err := pm.GetPlugin(jobType)
	if err != nil {
		return nil, err
	}
	
	// Get variables from context
	vars, ok := ctx.Value("variables").(map[string]interface{})
	if !ok {
		vars = make(map[string]interface{})
	}
	
	// Create a new context with variables if needed
	execCtx := ctx
	if len(vars) > 0 && ctx.Value("variables") == nil {
		execCtx = context.WithValue(ctx, "variables", vars)
	}
	
	// Validate plugin configuration
	if err := plg.Validate(execCtx, config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Execute the plugin
	result, err := plg.Execute(execCtx, config)
	return result, err
}

// ListPlugins returns all registered plugins
func (pm *Manager) ListPlugins() []plugin.Plugin {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	plugins := make([]plugin.Plugin, 0, len(pm.registry))
	for _, plg := range pm.registry {
		plugins = append(plugins, plg)
	}
	
	return plugins
} 