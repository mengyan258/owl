package main

import (
	"fmt"
	"os"

	"bit-labs.cn/owl/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "owl",
	Short: "Owl Framework CLI - A powerful Go web framework scaffolding tool",
	Long: `Owl CLI is a command-line tool for creating and managing Owl framework projects.
It provides interactive project generation similar to Vue CLI.`,
	Version: "1.0.0",
}

func main() {
	// 添加子命令
	rootCmd.AddCommand(cli.NewCreateCommand())
	rootCmd.AddCommand(cli.NewVersionCommand())
	rootCmd.AddCommand(cli.NewRouteScanCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
