package conf

// Repository 接口转换成 Go 语言中的结构体和方法实现
type Repository struct {
	// 假设这里有一个字段存放配置数据，例如 map[string]interface{}
	config map[string]interface{}
}

// NewRepository 创建一个新的 Repository 实例
func NewRepository() *Repository {
	return &Repository{config: make(map[string]interface{})}
}

// Has 方法检查给定的配置键是否存在
func (r *Repository) Has(key string) bool {
	_, exists := r.config[key]
	return exists
}

// Get 方法获取指定的配置值，如果不存在则返回默认值
func (r *Repository) Get(key interface{}, defaultVal interface{}) interface{} {
	var value interface{}

	switch k := key.(type) {
	case string:
		value, _ = r.config[k]
	case []string:
		value = r.config[k[0]]
		for i := 1; i < len(k); i++ {
			if subMap, ok := value.(map[string]interface{}); ok {
				value = subMap[k[i]]
			} else {
				break
			}
		}
	}

	if value == nil {
		return defaultVal
	}
	return value
}

// All 返回所有配置项
func (r *Repository) All() map[string]interface{} {
	return r.config
}

// Set 设置给定的配置值
func (r *Repository) Set(key interface{}, value interface{}) {
	switch k := key.(type) {
	case string:
		r.config[k] = value
	case []string:
		currentLevel := r.config
		for i, part := range k {
			if i == len(k)-1 {
				currentLevel[part] = value
				break
			}
			if nextLevel, ok := currentLevel[part].(map[string]interface{}); ok {
				currentLevel = nextLevel
			} else {
				currentLevel[part] = make(map[string]interface{})
				currentLevel = currentLevel[part].(map[string]interface{})
			}
		}
	}
}

// Prepend 在数组配置值前插入一个值
func (r *Repository) Prepend(key string, value interface{}) {
	if existingValue, exists := r.config[key]; exists {
		switch v := existingValue.(type) {
		case []interface{}:
			r.config[key] = append([]interface{}{value}, v...)
		}
	} else {
		r.config[key] = []interface{}{value}
	}
}

// Push 在数组配置值后追加一个值
func (r *Repository) Push(key string, value interface{}) {
	if existingValue, exists := r.config[key]; exists {
		switch v := existingValue.(type) {
		case []interface{}:
			r.config[key] = append(v, value)
		}
	} else {
		r.config[key] = []interface{}{value}
	}
}
