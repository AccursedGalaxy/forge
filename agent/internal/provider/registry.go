package provider

import (
	"fmt"
	"sync"
)

// Registry is a thread-safe store of named AgentProvider implementations.
// Providers are registered at startup; the default provider is used when
// no specific provider is requested.
type Registry struct {
	mu          sync.RWMutex
	providers   map[string]AgentProvider
	defaultName string
}

// NewRegistry returns an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]AgentProvider),
	}
}

// Register adds a provider under the given name.
// The first registered provider becomes the default automatically.
func (r *Registry) Register(name string, p AgentProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = p
	if r.defaultName == "" {
		r.defaultName = name
	}
}

// SetDefault changes which registered provider is used by Default().
func (r *Registry) SetDefault(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[name]; !ok {
		return fmt.Errorf("provider %q is not registered", name)
	}
	r.defaultName = name
	return nil
}

// Get returns the provider registered under name, or an error if absent.
func (r *Registry) Get(name string) (AgentProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q is not registered", name)
	}
	return p, nil
}

// Default returns the default provider.
func (r *Registry) Default() (AgentProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.defaultName == "" {
		return nil, fmt.Errorf("no providers registered")
	}
	return r.providers[r.defaultName], nil
}

// List returns the names of all registered providers.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// DefaultName returns the name of the current default provider.
func (r *Registry) DefaultName() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultName
}
