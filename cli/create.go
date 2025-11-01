package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bit-labs.cn/owl/cli/generator"
	"bit-labs.cn/owl/cli/prompt"
	"github.com/spf13/cobra"
)

// NewCreateCommand 创建项目命令
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [project-name]",
		Short: "Create a new Owl framework project",
		Long: `Create a new Owl framework project with interactive setup.
This command will guide you through the project configuration process.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runCreate,
	}

	cmd.Flags().BoolP("force", "f", false, "Force create project even if directory exists")
	cmd.Flags().BoolP("skip-git", "", false, "Skip git repository initialization")
	cmd.Flags().BoolP("skip-install", "", false, "Skip dependency installation")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	var projectName string

	// 获取项目名称
	if len(args) > 0 {
		projectName = args[0]
	} else {
		var err error
		projectName, err = prompt.Input("Project name", "my-owl-project", validateProjectName)
		if err != nil {
			return err
		}
	}

	// 获取命令行参数
	force, _ := cmd.Flags().GetBool("force")
	skipGit, _ := cmd.Flags().GetBool("skip-git")
	skipInstall, _ := cmd.Flags().GetBool("skip-install")

	// 检查目录是否存在
	projectPath := filepath.Join(".", projectName)
	if _, err := os.Stat(projectPath); err == nil && !force {
		overwrite, err := prompt.Confirm(fmt.Sprintf("Directory '%s' already exists. Overwrite?", projectName), false)
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println("Project creation cancelled.")
			return nil
		}
	}

	// 交互式配置收集
	config, err := collectProjectConfig()
	if err != nil {
		return err
	}

	config.ProjectName = projectName
	config.ProjectPath = projectPath
	config.SkipGit = skipGit
	config.SkipInstall = skipInstall

	// 生成项目
	fmt.Printf("\n🚀 Creating project '%s'...\n", projectName)

	gen := generator.NewProjectGenerator(config)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	fmt.Printf("\n🎉 项目 '%s' 创建成功！\n", config.ProjectName)
	fmt.Printf("📁 项目路径: %s\n", config.ProjectPath)
	if config.Database != "" {
		fmt.Printf("🗄️  数据库: %s\n", config.Database)
	}
	fmt.Printf("🌐 端口: %s\n", config.Port)

	fmt.Printf("\n📝 下一步操作:\n")
	fmt.Printf("   1. cd %s\n", config.ProjectName)
	fmt.Printf("   2. 复制 .env.example 为 .env 并配置数据库\n")
	fmt.Printf("   3. go mod tidy\n")
	fmt.Printf("   4. go run main.go\n")

	return nil
}

func collectProjectConfig() (*generator.ProjectConfig, error) {
	config := &generator.ProjectConfig{}

	// 使用默认值，不再询问用户
	config.Description = "A new Owl framework project"
	config.Author = "developer"
	config.Port = "8080"
	config.Database = "mysql" // 默认使用mysql

	return config, nil
}

func validateProjectName(input string) error {
	if len(input) == 0 {
		return fmt.Errorf("project name cannot be empty")
	}
	if strings.Contains(input, " ") {
		return fmt.Errorf("project name cannot contain spaces")
	}
	return nil
}

func validatePort(input string) error {
	if len(input) == 0 {
		return fmt.Errorf("port cannot be empty")
	}
	// 简单的端口验证
	return nil
}

func showCompletionMessage(config *generator.ProjectConfig) {
	fmt.Printf("\n✅ Project '%s' created successfully!\n\n", config.ProjectName)
	fmt.Println("📁 Project structure:")
	fmt.Printf("   %s/\n", config.ProjectName)
	fmt.Println("   ├── app/")
	fmt.Println("   ├── conf/")
	fmt.Println("   ├── go.mod")
	fmt.Println("   └── main.go")

	fmt.Println("\n🚀 Next steps:")
	fmt.Printf("   cd %s\n", config.ProjectName)

	if !config.SkipInstall {
		fmt.Println("   go mod tidy")
	}

	fmt.Println("   go run main.go")

	fmt.Printf("\n🌐 Your application will be available at: http://localhost:%s\n", config.Port)
}
