package domain

import "testing"

func TestStatusCountsIncludesOnlyObservedStatuses(t *testing.T) {
	counts := StatusCounts([]DroneTask{{Status: DroneTaskAccepted}, {Status: DroneTaskAccepted}, {Status: DroneTaskRejected}})
	if counts[DroneTaskAccepted] != 2 || counts[DroneTaskRejected] != 1 || counts[DroneTaskVerified] != 0 {
		t.Fatalf("counts=%v", counts)
	}
}

func TestMissionProgressCompleteRequiresNoRejectedDroneTasks(t *testing.T) {
	progress := MissionProgress{Required: 2, Completed: 2}
	if !progress.Complete() {
		t.Fatal("complete progress rejected")
	}
	progress.Rejected = 1
	if progress.Complete() {
		t.Fatal("rejected progress marked complete")
	}
}

func TestFleetOperatorCanAction(t *testing.T) {
	operator := FleetOperator{ID: "o", Name: "Supervisor", Role: RoleSafetySupervisor}
	if !operator.Can("archive") || !operator.Can("review_telemetry") {
		t.Fatal("safety_supervisor permissions missing")
	}
	field := FleetOperator{ID: "f", Name: "Field", Role: RoleDroneOperator}
	if field.Can("archive") {
		t.Fatal("drone_operator can archive")
	}
}
