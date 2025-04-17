package controlloop

import (
	"context"
	"time"
)

type Result struct {
	RequeueAfter time.Duration
	Requeue      bool
}

type ResourceObject[T any] interface {
	GetConditions() []Condition
	GetCondition(name string) (Condition, bool)
	GetGeneration() int64
	IncGeneration()
	SetKillTimestamp(time time.Time)
	GetKillTimestamp() string
	SetDeletionTimestamp(time.Time)
	GetDeletionTimestamp() string
	GetName() ObjectKey
	DeepCopy() T
}

type Reconcile[T ResourceObject[T]] interface {
	Reconcile(context.Context, T) (Result, error)
}
