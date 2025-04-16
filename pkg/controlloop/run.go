package controlloop

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/client-go/util/workqueue"
	"sync"
	"sync/atomic"
	"time"
)

const defaultReconcileTime = time.Second * 30
const errorReconcileTime = time.Second * 5

type ControlLoop struct {
	r           Reconcile
	stopChannel chan struct{}
	exitChannel chan struct{}
	l           Logger
	concurrency int
	Storage     Storage
	Queue       *Queue[ResourceObject]
}

func New(r Reconcile, options ...ClOption) *ControlLoop {
	currentOptions := &opts{}
	for _, o := range options {
		o(currentOptions)
	}
	typedRateLimitingQueueConfig := workqueue.TypedRateLimitingQueueConfig[ResourceObject]{}
	typedRateLimitingQueueConfig.DelayingQueue = workqueue.NewTypedDelayingQueue[ResourceObject]()
	queue := NewQueue()
	controlLoop := &ControlLoop{
		r:           r,
		stopChannel: make(chan struct{}),
		exitChannel: make(chan struct{}),
		Storage:     NewMemoryStorage(queue),
		Queue:       queue,
	}

	if currentOptions.logger != nil {
		controlLoop.l = currentOptions.logger
	} else {
		controlLoop.l = &logger{}
	}

	if currentOptions.concurrency > 0 {
		controlLoop.concurrency = currentOptions.concurrency
	} else {
		controlLoop.concurrency = 1
	}
	return controlLoop
}

func (cl *ControlLoop) Run() {
	stopping := atomic.Bool{}
	stopping.Store(false)

	go func() {
		<-cl.stopChannel
		delayQueueLen := cl.Queue.len()
		if delayQueueLen > 0 {
			stopping.Store(true)
			for object, _ := range cl.Queue.getExistedItems() {
				cl.Queue.queue.Add(object)
			}
		} else {
			cl.Queue.queue.ShutDownWithDrain()
		}
	}()

	f := func(wg *sync.WaitGroup) {
		defer wg.Done()

		r := cl.r
		ctx := context.Background()
		for {

			object, exit, err := cl.Storage.getLast()
			if exit {
				return
			}
			if err != nil {
				// object already deleted
				if errors.Is(err, KetNotExist) {
					continue
				}
			}

			if stopping.Load() && object.GetKillTimestamp() == "" {
				object.SetKillTimestamp(time.Now())
				err := cl.Storage.Update(object)
				if errors.Is(err, AlreadyUpdated) {
					cl.Queue.queue.Add(object.GetName())
				}
				continue
			}

			result, err := cl.reconcile(ctx, r, object)
			switch {
			case err != nil:
				cl.l.Error(err)
				cl.Queue.addRateLimited(object)
			case result.RequeueAfter > 0:
				cl.Queue.addAfter(object, result.RequeueAfter)
			case result.Requeue:
				cl.Queue.add(object)
			default:
				cl.Queue.finalize(object)
			}

			cl.Queue.done(object)
			if stopping.Load() && cl.Queue.len() == 0 {
				cl.Queue.queue.ShutDownWithDrain()
			}
		}
	}

	wg := &sync.WaitGroup{}

	wg.Add(cl.concurrency)
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		cl.exitChannel <- struct{}{}
	}(wg)

	for range cl.concurrency {
		go f(wg)
	}
}

func (cl *ControlLoop) Stop() {
	cl.stopChannel <- struct{}{}
	<-cl.exitChannel
}

func (cl *ControlLoop) reconcile(ctx context.Context, r Reconcile, object ResourceObject) (Result, error) {
	defer func() {
		if r := recover(); r != nil {
			cl.l.Error(fmt.Errorf("Recovered from panic: %v ", r))
		}
	}()
	return r.Reconcile(ctx, object)
}
