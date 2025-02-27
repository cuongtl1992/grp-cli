package engine

import (
	"github.com/cuongtl1992/grp-cli/internal/models"
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

// HasCycles checks if the dependency graph has cycles
func (g *JobGraph) HasCycles() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	// Check each node
	for name := range g.jobs {
		if !visited[name] {
			if g.hasCyclesDFS(name, visited, recStack) {
				return true
			}
		}
	}
	
	return false
}

// hasCyclesDFS performs depth-first search to detect cycles
func (g *JobGraph) hasCyclesDFS(node string, visited, recStack map[string]bool) bool {
	// Mark current node as visited and add to recursion stack
	visited[node] = true
	recStack[node] = true
	
	// Check all dependencies
	for _, dep := range g.dependencies[node] {
		// If not visited, check recursively
		if !visited[dep] {
			if g.hasCyclesDFS(dep, visited, recStack) {
				return true
			}
		} else if recStack[dep] {
			// If already in recursion stack, we found a cycle
			return true
		}
	}
	
	// Remove from recursion stack
	recStack[node] = false
	return false
} 