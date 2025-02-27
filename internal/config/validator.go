package config

import (
	"fmt"

	"github.com/yourusername/grp-cli/internal/models"
)

// Validator handles validation of release plans
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidatePlan checks if a plan is valid
func (v *Validator) ValidatePlan(plan *models.Plan) error {
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