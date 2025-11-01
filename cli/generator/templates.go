package generator

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed stub
var stubFS embed.FS

// TemplateFile 模板文件结构
type TemplateFile struct {
	Name    string
	Path    string
	Content string
}

// getTemplateFiles 根据项目配置获取模板文件
func (g *ProjectGenerator) getTemplateFiles(config *ProjectConfig) []TemplateFile {
	var templates []TemplateFile

	// 基础模板文件
	baseTemplates := g.getBaseTemplates()
	templates = append(templates, baseTemplates...)

	// 通用模板（包含API和Web功能）
	universalTemplates := g.getUniversalTemplates()
	templates = append(templates, universalTemplates...)

	return templates
}

// getBaseTemplates 获取基础模板文件
func (g *ProjectGenerator) getBaseTemplates() []TemplateFile {
	baseTemplates := []TemplateFile{
		g.loadStubTemplate("base/main.go.stub", "main.go", "main.go"),
		g.loadStubTemplate("base/go.mod.stub", "go.mod", "go.mod"),
		g.loadStubTemplate("base/README.md.stub", "README.md", "README.md"),
		g.loadStubTemplate("app/database/database.go.stub", "database.go", "app/database/database.go"),
		g.loadStubTemplate("app/model/base.go.stub", "base.go", "app/model/base.go"),
		g.loadStubTemplate("app/model/user.go.stub", "user.go", "app/model/user.go"),
		g.loadStubTemplate("app/service/user_service.go.stub", "user_service.go", "app/service/user_service.go"),
		g.loadStubTemplate("app/repository/user_repository.go.stub", "user_repository.go", "app/repository/user_repository.go"),
	}

	return baseTemplates
}

// getUniversalTemplates 获取通用模板文件（包含API和Web功能）
func (g *ProjectGenerator) getUniversalTemplates() []TemplateFile {
	universalTemplates := []TemplateFile{
		g.loadStubTemplate("app/app.go.stub", "app.go", "app/app.go"),
		g.loadStubTemplate("app/handle/web/web_handle.go.stub", "web_handle.go", "app/handle/web/web_handle.go"),
		g.loadStubTemplate("app/route/api_route.go.stub", "api_route.go", "app/route/api.go"),
		g.loadStubTemplate("app/route/web_route.go.stub", "web_route.go", "app/route/web.go"),
		g.loadStubTemplate("app/handle/v1/api_handle.go.stub", "api_handle.go", "app/handle/v1/api_handle.go"),
		g.loadStubTemplate("templates/index.html.stub", "index.html", "templates/index.html"),
		g.loadStubTemplate("templates/users/list.html.stub", "users_list.html", "templates/users/list.html"),
		g.loadStubTemplate("templates/users/create.html.stub", "users_create.html", "templates/users/create.html"),
		g.loadStubTemplate("templates/users/show.html.stub", "users_show.html", "templates/users/show.html"),
		g.loadStubTemplate("templates/users/edit.html.stub", "users_edit.html", "templates/users/edit.html"),
		g.loadStubTemplate("templates/error.html.stub", "error.html", "templates/error.html"),
	}

	return universalTemplates
}

// loadStubTemplate 从stub文件加载模板内容
func (g *ProjectGenerator) loadStubTemplate(stubPath, name, targetPath string) TemplateFile {
	// embed文件系统使用正斜杠路径
	fullPath := "stub/" + strings.ReplaceAll(stubPath, "\\", "/")
	content, err := stubFS.ReadFile(fullPath)
	if err != nil {
		// 如果读取失败，返回空内容，但保留结构
		return TemplateFile{
			Name:    name,
			Path:    targetPath,
			Content: "",
		}
	}

	return TemplateFile{
		Name:    name,
		Path:    targetPath,
		Content: string(content),
	}
}

// getStubFiles 获取所有stub文件列表
func (g *ProjectGenerator) getStubFiles() []string {
	var files []string

	err := fs.WalkDir(stubFS, "stub", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".stub") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return []string{}
	}

	return files
}
