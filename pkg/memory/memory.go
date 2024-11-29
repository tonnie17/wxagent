package memory

import (
	"github.com/tonnie17/wxagent/pkg/llm"
	"sync"
	"sync/atomic"
)

type Memory interface {
	Update(messages []*llm.ChatMessage) error
	History() ([]*llm.ChatMessage, error)
}

type Lock interface {
	Lock() bool
	Release()
	IsLocked() bool
}

type BaseLock struct {
	isLock int32
	lock   sync.Mutex
}

func (m *BaseLock) Lock() bool {
	lockSuccess := m.lock.TryLock()
	if lockSuccess {
		atomic.StoreInt32(&m.isLock, 1)
	}
	return lockSuccess
}

func (m *BaseLock) IsLocked() bool {
	return atomic.LoadInt32(&m.isLock) == 1
}

func (m *BaseLock) Release() {
	atomic.StoreInt32(&m.isLock, 0)
	m.lock.Unlock()
}
