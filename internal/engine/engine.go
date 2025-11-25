package engine

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"kafka-bridge/internal/store"
)

// Matcher coordinates reference caching and source matching per route.
type Matcher struct {
	routeID string
	fields  []string
	store   *store.MatchStore
}

// NewMatcher constructs a matcher for a specific route.
func NewMatcher(routeID string, fields []string, store *store.MatchStore) *Matcher {
	return &Matcher{
		routeID: routeID,
		fields:  fields,
		store:   store,
	}
}

// ProcessReference ingests a reference payload and stores its fingerprint.
func (m *Matcher) ProcessReference(payload []byte) (bool, error) {
	fp, err := fingerprintMessage(payload, m.fields)
	if err != nil {
		return false, err
	}
	return m.store.Add(m.routeID, fp), nil
}

// ShouldForward evaluates a source payload against cached fingerprints.
func (m *Matcher) ShouldForward(payload []byte) (bool, error) {
	fp, err := fingerprintMessage(payload, m.fields)
	if err != nil {
		return false, err
	}
	return m.store.Contains(m.routeID, fp), nil
}

// Size returns the number of cached fingerprints for the route.
func (m *Matcher) Size() int {
	return m.store.Size(m.routeID)
}

func fingerprintMessage(value []byte, fields []string) (string, error) {
	var payload map[string]any
	if err := json.Unmarshal(value, &payload); err != nil {
		return "", err
	}

	entries := make([]fieldValue, len(fields))
	for i, field := range fields {
		val, err := lookupField(payload, field)
		if err != nil {
			return "", err
		}
		entries[i] = fieldValue{Path: field, Value: val}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	data, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func lookupField(payload map[string]any, field string) (any, error) {
	parts := strings.Split(field, ".")
	switch len(parts) {
	case 1:
		if val, ok := payload[parts[0]]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("field %s not found", field)
	case 2:
		child, ok := payload[parts[0]].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("field %s missing nested object", field)
		}
		if val, ok := child[parts[1]]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("field %s not found", field)
	default:
		return nil, fmt.Errorf("field %s depth unsupported", field)
	}
}

type fieldValue struct {
	Path  string `json:"path"`
	Value any    `json:"value"`
}
