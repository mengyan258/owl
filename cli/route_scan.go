package cli

import (
	"fmt"
	"path/filepath"

	"bit-labs.cn/owl/cli/generator"
	"github.com/spf13/cobra"
)

// routeScanCmd 路由扫描命令
var routeScanCmd = &cobra.Command{
	Use:   "route:scan [project-path]",
	Short: "扫描项目中的 Swagger 注释并生成路由注册代码",
	Long: `扫描项目中所有 handle 文件的 Swagger 注释，
自动生成路由注册代码。

示例:
  owl route:scan ./my-project
  owl route:scan . --output ./app/route/auto_generated.go`,
	Args: cobra.MaximumNArgs(1),
	Run:  runRouteScan,
}

var (
	outputFile  string
	showRoutes  bool
	packageName string
)

func init() {
	routeScanCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出路由文件路径 (默认: app/route/auto_generated.go)")
	routeScanCmd.Flags().BoolVarP(&showRoutes, "list", "l", false, "只显示扫描到的路由，不生成文件")
	routeScanCmd.Flags().StringVarP(&packageName, "package", "p", "route", "生成文件的包名")
}

// NewRouteScanCommand 创建路由扫描命令
func NewRouteScanCommand() *cobra.Command {
	return routeScanCmd
}

func runRouteScan(cmd *cobra.Command, args []string) {
	// 确定项目路径
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		fmt.Printf("❌ 获取项目路径失败: %v\n", err)
		return
	}

	fmt.Printf("🔍 扫描项目: %s\n", absPath)

	// 创建扫描器
	scanner := generator.NewRouteScanner(absPath)

	// 扫描路由
	if err := scanner.ScanHandles(); err != nil {
		fmt.Printf("❌ 路由扫描失败: %v\n", err)
		return
	}

	routes := scanner.GetRoutes()
	fmt.Printf("✅ 路由扫描完成，发现 %d 个路由\n", len(routes))

	// 扫描 Binds
	if err := scanner.ScanBinds(); err != nil {
		fmt.Printf("❌ Binds 扫描失败: %v\n", err)
		return
	}

	binds := scanner.GetBinds()
	fmt.Printf("✅ Binds 扫描完成，发现 %d 个构造函数\n", len(binds))

	// 如果只是显示路由列表
	if showRoutes {
		displayRoutes(routes)
		return
	}

	// 生成路由文件
	if err := generateRouteFile(scanner, absPath); err != nil {
		fmt.Printf("❌ 生成路由文件失败: %v\n", err)
		return
	}

	// 生成 Binds 文件
	if err := generateBindsFile(scanner, absPath); err != nil {
		fmt.Printf("❌ 生成 Binds 文件失败: %v\n", err)
		return
	}

	fmt.Println("🎉 路由扫描和生成完成！")
}

// displayRoutes 显示路由列表
func displayRoutes(routes []generator.RouteInfo) {
	if len(routes) == 0 {
		fmt.Println("📝 没有发现任何路由")
		return
	}

	fmt.Println("\n📋 发现的路由:")
	fmt.Println("┌────────────┬─────────────────────────────────┬──────────────────┬─────────────────┐")
	fmt.Println("│   方法     │              路径               │      处理器      │      名称       │")
	fmt.Println("├────────────┼─────────────────────────────────┼──────────────────┼─────────────────┤")

	for _, route := range routes {
		method := route.Method
		if method == "" {
			method = "N/A"
		}

		path := route.Path
		if path == "" {
			path = "N/A"
		}

		name := route.Name
		if name == "" {
			name = route.Summary
		}
		if name == "" {
			name = "N/A"
		}

		fmt.Printf("│ %-10s │ %-31s │ %-16s │ %-15s │\n",
			method, truncateString(path, 31),
			truncateString(fmt.Sprintf("%s.%s", route.Package, route.HandlerMethod), 16), truncateString(name, 15))
	}

	fmt.Println("└────────────┴─────────────────────────────────┴──────────────────┴─────────────────┘")
}

// generateRouteFile 生成路由文件
func generateRouteFile(scanner *generator.RouteScanner, projectPath string) error {
	// 确定输出文件路径
	output := outputFile
	if output == "" {
		output = filepath.Join(projectPath, "app", "route", "auto_generated.go")
	}

	fmt.Printf("📝 生成路由文件: %s\n", output)

	// 获取路由信息并创建生成器
	routes := scanner.GetRoutes()
	routeGenerator := generator.NewRouteGenerator(routes)

	// 生成文件
	if err := routeGenerator.Generate(output); err != nil {
		return err
	}

	fmt.Printf("✅ 路由文件生成成功: %s\n", output)
	return nil
}

// generateBindsFile 生成 Binds 文件
func generateBindsFile(scanner *generator.RouteScanner, projectPath string) error {
	// 确定输出路径
	bindsFile := filepath.Join(projectPath, "app", "auto_generated_binds.go")

	// 获取绑定信息并创建生成器
	binds := scanner.GetBinds()
	bindsGenerator := generator.NewBindsGenerator(binds, projectPath)

	// 生成文件
	if err := bindsGenerator.Generate(bindsFile); err != nil {
		return err
	}

	fmt.Printf("📄 Binds 文件已生成: %s\n", bindsFile)
	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
