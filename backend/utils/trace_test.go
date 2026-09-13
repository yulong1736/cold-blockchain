package utils

import (
	"testing"
	"time"
)

func TestGenerateTraceID_UniquePerCall(t *testing.T) {
	a := GenerateTraceID("202604", "8", "signature")
	b := GenerateTraceID("202604", "8", "signature")
	if a == b {
		t.Fatalf("expected different trace ids, got duplicate: %s", a)
	}
	if len(a) != 64 {
		t.Fatalf("expected 64 hex chars, got len=%d", len(a))
	}
}

func TestGenerateTemperatureRecordHash_Deterministic(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2026-04-07T12:00:00Z")
	h1 := GenerateTemperatureRecordHash(1, "trace-a", -2.5, 55, "loc", ts, 10, "wh", false)
	h2 := GenerateTemperatureRecordHash(1, "trace-a", -2.5, 55, "loc", ts, 10, "wh", false)
	if h1 != h2 {
		t.Fatalf("same inputs should yield same hash: %s vs %s", h1, h2)
	}
	h3 := GenerateTemperatureRecordHash(2, "trace-a", -2.5, 55, "loc", ts, 10, "wh", false)
	if h1 == h3 {
		t.Fatal("different id should change hash")
	}
	if len(h1) != 64 {
		t.Fatalf("expected sha256 hex len 64, got %d", len(h1))
	}
}

func TestGenerateTransportNodeHash_Deterministic(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2026-04-07T15:30:00Z")
	h1 := GenerateTransportNodeHash(3, "t1", "节点", "上海", ts, time.Time{}, 20, "物流员", "in_transit")
	h2 := GenerateTransportNodeHash(3, "t1", "节点", "上海", ts, time.Time{}, 20, "物流员", "in_transit")
	if h1 != h2 {
		t.Fatalf("same inputs should yield same hash: %s vs %s", h1, h2)
	}
	dep, _ := time.Parse(time.RFC3339, "2026-04-07T18:00:00Z")
	h3 := GenerateTransportNodeHash(3, "t1", "节点", "上海", ts, dep, 20, "物流员", "in_transit")
	if h1 == h3 {
		t.Fatal("departure time should change hash")
	}
}
