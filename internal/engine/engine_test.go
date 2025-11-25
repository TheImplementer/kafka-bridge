package engine

import (
	"encoding/json"
	"testing"

	"kafka-bridge/internal/store"
)

func TestFingerprintDeterministic(t *testing.T) {
	payload := map[string]any{
		"fieldA": "value1",
		"sub": map[string]any{
			"fieldB": "value2",
		},
	}
	data, _ := json.Marshal(payload)

	fp1, err := fingerprintMessage(data, []string{"sub.fieldB", "fieldA"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fp2, err := fingerprintMessage(data, []string{"fieldA", "sub.fieldB"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `[{"path":"fieldA","value":"value1"},{"path":"sub.fieldB","value":"value2"}]`

	if fp1 != expected || fp2 != expected {
		t.Fatalf("fingerprint mismatch\nwant: %s\ngot1: %s\ngot2: %s", expected, fp1, fp2)
	}
}

func TestFingerprintMissingField(t *testing.T) {
	payload := map[string]any{"fieldA": "value1"}
	data, _ := json.Marshal(payload)

	_, err := fingerprintMessage(data, []string{"missing"})
	if err == nil {
		t.Fatalf("expected error for missing field")
	}
}

func TestMatcherReferenceAndForward(t *testing.T) {
	s := store.NewMatchStore()
	m := NewMatcher("route", []string{"fieldA", "sub.fieldB"}, s)

	refPayload := map[string]any{
		"fieldA": "value1",
		"sub":    map[string]any{"fieldB": "value2"},
	}
	refBytes, _ := json.Marshal(refPayload)

	added, err := m.ProcessReference(refBytes)
	if err != nil || !added {
		t.Fatalf("expected reference to be added, err=%v", err)
	}

	srcPayload := map[string]any{
		"fieldA": "value1",
		"sub":    map[string]any{"fieldB": "value2"},
	}
	srcBytes, _ := json.Marshal(srcPayload)

	forward, err := m.ShouldForward(srcBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !forward {
		t.Fatalf("expected source to forward after reference cached")
	}

	nonMatching := map[string]any{
		"fieldA": "value1",
		"sub":    map[string]any{"fieldB": "other"},
	}
	nonBytes, _ := json.Marshal(nonMatching)
	forward, _ = m.ShouldForward(nonBytes)
	if forward {
		t.Fatalf("expected non-matching source to be blocked")
	}
}
