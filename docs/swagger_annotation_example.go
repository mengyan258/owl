package example

import "github.com/gin-gonic/gin"

// 这是一个展示标准 Swagger 注释格式的示例文件
// 请按照以下格式在您的 handle 文件中添加注释

// CreateUser 创建用户
// @Summary 创建新用户
// @Description 创建一个新的用户账户，需要提供用户名、邮箱和密码
// @Tags 用户管理
// @Router /api/v1/users [POST]
// @Access Authorized
// @Name 创建用户
// @Param username body string true "用户名" minlength(3) maxlength(50)
// @Param email body string true "邮箱地址" format(email)
// @Param password body string true "密码" minlength(6)
// @Param role_ids body []int false "角色ID列表"
// @Success 200 {object} model.User "用户创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 409 {object} ErrorResponse "用户名或邮箱已存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
func (h *UserHandle) CreateUser(ctx *gin.Context) {
	// 实现代码...
}

// GetUser 获取用户详情
// @Summary 根据ID获取用户详情
// @Description 通过用户ID获取用户的详细信息
// @Tags 用户管理
// @Router /api/v1/users/{id} [GET]
// @Access Authorized
// @Name 获取用户详情
// @Param id path int true "用户ID" minimum(1)
// @Param include query string false "包含的关联数据" Enums(roles,permissions,profile)
// @Success 200 {object} model.User "用户信息"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
func (h *UserHandle) GetUser(ctx *gin.Context) {
	// 实现代码...
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 更新指定用户的信息，支持部分更新
// @Tags 用户管理
// @Router /api/v1/users/{id} [PUT]
// @Access Authorized
// @Name 更新用户信息
// @Param id path int true "用户ID" minimum(1)
// @Param username body string false "用户名" minlength(3) maxlength(50)
// @Param email body string false "邮箱地址" format(email)
// @Param status body string false "用户状态" Enums(active,inactive,suspended)
// @Success 200 {object} model.User "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 409 {object} ErrorResponse "用户名或邮箱已存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
func (h *UserHandle) UpdateUser(ctx *gin.Context) {
	// 实现代码...
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 软删除指定的用户账户
// @Tags 用户管理
// @Router /api/v1/users/{id} [DELETE]
// @Access Authorized
// @Name 删除用户
// @Param id path int true "用户ID" minimum(1)
// @Success 200 {object} SuccessResponse "删除成功"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
func (h *UserHandle) DeleteUser(ctx *gin.Context) {
	// 实现代码...
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表，支持搜索和过滤
// @Tags 用户管理
// @Router /api/v1/users [GET]
// @Access Authorized
// @Name 获取用户列表
// @Param page query int false "页码" minimum(1) default(1)
// @Param size query int false "每页数量" minimum(1) maximum(100) default(10)
// @Param search query string false "搜索关键词" maxlength(100)
// @Param status query string false "用户状态过滤" Enums(active,inactive,suspended)
// @Param role_id query int false "角色ID过滤" minimum(1)
// @Param sort query string false "排序字段" Enums(id,username,email,created_at) default(id)
// @Param order query string false "排序方向" Enums(asc,desc) default(desc)
// @Success 200 {object} PaginatedResponse{data=[]model.User} "用户列表"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
func (h *UserHandle) ListUsers(ctx *gin.Context) {
	// 实现代码...
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户通过用户名/邮箱和密码进行登录认证
// @Tags 认证
// @Router /api/v1/auth/login [POST]
// @Access Public
// @Name 用户登录
// @Param login body string true "登录名（用户名或邮箱）" minlength(3)
// @Param password body string true "密码" minlength(6)
// @Param remember body bool false "记住登录状态" default(false)
// @Success 200 {object} LoginResponse "登录成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "用户名或密码错误"
// @Failure 423 {object} ErrorResponse "账户已被锁定"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
func (h *UserHandle) Login(ctx *gin.Context) {
	// 实现代码...
}

// ChangePassword 修改密码
// @Summary 修改用户密码
// @Description 用户修改自己的登录密码
// @Tags 用户管理
// @Router /api/v1/users/me/password [PUT]
// @Access Authenticated
// @Name 修改密码
// @Param old_password body string true "原密码" minlength(6)
// @Param new_password body string true "新密码" minlength(6)
// @Param confirm_password body string true "确认新密码" minlength(6)
// @Success 200 {object} SuccessResponse "密码修改成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "原密码错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
func (h *UserHandle) ChangePassword(ctx *gin.Context) {
	// 实现代码...
}

/*
Swagger 注释标签说明:

必需标签:
- @Router: 完整的路由路径和HTTP方法，格式: /api/v1/path [METHOD]
- @Access: 访问级别 (Public, Authenticated, Authorized, AdminOnly)

推荐标签:
- @Summary: 简短的接口描述
- @Description: 详细的接口描述
- @Tags: 接口分组标签
- @Name: 接口名称（用于权限控制）

参数标签:
- @Param: 参数定义
  格式: name in type required "description" [constraints]
  - name: 参数名
  - in: 参数位置 (path, query, body, header, formData)
  - type: 参数类型 (string, int, bool, array, object)
  - required: 是否必需 (true/false)
  - description: 参数描述
  - constraints: 约束条件 (可选)
    - minlength(n): 最小长度
    - maxlength(n): 最大长度
    - minimum(n): 最小值
    - maximum(n): 最大值
    - format(email): 格式验证
    - Enums(a,b,c): 枚举值
    - default(value): 默认值

响应标签:
- @Success: 成功响应
  格式: code {type} model "description"
- @Failure: 失败响应
  格式: code {type} model "description"

访问级别说明:
- Public: 开放接口，无需认证
- Authenticated: 需要登录认证
- Authorized: 需要权限授权
- AdminOnly: 仅超级管理员

路径规范:
- 使用完整的 API 路径，包含版本号
- 路径参数使用 {param} 格式
- 示例: /api/v1/users/{id}/roles

标签分组建议:
- 用户管理
- 角色管理
- 权限管理
- 系统管理
- 认证
- 文件管理
*/
