# Technical Specification: DevOps Release CLI Tool

## 1. Overview

This document outlines the technical specification for a comprehensive DevOps release CLI tool that orchestrates complex and flexible release workflows across multiple deployment targets (VMs, Docker containers, and Kubernetes). The tool supports multiple release strategies (Canary, Blue/Green, Shadow, A/B) and provides integration capabilities with various external tools through a plugin architecture.

## 2. Goals and Requirements

### 2.1 Core Requirements

- Cross-platform compatibility (Linux, macOS, Windows) and CPU architectures
- Declarative release plans using YAML configuration
- Modular plugin architecture for integrations
- Support for multiple deployment targets (VM, Docker, Kubernetes)
- Multiple release strategies implementation
- REST API for UI integration
- Rollback capabilities

### 2.2 Supported Release Strategies

- **Canary Deployment**: Gradual traffic shifting with automated health checking
- **Blue/Green Deployment**: Complete environment switching with instant rollback capability
- **Shadow Deployment**: Duplicate traffic to new service without affecting users
- **A/B Testing**: Split traffic with metric collection for decision making

### 2.3 Workflow Stages

1. **Verify Plan**: Validate configuration and prerequisites
2. **Approver Gate**: Request and manage approvals before proceeding
3. **Process Plan**: Execute release plan with proper orchestration
4. **Report Result**: Generate comprehensive reports on release outcomes

## 3. Architecture

### 3.1 High-Level Architecture

```
┌─────────────────┐     ┌───────────────────┐     ┌────────────────────┐
│                 │     │                   │     │                    │
│  Command Line   │────▶│  Execution Engine │────▶│  Plugin System     │
│  Interface      │     │                   │     │                    │
│                 │     │                   │     │                    │
└─────────────────┘     └───────────────────┘     └────────────────────┘
        │                        │                         │
        │                        │                         │
        ▼                        ▼                         ▼
┌─────────────────┐     ┌───────────────────┐     ┌────────────────────┐
│                 │     │                   │     │                    │
│  Configuration  │     │  State Management │     │  Strategy Executor │
│  Management     │     │                   │     │                    │
│                 │     │                   │     │                    │
└─────────────────┘     └───────────────────┘     └────────────────────┘
                                                          │
        ┌───────────────────────────────────────┐         │
        │                                       │         │
        ▼                                       ▼         ▼
┌─────────────────┐                   ┌───────────────────────┐
│                 │                   │                       │
│  REST API       │                   │  External Systems     │
│  Server         │                   │  (APISIX, K8s, etc.)  │
│                 │                   │                       │
└─────────────────┘                   └───────────────────────┘
```

### 3.2 Package Structure

```
cmd/                 # CLI commands implementation
  ├── root.go        # Root command
  ├── init.go        # Init command
  ├── validate.go    # Validate command
  ├── run.go         # Run command
  └── ...
internal/            # Internal packages
  ├── config/        # Configuration handling
  │     ├── parser.go     # YAML parser
  │     ├── validator.go  # Configuration validator
  │     └── merger.go     # Configuration merger
  ├── engine/        # Core execution engine
  │     ├── orchestrator.go  # Job orchestration
  │     ├── executor.go      # Job execution
  │     ├── state.go         # State management
  │     └── rollback.go      # Rollback handling
  ├── plugins/       # Plugin system
  │     ├── manager.go       # Plugin manager
  │     ├── registry.go      # Plugin registry
  │     └── loader.go        # Plugin loader
  ├── strategies/    # Deployment strategies
  │     ├── canary.go        # Canary implementation
  │     ├── bluegreen.go     # Blue-Green implementation
  │     ├── shadow.go        # Shadow deployment
  │     └── ab.go            # A/B testing
  ├── api/           # REST API
  │     ├── server.go        # API server
  │     ├── handlers.go      # Request handlers
  │     └── middleware.go    # API middleware
  └── models/        # Shared data models
        ├── plan.go          # Release plan model
        ├── job.go           # Job model
        └── result.go        # Result model
pkg/                 # Public packages
  ├── plugin/        # Plugin interfaces
  │     ├── types.go         # Plugin type definitions
  │     └── builder.go       # Plugin builder helpers
  ├── client/        # API client for Go applications
  └── utils/         # Shared utilities
```

## 4. Core Components

### 4.1 Command Line Interface

#### 4.1.1 Command Structure

```
releasectl [global options] command [command options] [arguments...]
```

#### 4.1.2 Global Commands

```
releasectl
  ├── init        # Initialize a new release plan
  ├── validate    # Validate a release plan
  ├── run         # Execute a release plan
  ├── rollback    # Execute rollback for a release
  ├── status      # Check status of running/completed releases
  ├── list        # List releases
  ├── plugins     # Manage plugins
  │     ├── list       # List available plugins
  │     ├── install    # Install a plugin
  │     ├── update     # Update a plugin
  │     ├── remove     # Remove a plugin
  │     └── info       # Show plugin details
  ├── server      # Start the REST API server
  └── version     # Show version information
```

#### 4.1.3 Implementation (cmd/root.go)

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "releasectl",
	Short: "A flexible DevOps release automation tool",
	Long: `A comprehensive CLI tool for DevOps to automate and manage complex release workflows
across multiple environments and deployment targets including VMs, Docker containers, and Kubernetes.
It supports various release strategies like Canary, Blue/Green, Shadow, and A/B testing.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.releasectl.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	
	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".releasectl" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigName(".releasectl")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
```

#### 4.1.4 Run Command Implementation (cmd/run.go)

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	
	"github.com/yourusername/releasectl/internal/engine"
	"github.com/yourusername/releasectl/internal/config"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [plan file]",
	Short: "Execute a release plan",
	Long: `Execute a release plan defined in YAML format. This command will:
1. Validate the plan file
2. Process required approvals
3. Execute all stages and jobs
4. Generate a report of the results`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planFile := args[0]
		
		// Create a context that can be canceled
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		// Handle signals for graceful shutdown
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			fmt.Println("Received signal, attempting graceful shutdown...")
			cancel()
		}()
		
		// Load the plan
		planLoader := config.NewLoader()
		plan, err := planLoader.LoadPlan(planFile)
		if err != nil {
			return fmt.Errorf("failed to load plan: %w", err)
		}
		
		// Get execution options from flags
		autoRollback, _ := cmd.Flags().GetBool("auto-rollback")
		skipApproval, _ := cmd.Flags().GetBool("skip-approval")
		
		// Create orchestrator
		orchestrator := engine.NewOrchestrator()
		
		// Execute the plan
		options := engine.ExecuteOptions{
			AutoRollback: autoRollback,
			SkipApproval: skipApproval,
		}
		
		result, err := orchestrator.ExecutePlan(ctx, plan, options)
		if err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}
		
		// Display result summary
		fmt.Printf("Execution completed successfully. ID: %s\n", result.ID)
		fmt.Printf("Total stages: %d, Jobs: %d\n", result.TotalStages, result.TotalJobs)
		fmt.Printf("Duration: %s\n", result.Duration)
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	
	// Local flags
	runCmd.Flags().Bool("auto-rollback", false, "Automatically rollback on failure")
	runCmd.Flags().Bool("skip-approval", false, "Skip approval steps")
	runCmd.Flags().Bool("dry-run", false, "Validate and simulate execution without making changes")
}
```

### 4.2 Configuration Management

#### 4.2.1 YAML Format

The release tool uses a structured YAML format for defining release plans. Plans can be modular, with a main file including referenced components.

##### Master Plan File Example:

```yaml
apiVersion: "devops.release/v1"
kind: "ReleasePlan"
metadata:
  name: "service-x-release-v1.2.3"
  description: "Service X Release with new payment feature"
  owner: "platform-team@example.com"
  version: "1.2.3"
  
# Reference to shared configurations
includes:
  - path: "./common/variables.yaml"
  - path: "./common/notifications.yaml"
  - path: "./environments/production.yaml"
  - path: "./strategies/canary.yaml"

# Global variables
variables:
  releaseVersion: "1.2.3"
  approvers:
    - "lead@example.com"
    - "manager@example.com"

# Stages of the release
stages:
  - name: "pre-release"
    description: "Documentation and preparation"
    jobs:
      - name: "update-confluence"
        type: "confluence-doc"
        config:
          template: "release-notes-template"
          spaceKey: "PROJ"
          parentPage: "Releases"
          title: "Release ${variables.releaseVersion}"
        
  - name: "deployment"
    description: "Main deployment activities"
    requireApproval: true
    approvers: ${variables.approvers}
    jobs:
      - name: "configure-routes"
        type: "apisix-gateway"
        config: 
          strategy: ${includes.strategies.canary}
          serviceId: "payment-service"
          
      - name: "deploy-k8s"
        type: "kubernetes"
        dependsOn: ["configure-routes"]
        config:
          namespace: "production"
          manifests: "./k8s/deployment.yaml"
          valuesFile: "./k8s/values.yaml"
          
  - name: "post-release"
    description: "Verification and notification"
    jobs:
      - name: "health-check"
        type: "validation"
        config:
          endpoints:
            - url: "https://api.example.com/health"
              expectedStatus: 200
              
      - name: "notify-stakeholders"
        type: "notification"
        dependsOn: ["health-check"]
        config: ${includes.notifications.release-complete}

# Rollback plan if needed
rollback:
  stages:
    - name: "revert-deployment"
      jobs:
        - name: "rollback-k8s"
          type: "kubernetes"
          config:
            namespace: "production"
            manifests: "./k8s/previous-deployment.yaml"
            
        - name: "restore-routes"
          type: "apisix-gateway"
          dependsOn: ["rollback-k8s"]
          config:
            restore: true
            serviceId: "payment-service"
```

#### 4.2.2 Configuration Loader Implementation (internal/config/loader.go)

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
	
	"github.com/yourusername/releasectl/internal/models"
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
```

#### 4.2.3 Variable Resolver Implementation (internal/config/resolver.go)

```go
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
```

### 4.3 Plugin System

#### 4.3.1 Plugin Interface (pkg/plugin/types.go)

```go
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
```

#### 4.3.2 Plugin Manager Implementation (internal/plugins/manager.go)

```go
package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
	
	"github.com/yourusername/releasectl/pkg/plugin"
)

// Manager handles plugin discovery, loading and execution
type Manager struct {
	registry       map[string]plugin.Plugin
	pluginDir      string
	configResolver *config.Resolver
	mutex          sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager(pluginDir string) *Manager {
	return &Manager{
		registry:       make(map[string]plugin.Plugin),
		pluginDir:      pluginDir,
		configResolver: config.NewResolver(),
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
	plug, err := plugin.Open(path)
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
	
	// Resolve any template variables in config
	vars := ctx.Value("variables").(map[string]interface{})
	resolvedConfig, err := pm.configResolver.ResolveValues(config, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config: %w", err)
	}
	
	// Validate plugin configuration
	if err := plg.Validate(ctx, resolvedConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Execute the plugin
	result, err := plg.Execute(ctx, resolvedConfig)
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
```

### 4.4 Execution Engine

#### 4.4.1 Orchestrator Implementation (internal/engine/orchestrator.go)

```go
package engine

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	"github.com/yourusername/releasectl/internal/models"
	"github.com/yourusername/releasectl/internal/plugins"
	"github.com/yourusername/releasectl/internal/approvals"
	"github.com/yourusername/releasectl/internal/notifications"
	"github.com/yourusername/releasectl/internal/state"
)

// ExecuteOptions contains options for plan execution
type ExecuteOptions struct {
	AutoRollback bool
	SkipApproval bool
	DryRun       bool
}

// ExecutionResult contains the outcome of a plan execution
type ExecutionResult struct {
	ID            string
	Success       bool
	TotalStages   int
	TotalJobs     int
	CompletedJobs int
	FailedJobs    int
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	Stages        []StageResult
}

// StageResult contains the outcome of a stage execution
type StageResult struct {
	Name     string
	Success  bool
	Jobs     []JobResult
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// JobResult contains the outcome of a job execution
type JobResult struct {
	Name     string
	Type     string
	Success  bool
	Message  string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// Orchestrator manages the execution of a release plan
type Orchestrator struct {
	pluginManager   *plugins.Manager
	stateManager    *state.Manager
	notifier        *notifications.Manager
	approvalManager *approvals.Manager
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		pluginManager:   plugins.NewManager("./plugins"),
		stateManager:    state.NewManager(),
		notifier:        notifications.NewManager(),
		approvalManager: approvals.NewManager(),
	}
}

// ExecutePlan runs a release plan
func (o *Orchestrator) ExecutePlan(ctx context.Context, plan *models.Plan, options ExecuteOptions) (*ExecutionResult, error) {
	// Generate unique execution ID
	executionID := uuid.New().String()
	
	// Create execution context with variables
	execCtx := context.WithValue(ctx, "executionID", executionID)
	execCtx = context.WithValue(execCtx, "variables", plan.Variables)
	
	// Create execution result
	result := &ExecutionResult{
		ID:         executionID,
		StartTime:  time.Now(),
		TotalStages: len(plan.Stages),
		TotalJobs:   o.countTotalJobs(plan),
	}
	
	// Initialize state
	o.stateManager.InitializeState(execCtx, executionID, plan)
	
	// Execute stages sequentially
	for i, stage := range plan.Stages {
		stageResult := StageResult{
			Name:      stage.Name,
			StartTime: time.Now(),
		}
		
		// Check if approval is required
		if stage.RequireApproval && !options.SkipApproval {
			approved, err := o.approvalManager.RequestApproval(execCtx, stage)
			if err != nil || !approved {
				stageResult.Success = false
				stageResult.EndTime = time.Now()
				stageResult.Duration = stageResult.EndTime.Sub(stageResult.StartTime)
				result.Stages = append(result.Stages, stageResult)
				return o.finalizeResult(result, false, "Stage approval failed")
			}
		}
		
		// Execute the stage
		stageErr := o.executeStage(execCtx, stage, &stageResult, options)
		
		// Update stage result
		stageResult.EndTime = time.Now()
		stageResult.Duration = stageResult.EndTime.Sub(stageResult.StartTime)
		stageResult.Success = stageErr == nil
		result.Stages = append(result.Stages, stageResult)
		
		// Handle stage failure
		if stageErr != nil {
			// Execute rollback if configured
			if options.AutoRollback && plan.Rollback != nil {
				o.executeRollback(execCtx, plan.Rollback)
			}
			
			return o.finalizeResult(result, false, fmt.Sprintf("Stage %s failed: %v", stage.Name, stageErr))
		}
		
		// Update state
		o.stateManager.UpdateStageStatus(execCtx, executionID, i, "completed")
	}
	
	// All stages completed successfully
	return o.finalizeResult(result, true, "Plan execution completed successfully")
}

// executeStage runs all jobs in a stage with proper dependency handling
func (o *Orchestrator) executeStage(ctx context.Context, stage *models.Stage, result *StageResult, options ExecuteOptions) error {
	// Build job dependency graph
	graph := buildDependencyGraph(stage.Jobs)
	
	// Create a new execution context for this stage
	stageCtx := context.WithValue(ctx, "stageName", stage.Name)
	
	// Execute jobs in dependency order
	executor := NewExecutor(o.pluginManager)
	return executor.ExecuteGraph(stageCtx, graph, result, options.DryRun)
}

// executeRollback runs the rollback plan
func (o *Orchestrator) executeRollback(ctx context.Context, rollback *models.Rollback) error {
	// Log rollback start
	fmt.Println("Starting rollback execution...")
	
	// Execute rollback stages
	for _, stage := range rollback.Stages {
		// Build job dependency graph
		graph := buildDependencyGraph(stage.Jobs)
		
		// Execute jobs in dependency order
		executor := NewExecutor(o.pluginManager)
		stageResult := &StageResult{Name: stage.Name}
		if err := executor.ExecuteGraph(ctx, graph, stageResult, false); err != nil {
			fmt.Printf("Rollback stage %s failed: %v\n", stage.Name, err)
			// Continue with other rollback stages even if one fails
		}
	}
	
	fmt.Println("Rollback execution completed")
	return nil
}

// finalizeResult completes the execution result
func (o *Orchestrator) finalizeResult(result *ExecutionResult, success bool, message string) (*ExecutionResult, error) {
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = success
	
	// Count completed and failed jobs
	for _, stage := range result.Stages {
		for _, job := range stage.Jobs {
			if job.Success {
				result.CompletedJobs++
			} else {
				result.FailedJobs++
			}
		}
	}
	
	if !success {
		return result, fmt.Errorf(message)
	}
	
	return result, nil
}

// countTotalJobs counts the total number of jobs in a plan
func (o *Orchestrator) countTotalJobs(plan *models.Plan) int {
	count := 0
	for _, stage := range plan.Stages {
		count += len(stage.Jobs)
	}
	return count
}

// buildDependencyGraph creates a graph of jobs based on dependencies
func buildDependencyGraph(jobs []models.Job) *JobGraph {
	graph := NewJobGraph()
	
	// Add all jobs to the graph
	for _, job := range jobs {
		graph.AddJob(job)
	}
	
	// Add dependencies
	for _, job := range jobs {
		for _, depName := range job.DependsOn {
			graph.AddDependency(job.Name, depName)
		}
	}
	
	return graph
}
```

#### 4.4.2 Job Executor Implementation (internal/engine/executor.go)

```go
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/yourusername/releasectl/internal/models"
	"github.com/yourusername/releasectl/internal/plugins"
)

// Executor handles the execution of jobs
type Executor struct {
	pluginManager *plugins.Manager
}

// NewExecutor creates a new executor
func NewExecutor(pluginManager *plugins.Manager) *Executor {
	return &Executor{
		pluginManager: pluginManager,
	}
}

// ExecuteGraph runs jobs in the order defined by the dependency graph
func (e *Executor) ExecuteGraph(ctx context.Context, graph *JobGraph, stageResult *StageResult, dryRun bool) error {
	// Get ready jobs (those with no dependencies)
	readyJobs := graph.GetReadyJobs()
	
	// Process until no more jobs are available
	for len(readyJobs) > 0 {
		var wg sync.WaitGroup
		jobResults := make([]JobResult, len(readyJobs))
		
		// Execute ready jobs in parallel
		for i, job := range readyJobs {
			wg.Add(1)
			
			go func(i int, job models.Job) {
				defer wg.Done()
				
				// Execute the job
				result := JobResult{
					Name:      job.Name,
					Type:      job.Type,
					StartTime: time.Now(),
				}
				
				if dryRun {
					// Simulate execution in dry-run mode
					time.Sleep(100 * time.Millisecond)
					result.Success = true
					result.Message = "Dry run simulation"
				} else {
					// Actual execution
					success, message := e.executeJob(ctx, job)
					result.Success = success
					result.Message = message
				}
				
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
				jobResults[i] = result
			}(i, job)
		}
		
		// Wait for all jobs to complete
		wg.Wait()
		
		// Process results
		for _, result := range jobResults {
			stageResult.Jobs = append(stageResult.Jobs, result)
			
			// Mark job as complete in the graph
			if result.Success {
				graph.MarkCompleted(result.Name)
			} else {
				// If a job fails, stop execution
				return fmt.Errorf("job %s failed: %s", result.Name, result.Message)
			}
		}
		
		// Get next batch of ready jobs
		readyJobs = graph.GetReadyJobs()
	}
	
	return nil
}

// executeJob runs a single job using the appropriate plugin
func (e *Executor) executeJob(ctx context.Context, job models.Job) (bool, string) {
	fmt.Printf("Executing job: %s (type: %s)\n", job.Name, job.Type)
	
	// Execute the job using the plugin manager
	result, err := e.pluginManager.ExecutePlugin(ctx, job.Type, job.Config)
	if err != nil {
		return false, fmt.Sprintf("Failed to execute job: %v", err)
	}
	
	if !result.Success {
		return false, result.Message
	}
	
	return true, result.Message
}
```

#### 4.4.3 Job Graph Implementation (internal/engine/jobgraph.go)

```go
package engine

import (
	"github.com/yourusername/releasectl/internal/models"
)

// JobGraph represents a dependency graph of jobs
type JobGraph struct {
	jobs           map[string]models.Job
	dependencies   map[string][]string
	dependents     map[string][]string
	completed      map[string]bool
}

// NewJobGraph creates a new job graph
func NewJobGraph() *JobGraph {
	return &JobGraph{
		jobs:         make(map[string]models.Job),
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
		completed:    make(map[string]bool),
	}
}

// AddJob adds a job to the graph
func (g *JobGraph) AddJob(job models.Job) {
	g.jobs[job.Name] = job
	
	// Initialize empty dependency lists if they don't exist
	if _, exists := g.dependencies[job.Name]; !exists {
		g.dependencies[job.Name] = []string{}
	}
	
	if _, exists := g.dependents[job.Name]; !exists {
		g.dependents[job.Name] = []string{}
	}
}

// AddDependency adds a dependency between jobs
func (g *JobGraph) AddDependency(jobName, dependsOn string) {
	// Add the dependency
	g.dependencies[jobName] = append(g.dependencies[jobName], dependsOn)
	
	// Add the dependent relationship (reverse direction)
	g.dependents[dependsOn] = append(g.dependents[dependsOn], jobName)
}

// GetReadyJobs returns jobs that are ready to be executed
func (g *JobGraph) GetReadyJobs() []models.Job {
	var readyJobs []models.Job
	
	for name, job := range g.jobs {
		// Skip already completed jobs
		if g.completed[name] {
			continue
		}
		
		// Check if all dependencies are completed
		allDepsCompleted := true
		for _, dep := range g.dependencies[name] {
			if !g.completed[dep] {
				allDepsCompleted = false
				break
			}
		}
		
		if allDepsCompleted {
			readyJobs = append(readyJobs, job)
		}
	}
	
	return readyJobs
}

// MarkCompleted marks a job as completed
func (g *JobGraph) MarkCompleted(jobName string) {
	g.completed[jobName] = true
}

// IsCompleted returns true if all jobs are completed
func (g *JobGraph) IsCompleted() bool {
	for name := range g.jobs {
		if !g.completed[name] {
			return false
		}
	}
	return true
}

// GetRemainingJobs returns jobs that are not yet completed
func (g *JobGraph) GetRemainingJobs() []models.Job {
	var remainingJobs []models.Job
	
	for name, job := range g.jobs {
		if !g.completed[name] {
			remainingJobs = append(remainingJobs, job)
		}
	}
	
	return remainingJobs
}
```

### 4.5 Data Models

#### 4.5.1 Plan Model (internal/models/plan.go)

```go
package models

// Plan represents a release plan
type Plan struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   Metadata               `yaml:"metadata"`
	Includes   []Include              `yaml:"includes,omitempty"`
	Variables  map[string]interface{} `yaml:"variables,omitempty"`
	Stages     []Stage                `yaml:"stages"`
	Rollback   *Rollback              `yaml:"rollback,omitempty"`
}

// Metadata contains information about the plan
type Metadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Owner       string `yaml:"owner,omitempty"`
	Version     string `yaml:"version,omitempty"`
}

// Include represents a reference to an external file
type Include struct {
	Path string `yaml:"path"`
}

// Stage represents a stage in the release plan
type Stage struct {
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description,omitempty"`
	RequireApproval bool     `yaml:"requireApproval,omitempty"`
	Approvers      []string `yaml:"approvers,omitempty"`
	Jobs           []Job    `yaml:"jobs"`
}

// Job represents a job to be executed
type Job struct {
	Name      string                 `yaml:"name"`
	Type      string                 `yaml:"type"`
	DependsOn []string               `yaml:"dependsOn,omitempty"`
	Timeout   string                 `yaml:"timeout,omitempty"`
	Retries   int                    `yaml:"retries,omitempty"`
	Config    map[string]interface{} `yaml:"config"`
}

// Rollback represents a rollback plan
type Rollback struct {
	Stages []Stage `yaml:"stages"`
}
```

### 4.6 Deployment Strategies

#### 4.6.1 Strategy Executor (internal/strategies/executor.go)

```go
package strategies

import (
	"context"
	"fmt"
	"time"
	
	"github.com/yourusername/releasectl/internal/health"
	"github.com/yourusername/releasectl/internal/metrics"
	"github.com/yourusername/releasectl/internal/routes"
)

// StrategyExecutor handles deployment strategy execution
type StrategyExecutor struct {
	healthChecker   *health.Checker
	metricCollector *metrics.Collector
	routeManager    *routes.Manager
}

// NewStrategyExecutor creates a new strategy executor
func NewStrategyExecutor() *StrategyExecutor {
	return &StrategyExecutor{
		healthChecker:   health.NewChecker(),
		metricCollector: metrics.NewCollector(),
		routeManager:    routes.NewManager(),
	}
}

// ExecuteStrategy runs the appropriate deployment strategy
func (s *StrategyExecutor) ExecuteStrategy(ctx context.Context, strategy string, config map[string]interface{}) error {
	switch strategy {
	case "canary":
		return s.executeCanaryDeployment(ctx, config)
	case "bluegreen":
		return s.executeBlueGreenDeployment(ctx, config)
	case "shadow":
		return s.executeShadowDeployment(ctx, config)
	case "ab":
		return s.executeABDeployment(ctx, config)
	default:
		return fmt.Errorf("unknown deployment strategy: %s", strategy)
	}
}

// CanaryPhase represents a phase in canary deployment
type CanaryPhase struct {
	Percentage     int                    `yaml:"percentage"`
	Duration       time.Duration          `yaml:"duration"`
	SuccessCriteria map[string]string     `yaml:"successCriteria"`
}

// executeCanaryDeployment implements canary deployment strategy
func (s *StrategyExecutor) executeCanaryDeployment(ctx context.Context, config map[string]interface{}) error {
	// Extract config
	serviceID, _ := config["serviceId"].(string)
	newVersion, _ := config["newVersion"].(string)
	
	// Extract phases
	phasesConfig, _ := config["phases"].([]interface{})
	phases := make([]CanaryPhase, len(phasesConfig))
	
	for i, phaseConfig := range phasesConfig {
		if phaseMap, ok := phaseConfig.(map[string]interface{}); ok {
			percentage, _ := phaseMap["percentage"].(int)
			durationStr, _ := phaseMap["duration"].(string)
			duration, _ := time.ParseDuration(durationStr)
			
			criteriaMap, _ := phaseMap["successCriteria"].(map[string]interface{})
			criteria := make(map[string]string)
			for k, v := range criteriaMap {
				criteria[k] = fmt.Sprintf("%v", v)
			}
			
			phases[i] = CanaryPhase{
				Percentage:     percentage,
				Duration:       duration,
				SuccessCriteria: criteria,
			}
		}
	}
	
	// 1. Initialize the deployment
	if err := s.routeManager.InitializeCanary(ctx, serviceID, newVersion); err != nil {
		return fmt.Errorf("failed to initialize canary: %w", err)
	}
	
	// 2. Execute phased rollout
	for i, phase := range phases {
		// Update traffic split
		err := s.routeManager.UpdateTrafficSplit(ctx, serviceID, phase.Percentage)
		if err != nil {
			return fmt.Errorf("failed to update traffic split: %w", err)
		}
		
		fmt.Printf("Canary phase %d: %d%% traffic to new version\n", i+1, phase.Percentage)
		
		// Wait for specified duration
		if phase.Duration > 0 {
			fmt.Printf("Waiting for %s before evaluating health...\n", phase.Duration)
			select {
			case <-time.After(phase.Duration):
				// Continue after duration
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		
		// Check health metrics against criteria
		healthy, metrics := s.evaluateHealthCriteria(ctx, serviceID, phase.SuccessCriteria)
		if !healthy {
			// Trigger rollback
			fmt.Printf("Phase %d failed health checks: %v\n", i+1, metrics)
			s.routeManager.RollbackDeployment(ctx, serviceID)
			return fmt.Errorf("health criteria not met in phase %d", i+1)
		}
		
		fmt.Printf("Successfully completed phase %d (%d%%)\n", i+1, phase.Percentage)
	}
	
	// 3. Finalize the deployment - full traffic cutover
	return s.routeManager.FinalizeDeployment(ctx, serviceID)
}

// executeBlueGreenDeployment implements blue/green deployment strategy
func (s *StrategyExecutor) executeBlueGreenDeployment(ctx context.Context, config map[string]interface{}) error {
	// Implementation details...
	return nil
}

// executeShadowDeployment implements shadow deployment strategy
func (s *StrategyExecutor) executeShadowDeployment(ctx context.Context, config map[string]interface{}) error {
	// Implementation details...
	return nil
}

// executeABDeployment implements A/B testing deployment strategy
func (s *StrategyExecutor) executeABDeployment(ctx context.Context, config map[string]interface{}) error {
	// Implementation details...
	return nil
}

// evaluateHealthCriteria checks if health metrics meet defined criteria
func (s *StrategyExecutor) evaluateHealthCriteria(ctx context.Context, serviceID string, criteria map[string]string) (bool, map[string]interface{}) {
	// Collect current metrics
	metrics, err := s.metricCollector.CollectMetrics(ctx, serviceID)
	if err != nil {
		return false, map[string]interface{}{"error": err.Error()}
	}
	
	// Evaluate each criterion
	for metricName, threshold := range criteria {
		if !s.metricMeetsCriterion(metrics, metricName, threshold) {
			return false, metrics
		}
	}
	
	return true, metrics
}

// metricMeetsCriterion checks if a metric meets a specified criterion
func (s *StrategyExecutor) metricMeetsCriterion(metrics map[string]interface{}, metricName, threshold string) bool {
	// Implementation details...
	return true
}
```

### 4.7 REST API

#### 4.7.1 API Server Implementation (internal/api/server.go)

```go
package api

import (
	"context"
	"fmt"
	"net/http"
	
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	
	"github.com/yourusername/releasectl/internal/engine"
	"github.com/yourusername/releasectl/internal/state"
	"github.com/yourusername/releasectl/internal/plugins"
)

// Server provides a REST API for the release system
type Server struct {
	router         *echo.Echo
	orchestrator   *engine.Orchestrator
	stateManager   *state.Manager
	pluginManager  *plugins.Manager
	config         *Config
}

// Config contains server configuration
type Config struct {
	Port           int
	AuthEnabled    bool
	AuthToken      string
	AllowedOrigins []string
}

// NewServer creates a new API server
func NewServer(config *Config) *Server {
	e := echo.New()
	
	return &Server{
		router:        e,
		orchestrator:  engine.NewOrchestrator(),
		stateManager:  state.NewManager(),
		pluginManager: plugins.NewManager("./plugins"),
		config:        config,
	}
}

// Start initializes and starts the server
func (s *Server) Start() error {
	// Configure middleware
	s.router.Use(middleware.Logger())
	s.router.Use(middleware.Recover())
	s.router.Use(middleware.CORS(&middleware.CORSConfig{
		AllowOrigins: s.config.AllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	
	// Add authentication if enabled
	if s.config.AuthEnabled {
		s.router.Use(s.authMiddleware)
	}
	
	// Set up routes
	s.SetupRoutes()
	
	// Start server
	addr := fmt.Sprintf(":%d", s.config.Port)
	return s.router.Start(addr)
}

// SetupRoutes configures all API endpoints
func (s *Server) SetupRoutes() {
	// Health check
	s.router.GET("/health", s.HealthCheck)
	
	// API versioning
	v1 := s.router.Group("/api/v1")
	
	// Plans
	plans := v1.Group("/plans")
	plans.GET("", s.ListPlans)
	plans.POST("", s.CreatePlan)
	plans.GET("/:id", s.GetPlan)
	plans.DELETE("/:id", s.DeletePlan)
	plans.POST("/:id/validate", s.ValidatePlan)
	
	// Executions
	executions := v1.Group("/executions")
	executions.POST("", s.StartExecution)
	executions.GET("", s.ListExecutions)
	executions.GET("/:id", s.GetExecution)
	executions.POST("/:id/stop", s.StopExecution)
	executions.POST("/:id/rollback", s.RollbackExecution)
	
	// Approvals
	approvals := v1.Group("/approvals")
	approvals.GET("", s.ListPendingApprovals)
	approvals.POST("/:id/approve", s.ApproveRequest)
	approvals.POST("/:id/reject", s.RejectRequest)
	
	// Plugins
	plugins := v1.Group("/plugins")
	plugins.GET("", s.ListPlugins)
	plugins.GET("/:id", s.GetPluginInfo)
}

// ErrorResponse is the standard error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// authMiddleware provides token-based authentication
func (s *Server) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token != fmt.Sprintf("Bearer %s", s.config.AuthToken) {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		}
		return next(c)
	}
}

// HealthCheck handles health check requests
func (s *Server) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// StartExecution handles the request to start a new release
func (s *Server) StartExecution(c echo.Context) error {
	var req struct {
		PlanID  string                 `json:"planId"`
		Options map[string]interface{} `json:"options"`
	}
	
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
	}
	
	// Convert to execution options
	options := engine.ExecuteOptions{
		AutoRollback: getOptionBool(req.Options, "autoRollback", false),
		SkipApproval: getOptionBool(req.Options, "skipApproval", false),
		DryRun:       getOptionBool(req.Options, "dryRun", false),
	}
	
	// Start the execution in a goroutine
	executionID := ""
	go func() {
		ctx := context.Background()
		// Load the plan
		// Execute the plan
		// Implementation details...
	}()
	
	return c.JSON(http.StatusAccepted, map[string]string{
		"status": "Execution started",
		"id":     executionID,
	})
}

// Helper function to extract boolean options
func getOptionBool(options map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := options[key].(bool); ok {
		return val
	}
	return defaultValue
}

// Additional API handlers...
```

## 5. Plugin Implementation Examples

### 5.1 APISIX Gateway Plugin

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	
	"github.com/yourusername/releasectl/pkg/plugin"
)

type APISIXGatewayPlugin struct{}

// Name returns the plugin identifier
func (p *APISIXGatewayPlugin) Name() string {
	return "apisix-gateway"
}

// Description provides plugin information
func (p *APISIXGatewayPlugin) Description() string {
	return "Manages APISIX API gateway routes and configurations for deployment strategies"
}

// Version returns the plugin version
func (p *APISIXGatewayPlugin) Version() string {
	return "1.0.0"
}

// ConfigSchema defines configuration structure
func (p *APISIXGatewayPlugin) ConfigSchema() *plugin.JSONSchema {
	return &plugin.JSONSchema{
		Type: "object",
		Properties: map[string]*plugin.JSONSchema{
			"apiEndpoint": {Type: "string"},
			"adminKey": {Type: "string"},
			"serviceId": {Type: "string"},
			"strategy": {
				Type: "object",
				Properties: map[string]*plugin.JSONSchema{
					"type": {Type: "string"},
					"phases": {
						Type: "array",
						Items: &plugin.JSONSchema{
							Type: "object",
							Properties: map[string]*plugin.JSONSchema{
								"percentage": {Type: "number"},
								"duration": {Type: "string"},
							},
						},
					},
				},
			},
		},
		Required: []string{"apiEndpoint", "adminKey", "serviceId"},
	}
}

// Validate checks configuration
func (p *APISIXGatewayPlugin) Validate(ctx context.Context, config map[string]interface{}) error {
	// Validation logic
	requiredKeys := []string{"apiEndpoint", "adminKey", "serviceId"}
	for _, key := range requiredKeys {
		if _, exists := config[key]; !exists {
			return fmt.Errorf("missing required config: %s", key)
		}
	}
	
	return nil
}

// Execute implements the plugin logic
func (p *APISIXGatewayPlugin) Execute(ctx context.Context, config map[string]interface{}) (*plugin.Result, error) {
	apiEndpoint := config["apiEndpoint"].(string)
	adminKey := config["adminKey"].(string)
	serviceId := config["serviceId"].(string)
	
	// Get strategy configuration
	strategyConfig, ok := config["strategy"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid strategy configuration")
	}
	
	strategyType, _ := strategyConfig["type"].(string)
	
	// Execute the appropriate strategy
	switch strategyType {
	case "canary":
		return p.configureCanary(ctx, apiEndpoint, adminKey, serviceId, strategyConfig)
	case "bluegreen":
		return p.configureBlueGreen(ctx, apiEndpoint, adminKey, serviceId, strategyConfig)
	case "shadow":
		return p.configureShadow(ctx, apiEndpoint, adminKey, serviceId, strategyConfig)
	default:
		return nil, fmt.Errorf("unsupported strategy type: %s", strategyType)
	}
}

// configureCanary sets up canary routing in APISIX
func (p *APISIXGatewayPlugin) configureCanary(ctx context.Context, apiEndpoint, adminKey, serviceId string, config map[string]interface{}) (*plugin.Result, error) {
	// Get original route configuration
	route, err := p.getRouteConfig(apiEndpoint, adminKey, serviceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get route config: %w", err)
	}
	
	// Create canary route configuration
	canaryRoute := map[string]interface{}{
		"uri": route["uri"],
		"plugins": map[string]interface{}{
			"traffic-split": map[string]interface{}{
				"rules": []map[string]interface{}{
					{
						"match": []map[string]interface{}{},
						"weighted_upstreams": []map[string]interface{}{
							{
								"upstream_id": route["upstream_id"],
								"weight": 90,
							},
							{
								"upstream_id": config["newUpstreamId"].(string),
								"weight": 10,
							},
						},
					},
				},
			},
		},
	}
	
	// Update route with canary configuration
	if err := p.updateRouteConfig(apiEndpoint, adminKey, serviceId, canaryRoute); err != nil {
		return nil, fmt.Errorf("failed to update route config: %w", err)
	}
	
	return &plugin.Result{
		Success: true,
		Message: "Canary routing configured successfully",
		Data: map[string]interface{}{
			"routeId": serviceId,
		},
		ExecutionID: ctx.Value("executionID").(string),
	}, nil
}

// Implement other strategy configurations...

// getRouteConfig retrieves route configuration from APISIX
func (p *APISIXGatewayPlugin) getRouteConfig(apiEndpoint, adminKey, routeId string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/apisix/admin/routes/%s", apiEndpoint, routeId)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-API-KEY", adminKey)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}
	
	var result struct {
		Value map[string]interface{} `json:"value"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Value, nil
}

// updateRouteConfig updates route configuration in APISIX
func (p *APISIXGatewayPlugin) updateRouteConfig(apiEndpoint, adminKey, routeId string, config map[string]interface{}) error {
	url := fmt.Sprintf("%s/apisix/admin/routes/%s", apiEndpoint, routeId)
	
	body, err := json.Marshal(config)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("PUT", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	
	req.Header.Set("X-API-KEY", adminKey)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// Rollback reverts changes if execution fails
func (p *APISIXGatewayPlugin) Rollback(ctx context.Context, executionID string) error {
	// Implement rollback logic...
	return nil
}

// Export the plugin
var Plugin APISIXGatewayPlugin
```

### 5.2 Kubernetes Deployment Plugin

```go
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	
	"github.com/yourusername/releasectl/pkg/plugin"
)

type KubernetesPlugin struct{}

// Name returns the plugin identifier
func (p *KubernetesPlugin) Name() string {
	return "kubernetes"
}

// Description provides plugin information
func (p *KubernetesPlugin) Description() string {
	return "Manages Kubernetes deployments"
}

// Version returns the plugin version
func (p *KubernetesPlugin) Version() string {
	return "1.0.0"
}

// ConfigSchema defines configuration structure
func (p *KubernetesPlugin) ConfigSchema() *plugin.JSONSchema {
	return &plugin.JSONSchema{
		Type: "object",
		Properties: map[string]*plugin.JSONSchema{
			"kubeconfig": {Type: "string"},
			"namespace": {Type: "string"},
			"manifests": {
				Type: "string",
				Description: "Path to Kubernetes manifest file or directory",
			},
			"valuesFile": {
				Type: "string",
				Description: "Path to values file for template processing",
			},
			"kubectl": {
				Type: "string",
				Description: "Path to kubectl binary (optional)",
			},
		},
		Required: []string{"namespace", "manifests"},
	}
}

// Validate checks configuration
func (p *KubernetesPlugin) Validate(ctx context.Context, config map[string]interface{}) error {
	// Check required fields
	requiredFields := []string{"namespace", "manifests"}
	for _, field := range requiredFields {
		if _, exists := config[field]; !exists {
			return fmt.Errorf("missing required config field: %s", field)
		}
	}
	
	// Check if manifests path exists
	manifestsPath, _ := config["manifests"].(string)
	if _, err := ioutil.ReadFile(manifestsPath); err != nil {
		return fmt.Errorf("cannot access manifests file: %w", err)
	}
	
	// If values file is specified, check if it exists
	if valuesFile, ok := config["valuesFile"].(string); ok {
		if _, err := ioutil.ReadFile(valuesFile); err != nil {
			return fmt.Errorf("cannot access values file: %w", err)
		}
	}
	
	// Check kubectl availability
	kubectlPath := "kubectl" // Default to PATH
	if path, ok := config["kubectl"].(string); ok {
		kubectlPath = path
	}
	
	_, err := exec.LookPath(kubectlPath)
	if err != nil {
		return fmt.Errorf("kubectl not found: %w", err)
	}
	
	return nil
}

// Execute implements the plugin logic
func (p *KubernetesPlugin) Execute(ctx context.Context, config map[string]interface{}) (*plugin.Result, error) {
	namespace := config["namespace"].(string)
	manifestsPath := config["manifests"].(string)
	
	// Determine kubectl path
	kubectlPath := "kubectl"
	if path, ok := config["kubectl"].(string); ok {
		kubectlPath = path
	}
	
	// Set kubeconfig if provided
	args := []string{"apply"}
	if kubeconfig, ok := config["kubeconfig"].(string); ok {
		args = append(args, "--kubeconfig", kubeconfig)
	}
	
	// Set namespace
	args = append(args, "-n", namespace)
	
	// Apply manifests
	args = append(args, "-f", manifestsPath)
	
	// Execute kubectl command
	cmd := exec.CommandContext(ctx, kubectlPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("kubectl apply failed: %s: %w", string(output), err)
	}
	
	// Wait for deployment to complete
	if err := p.waitForDeployment(ctx, kubectlPath, namespace, manifestsPath); err != nil {
		return nil, err
	}
	
	return &plugin.Result{
		Success: true,
		Message: "Kubernetes deployment completed successfully",
		Data: map[string]interface{}{
			"namespace":  namespace,
			"manifests":  manifestsPath,
			"kubeOutput": string(output),
		},
		ExecutionID: ctx.Value("executionID").(string),
	}, nil
}

// waitForDeployment waits for all resources to be ready
func (p *KubernetesPlugin) waitForDeployment(ctx context.Context, kubectlPath, namespace, manifestsPath string) error {
	// Get resource types from manifests
	resources, err := p.extractResources(manifestsPath)
	if err != nil {
		return err
	}
	
	// Wait for deployments, statefulsets, etc.
	for _, resource := range resources {
		if resource.Kind == "Deployment" || resource.Kind == "StatefulSet" || resource.Kind == "DaemonSet" {
			args := []string{"rollout", "status", resource.Kind, resource.Name, "-n", namespace}
			
			cmd := exec.CommandContext(ctx, kubectlPath, args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("resource %s/%s not ready: %s: %w", 
					resource.Kind, resource.Name, string(output), err)
			}
		}
	}
	
	return nil
}

// extractResources parses manifests to extract resources
func (p *KubernetesPlugin) extractResources(manifestsPath string) ([]Resource, error) {
	// Implementation details...
	return nil, nil
}

// Resource represents a Kubernetes resource
type Resource struct {
	Kind string
	Name string
}

// Rollback reverts changes if execution fails
func (p *KubernetesPlugin) Rollback(ctx context.Context, executionID string) error {
	// Implementation details...
	return nil
}

// Export the plugin
var Plugin KubernetesPlugin
```

## 6. Implementation Guidelines

### 6.1 Project Setup

1. Initialize Go module:
   ```bash
   mkdir releasectl
   cd releasectl
   go mod init github.com/yourusername/releasectl
   ```

2. Install dependencies:
   ```bash
   go get -u github.com/spf13/cobra
   go get -u github.com/spf13/viper
   go get -u gopkg.in/yaml.v3
   go get -u github.com/google/uuid
   go get -u github.com/labstack/echo/v4
   ```

3. Create directory structure:
   ```bash
   mkdir -p cmd internal/{config,engine,plugins,strategies,api,models,health,metrics,routes,state,approvals,notifications} pkg/{plugin,client,utils}
   ```

### 6.2 Build Process

1. Create a Makefile for common tasks:
   ```makefile
   .PHONY: build test clean plugins

   BINARY=releasectl
   PLUGINS_DIR=plugins
   VERSION=$(shell git describe --tags --always --dirty)
   BUILD_TIME=$(shell date +%FT%T%z)
   LDFLAGS=-ldflags "-X github.com/yourusername/releasectl/cmd.Version=${VERSION} -X github.com/yourusername/releasectl/cmd.BuildTime=${BUILD_TIME}"

   build:
       go build ${LDFLAGS} -o ${BINARY} main.go

   test:
       go test ./...

   plugins:
       mkdir -p ${PLUGINS_DIR}
       go build -buildmode=plugin -o ${PLUGINS_DIR}/apisix.so ./plugins/apisix/main.go
       go build -buildmode=plugin -o ${PLUGINS_DIR}/kubernetes.so ./plugins/kubernetes/main.go
       # Add other plugins here

   clean:
       rm -f ${BINARY}
       rm -f ${PLUGINS_DIR}/*.so

   all: clean build plugins
   ```

2. Create a Dockerfile for containerization:
   ```dockerfile
   FROM golang:1.20-alpine AS builder

   WORKDIR /app
   COPY . .
   RUN apk add --no-cache make git
   RUN make all

   FROM alpine:3.17

   RUN apk add --no-cache ca-certificates

   WORKDIR /app
   COPY --from=builder /app/releasectl /app/
   COPY --from=builder /app/plugins /app/plugins

   ENTRYPOINT ["/app/releasectl"]
   ```

### 6.3 Plugin Development Guidelines

1. Each plugin should implement the Plugin interface from pkg/plugin/types.go
2. Plugins should be compiled as Go plugins with `go build -buildmode=plugin`
3. Plugin configuration should be validated against a schema
4. Plugins should implement proper error handling and logging
5. All plugins should support the rollback operation
6. Integration points should use standard protocols (HTTP, gRPC) when possible

### 6.4 Testing Strategy

1. Unit tests for core components
2. Integration tests for plugin interactions
3. End-to-end tests for complete workflows
4. Mock external dependencies for testing

### 6.5 Cross-Platform Compatibility

1. Use Go's standard library for file operations
2. Avoid platform-specific system calls
3. Handle path separators correctly for cross-platform compatibility
4. Use os-agnostic utilities for executing commands
5. Provide binaries for all major platforms (Linux, macOS, Windows)

## 7. Example Usage

### 7.1 Creating a Release Plan

```yaml
apiVersion: "devops.release/v1"
kind: "ReleasePlan"
metadata:
  name: "payment-service-v1.2.3"
  description: "Release payment service with new features"
  owner: "platform-team@example.com"
  version: "1.2.3"
  
variables:
  releaseVersion: "1.2.3"
  environment: "production"
  approvers:
    - "lead@example.com"
    - "manager@example.com"

stages:
  - name: "preparation"
    description: "Documentation and notifications"
    jobs:
      - name: "create-release-notes"
        type: "confluence-doc"
        config:
          spaceKey: "PROJ"
          parentPage: "Releases"
          template: "release-template"
          title: "Release ${variables.releaseVersion}"
          
      - name: "notify-stakeholders"
        type: "notification"
        config:
          channels: ["slack", "email"]
          recipients: ["platform-team", "product-team"]
          template: "pre-release-notification"
  
  - name: "deployment"
    description: "Production deployment with canary strategy"
    requireApproval: true
    approvers: ${variables.approvers}
    jobs:
      - name: "configure-gateway"
        type: "apisix-gateway"
        config:
          apiEndpoint: "https://apisix-admin.example.com"
          adminKey: "${env.APISIX_ADMIN_KEY}"
          serviceId: "payment-service"
          strategy:
            type: "canary"
            phases:
              - percentage: 5
                duration: "15m"
                successCriteria:
                  errorRate: "<1%"
              - percentage: 20
                duration: "30m"
                successCriteria:
                  errorRate: "<1%"
              - percentage: 50
                duration: "1h"
                successCriteria:
                  errorRate: "<0.5%"
              - percentage: 100
                successCriteria:
                  errorRate: "<0.5%"
          
      - name: "deploy-kubernetes"
        type: "kubernetes"
        dependsOn: ["configure-gateway"]
        config:
          namespace: "${variables.environment}"
          manifests: "./k8s/payment-service/"
          valuesFile: "./k8s/values-${variables.environment}.yaml"
          
  - name: "validation"
    description: "Post-deployment validation"
    jobs:
      - name: "run-tests"
        type: "test-runner"
        config:
          testSuite: "payment-api-tests"
          environment: "${variables.environment}"
          
      - name: "monitor-metrics"
        type: "prometheus-monitor"
        config:
          duration: "10m"
          queries:
            - name: "error-rate"
              query: 'sum(rate(http_requests_total{status=~"5..",service="payment"}[5m])) / sum(rate(http_requests_total{service="payment"}[5m]))'
              threshold: "< 0.01"
            - name: "latency-p95"
              query: 'histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{service="payment"}[5m])) by (le))'
              threshold: "< 0.5"

rollback:
  stages:
    - name: "revert-deployment"
      jobs:
        - name: "rollback-kubernetes"
          type: "kubernetes"
          config:
            namespace: "${variables.environment}"
            manifests: "./k8s/payment-service-previous/"
            
        - name: "restore-gateway"
          type: "apisix-gateway"
          dependsOn: ["rollback-kubernetes"]
          config:
            apiEndpoint: "https://apisix-admin.example.com"
            adminKey: "${env.APISIX_ADMIN_KEY}"
            serviceId: "payment-service"
            restore: true
```

### 7.2 Running the Release

```bash
# Validate the release plan
releasectl validate release-plan.yaml

# Execute the release plan
releasectl run release-plan.yaml --auto-rollback

# Check status of the release
releasectl status <execution-id>

# Manually trigger rollback
releasectl rollback <execution-id>
```

### 7.3 Using the REST API

#### Start Release Execution

```bash
curl -X POST http://localhost:8080/api/v1/executions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "planId": "payment-service-v1.2.3",
    "options": {
      "autoRollback": true,
      "skipApproval": false
    }
  }'
```

#### Check Release Status

```bash
curl -X GET http://localhost:8080/api/v1/executions/<execution-id> \
  -H "Authorization: Bearer <token>"
```

## 8. Conclusion

This technical specification outlines a comprehensive DevOps release CLI tool that supports multiple deployment targets, release strategies, and integration capabilities. The modular plugin architecture allows for extensibility and the REST API enables UI integration. The tool is designed to be cross-platform, robust, and flexible to fit various DevOps workflows.

Implementation should follow Go best practices, with strong emphasis on error handling, testing, and maintainability. The plugin system is key to the tool's flexibility, allowing new integrations to be added without modifying the core code.
