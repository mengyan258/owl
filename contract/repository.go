package contract

import "gorm.io/gorm"

type Repository[T any] interface {
	Create(data *T) error
	Update(data *T) error
	Delete(id uint) error
	Detail(id any) (*T, error)
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []T, err error)
}
