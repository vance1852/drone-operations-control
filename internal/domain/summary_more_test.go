package domain

import "testing"

func TestSearchEmptyPageAndLimits(t *testing.T) {
	items := SearchDroneTasks(nil, SearchRequest{Offset: 10, Limit: 2})
	if len(items) != 0 {
		t.Fatalf("items=%v", items)
	}
}

func TestRedactionShortValuesAreFullyHidden(t *testing.T) {
	if RedactTaskCode("abc") != "****" || RedactRoomLabel("ab") != "***" {
		t.Fatal("short values were not hidden")
	}
}

func TestStateMachineUnknownStateIsUnreachable(t *testing.T) {
	if DefaultMissionMachine().Allows("missing", "draft") {
		t.Fatal("unknown state has transition")
	}
}

func TestBusinessCodeAcceptsUnderscoreAndDigits(t *testing.T) {
	if err := ValidateBusinessCode("ROBOT_001"); err != nil {
		t.Fatal(err)
	}
}

func TestStatusCountsEmptyInputIsEmpty(t *testing.T) {
	if len(StatusCounts(nil)) != 0 {
		t.Fatal("empty status counts should be empty")
	}
}

func TestEnsureLimitAllowsBoundary(t *testing.T) {
	if err := EnsureLimit(3, 3, "drone_tasks"); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureLimitRejectsNameOnlyForMessage(t *testing.T) {
	if err := EnsureLimit(4, 3, "drone_tasks"); err == nil || err.Error() == "" {
		t.Fatal("limit error missing")
	}
}
