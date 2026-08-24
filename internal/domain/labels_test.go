package domain

import "testing"

func TestNormalizeLabelsCopiesAndCleansValues(t *testing.T) {
	labels := map[string]string{" Drone ": " North ", "": "ignored", "empty": " "}
	got := NormalizeLabels(labels)
	if got["drone"] != "North" || len(got) != 1 {
		t.Fatalf("labels=%v", got)
	}
	got["drone"] = "changed"
	if labels[" Drone "] != " North " {
		t.Fatal("source labels were mutated")
	}
}
