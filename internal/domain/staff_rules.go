package domain

import "fmt"

func CanAssign(operator FleetOperator, assignment Assignment) error {
	if err := operator.Validate(); err != nil {
		return err
	}
	if assignment.FleetOperatorID != operator.ID {
		return fmt.Errorf("assignment operator mismatch: %w", ErrConflict)
	}
	if operator.Role != RoleDroneOperator && operator.Role != RoleSafetySupervisor {
		return fmt.Errorf("operator cannot receive a drone drone assignment: %w", ErrConflict)
	}
	return assignment.Validate()
}

func CanReview(operator FleetOperator, telemetry ObservationStatus) error {
	if !operator.Has(PermissionObservationReview) {
		return fmt.Errorf("operator cannot review telemetry_events: %w", ErrConflict)
	}
	if telemetry != ObservationPending {
		return fmt.Errorf("telemetry is already reviewed: %w", ErrConflict)
	}
	return nil
}
