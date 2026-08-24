package audit

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestExportCSVRejectsInvalidEvent(t *testing.T) {
	var output strings.Builder
	err := ExportCSV(&output, []Event{{ObjectType: "task"}})
	if err == nil {
		t.Fatal("invalid event exported")
	}
}

func TestFilterRequiresRequestCorrelation(t *testing.T) {
	event := Event{ObjectType: "task", ObjectID: "s1", Action: "create", Outcome: "success", CreatedAt: time.Now()}
	if (Filter{}).Match(event) {
		t.Fatal("uncorrelated event matched")
	}
}

func TestChainRejectsInvalidEvent(t *testing.T) {
	var chain Chain
	if _, err := chain.Append(Event{}); err == nil {
		t.Fatal("invalid event appended")
	}
}

// TestChainRejectedAppendLeavesHeadIsolated guards the integrity contract: an
// event that passes field validation but cannot be JSON-encoded must not
// advance the chain head. The subsequent normal append must link from the
// same prior node as a chain that never observed the rejection, otherwise
// different instances of the same event stream diverge into different heads.
func TestChainRejectedAppendLeavesHeadIsolated(t *testing.T) {
	unencodable := validEvent("create", "success")
	unencodable.Detail = map[string]any{"bad": math.NaN()}

	reject := func(c *Chain) {
		if _, err := c.Append(unencodable); err == nil {
			t.Fatal("unencodable event was appended")
		}
	}

	var withRejection, withoutRejection Chain
	reject(&withRejection)

	if withRejection.Head() != withoutRejection.Head() {
		t.Fatalf("rejected append advanced head: got %s, want %s", withRejection.Head(), withoutRejection.Head())
	}

	rejectedHead, err := withRejection.Append(validEvent("complete", "success"))
	if err != nil {
		t.Fatal(err)
	}
	cleanHead, err := withoutRejection.Append(validEvent("complete", "success"))
	if err != nil {
		t.Fatal(err)
	}
	if rejectedHead != cleanHead {
		t.Fatalf("chain diverged after rejected append: rejected=%s clean=%s", rejectedHead, cleanHead)
	}
}
