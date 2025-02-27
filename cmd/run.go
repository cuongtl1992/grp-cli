package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/cuongtl1992/grp-cli/internal/config"
	"github.com/cuongtl1992/grp-cli/internal/engine"
	"github.com/cuongtl1992/grp-cli/internal/plugins"
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
		loader := config.NewLoader()
		plan, err := loader.LoadPlan(planFile)
		if err != nil {
			return fmt.Errorf("failed to load plan: %w", err)
		}
		
		// Validate the plan
		validator := config.NewValidator()
		if err := validator.ValidatePlan(plan); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		
		// Get execution options from flags
		autoRollback, _ := cmd.Flags().GetBool("auto-rollback")
		skipApproval, _ := cmd.Flags().GetBool("skip-approval")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		
		// Initialize plugin manager
		pluginDir, _ := cmd.Flags().GetString("plugin-dir")
		if pluginDir == "" {
			// Default to plugins directory in current working directory
			pluginDir = "./plugins"
		}
		
		pluginManager := plugins.NewManager(pluginDir)
		
		// Load plugins
		if err := pluginManager.LoadPlugins(); err != nil {
			fmt.Printf("Warning: Failed to load plugins: %v\n", err)
		}
		
		// Create orchestrator
		orchestrator := engine.NewOrchestrator(pluginManager)
		
		// Execute the plan
		options := engine.ExecuteOptions{
			AutoRollback: autoRollback,
			SkipApproval: skipApproval,
			DryRun:       dryRun,
		}
		
		fmt.Printf("Starting execution of plan: %s\n", plan.Metadata.Name)
		startTime := time.Now()
		
		result, err := orchestrator.ExecutePlan(ctx, plan, options)
		if err != nil {
			fmt.Printf("Execution failed: %v\n", err)
			return err
		}
		
		// Display result summary
		fmt.Printf("\nExecution completed successfully in %s\n", time.Since(startTime))
		fmt.Printf("ID: %s\n", result.ID)
		fmt.Printf("Total stages: %d, Jobs: %d\n", result.TotalStages, result.TotalJobs)
		fmt.Printf("Completed jobs: %d, Failed jobs: %d\n", result.CompletedJobs, result.FailedJobs)
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	
	// Local flags
	runCmd.Flags().Bool("auto-rollback", false, "Automatically rollback on failure")
	runCmd.Flags().Bool("skip-approval", false, "Skip approval steps")
	runCmd.Flags().Bool("dry-run", false, "Validate and simulate execution without making changes")
	runCmd.Flags().String("plugin-dir", "", "Directory containing plugins (default: ./plugins)")
} 