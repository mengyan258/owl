package db

import (
	"time"

	"bit-labs.cn/owl/utils"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint            `gorm:"primarykey" json:"id,string"`
	CreatedAt *time.Time      `json:"createdAt,omitempty"`
	UpdatedAt *time.Time      `json:"updatedAt,omitempty"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`       // 删除时间
	CreatorID string          `gorm:"comment:创建人" json:"creatorID,omitempty"` // 创建人
	UpdaterID string          `gorm:"comment:修改人" json:"updaterID,omitempty"` // 修改人
}

func (i *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	idGenerator := utils.NewWorker(1, 3)
	id, err := idGenerator.NextID()
	if err != nil {
		return err
	}
	// 如果手动设置了 ID，则不进行设置
	if i.ID == 0 {
		i.ID = cast.ToUint(id)
	}
	return
}
