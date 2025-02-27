package container

import (
	"errors"
)

// Container 是一个依赖注入容器的接口
type Container interface {
	// Bound 确定给定的抽象类型是否已被绑定
	Bound(abstract string) bool

	// Alias 为一个类型设置别名
	Alias(abstract string, alias string) error

	// Tag 为给定的绑定分配一组标签
	Tag(abstracts []string, tags ...interface{}) error

	// Tagged 解析给定标签的所有绑定
	Tagged(tag string) []interface{}

	// Bind 注册一个绑定到容器
	Bind(abstract string, concrete interface{}, shared bool) error

	// BindIf 如果尚未注册，则注册一个绑定
	BindIf(abstract string, concrete interface{}, shared bool) error

	// Singleton 注册一个共享绑定到容器
	Singleton(abstract string, concrete interface{}) error

	// SingletonIf 如果尚未注册，则注册一个共享绑定
	SingletonIf(abstract string, concrete interface{}) error

	// Extend 扩展容器中的抽象类型
	Extend(abstract string, closure func(interface{}) (interface{}, error)) error

	// Instance 注册一个现有的实例作为共享实例到容器
	Instance(abstract string, instance interface{}) error

	// AddContextualBinding 添加一个上下文绑定到容器
	AddContextualBinding(concrete string, abstract string, implementation interface{}) error

	// When 定义一个上下文绑定
	When(concrete interface{}) ContextualBindingBuilder

	// Factory 获取一个闭包来从容器中解析给定类型
	Factory(abstract string) func() (interface{}, error)

	// Flush 清空容器中的所有绑定和已解析的实例
	Flush()

	// Make 从容器中解析给定类型
	Make(abstract string, parameters ...interface{}) (interface{}, error)

	// Call 调用给定的闭包/类方法并注入其依赖项
	Call(callback interface{}, parameters ...interface{}) (interface{}, error)

	// Resolved 确定给定的抽象类型是否已被解析
	Resolved(abstract string) bool

	// Resolving 注册一个新的解析回调
	Resolving(abstract interface{}, callback func(interface{})) error

	// AfterResolving 注册一个新的解析后回调
	AfterResolving(abstract interface{}, callback func(interface{})) error
}

// ContextualBindingBuilder 是用于构建上下文绑定的接口
type ContextualBindingBuilder interface {
	// Needs 指定需要的抽象类型
	Needs(abstract string) ContextualBindingBuilder

	// Give 提供实现
	Give(implementation interface{}) error
}

// 默认错误
var (
	ErrAlreadyBound          = errors.New("already bound")
	ErrNotBound              = errors.New("not bound")
	ErrAlreadyResolved       = errors.New("already resolved")
	ErrNotResolved           = errors.New("not resolved")
	ErrInvalidImplementation = errors.New("invalid implementation")
)
