package domain

import "testing"

func TestObservationRecordingRequiresTelemetryOperatorRole(t *testing.T) {
	tests := []struct {
		role    FleetOperatorRole
		allowed bool
	}{
		{role: RoleDroneOperator, allowed: false},
		{role: RoleQualityReviewer, allowed: false},
		{role: RoleTelemetryOperator, allowed: true},
		{role: RoleSafetySupervisor, allowed: true},
	}
	for _, test := range tests {
		operator := FleetOperator{ID: "operator", Name: "FleetOperator", Role: test.role}
		if actual := operator.CanRecordObservation(); actual != test.allowed {
			t.Fatalf("role=%s allowed=%v", test.role, actual)
		}
	}
}
