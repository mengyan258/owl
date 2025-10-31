package contract

import "gorm.io/gorm"

type Repository[T any] interface {
	Create(data *T) error
	Update(data *T) error
	Delete(id uint) error
	Detail(id any) (*T, error)
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []T, err error)
}

type PageReq struct {
	Page     int
	PageSize int
}

type PageResp struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}
