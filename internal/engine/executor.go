package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cuongtl1992/grp-cli/internal/models"
	"github.com/cuongtl1992/grp-cli/internal/plugins"
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
func (e *Executor) ExecuteGraph(ctx context.Context, graph *JobGraph, stageResult *models.StageResult, dryRun bool) error {
	// Check for cycles in the dependency graph
	if graph.HasCycles() {
		return fmt.Errorf("dependency cycle detected in job graph")
	}

	// Get ready jobs (those with no dependencies)
	readyJobs := graph.GetReadyJobs()

	// Process until no more jobs are available
	for len(readyJobs) > 0 {
		var wg sync.WaitGroup
		jobResults := make([]models.JobResult, len(readyJobs))

		// Execute ready jobs in parallel
		for i, job := range readyJobs {
			wg.Add(1)

			go func(i int, job models.Job) {
				defer wg.Done()

				// Execute the job
				result := models.JobResult{
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
					success, message, data := e.executeJob(ctx, job)
					result.Success = success
					result.Message = message
					result.Data = data
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
func (e *Executor) executeJob(ctx context.Context, job models.Job) (bool, string, map[string]interface{}) {
	fmt.Printf("Executing job: %s (type: %s)\n", job.Name, job.Type)

	// Execute the job using the plugin manager
	result, err := e.pluginManager.ExecutePlugin(ctx, job.Type, job.Config)
	if err != nil {
		return false, fmt.Sprintf("Failed to execute job: %v", err), nil
	}

	if !result.Success {
		return false, result.Message, result.Data
	}

	return true, result.Message, result.Data
}
