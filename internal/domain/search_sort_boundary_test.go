package domain

import "testing"

func TestSearchUnknownSortFieldDoesNotPanic(t *testing.T) {
	items := []DroneTask{
		{ID: "b", TaskCode: "B", Status: DroneTaskAccepted},
		{ID: "a", TaskCode: "A", Status: DroneTaskAccepted},
	}
	got := SearchDroneTasks(items, SearchRequest{Sort: SortField("unknown_field"), Limit: 10})
	if len(got) != 2 {
		t.Fatalf("expected stable result, got=%+v", got)
	}
}

func TestNormalizeCoercesUnknownSortToDefault(t *testing.T) {
	normalized := SearchRequest{Sort: SortField("bogus")}.Normalize()
	if normalized.Sort != SortCreated {
		t.Fatalf("unknown sort should fall back to SortCreated, got=%s", normalized.Sort)
	}
	for _, known := range []SortField{SortCreated, SortExpiry, SortCode} {
		got := SearchRequest{Sort: known}.Normalize()
		if got.Sort != known {
			t.Fatalf("known sort %s was coerced to %s", known, got.Sort)
		}
	}
}
