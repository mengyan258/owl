package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RouteInfo 路由信息
type RouteInfo struct {
	Method        string      `json:"method"`         // HTTP 方法
	Path          string      `json:"path"`           // 完整路径
	CleanPath     string      `json:"clean_path"`     // 清理后的路径（移除前缀）
	Summary       string      `json:"summary"`        // 摘要
	Description   string      `json:"description"`    // 描述
	Tags          []string    `json:"tags"`           // 标签
	Access        string      `json:"access"`         // 访问级别
	Name          string      `json:"name"`           // 名称
	Parameters    []ParamInfo `json:"parameters"`     // 参数
	Responses     []string    `json:"responses"`      // 响应
	HandlerMethod string      `json:"handler_method"` // 处理器方法名
	HandlerType   string      `json:"handler_type"`   // 处理器类型
	HandlerVar    string      `json:"handler_var"`    // 处理器变量名
	HandlerCall   string      `json:"handler_call"`   // 处理器调用方法
	Package       string      `json:"package"`        // 包名
}

// ParamInfo 参数信息
type ParamInfo struct {
	Name        string `json:"name"`        // 参数名
	Type        string `json:"type"`        // 参数类型
	In          string `json:"in"`          // 参数位置: path, query, body, header
	Required    bool   `json:"required"`    // 是否必需
	Description string `json:"description"` // 参数描述
}

// BindInfo 绑定信息
type BindInfo struct {
	ConstructorName string `json:"constructor_name"` // 构造函数名，如 v1.NewApiHandle
	Package         string `json:"package"`          // 包名，如 v1, oauth, service, repository
	Type            string `json:"type"`             // 类型，如 handle, service, repository
	FilePath        string `json:"file_path"`        // 文件路径
}

// RouteScanner 路由扫描器
type RouteScanner struct {
	projectPath string
	routes      []RouteInfo
	binds       []BindInfo
}

// NewRouteScanner 创建新的路由扫描器
func NewRouteScanner(projectPath string) *RouteScanner {
	return &RouteScanner{
		projectPath: projectPath,
		routes:      make([]RouteInfo, 0),
	}
}

// ScanHandles 扫描所有 handle 文件
func (rs *RouteScanner) ScanHandles() error {
	handleDir := filepath.Join(rs.projectPath, "app", "handle")

	return filepath.Walk(handleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return rs.scanFile(path)
	})
}

// scanFile 扫描单个文件
func (rs *RouteScanner) scanFile(filePath string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("解析文件 %s 失败: %v", filePath, err)
	}

	// 遍历所有函数
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Doc != nil {
				route := rs.parseSwaggerComments(funcDecl, filePath, node.Name.Name)
				if route != nil {
					rs.routes = append(rs.routes, *route)
				}
			}
		}
	}

	return nil
}

// parseSwaggerComments 解析 Swagger 注释
func (rs *RouteScanner) parseSwaggerComments(funcDecl *ast.FuncDecl, filePath, packageName string) *RouteInfo {
	if funcDecl.Doc == nil {
		return nil
	}

	// 根据文件路径推断处理器类型
	handlerType := rs.inferHandlerTypeFromFile(filePath, packageName)

	route := &RouteInfo{
		HandlerMethod: funcDecl.Name.Name,
		HandlerType:   handlerType,
		Package:       packageName,
		Parameters:    make([]ParamInfo, 0),
		Responses:     make([]string, 0),
		Tags:          make([]string, 0),
	}

	var hasSwagger bool

	for _, comment := range funcDecl.Doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

		// 检查是否是 Swagger 注释
		if strings.HasPrefix(text, "@") {
			hasSwagger = true
			rs.parseSwaggerTag(text, route)
		}
	}

	if !hasSwagger {
		return nil
	}

	// 设置清理后的路径
	route.CleanPath = rs.cleanPath(route.Path)

	return route
}

// parseSwaggerTag 解析 Swagger 标签
func (rs *RouteScanner) parseSwaggerTag(text string, route *RouteInfo) {
	// 移除 @ 符号
	text = strings.TrimPrefix(text, "@")

	// 分割标签和值
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 1 {
		return
	}

	tag := strings.ToLower(parts[0])
	value := ""
	if len(parts) > 1 {
		value = strings.TrimSpace(parts[1])
	}

	switch tag {
	case "router":
		rs.parseRouterTag(value, route)
	case "summary":
		route.Summary = value
	case "description":
		route.Description = value
	case "tags":
		route.Tags = strings.Split(value, ",")
		for i, tag := range route.Tags {
			route.Tags[i] = strings.TrimSpace(tag)
		}
	case "param":
		param := rs.parseParamTag(value)
		if param.Name != "" {
			route.Parameters = append(route.Parameters, param)
		}
	case "success", "failure":
		rs.parseResponseTag(value, route, tag)
	case "access":
		route.Access = value
	case "name":
		route.Name = value
	}
}

// parseRouterTag 解析路由标签
// 格式: @Router /api/v1/users [POST]
func (rs *RouteScanner) parseRouterTag(value string, route *RouteInfo) {
	// 使用正则表达式解析
	re := regexp.MustCompile(`^(.+?)\s+\[(\w+)\]$`)
	matches := re.FindStringSubmatch(value)

	if len(matches) == 3 {
		route.Path = strings.TrimSpace(matches[1])
		route.Method = strings.ToUpper(strings.TrimSpace(matches[2]))
	}
}

// cleanPath 清理路径，移除 /api/v1 前缀
func (rs *RouteScanner) cleanPath(path string) string {
	if strings.HasPrefix(path, "/api/v1") {
		return strings.TrimPrefix(path, "/api/v1")
	}
	return path
}

// parseParamTag 解析参数标签
// 格式: @Param id path int true "用户ID"
func (rs *RouteScanner) parseParamTag(value string) ParamInfo {
	parts := strings.Fields(value)
	if len(parts) < 4 {
		return ParamInfo{}
	}

	required := parts[3] == "true"
	description := ""
	if len(parts) > 4 {
		description = strings.Trim(strings.Join(parts[4:], " "), "\"")
	}

	return ParamInfo{
		Name:        parts[0],
		In:          parts[1],
		Type:        parts[2],
		Required:    required,
		Description: description,
	}
}

// parseResponseTag 解析响应标签
// 格式: @Success 200 {object} User "用户信息"
// 或: @Failure 400 {object} Error "错误信息"
func (rs *RouteScanner) parseResponseTag(value string, route *RouteInfo, responseType string) {
	parts := strings.Fields(value)
	if len(parts) < 2 {
		return
	}

	description := ""
	if len(parts) > 2 {
		// 提取描述部分（可能包含引号）
		descParts := parts[2:]
		description = strings.Join(descParts, " ")
		description = strings.Trim(description, `"`)
	}

	route.Responses = append(route.Responses, fmt.Sprintf("%s: %s", responseType, value))
}

// GetRoutes 获取扫描到的路由
func (rs *RouteScanner) GetRoutes() []RouteInfo {
	return rs.routes
}

// inferHandlerTypeFromFile 根据文件路径推断处理器类型
func (rs *RouteScanner) inferHandlerTypeFromFile(filePath, packageName string) string {
	// 获取文件名（不含扩展名）
	fileName := filepath.Base(filePath)
	fileName = strings.TrimSuffix(fileName, ".go")

	// 将下划线分隔的文件名转换为驼峰命名
	// 例如：api_handle -> ApiHandle, dept_handle -> DeptHandle
	parts := strings.Split(fileName, "_")
	var typeName strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			// 首字母大写
			typeName.WriteString(strings.ToUpper(string(part[0])))
			if len(part) > 1 {
				typeName.WriteString(part[1:])
			}
		}
	}

	return fmt.Sprintf("%s.%s", packageName, typeName.String())
}

// GenerateRouteFile 生成路由注册文件
// ScanBinds 扫描所有的构造函数
func (rs *RouteScanner) ScanBinds() error {
	rs.binds = []BindInfo{}

	// 扫描 handle 目录
	handlePath := filepath.Join(rs.projectPath, "app", "handle")
	if err := rs.scanBindsInDirectory(handlePath, "handle"); err != nil {
		return fmt.Errorf("扫描 handle 目录失败: %v", err)
	}

	// 扫描 service 目录
	servicePath := filepath.Join(rs.projectPath, "app", "service")
	if err := rs.scanBindsInDirectory(servicePath, "service"); err != nil {
		return fmt.Errorf("扫描 service 目录失败: %v", err)
	}

	// 扫描 repository 目录
	repositoryPath := filepath.Join(rs.projectPath, "app", "repository")
	if err := rs.scanBindsInDirectory(repositoryPath, "repository"); err != nil {
		return fmt.Errorf("扫描 repository 目录失败: %v", err)
	}

	return nil
}

// scanBindsInDirectory 扫描指定目录中的构造函数
func (rs *RouteScanner) scanBindsInDirectory(dirPath, bindType string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return rs.scanBindsInFile(path, bindType)
	})
}

// scanBindsInFile 扫描文件中的构造函数
func (rs *RouteScanner) scanBindsInFile(filePath, bindType string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	packageName := node.Name.Name

	// 遍历所有函数声明
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if rs.isConstructorFunction(funcDecl) {
				constructorName := rs.buildConstructorName(packageName, funcDecl.Name.Name)

				bindInfo := BindInfo{
					ConstructorName: constructorName,
					Package:         packageName,
					Type:            bindType,
					FilePath:        filePath,
				}

				rs.binds = append(rs.binds, bindInfo)
			}
		}
	}

	return nil
}

// isConstructorFunction 判断是否为构造函数
func (rs *RouteScanner) isConstructorFunction(funcDecl *ast.FuncDecl) bool {
	// 构造函数通常以 New 开头
	return strings.HasPrefix(funcDecl.Name.Name, "New")
}

// buildConstructorName 构建构造函数名称
func (rs *RouteScanner) buildConstructorName(packageName, funcName string) string {
	return fmt.Sprintf("%s.%s", packageName, funcName)
}

// GetBinds 获取所有绑定信息
func (rs *RouteScanner) GetBinds() []BindInfo {
	return rs.binds
}

func (rs *RouteScanner) GenerateRouteFile(outputPath string) error {
	generator := NewRouteGenerator(rs.routes)
	return generator.Generate(outputPath)
}

// GenerateBindsFile 生成 Binds 文件
func (rs *RouteScanner) GenerateBindsFile(outputPath string) error {
	generator := NewBindsGenerator(rs.binds, rs.projectPath)
	return generator.Generate(outputPath)
}
