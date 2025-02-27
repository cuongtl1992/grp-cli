package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/cuongtl1992/grp-cli/internal/models"
	"github.com/cuongtl1992/grp-cli/internal/plugins"
)

// ExecuteOptions contains options for plan execution
type ExecuteOptions struct {
	AutoRollback bool
	SkipApproval bool
	DryRun       bool
}

// Orchestrator manages the execution of a release plan
type Orchestrator struct {
	pluginManager *plugins.Manager
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(pluginManager *plugins.Manager) *Orchestrator {
	return &Orchestrator{
		pluginManager: pluginManager,
	}
}

// ExecutePlan runs a release plan
func (o *Orchestrator) ExecutePlan(ctx context.Context, plan *models.Plan, options ExecuteOptions) (*models.ExecutionResult, error) {
	// Generate unique execution ID
	executionID := uuid.New().String()
	
	// Create execution context with variables
	execCtx := context.WithValue(ctx, "executionID", executionID)
	execCtx = context.WithValue(execCtx, "variables", plan.Variables)
	
	// Create execution result
	result := &models.ExecutionResult{
		ID:         executionID,
		StartTime:  time.Now(),
		TotalStages: len(plan.Stages),
		TotalJobs:   o.countTotalJobs(plan),
	}
	
	// Execute stages sequentially
	for _, stage := range plan.Stages {
		stageResult := models.StageResult{
			Name:      stage.Name,
			StartTime: time.Now(),
		}
		
		// Check if approval is required
		if stage.RequireApproval && !options.SkipApproval {
			// In a real implementation, this would call an approval service
			fmt.Printf("Stage %s requires approval. Waiting for approval...\n", stage.Name)
			
			// For now, we'll just simulate approval
			fmt.Printf("Stage %s approved.\n", stage.Name)
		}
		
		// Execute the stage
		stageErr := o.executeStage(execCtx, &stage, &stageResult, options)
		
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
		
		fmt.Printf("Stage %s completed successfully.\n", stage.Name)
	}
	
	// All stages completed successfully
	return o.finalizeResult(result, true, "Plan execution completed successfully")
}

// executeStage runs all jobs in a stage with proper dependency handling
func (o *Orchestrator) executeStage(ctx context.Context, stage *models.Stage, result *models.StageResult, options ExecuteOptions) error {
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
		stageResult := &models.StageResult{Name: stage.Name}
		if err := executor.ExecuteGraph(ctx, graph, stageResult, false); err != nil {
			fmt.Printf("Rollback stage %s failed: %v\n", stage.Name, err)
			// Continue with other rollback stages even if one fails
		}
	}
	
	fmt.Println("Rollback execution completed")
	return nil
}

// finalizeResult completes the execution result
func (o *Orchestrator) finalizeResult(result *models.ExecutionResult, success bool, message string) (*models.ExecutionResult, error) {
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