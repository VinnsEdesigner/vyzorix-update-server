// Package lifecycle coordinates background services with dependency-ordered
// start/stop and per-service health for the /healthz endpoint.
package lifecycle

import (
	"errors"
	"sort"
	"sync"
)

// BackgroundService is one named long-running service.
type BackgroundService interface {
	// Start launches the service (usually a goroutine).
	Start()
	// Stop shuts the service down gracefully.
	Stop()
	// Healthy reports whether the service is running correctly.
	Healthy() bool
}

type namedService struct {
	svc       BackgroundService
	dependsOn []string
}

// Registry holds ordered services and their dependencies.
type Registry struct {
	services map[string]*namedService
	ordered  []string
	mu       sync.RWMutex
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{services: make(map[string]*namedService)}
}

// ErrCircularDependency is returned when a cycle is introduced in dependsOn.
var ErrCircularDependency = errors.New("circular dependency in service graph")

// Get returns the service for a name, or nil.
func (r *Registry) Get(name string) BackgroundService {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if svc, ok := r.services[name]; ok {
		return svc.svc
	}
	return nil
}

// Register adds a service with its dependencies. Duplicates are replaced.
func (r *Registry) Register(name string, svc BackgroundService, dependsOn ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[name] = &namedService{dependsOn: dependsOn, svc: svc}
	r.ordered = nil
}

// computeOrder returns services in dependency order, or lexicographic for
// independent ones. Errors on circular dependencies.
func (r *Registry) computeOrder() ([]string, error) {
	order := make([]string, 0, len(r.services))
	visited := make(map[string]bool, len(r.services))
	visiting := make(map[string]bool, len(r.services))
	var add func(name string) error
	add = func(name string) error {
		if visiting[name] {
			return ErrCircularDependency
		}
		if visited[name] {
			return nil
		}
		visiting[name] = true
		for _, dep := range r.services[name].dependsOn {
			if err := add(dep); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}
	for name := range r.services {
		if !visited[name] {
			if err := add(name); err != nil {
				return nil, err
			}
		}
	}
	return order, nil
}

// StartAll starts every service in dependency order. Errors propagate on
// circular dependencies.
func (r *Registry) StartAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	order, err := r.computeOrder()
	if err != nil {
		return err
	}
	for _, name := range order {
		r.services[name].svc.Start()
	}
	r.ordered = order
	return nil
}

// StopAll stops every service in reverse start order.
func (r *Registry) StopAll() {
	r.mu.RLock()
	order := append([]string(nil), r.ordered...)
	r.mu.RUnlock()
	for i := len(order) - 1; i >= 0; i-- {
		if svc, ok := r.services[order[i]]; ok {
			svc.svc.Stop()
		}
	}
}

// Health returns a snapshot of service health by name.
func (r *Registry) Health() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snapshot := make(map[string]bool, len(r.services))
	for name, svc := range r.services {
		snapshot[name] = svc.svc.Healthy()
	}
	return snapshot
}

// Services returns registered service names sorted for stable output.
func (r *Registry) Services() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
