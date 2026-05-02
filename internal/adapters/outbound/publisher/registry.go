package publisher

import (
	"fmt"
	"sync"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
)

// DestinationRegistry manages registered publishers for different destinations
type DestinationRegistry struct {
	publishers map[string]outbound.PublisherService
	mu         sync.RWMutex
}

// DestinationInfo provides metadata about a registered destination
type DestinationInfo struct {
	Name                  string
	SupportedEnvironments []string
}

// NewDestinationRegistry creates a new empty registry
func NewDestinationRegistry() *DestinationRegistry {
	return &DestinationRegistry{
		publishers: make(map[string]outbound.PublisherService),
	}
}

func (r *DestinationRegistry) NumberOfPublishers() int {
	return len(r.publishers)
}

// Register adds a publisher for a destination
// Returns error if destination is already registered
func (r *DestinationRegistry) Register(destination string, publisher outbound.PublisherService) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.publishers[destination]; exists {
		return fmt.Errorf("destination %s is already registered", destination)
	}

	r.publishers[destination] = publisher
	return nil
}

// GetPublisher retrieves a publisher for a destination
// Returns error if destination is not registered
func (r *DestinationRegistry) GetPublisher(destination string) (outbound.PublisherService, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	publisher, exists := r.publishers[destination]
	if !exists {
		return nil, fmt.Errorf("destination %s is not registered", destination)
	}

	return publisher, nil
}

// ListDestinations returns information about all registered destinations
func (r *DestinationRegistry) ListDestinations() []DestinationInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	destinations := make([]DestinationInfo, 0, len(r.publishers))
	for name, pub := range r.publishers {
		destinations = append(destinations, DestinationInfo{
			Name:                  name,
			SupportedEnvironments: pub.SupportedEnvironments(),
		})
	}

	return destinations
}

// IsRegistered checks if a destination is registered
func (r *DestinationRegistry) IsRegistered(destination string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.publishers[destination]
	return exists
}
