package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cuongtl1992/grp-cli/internal/config"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [plan file]",
	Short: "Validate a release plan",
	Long: `Validate a release plan defined in YAML format. This command will:
1. Check the syntax of the plan file
2. Validate the structure against the schema
3. Verify that all references are valid
4. Check for circular dependencies`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planFile := args[0]
		
		 // Check if file exists before proceeding
		if _, err := os.Stat(planFile); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("plan file not found: %s", planFile)
			}
			return fmt.Errorf("error accessing plan file: %w", err)
		}
		
		// Create loader and validator
		loader := config.NewLoader()
		validator := config.NewValidator()
		
		// Load the plan
		plan, err := loader.LoadPlan(planFile)
		if err != nil {
			return fmt.Errorf("failed to load plan: %w", err)
		}
		
		// Validate the plan
		if err := validator.ValidatePlan(plan); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		
		fmt.Println("Plan validation successful!")
		if plan.Metadata.Version != "" {
			fmt.Printf("Plan: %s (version: %s)\n", plan.Metadata.Name, plan.Metadata.Version)
		} else {
			fmt.Printf("Plan: %s (version: not specified)\n", plan.Metadata.Name)
		}
		fmt.Printf("Stages: %d\n", len(plan.Stages))
		
		// Print stage information if verbose
		verboseFlag := cmd.Flag("verbose")
		if verboseFlag != nil && verboseFlag.Value.String() == "true" {
			for i, stage := range plan.Stages {
				fmt.Printf("Stage %d: %s (%d jobs)\n", i+1, stage.Name, len(stage.Jobs))
				
				// Print job information
				for j, job := range stage.Jobs {
					fmt.Printf("  Job %d: %s (type: %s)\n", j+1, job.Name, job.Type)
					if len(job.DependsOn) > 0 {
						fmt.Printf("    Dependencies: %v\n", job.DependsOn)
					}
				}
			}
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolP("verbose", "v", false, "Show detailed validation information")
}