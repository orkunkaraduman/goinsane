package goinsane

import (
	"sync"
)

type NamedLock struct {
	mutex   sync.Mutex
	holders map[string]*namedLockHolder
}

type NamedLocker interface {
	Name() string
	Lock()
	TryLock() bool
	Unlock()
	RLock()
	TryRLock() bool
	RUnlock()
}

func NewNamedLock() *NamedLock {
	return &NamedLock{
		holders: make(map[string]*namedLockHolder),
	}
}

func (l *NamedLock) Locker(name string) NamedLocker {
	return &namedLocker{
		owner: l,
		name:  name,
	}
}

func (l *NamedLock) holder(name string) *namedLockHolder {
	l.mutex.Lock()
	h := l.holders[name]
	l.mutex.Unlock()
	return h
}

func (l *NamedLock) increase(name string) *namedLockHolder {
	l.mutex.Lock()
	h := l.holders[name]
	if h == nil {
		h = &namedLockHolder{
			owner: l,
			name:  name,
		}
		l.holders[name] = h
	}
	h.count++
	l.mutex.Unlock()
	return h
}

type namedLockHolder struct {
	owner *NamedLock
	name  string
	count int
	mutex sync.RWMutex
}

func (h *namedLockHolder) decrease() {
	h.owner.mutex.Lock()
	h.count--
	if h.count <= 0 {
		delete(h.owner.holders, h.name)
	}
	h.owner.mutex.Unlock()
}

type namedLocker struct {
	owner *NamedLock
	name  string
}

func (l *namedLocker) Name() string {
	return l.name
}

func (l *namedLocker) Lock() {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// lock
	h.mutex.Lock()
}

func (l *namedLocker) TryLock() bool {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// try  lock
	ok := h.mutex.TryLock()
	if !ok {
		// decrease count, delete holder if needed
		h.decrease()
	}
	return ok
}

func (l *namedLocker) Unlock() {
	// get holder
	h := l.owner.holder(l.name)
	if h == nil {
		panic("unlock of unlocked mutex")
	}
	// unlock
	h.mutex.Unlock()
	// decrease count, delete holder
	h.decrease()
}

func (l *namedLocker) RLock() {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// rlock
	h.mutex.RLock()
}

func (l *namedLocker) TryRLock() bool {
	// increase count, get holder
	h := l.owner.increase(l.name)
	// try  rlock
	ok := h.mutex.TryRLock()
	if !ok {
		// decrease count, delete holder if needed
		h.decrease()
	}
	return ok
}

func (l *namedLocker) RUnlock() {
	// get holder
	h := l.owner.holder(l.name)
	if h == nil {
		panic("runlock of unlocked mutex")
	}
	// runlock
	h.mutex.RUnlock()
	// decrease count, delete holder
	h.decrease()
}
