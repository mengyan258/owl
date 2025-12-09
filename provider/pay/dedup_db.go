package pay

import (
	"time"

	"gorm.io/gorm"
)

type dbStore struct {
	db  *gorm.DB
	ttl time.Duration
}

type DedupEntry struct {
	Key       string `gorm:"primaryKey;size:128"`
	CreatedAt time.Time
}

func InitDBDedupStore(db *gorm.DB, ttl time.Duration) {
	if db == nil {
		return
	}
	_ = db.AutoMigrate(&DedupEntry{})
	SetDedupStore(&dbStore{db: db, ttl: ttl})
}

func (s *dbStore) IsDuplicate(key string) bool {
	if key == "" {
		return false
	}
	var e DedupEntry
	if err := s.db.First(&e, "key = ?", key).Error; err == nil {
		if time.Since(e.CreatedAt) < s.ttl {
			return true
		}
		s.db.Delete(&DedupEntry{Key: key})
	}
	return false
}

func (s *dbStore) MarkKey(key string) {
	if key == "" {
		return
	}
	_ = s.db.Save(&DedupEntry{Key: key, CreatedAt: time.Now()}).Error
}
