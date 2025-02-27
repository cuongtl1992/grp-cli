package plugins

import (
	"context"
	"os"
	"testing"

	"github.com/cuongtl1992/grp-cli/pkg/plugin"
)

// MockPlugin implements the Plugin interface for testing
type MockPlugin struct {
	name        string
	validateErr error
	executeErr  error
}

func (m *MockPlugin) Name() string                                       { return m.name }
func (m *MockPlugin) Description() string                               { return "Mock plugin for testing" }
func (m *MockPlugin) Version() string                                   { return "1.0.0" }
func (m *MockPlugin) ConfigSchema() *plugin.JSONSchema                  { return nil }
func (m *MockPlugin) Validate(ctx context.Context, config map[string]interface{}) error { 
	return m.validateErr 
}
func (m *MockPlugin) Execute(ctx context.Context, config map[string]interface{}) (*plugin.Result, error) {
	if m.executeErr != nil {
		return nil, m.executeErr
	}
	return &plugin.Result{Success: true}, nil
}
func (m *MockPlugin) Rollback(ctx context.Context, executionID string) error { return nil }

func TestNewManager(t *testing.T) {
	manager := NewManager("./plugins")
	if manager == nil {
		t.Fatal("Expected non-nil manager")
		return
	}
	if manager.pluginDir != "./plugins" {
		t.Errorf("Expected plugin dir to be './plugins', got %s", manager.pluginDir)
	}
}

func TestRegisterPlugin(t *testing.T) {
	manager := NewManager("./plugins")
	mockPlugin := &MockPlugin{name: "test-plugin"}

	// Test successful registration
	err := manager.RegisterPlugin(mockPlugin)
	if err != nil {
		t.Errorf("Failed to register plugin: %v", err)
	}

	// Test duplicate registration
	err = manager.RegisterPlugin(mockPlugin)
	if err == nil {
		t.Error("Expected error on duplicate registration")
	}
}

func TestGetPlugin(t *testing.T) {
	manager := NewManager("./plugins")
	mockPlugin := &MockPlugin{name: "test-plugin"}
	manager.RegisterPlugin(mockPlugin)

	// Test successful retrieval
	plugin, err := manager.GetPlugin("test-plugin")
	if err != nil {
		t.Errorf("Failed to get plugin: %v", err)
	}
	if plugin.Name() != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got %s", plugin.Name())
	}

	// Test non-existent plugin
	_, err = manager.GetPlugin("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}
}

func TestExecutePlugin(t *testing.T) {
	manager := NewManager("./plugins")
	ctx := context.Background()

	tests := []struct {
		name      string
		plugin    *MockPlugin
		config    map[string]interface{}
		expectErr bool
	}{
		{
			name:      "successful execution",
			plugin:    &MockPlugin{name: "test-plugin"},
			config:    map[string]interface{}{"key": "value"},
			expectErr: false,
		},
		{
			name:      "validation error",
			plugin:    &MockPlugin{name: "validation-error", validateErr: os.ErrInvalid},
			config:    map[string]interface{}{},
			expectErr: true,
		},
		{
			name:      "execution error",
			plugin:    &MockPlugin{name: "execution-error", executeErr: os.ErrInvalid},
			config:    map[string]interface{}{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the registry for each test to avoid interference
			manager = NewManager("./plugins")
			
			// Register the plugin for this test
			err := manager.RegisterPlugin(tt.plugin)
			if err != nil {
				t.Fatalf("Failed to register plugin: %v", err)
			}
			
			// Execute the plugin
			_, err = manager.ExecutePlugin(ctx, tt.plugin.name, tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("ExecutePlugin() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestListPlugins(t *testing.T) {
	manager := NewManager("./plugins")
	plugins := []plugin.Plugin{
		&MockPlugin{name: "plugin1"},
		&MockPlugin{name: "plugin2"},
		&MockPlugin{name: "plugin3"},
	}

	// Register test plugins
	for _, p := range plugins {
		manager.RegisterPlugin(p)
	}

	// Test listing
	list := manager.ListPlugins()
	if len(list) != len(plugins) {
		t.Errorf("Expected %d plugins, got %d", len(plugins), len(list))
	}

	// Verify all plugins are present
	found := make(map[string]bool)
	for _, p := range list {
		found[p.Name()] = true
	}
	for _, p := range plugins {
		if !found[p.Name()] {
			t.Errorf("Plugin %s not found in list", p.Name())
		}
	}
}
