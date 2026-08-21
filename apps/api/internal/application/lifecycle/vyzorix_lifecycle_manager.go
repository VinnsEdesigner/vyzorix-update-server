// Package lifecycle wires the worker registry for the API server.
package lifecycle

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/lifecycle"
)

// Manager registers and coordinates background workers.
type Manager struct {
	registry *lifecycle.Registry
}

// NewManager creates a Manager.
func NewManager() *Manager {
	return &Manager{registry: lifecycle.NewRegistry()}
}

// Registry exposes the underlying service registry for handler registration.
func (m *Manager) Registry() *lifecycle.Registry {
	return m.registry
}

// Register adds a service with a name and its dependencies.
func (m *Manager) Register(name string, svc lifecycle.BackgroundService, dependsOn ...string) {
	m.registry.Register(name, svc, dependsOn...)
}

// StartAll starts all services in dependency order.
func (m *Manager) StartAll() error {
	return m.registry.StartAll()
}

// StopAll stops all services in reverse order.
func (m *Manager) StopAll() {
	m.registry.StopAll()
}

// Health returns per-service health status.
func (m *Manager) Health() map[string]bool {
	return m.registry.Health()
}
