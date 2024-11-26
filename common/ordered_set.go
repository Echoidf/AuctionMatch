package common

import (
	"sort"
	"sync"
)

// OrderedSet 是一个线程安全的有序集合
type OrderedSet struct {
	items   map[int64]struct{}
	sorted  []int64
	isDirty bool
	mu      sync.RWMutex
}

// NewOrderedSet 创建新的OrderedSet
func NewOrderedSet() *OrderedSet {
	return &OrderedSet{
		items: make(map[int64]struct{}),
	}
}

// Add 添加元素
func (s *OrderedSet) Add(item int64) {
	s.mu.Lock()
	if _, exists := s.items[item]; !exists {
		s.items[item] = struct{}{}
		s.isDirty = true
	}
	s.mu.Unlock()
}

func (s *OrderedSet) Len() int {
	return len(s.items)
}

// GetSorted 获取排序后的元素列表
func (s *OrderedSet) GetSorted(desc bool) []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isDirty {
		s.sorted = make([]int64, 0, len(s.items))
		for item := range s.items {
			s.sorted = append(s.sorted, item)
		}
		sort.Slice(s.sorted, func(i, j int) bool {
			return (s.sorted[i] < s.sorted[j]) != desc
		})
		s.isDirty = false
	}

	// 返回副本以保证线程安全
	result := make([]int64, len(s.sorted))
	copy(result, s.sorted)
	return result
}
