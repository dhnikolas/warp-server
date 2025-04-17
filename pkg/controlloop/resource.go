package controlloop

import (
	"sync/atomic"
	"time"
)

type ObjectKey = string

type Resource struct {
	Conditions        []Condition
	KillTimestamp     string
	DeletionTimestamp string
	Generation        int64
	Name              ObjectKey
}

type Condition struct {
	Type    string
	Status  string
	Reason  string
	Message string
}

func NewResource(name ObjectKey) *Resource {
	return &Resource{Name: name}
}

func (r *Resource) SetKillTimestamp(time time.Time) {
	if r.KillTimestamp == "" {
		r.KillTimestamp = time.Format("2006-01-02 15:04:05")
	}
}

func (r *Resource) GetKillTimestamp() string {
	return r.KillTimestamp
}

func (r *Resource) SetDeletionTimestamp(time time.Time) {
	if r.DeletionTimestamp == "" {
		r.DeletionTimestamp = time.Format("2006-01-02 15:04:05")
	}
}

func (r *Resource) GetDeletionTimestamp() string {
	return r.DeletionTimestamp
}

func (r *Resource) IncGeneration() {
	atomic.AddInt64(&r.Generation, 1)
}

func (r *Resource) GetGeneration() int64 {
	return atomic.LoadInt64(&r.Generation)
}

func (r *Resource) setCondition(name, status, reason, message string) {
	exist := func(slice []Condition, name string) bool {
		for _, item := range slice {
			if item.Type == name {
				return true
			}
		}
		return false
	}
	if exist(r.Conditions, name) {
		for i := range r.Conditions {
			if r.Conditions[i].Type == name {
				r.Conditions[i].Reason = reason
				r.Conditions[i].Message = message
				r.Conditions[i].Status = status
			}
		}
		return
	}
	r.Conditions = append(r.Conditions, Condition{Type: name, Status: status, Reason: reason, Message: message})
}

func (r *Resource) GetConditions() []Condition {
	return r.Conditions
}

func (r *Resource) GetName() ObjectKey {
	return r.Name
}

func (r *Resource) GetCondition(name string) (Condition, bool) {
	for _, c := range r.GetConditions() {
		if c.Type == name {
			return c, true
		}
	}
	return Condition{}, false
}

func (r *Resource) MarkFalse(name, reason, message string) {
	r.setCondition(name, "False", reason, message)
}

func (r *Resource) MarkTrue(name string) {
	r.setCondition(name, "True", "", "")
}
