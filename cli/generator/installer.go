package generator

import (
	"fmt"
	"os"
	"os/exec"
)

// Installer 负责项目的安装和初始化
type Installer struct {
	projectPath string
	config      *ProjectConfig
}

// NewInstaller 创建新的安装器
func NewInstaller(projectPath string, config *ProjectConfig) *Installer {
	return &Installer{
		projectPath: projectPath,
		config:      config,
	}
}

// InstallDependencies 安装项目依赖
func (i *Installer) InstallDependencies() error {
	fmt.Println("📦 正在安装项目依赖...")

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = i.projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	fmt.Println("✅ 依赖安装完成")
	return nil
}

// InitializeGit 初始化 Git 仓库
func (i *Installer) InitializeGit() error {
	if !i.isGitInstalled() {
		fmt.Println("⚠️  Git 未安装，跳过 Git 仓库初始化")
		return nil
	}

	fmt.Println("📁 正在初始化 Git 仓库...")

	// 初始化 Git 仓库
	cmd := exec.Command("git", "init")
	cmd.Dir = i.projectPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// 添加所有文件
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = i.projectPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// 创建初始提交
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = i.projectPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	fmt.Println("✅ Git 仓库初始化完成")
	return nil
}

// isGitInstalled 检查是否安装了 Git
func (i *Installer) isGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}
