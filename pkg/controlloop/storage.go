package controlloop

import (
	"sync"
)

type Storage[T ResourceObject[T]] interface {
	Add(item T)
	Get(item T) T
	Update(item T) error
	Delete(item T)
	getLast() (T, bool, error)
}

func NewMemoryStorage[T ResourceObject[T]](q *Queue[T]) *MemoryStorage[T] {
	return &MemoryStorage[T]{
		m:       &sync.RWMutex{},
		Queue:   q,
		objects: make(map[ObjectKey]T),
	}
}

type MemoryStorage[T ResourceObject[T]] struct {
	m       *sync.RWMutex
	Queue   *Queue[T]
	objects map[ObjectKey]T
}

func (s *MemoryStorage[T]) Add(item T) {
	s.m.Lock()
	defer s.m.Unlock()
	s.objects[item.GetName()] = item
	s.Queue.add(item)
}

func (s *MemoryStorage[T]) Get(item T) T {
	s.m.RLock()
	defer s.m.RUnlock()
	var zero T
	val, exist := s.objects[item.GetName()]
	if !exist {
		return zero
	}
	return val.DeepCopy()
}

func (s *MemoryStorage[T]) Update(item T) error {
	s.m.Lock()
	defer s.m.Unlock()
	curr, exist := s.objects[item.GetName()]
	if !exist {
		return KetNotExist
	}
	if curr.GetGeneration() > item.GetGeneration() {
		return AlreadyUpdated
	}
	curr.IncGeneration()
	s.objects[item.GetName()] = item
	s.Queue.add(item)
	s.Queue.done(item)
	return nil
}

func (s *MemoryStorage[T]) Delete(item T) {
	s.m.Lock()
	defer s.m.Unlock()
	s.Queue.finalize(item)
	delete(s.objects, item.GetName())
}

func (s *MemoryStorage[T]) getLast() (T, bool, error) {
	s.m.Lock()
	defer s.m.Unlock()
	var zero T
	name, shutdown := s.Queue.get()
	if shutdown {
		return zero, true, nil
	}
	// object already deleted
	if _, exist := s.objects[name]; !exist {
		return zero, false, KetNotExist
	}
	return s.objects[name].DeepCopy(), false, nil
}
