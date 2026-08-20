package shared

import "sync"

type LifecycleEvent struct {
	Type         string
	OrgID        string
	OperatorID   string
	InvitationID string
	Role         string
}

type LifecycleBus struct {
	subs []func(LifecycleEvent)
	mu   sync.RWMutex
}

func NewLifecycleBus() *LifecycleBus {
	return &LifecycleBus{}
}

func (b *LifecycleBus) Subscribe(fn func(LifecycleEvent)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, fn)
}

func (b *LifecycleBus) Publish(ev LifecycleEvent) {
	b.mu.RLock()
	subs := b.subs
	b.mu.RUnlock()
	for _, fn := range subs {
		fn(ev)
	}
}

type LifecyclePublisher interface {
	Publish(ev LifecycleEvent)
}
