package pay

import (
	"sync"
	"time"
)

type DedupStore interface {
	IsDuplicate(key string) bool
	MarkKey(key string)
}

type memoryStore struct {
	m   sync.Map
	ttl time.Duration
}

func NewMemoryDedupStore(ttl time.Duration) DedupStore {
	return &memoryStore{ttl: ttl}
}

func (s *memoryStore) IsDuplicate(key string) bool {
	if key == "" {
		return false
	}
	if v, ok := s.m.Load(key); ok {
		if t, ok2 := v.(time.Time); ok2 {
			if time.Since(t) < s.ttl {
				return true
			}
			s.m.Delete(key)
		}
	}
	return false
}

func (s *memoryStore) MarkKey(key string) {
	if key == "" {
		return
	}
	s.m.Store(key, time.Now())
}

var dstore DedupStore = NewMemoryDedupStore(24 * time.Hour)

func SetDedupStore(store DedupStore) {
	if store != nil {
		dstore = store
	}
}

func isDuplicate(key string) bool {
	if dstore == nil {
		return false
	}
	return dstore.IsDuplicate(key)
}

func markKey(key string) {
	if dstore == nil {
		return
	}
	dstore.MarkKey(key)
}
