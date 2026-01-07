package validator

import (
	"reflect"
	"strings"

	"bit-labs.cn/owl/contract/foundation"
	"github.com/go-playground/validator/v10"
)

var _ foundation.ServiceProvider = (*ValidatorServiceProvider)(nil)

type ValidatorServiceProvider struct {
	app foundation.Application
}

func (i *ValidatorServiceProvider) Register() {
	i.app.Register(func() *validator.Validate {
		v := validator.New(validator.WithRequiredStructEnabled())
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("json")
			if name == "" {
				return fld.Name
			}
			name = strings.Split(name, ",")[0]
			if name == "" || name == "-" {
				return fld.Name
			}
			return name
		})
		return v
	})
}

func (i *ValidatorServiceProvider) Boot() {}

func (i *ValidatorServiceProvider) GenerateConf() map[string]string { return nil }
