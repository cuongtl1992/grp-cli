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
	Name            string   `yaml:"name"`
	Description     string   `yaml:"description,omitempty"`
	RequireApproval bool     `yaml:"requireApproval,omitempty"`
	Approvers       []string `yaml:"approvers,omitempty"`
	Jobs            []Job    `yaml:"jobs"`
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