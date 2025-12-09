package db

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	CreatedAt *time.Time      `json:"createdAt,omitempty"`
	UpdatedAt *time.Time      `json:"updatedAt,omitempty"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`       // 删除时间
	CreatorID string          `gorm:"comment:创建人" json:"creatorID,omitempty"` // 创建人
	UpdaterID string          `gorm:"comment:修改人" json:"updaterID,omitempty"` // 修改人
}

func (i *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	//idGenerator, err := util.NewSnowflake(1, 3)
	//if err != nil {
	//	return err
	//}
	//i.ID = cast.ToUint(idGenerator.NextVal())
	//i.ID = uint(rand.Int())
	return
}
