package web

import (
	"github.com/tonni17/wxagent/pkg/memory"
	"sync"
	"time"
)

type UserMemory struct {
	mem        memory.Memory
	lastAccess time.Time
}

type UserMemoryStore struct {
	memTTL     time.Duration
	once       sync.Once
	userMemory sync.Map
}

func NewUserMemoryStore(memTTL time.Duration) *UserMemoryStore {
	return &UserMemoryStore{
		memTTL: memTTL,
	}
}

func (s *UserMemoryStore) GetOrNew(userID string, memFactory func() memory.Memory) memory.Memory {
	v, ok := s.userMemory.Load(userID)
	if !ok {
		mem := memFactory()
		userMem := &UserMemory{
			mem:        mem,
			lastAccess: time.Now(),
		}
		s.userMemory.Store(userID, userMem)
		return mem
	}
	userMem := v.(*UserMemory)
	userMem.lastAccess = time.Now()
	return userMem.mem
}

func (s *UserMemoryStore) CheckAndClear(interval time.Duration) {
	worker := func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			s.userMemory.Range(func(key, value any) bool {
				userMem := value.(*UserMemory)
				if time.Now().Sub(userMem.lastAccess) > s.memTTL {
					s.userMemory.Delete(key)
				}
				return true
			})
		}
	}
	s.once.Do(func() {
		go worker()
	})
}
