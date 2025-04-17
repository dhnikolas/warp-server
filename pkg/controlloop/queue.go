package controlloop

import (
	"k8s.io/client-go/util/workqueue"
	"sync"
	"time"
)

type Queue[T ResourceObject[T]] struct {
	queue        workqueue.TypedRateLimitingInterface[ObjectKey]
	existedItems map[ObjectKey]ResourceObject[T]
	m            *sync.RWMutex
}

func NewQueue[T ResourceObject[T]]() *Queue[T] {
	rateLimitingConfig := workqueue.TypedRateLimitingQueueConfig[ObjectKey]{}
	rateLimitingConfig.DelayingQueue = workqueue.NewTypedDelayingQueue[ObjectKey]()
	queue := workqueue.NewTypedRateLimitingQueueWithConfig[ObjectKey](workqueue.NewTypedMaxOfRateLimiter[ObjectKey](), rateLimitingConfig)
	return &Queue[T]{queue: queue, existedItems: make(map[ObjectKey]ResourceObject[T]), m: &sync.RWMutex{}}
}

func (q *Queue[T]) getExistedItems() map[ObjectKey]ResourceObject[T] {
	q.m.RLock()
	defer q.m.RUnlock()
	return q.existedItems
}

func (q *Queue[T]) len() int {
	q.m.RLock()
	defer q.m.RUnlock()
	return len(q.existedItems)
}

func (q *Queue[T]) add(item ResourceObject[T]) {
	q.m.Lock()
	defer q.m.Unlock()
	q.existedItems[item.GetName()] = item
	q.queue.Add(item.GetName())
}

func (q *Queue[T]) finalize(item ResourceObject[T]) {
	q.m.Lock()
	defer q.m.Unlock()
	delete(q.existedItems, item.GetName())
}

func (q *Queue[T]) done(item ResourceObject[T]) {
	q.queue.Done(item.GetName())
}

func (q *Queue[T]) addAfter(item ResourceObject[T], duration time.Duration) {
	q.queue.AddAfter(item.GetName(), duration)
}

func (q *Queue[T]) addRateLimited(item ResourceObject[T]) {
	q.queue.AddRateLimited(item.GetName())
}

func (q *Queue[t]) get() (ObjectKey, bool) {
	name, shutdown := q.queue.Get()
	if shutdown {
		return "", true
	}
	return name, false
}
