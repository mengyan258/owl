package generator

import (
	"strings"
)

// ProjectConfig 项目配置
type ProjectConfig struct {
	// 基本信息
	ProjectName string `json:"project_name"`
	ProjectPath string `json:"project_path"`
	Description string `json:"description"`
	Author      string `json:"author"`

	// 数据库配置
	Database    string `json:"database"` // mysql, postgresql, sqlite
	EnableRedis bool   `json:"enable_redis"`

	// 功能配置
	EnableAuth bool `json:"enable_auth"`

	// 服务配置
	Port string `json:"port"`

	// 构建选项
	SkipGit     bool `json:"skip_git"`
	SkipInstall bool `json:"skip_install"`
}

// GetModuleName 获取 Go 模块名
func (c *ProjectConfig) GetModuleName() string {
	return strings.ToLower(c.ProjectName)
}

// GetPackageName 获取包名
func (c *ProjectConfig) GetPackageName() string {
	return strings.ReplaceAll(strings.ToLower(c.ProjectName), "-", "")
}

// GetAppName 获取应用名称
func (c *ProjectConfig) GetAppName() string {
	// 将项目名转换为有效的Go结构体名称
	name := strings.ReplaceAll(c.ProjectName, "-", "")
	name = strings.ReplaceAll(name, "_", "")
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	return name
}

// GetDatabaseDriver 获取数据库驱动
func (c *ProjectConfig) GetDatabaseDriver() string {
	switch c.Database {
	case "mysql":
		return "mysql"
	case "postgresql":
		return "postgres"
	case "sqlite":
		return "sqlite"
	default:
		return "mysql"
	}
}

// GetDatabaseImport 获取数据库导入包
func (c *ProjectConfig) GetDatabaseImport() string {
	switch c.Database {
	case "mysql":
		return "gorm.io/driver/mysql"
	case "postgresql":
		return "gorm.io/driver/postgres"
	case "sqlite":
		return "gorm.io/driver/sqlite"
	default:
		return "gorm.io/driver/mysql"
	}
}

// GetDatabaseDSN 获取数据库连接字符串示例
func (c *ProjectConfig) GetDatabaseDSN() string {
	switch c.Database {
	case "mysql":
		return "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	case "postgresql":
		return "host=localhost user=user password=password dbname=dbname port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	case "sqlite":
		return "./data.db"
	default:
		return "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	}
}

// GetProjectStructure 获取项目目录结构
func (c *ProjectConfig) GetProjectStructure() []string {
	dirs := []string{
		"app",
		"app/database",
		"app/handle",
		"app/handle/v1",
		"app/model",
		"app/route",
		"app/service",
		"app/repository",
		"app/middleware",
		"app/provider",
		"script",
		"logs",
		"storage",
	}

	// 添加通用模板的所有目录
	dirs = append(dirs,
		"app/handle/web",
		"templates",
		"templates/static",
		"templates/static/css",
		"templates/static/js",
		"templates/static/images",
	)

	return dirs
}

// GetTemplateVars 获取模板变量
func (c *ProjectConfig) GetTemplateVars() map[string]interface{} {
	return map[string]interface{}{
		"ProjectName":    c.ProjectName,
		"ModuleName":     c.GetModuleName(),
		"PackageName":    c.GetPackageName(),
		"AppName":        c.GetAppName(),
		"Description":    c.Description,
		"Author":         c.Author,
		"Database":       c.Database,
		"DatabaseDriver": c.GetDatabaseDriver(),
		"DatabaseImport": c.GetDatabaseImport(),
		"DatabaseDSN":    c.GetDatabaseDSN(),
		"EnableRedis":    c.EnableRedis,
		"EnableAuth":     c.EnableAuth,
		"Port":           c.Port,
	}
}
