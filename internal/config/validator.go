package config

import (
	"fmt"

	"github.com/cuongtl1992/grp-cli/internal/models"
)

// Validator handles validation of release plans
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// checkCircularDependencies checks for circular dependencies in job dependencies
func (v *Validator) checkCircularDependencies(jobs []models.Job) error {
	visited := make(map[string]bool)
	stack := make(map[string]bool)

	var checkDeps func(job models.Job) error
	checkDeps = func(job models.Job) error {
		visited[job.Name] = true
		stack[job.Name] = true

		for _, depName := range job.DependsOn {
			if !visited[depName] {
				for _, j := range jobs {
					if j.Name == depName {
						if err := checkDeps(j); err != nil {
							return err
						}
					}
				}
			} else if stack[depName] {
				return fmt.Errorf("circular dependency detected: %s -> %s", job.Name, depName)
			}
		}

		stack[job.Name] = false
		return nil
	}

	for _, job := range jobs {
		if !visited[job.Name] {
			if err := checkDeps(job); err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidatePlan checks if a plan is valid
func (v *Validator) ValidatePlan(plan *models.Plan) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}

	// Check required fields
	if plan.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	
	if plan.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	
	if plan.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	
	if len(plan.Stages) == 0 {
		return fmt.Errorf("at least one stage is required")
	}
	
	// Validate stages
	stageNames := make(map[string]bool)
	for i, stage := range plan.Stages {
		if stage.Name == "" {
			return fmt.Errorf("stage[%d].name is required", i)
		}
		
		if stageNames[stage.Name] {
			return fmt.Errorf("duplicate stage name: %s", stage.Name)
		}
		stageNames[stage.Name] = true
		
		if len(stage.Jobs) == 0 {
			return fmt.Errorf("stage[%s] must have at least one job", stage.Name)
		}
		
		// Validate jobs
		jobNames := make(map[string]bool)
		for j, job := range stage.Jobs {
			if job.Name == "" {
				return fmt.Errorf("stage[%s].job[%d].name is required", stage.Name, j)
			}
			
			if jobNames[job.Name] {
				return fmt.Errorf("duplicate job name in stage %s: %s", stage.Name, job.Name)
			}
			jobNames[job.Name] = true
			
			if job.Type == "" {
				return fmt.Errorf("stage[%s].job[%s].type is required", stage.Name, job.Name)
			}
			
			// Validate job dependencies
			for _, depName := range job.DependsOn {
				if !jobNames[depName] {
					return fmt.Errorf("stage[%s].job[%s] depends on unknown job: %s", stage.Name, job.Name, depName)
				}
			}
		}

		// Check for circular dependencies in each stage
		if err := v.checkCircularDependencies(stage.Jobs); err != nil {
			return fmt.Errorf("in stage %s: %w", stage.Name, err)
		}
	}
	
	// Validate rollback if present
	if plan.Rollback != nil {
		if len(plan.Rollback.Stages) == 0 {
			return fmt.Errorf("rollback must have at least one stage")
		}
		
		// Validate rollback stages
		for i, stage := range plan.Rollback.Stages {
			if stage.Name == "" {
				return fmt.Errorf("rollback.stage[%d].name is required", i)
			}
			
			if len(stage.Jobs) == 0 {
				return fmt.Errorf("rollback.stage[%s] must have at least one job", stage.Name)
			}
			
			// Validate rollback jobs
			jobNames := make(map[string]bool)
			for j, job := range stage.Jobs {
				if job.Name == "" {
					return fmt.Errorf("rollback.stage[%s].job[%d].name is required", stage.Name, j)
				}
				
				if jobNames[job.Name] {
					return fmt.Errorf("duplicate job name in rollback stage %s: %s", stage.Name, job.Name)
				}
				jobNames[job.Name] = true
				
				if job.Type == "" {
					return fmt.Errorf("rollback.stage[%s].job[%s].type is required", stage.Name, job.Name)
				}
				
				// Validate job dependencies
				for _, depName := range job.DependsOn {
					if !jobNames[depName] {
						return fmt.Errorf("rollback.stage[%s].job[%s] depends on unknown job: %s", stage.Name, job.Name, depName)
					}
				}
			}
		}
	}
	
	return nil
}