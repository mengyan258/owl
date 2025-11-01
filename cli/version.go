package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	Version   = "1.0.0"
	BuildDate = "2024-01-01"
)

// NewVersionCommand 创建版本命令
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display the version of Owl CLI and runtime information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Owl CLI v%s\n", Version)
			fmt.Printf("Build Date: %s\n", BuildDate)
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
