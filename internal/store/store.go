package store

import "sync"

// MatchStore keeps allowed payload fingerprints per route.
type MatchStore struct {
	mu     sync.RWMutex
	values map[string]map[string]struct{}
}

// NewMatchStore creates an empty store.
func NewMatchStore() *MatchStore {
	return &MatchStore{values: make(map[string]map[string]struct{})}
}

// Add inserts the fingerprint for the given route.
func (s *MatchStore) Add(route string, fingerprint string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	routeMap, ok := s.values[route]
	if !ok {
		routeMap = make(map[string]struct{})
		s.values[route] = routeMap
	}
	if _, exists := routeMap[fingerprint]; exists {
		return false
	}
	routeMap[fingerprint] = struct{}{}
	return true
}

// Contains reports whether a fingerprint exists for the route.
func (s *MatchStore) Contains(route string, fingerprint string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	routeMap, ok := s.values[route]
	if !ok {
		return false
	}
	_, exists := routeMap[fingerprint]
	return exists
}

// Size returns the number of fingerprints stored for the route.
func (s *MatchStore) Size(route string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.values[route])
}
