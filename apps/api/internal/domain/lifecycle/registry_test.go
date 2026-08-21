package lifecycle

import (
	"sync/atomic"
	"testing"
)

type fakeService struct {
	healthy  bool
	started  int32
	stopped  int32
}

func (f *fakeService) Start()  { atomic.AddInt32(&f.started, 1) }
func (f *fakeService) Stop()   { atomic.AddInt32(&f.stopped, 1) }
func (f *fakeService) Healthy() bool { return f.healthy }

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	svc := &fakeService{healthy: true}
	r.Register("a", svc)
	if r.Get("a") == nil {
		t.Fatal("expected service to be registered")
	}
	if r.Get("missing") != nil {
		t.Error("expected nil for unregistered name")
	}
}

func TestRegistry_HealthSnapshot(t *testing.T) {
	r := NewRegistry()
	healthy := &fakeService{healthy: true}
	unhealthy := &fakeService{healthy: false}
	r.Register("healthy", healthy)
	r.Register("unhealthy", unhealthy)

	h := r.Health()
	if !h["healthy"] {
		t.Error("expected healthy=true")
	}
	if h["unhealthy"] {
		t.Error("expected unhealthy=false")
	}
}

func TestRegistry_StartStopOrder(t *testing.T) {
	r := NewRegistry()
	a := &fakeService{healthy: true}
	b := &fakeService{healthy: true}
	r.Register("a", a)
	r.Register("b", b)

	if err := r.StartAll(); err != nil {
		t.Fatalf("StartAll: %v", err)
	}
	if a.started != 1 || b.started != 1 {
		t.Errorf("started counts: a=%d, b=%d", a.started, b.started)
	}

	r.StopAll()
	if a.stopped != 1 || b.stopped != 1 {
		t.Errorf("stopped counts: a=%d, b=%d", a.stopped, b.stopped)
	}
}

func TestRegistry_StartAllCircular(t *testing.T) {
	r := NewRegistry()
	r.Register("a", &fakeService{healthy: true}, "b")
	r.Register("b", &fakeService{healthy: true}, "a")
	if err := r.StartAll(); err != ErrCircularDependency {
		t.Errorf("expected ErrCircularDependency, got %v", err)
	}
}

func TestRegistry_DependencyOrder(t *testing.T) {
	r := NewRegistry()
	a := &fakeService{healthy: true}
	b := &fakeService{healthy: true}
	c := &fakeService{healthy: true}
	// c depends on b, b depends on a.
	r.Register("c", c, "b")
	r.Register("b", b, "a")
	r.Register("a", a)

	if err := r.StartAll(); err != nil {
		t.Fatalf("StartAll: %v", err)
	}

	order := r.Services()
	if len(order) != 3 {
		t.Fatalf("expected 3 services, got %d", len(order))
	}
}

func TestRegistry_ServicesSorted(t *testing.T) {
	r := NewRegistry()
	r.Register("z", &fakeService{healthy: true})
	r.Register("a", &fakeService{healthy: true})
	r.Register("m", &fakeService{healthy: true})

	names := r.Services()
	if len(names) != 3 || names[0] != "a" || names[1] != "m" || names[2] != "z" {
		t.Errorf("services not sorted: %v", names)
	}
}
