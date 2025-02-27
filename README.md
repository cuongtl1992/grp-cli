# GRP-CLI: DevOps Release Automation Tool

A comprehensive CLI tool for DevOps to automate and manage complex release workflows across multiple environments and deployment targets including VMs, Docker containers, and Kubernetes. It supports various release strategies like Canary, Blue/Green, Shadow, and A/B testing.

## Features

- Cross-platform compatibility (Linux, macOS, Windows)
- Declarative release plans using YAML configuration
- Modular plugin architecture for integrations
- Support for multiple deployment targets (VM, Docker, Kubernetes)
- Multiple release strategies implementation
- Rollback capabilities
- Approval workflows

## Installation

### From Source

```bash
git clone https://github.com/cuongtl1992/grp-cli.git
cd grp-cli
go build -o grp-cli
```

### Using Go Install

```bash
go install github.com/cuongtl1992/grp-cli@latest
```

## Usage

### Basic Commands

```bash
# Show version information
grp-cli version

# Validate a release plan
grp-cli validate examples/kubernetes-deployment.yaml

# Execute a release plan
grp-cli run examples/kubernetes-deployment.yaml

# Execute with options
grp-cli run examples/kubernetes-deployment.yaml --dry-run --skip-approval
```

### Command Options

- `--auto-rollback`: Automatically rollback on failure
- `--skip-approval`: Skip approval steps
- `--dry-run`: Validate and simulate execution without making changes
- `--plugin-dir`: Directory containing plugins (default: ./plugins)
- `--verbose`: Enable verbose output
- `--debug`: Enable debug mode

## Release Plan Structure

Release plans are defined in YAML format with the following structure:

```yaml
apiVersion: v1
kind: ReleasePlan
metadata:
  name: example-plan
  description: Example release plan
  owner: DevOps Team
  version: 1.0.0

variables:
  app:
    name: example-app
    namespace: default

stages:
  - name: preparation
    description: Prepare the environment
    jobs:
      - name: job1
        type: plugin-type
        config:
          key: value

  - name: deployment
    description: Deploy the application
    requireApproval: true
    approvers:
      - user1@example.com
    jobs:
      - name: job2
        type: plugin-type
        dependsOn:
          - job1
        config:
          key: value

rollback:
  stages:
    - name: rollback-deployment
      jobs:
        - name: rollback-job
          type: plugin-type
          config:
            key: value
```

## Plugin Development

Plugins implement the `Plugin` interface defined in `pkg/plugin/types.go`:

```go
type Plugin interface {
    Name() string
    Description() string
    Version() string
    ConfigSchema() *JSONSchema
    Validate(ctx context.Context, config map[string]interface{}) error
    Execute(ctx context.Context, config map[string]interface{}) (*Result, error)
    Rollback(ctx context.Context, executionID string) error
}
```

See the example Kubernetes plugin in `plugins/kubernetes/kubernetes.go` for a reference implementation.

## License

MIT License
