package cmd

import (
	"envswitch/internal/config"
	"envswitch/internal/file"
	"envswitch/internal/project"
	"fmt"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [project] <env-name>",
	Short: "Switch to an environment",
	Long:  "Switch to the specified environment, replacing files according to the configuration",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var projectName, envName string

		if len(args) == 1 {
			// 使用默认项目
			projectName = config.GetDefaultProject()
			if projectName == "" {
				fmt.Println("No default project set. Please specify project name.")
				fmt.Println("Usage: envswitch switch <project> <env-name>")
				return
			}
			envName = args[0]
		} else {
			projectName = args[0]
			envName = args[1]
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		manager := project.NewManager()
		fileManager := file.NewManager()

		// 获取项目和环境信息
		proj, err := manager.GetProject(projectName)
		checkError(err)

		env, err := manager.GetEnvironment(projectName, envName)
		checkError(err)

		if len(env.Files) == 0 {
			fmt.Printf("Environment '%s' has no file configurations\n", envName)
			return
		}

		if dryRun {
			fmt.Printf("Dry run: Would switch to environment '%s' in project '%s'\n", envName, projectName)
			fmt.Printf("Files that would be switched:\n")
			for _, fileConfig := range env.Files {
				fmt.Printf("  %s -> %s\n", fileConfig.SourcePath, fileConfig.TargetPath)
			}
			return
		}

		fmt.Printf("Switching to environment '%s' in project '%s'...\n", envName, projectName)

		// 执行切换
		err = fileManager.SwitchEnvironment(proj.ID, env.ID)
		if err != nil {
			fmt.Printf("Failed to switch environment: %v\n", err)
			fmt.Println("You may need to run 'envswitch rollback' to restore previous state")
			return
		}

		fmt.Printf("Successfully switched to environment '%s'\n", envName)
		fmt.Printf("Switched %d files\n", len(env.Files))
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current environment status",
	Run: func(_ *cobra.Command, _ []string) {
		fileManager := file.NewManager()
		projectManager := project.NewManager()

		state, err := fileManager.GetCurrentState()
		checkError(err)

		if state.CurrentProject == "" {
			fmt.Println("No environment is currently active")
			return
		}

		// 获取项目和环境信息
		proj, err := projectManager.GetProject(state.CurrentProject)
		if err != nil {
			fmt.Printf("Warning: Could not load current project: %v\n", err)
			fmt.Printf("Current project ID: %s\n", state.CurrentProject)
			fmt.Printf("Current environment ID: %s\n", state.CurrentEnvironment)
			return
		}

		env, err := projectManager.GetEnvironment(proj.ID, state.CurrentEnvironment)
		if err != nil {
			fmt.Printf("Warning: Could not load current environment: %v\n", err)
			fmt.Printf("Current project: %s\n", proj.Name)
			fmt.Printf("Current environment ID: %s\n", state.CurrentEnvironment)
			return
		}

		fmt.Printf("Current Status:\n")
		fmt.Printf("Project: %s\n", proj.Name)
		fmt.Printf("Environment: %s\n", env.Name)

		if state.LastSwitchAt != nil {
			fmt.Printf("Last Switch: %s\n", state.LastSwitchAt.Format("2006-01-02 15:04:05"))
		}

		if state.BackupID != "" {
			fmt.Printf("Backup ID: %s\n", state.BackupID)
		}

		fmt.Printf("Active Files: %d\n", len(env.Files))

		if len(env.Files) > 0 {
			fmt.Println("\nActive file configurations:")
			for _, fileConfig := range env.Files {
				fmt.Printf("  %s -> %s\n", fileConfig.SourcePath, fileConfig.TargetPath)
			}
		}
	},
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback [backup-id]",
	Short: "Rollback to previous state",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileManager := file.NewManager()

		var backupID string
		if len(args) > 0 {
			backupID = args[0]
		} else {
			// 使用当前状态中的备份ID
			state, err := fileManager.GetCurrentState()
			checkError(err)

			if state.BackupID == "" {
				fmt.Println("No backup available for rollback")
				fmt.Println("You can specify a backup ID: envswitch rollback <backup-id>")
				return
			}
			backupID = state.BackupID
		}

		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to rollback to backup '%s'?\n", backupID)
			fmt.Print("Type 'yes' to confirm: ")
			var confirmation string
			_, _ = fmt.Scanln(&confirmation)
			if confirmation != "yes" {
				fmt.Println("Operation cancelled")
				return
			}
		}

		fmt.Printf("Rolling back to backup '%s'...\n", backupID)

		err := fileManager.RollbackFromBackup(backupID)
		checkError(err)

		fmt.Println("Rollback completed successfully")
	},
}

func init() {
	// switch flags
	switchCmd.Flags().BoolP("dry-run", "n", false, "Show what would be done without actually doing it")

	// rollback flags
	rollbackCmd.Flags().BoolP("force", "f", false, "Force rollback without confirmation")
}
