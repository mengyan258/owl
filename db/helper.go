package db

import (
	"bit-labs.cn/owl/utils"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

// GetMenuModelsByIDs 获取菜单模型，gorm 多对多关联写数据

func GetModelsByIDs[T any](modelIDs []string) []T {
	var models []T
	for _, id := range modelIDs {
		m := *new(T)
		value := reflect.ValueOf(&m).Elem()
		fieldType := value.FieldByName("ID").Type()
		switch fieldType.Kind() {
		case reflect.Uint:
			value.FieldByName("ID").SetUint(cast.ToUint64(id))
		case reflect.String:
			value.FieldByName("ID").SetString(id)
		default:
			panic("unhandled default case")
		}

		models = append(models, m)
	}
	return models
}

// AppendWhereFromStruct 根据结构体中的字段名，生成 where 条件
// NameLike = name like
// AgeGt = where age >
// AgeLt = where age <
func AppendWhereFromStruct(db *gorm.DB, s any) {
	result := make(map[string]interface{})
	v := reflect.ValueOf(s)
	if v.Equal(reflect.ValueOf(nil)) {
		return
	}

	// 兼容指针结构体和结构体
	if v.Kind() == reflect.Struct || v.Kind() == reflect.Ptr {
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
	}

	getNonZeroFields(v, result)

	// 字段名可能叫 NameLike, AgeGt ...
	for field, value := range result {
		realField, operator := splitStringAtLastUnderscore(field)
		switch operator {
		case "gt":
			db.Where(fmt.Sprintf("%s > ?", realField), value)
		case "gte":
			db.Where(fmt.Sprintf("%s >= ?", realField), value)
		case "lt":
			db.Where(fmt.Sprintf("%s < ?", realField), value)
		case "lte":
			db.Where(fmt.Sprintf("%s < ?", realField), value)
		case "in":
			str, ok := value.(string)

			if ok {
				var ids []string
				// 处理字符串 "["1", "2", "3"]"
				if strings.Contains(str, "[") {
					_ = jsoniter.Unmarshal([]byte(str), &ids)
				} else {
					// 处理字符串 "1", "2", "3"
					ids = strings.Split(str, ",")
				}

				db.Where(fmt.Sprintf("%s in ?", realField), ids)
			} else {
				db.Where(fmt.Sprintf("%s in ?", realField), value)
			}

		case "notin":
			db.Where(fmt.Sprintf("%s not in ?", realField), value)
		case "like":
			db.Where(fmt.Sprintf("%s like ?", realField), "%"+cast.ToString(value)+"%")
		case "between":
			valueArr := strings.Split(cast.ToString(value), ",")
			if len(valueArr) == 2 {
				db.Where(fmt.Sprintf("%s >= ? and %s <= ?", realField, realField), valueArr[0], valueArr[1])
			} else {
				//logx.Error("between operator must be two value")
			}

		default:
			db.Where(fmt.Sprintf("%s = ?", field), value)
		}
	}
}

func splitStringAtLastUnderscore(s string) (string, string) {
	// 找到最后一个 '_' 的位置
	lastIndex := strings.LastIndex(s, "_")

	// 判断是否有 '_' 存在
	if lastIndex == -1 {
		return s, ""
	}

	// 分割字符串
	prefix := s[:lastIndex]
	suffix := s[lastIndex+1:]

	return prefix, suffix
}

// 定义一个函数来递归获取结构体的所有字段，并排除零值的字段
func getNonZeroFields(v reflect.Value, result map[string]interface{}) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 过滤掉特定的字段
		if field.Name == "PageReq" {
			continue
		}

		// 检查字段是否为零值
		isZero := fieldValue.Interface() == reflect.Zero(fieldValue.Type()).Interface()
		if isZero {
			continue
		}

		// 如果字段是结构体类型，则递归调用
		if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Ptr {
			if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
				fieldValue = fieldValue.Elem()
			}
			getNonZeroFields(fieldValue, result)
		} else {
			// 将非零值字段添加到结果中
			result[utils.Cc2Udl(field.Name)] = fieldValue.Interface()
		}
	}
}
