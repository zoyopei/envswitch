package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/zoyopei/envswitch/internal"
	"github.com/zoyopei/envswitch/internal/config"
	"github.com/zoyopei/envswitch/internal/project"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  "Create, list, view, and manage projects",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		description, _ := cmd.Flags().GetString("description")

		manager := project.NewManager()
		proj, err := manager.CreateProject(name, description)
		checkError(err)

		fmt.Printf("Project '%s' created successfully (ID: %s)\n", proj.Name, proj.ID)
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Run: func(cmd *cobra.Command, _ []string) {
		manager := project.NewManager()
		projects, err := manager.ListProjects()
		checkError(err)

		if len(projects) == 0 {
			fmt.Println("No projects found")
			return
		}

		// 获取当前应用状态
		storage := manager.GetStorage()
		appState, err := storage.LoadAppState()
		if err != nil {
			fmt.Printf("Warning: failed to load app state: %v\n", err)
			appState = &internal.AppState{}
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tDESCRIPTION\tENVIRONMENTS\tCREATED\tUPDATED")

		for _, p := range projects {
			// 检查是否为当前项目
			marker := ""
			if appState.CurrentProject == p.Name || appState.CurrentProject == p.ID {
				marker = "*"
			}
			
			_, _ = fmt.Fprintf(w, "%s%s\t%s\t%d\t%s\t%s\n",
				marker,
				p.Name,
				truncateString(p.Description, 30),
				len(p.Environments),
				p.CreatedAt.Format("2006-01-02 15:04"),
				p.UpdatedAt.Format("2006-01-02 15:04"),
			)
		}
		_ = w.Flush()
		
		// 显示图例
		fmt.Println("\n* = Current project")
	},
}

var projectShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show project details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		identifier := args[0]

		manager := project.NewManager()
		proj, err := manager.GetProject(identifier)
		checkError(err)

		fmt.Printf("Project: %s\n", proj.Name)
		fmt.Printf("ID: %s\n", proj.ID)
		fmt.Printf("Description: %s\n", proj.Description)
		fmt.Printf("Created: %s\n", proj.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", proj.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Environments: %d\n", len(proj.Environments))

		if len(proj.Environments) > 0 {
			fmt.Println("\nEnvironments:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "  NAME\tDESCRIPTION\tFILES\tLAST SWITCH")

			for _, env := range proj.Environments {
				lastSwitch := "Never"
				if env.LastSwitchAt != nil {
					lastSwitch = env.LastSwitchAt.Format("2006-01-02 15:04")
				}

				_, _ = fmt.Fprintf(w, "  %s\t%s\t%d\t%s\n",
					env.Name,
					truncateString(env.Description, 25),
					len(env.Files),
					lastSwitch,
				)
			}
			_ = w.Flush()
		}
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		identifier := args[0]
		force, _ := cmd.Flags().GetBool("force")

		manager := project.NewManager()

		// 获取项目信息用于确认
		proj, err := manager.GetProject(identifier)
		checkError(err)

		if !force {
			fmt.Printf("Are you sure you want to delete project '%s'? This action cannot be undone.\n", proj.Name)
			fmt.Print("Type 'yes' to confirm: ")
			var confirmation string
			_, _ = fmt.Scanln(&confirmation)
			if confirmation != "yes" {
				fmt.Println("Operation cancelled")
				return
			}
		}

		err = manager.DeleteProject(identifier)
		checkError(err)

		fmt.Printf("Project '%s' deleted successfully\n", proj.Name)
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update <n>",
	Short: "Update a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		identifier := args[0]
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		// 检查是否至少有一个更新字段
		if name == "" && description == "" {
			fmt.Println("Error: At least one of --name or --description must be provided")
			return
		}

		manager := project.NewManager()

		// 构建更新映射
		updates := make(map[string]interface{})
		if name != "" {
			updates["name"] = name
		}
		if description != "" {
			updates["description"] = description
		}

		proj, err := manager.UpdateProject(identifier, updates)
		checkError(err)

		fmt.Printf("Project '%s' updated successfully\n", proj.Name)

		// 显示更新后的信息
		fmt.Printf("  Name: %s\n", proj.Name)
		fmt.Printf("  Description: %s\n", proj.Description)
		fmt.Printf("  Updated: %s\n", proj.UpdatedAt.Format("2006-01-02 15:04:05"))
	},
}

var projectSetDefaultCmd = &cobra.Command{
	Use:   "set-default <n>",
	Short: "Set the default project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		identifier := args[0]

		// 验证项目存在
		manager := project.NewManager()
		proj, err := manager.GetProject(identifier)
		checkError(err)

		// 设置默认项目
		err = config.SetDefaultProject(proj.Name)
		checkError(err)

		fmt.Printf("Default project set to '%s'\n", proj.Name)
	},
}

func init() {
	// project create
	projectCreateCmd.Flags().StringP("description", "d", "", "Project description")

	// project delete
	projectDeleteCmd.Flags().BoolP("force", "f", false, "Force delete without confirmation")

	// project update
	projectUpdateCmd.Flags().StringP("name", "n", "", "New project name")
	projectUpdateCmd.Flags().StringP("description", "d", "", "New project description")

	// 添加子命令
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectShowCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectSetDefaultCmd)
	projectCmd.AddCommand(projectUpdateCmd)
}

// 辅助函数
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
