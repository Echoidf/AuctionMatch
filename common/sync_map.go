package common

import "sync"

type SMap[K comparable, V any] struct {
	sync.RWMutex
	Map map[K]V
}

func (l *SMap[K, V]) ReadMap(key K) (V, bool) {
	l.RLock()
	value, ok := l.Map[key]
	l.RUnlock()
	return value, ok
}

func (l *SMap[K, V]) WriteMap(key K, value V) {
	l.Lock()
	l.Map[key] = value
	l.Unlock()
}

func NewSMap[K comparable, V any]() *SMap[K, V] {
	return &SMap[K, V]{
		Map: make(map[K]V),
	}
}
