package cmd

import (
	"envswitch/internal"
	"envswitch/internal/config"
	"envswitch/internal/file"
	"envswitch/internal/project"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environments",
	Long:  "Create, list, view, and manage environments within projects",
}

var envCreateCmd = &cobra.Command{
	Use:   "create <project> <env-name>",
	Short: "Create a new environment",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		envName := args[1]
		description, _ := cmd.Flags().GetString("description")
		tagsStr, _ := cmd.Flags().GetString("tags")

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}

		manager := project.NewManager()

		env := &internal.Environment{
			Name:        envName,
			Description: description,
			Tags:        tags,
			Files:       []internal.FileConfig{},
		}

		err := manager.AddEnvironment(projectName, env)
		checkError(err)

		fmt.Printf("Environment '%s' created in project '%s'\n", envName, projectName)
	},
}

var envListCmd = &cobra.Command{
	Use:   "list [project]",
	Short: "List environments",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		manager := project.NewManager()

		var projectName string
		if len(args) > 0 {
			projectName = args[0]
		} else {
			// 使用默认项目
			projectName = config.GetDefaultProject()
			if projectName == "" {
				fmt.Println("No project specified and no default project set")
				fmt.Println("Usage: envswitch env list <project>")
				return
			}
		}

		environments, err := manager.ListEnvironments(projectName)
		checkError(err)

		if len(environments) == 0 {
			fmt.Printf("No environments found in project '%s'\n", projectName)
			return
		}

		fmt.Printf("Environments in project '%s':\n\n", projectName)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION\tTAGS\tFILES\tCREATED\tLAST SWITCH")

		for _, env := range environments {
			tagsStr := strings.Join(env.Tags, ", ")
			if len(tagsStr) > 20 {
				tagsStr = tagsStr[:17] + "..."
			}

			lastSwitch := "Never"
			if env.LastSwitchAt != nil {
				lastSwitch = env.LastSwitchAt.Format("2006-01-02 15:04")
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
				env.Name,
				truncateString(env.Description, 25),
				tagsStr,
				len(env.Files),
				env.CreatedAt.Format("2006-01-02 15:04"),
				lastSwitch,
			)
		}
		w.Flush()
	},
}

var envShowCmd = &cobra.Command{
	Use:   "show <project> <env-name>",
	Short: "Show environment details",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		projectName := args[0]
		envName := args[1]

		manager := project.NewManager()
		env, err := manager.GetEnvironment(projectName, envName)
		checkError(err)

		fmt.Printf("Environment: %s\n", env.Name)
		fmt.Printf("ID: %s\n", env.ID)
		fmt.Printf("Description: %s\n", env.Description)
		fmt.Printf("Tags: %s\n", strings.Join(env.Tags, ", "))
		fmt.Printf("Created: %s\n", env.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", env.UpdatedAt.Format("2006-01-02 15:04:05"))

		if env.LastSwitchAt != nil {
			fmt.Printf("Last Switch: %s\n", env.LastSwitchAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("Last Switch: Never\n")
		}

		fmt.Printf("Files: %d\n", len(env.Files))

		if len(env.Files) > 0 {
			fmt.Println("\nFile Configurations:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "  SOURCE\tTARGET\tDESCRIPTION")

			for _, file := range env.Files {
				fmt.Fprintf(w, "  %s\t%s\t%s\n",
					file.SourcePath,
					file.TargetPath,
					file.Description,
				)
			}
			w.Flush()
		}
	},
}

var envUpdateCmd = &cobra.Command{
	Use:   "update <project> <env-name>",
	Short: "Update environment",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		envName := args[1]

		updates := make(map[string]interface{})

		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			updates["description"] = description
		}

		if cmd.Flags().Changed("tags") {
			tagsStr, _ := cmd.Flags().GetString("tags")
			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
			}
			updates["tags"] = tags
		}

		if len(updates) == 0 {
			fmt.Println("No updates specified")
			return
		}

		manager := project.NewManager()
		env, err := manager.UpdateEnvironment(projectName, envName, updates)
		checkError(err)

		fmt.Printf("Environment '%s' updated successfully\n", env.Name)
	},
}

var envDeleteCmd = &cobra.Command{
	Use:   "delete <project> <env-name>",
	Short: "Delete an environment",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		envName := args[1]
		force, _ := cmd.Flags().GetBool("force")

		manager := project.NewManager()

		// 获取环境信息用于确认
		env, err := manager.GetEnvironment(projectName, envName)
		checkError(err)

		if !force {
			fmt.Printf("Are you sure you want to delete environment '%s' from project '%s'?\n", env.Name, projectName)
			fmt.Print("Type 'yes' to confirm: ")
			var confirmation string
			_, _ = fmt.Scanln(&confirmation)
			if confirmation != "yes" {
				fmt.Println("Operation cancelled")
				return
			}
		}

		err = manager.RemoveEnvironment(projectName, envName)
		checkError(err)

		fmt.Printf("Environment '%s' deleted from project '%s'\n", env.Name, projectName)
	},
}

var envAddFileCmd = &cobra.Command{
	Use:   "add-file <project> <env-name> <source> <target>",
	Short: "Add file configuration to environment",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		envName := args[1]
		sourcePath := args[2]
		targetPath := args[3]
		description, _ := cmd.Flags().GetString("description")

		manager := project.NewManager()

		// 获取项目和环境ID
		proj, err := manager.GetProject(projectName)
		checkError(err)

		env, err := manager.GetEnvironment(projectName, envName)
		checkError(err)

		fileManager := file.NewManager()
		err = fileManager.AddFileConfig(proj.ID, env.ID, sourcePath, targetPath, description)
		checkError(err)

		fmt.Printf("File configuration added to environment '%s'\n", envName)
		fmt.Printf("Source: %s\n", sourcePath)
		fmt.Printf("Target: %s\n", targetPath)
	},
}

var envRemoveFileCmd = &cobra.Command{
	Use:   "remove-file <project> <env-name> <file-id>",
	Short: "Remove file configuration from environment",
	Args:  cobra.ExactArgs(3),
	Run: func(_ *cobra.Command, args []string) {
		projectName := args[0]
		envName := args[1]
		fileID := args[2]

		manager := project.NewManager()

		// 获取项目和环境ID
		proj, err := manager.GetProject(projectName)
		checkError(err)

		env, err := manager.GetEnvironment(projectName, envName)
		checkError(err)

		fileManager := file.NewManager()
		err = fileManager.RemoveFileConfig(proj.ID, env.ID, fileID)
		checkError(err)

		fmt.Printf("File configuration removed from environment '%s'\n", envName)
	},
}

func init() {
	// env create
	envCreateCmd.Flags().StringP("description", "d", "", "Environment description")
	envCreateCmd.Flags().StringP("tags", "t", "", "Comma-separated tags")

	// env update
	envUpdateCmd.Flags().StringP("description", "d", "", "New description")
	envUpdateCmd.Flags().StringP("tags", "t", "", "New comma-separated tags")

	// env delete
	envDeleteCmd.Flags().BoolP("force", "f", false, "Force delete without confirmation")

	// env add-file
	envAddFileCmd.Flags().StringP("description", "d", "", "File configuration description")

	// 添加子命令
	envCmd.AddCommand(envCreateCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envShowCmd)
	envCmd.AddCommand(envUpdateCmd)
	envCmd.AddCommand(envDeleteCmd)
	envCmd.AddCommand(envAddFileCmd)
	envCmd.AddCommand(envRemoveFileCmd)
}
