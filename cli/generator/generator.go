package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ProjectGenerator 项目生成器
type ProjectGenerator struct {
	config *ProjectConfig
}

// NewProjectGenerator 创建项目生成器
func NewProjectGenerator(config *ProjectConfig) *ProjectGenerator {
	return &ProjectGenerator{
		config: config,
	}
}

// Generate 生成项目
func (g *ProjectGenerator) Generate() error {
	// 创建项目目录
	if err := g.createProjectDirectories(); err != nil {
		return err
	}

	// 生成文件
	if err := g.generateFiles(); err != nil {
		return err
	}

	// 创建安装器
	installer := NewInstaller(g.config.ProjectPath, g.config)

	// 初始化 Git 仓库
	if !g.config.SkipGit {
		if err := installer.InitializeGit(); err != nil {
			return err
		}
	}

	// 安装依赖
	if !g.config.SkipInstall {
		if err := installer.InstallDependencies(); err != nil {
			return err
		}
	}

	return nil
}

// createProjectDirectories 创建项目目录结构
func (g *ProjectGenerator) createProjectDirectories() error {
	fmt.Println("📁 Creating project directories...")

	// 创建根目录
	if err := os.MkdirAll(g.config.ProjectPath, 0755); err != nil {
		return err
	}

	// 创建子目录
	for _, dir := range g.config.GetProjectStructure() {
		dirPath := filepath.Join(g.config.ProjectPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}

	return nil
}

// generateFiles 生成项目文件
func (g *ProjectGenerator) generateFiles() error {
	fmt.Println("📝 Generating project files...")

	templateVars := g.config.GetTemplateVars()

	// 生成基础文件
	files := g.getTemplateFiles(g.config)

	for _, file := range files {
		if err := g.generateFile(file, templateVars); err != nil {
			return fmt.Errorf("failed to generate file %s: %w", file.Path, err)
		}
	}

	return nil
}

// generateFile 生成单个文件
func (g *ProjectGenerator) generateFile(file TemplateFile, vars map[string]interface{}) error {
	// 解析模板
	tmpl, err := template.New(file.Name).Parse(file.Content)
	if err != nil {
		return err
	}

	// 创建文件
	filePath := filepath.Join(g.config.ProjectPath, file.Path)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	// 创建文件
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 执行模板
	return tmpl.Execute(f, vars)
}
