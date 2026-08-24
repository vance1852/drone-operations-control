package domain

import (
	"fmt"
	"strings"
)

type FleetOperatorRole string

const (
	RoleDroneOperator     FleetOperatorRole = "drone_operator"
	RoleTelemetryOperator FleetOperatorRole = "telemetry_operator"
	RoleQualityReviewer   FleetOperatorRole = "quality_reviewer"
	RoleSafetySupervisor  FleetOperatorRole = "safety_supervisor"
)

type FleetOperator struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Role FleetOperatorRole `json:"role"`
}

func (o FleetOperator) Validate() error {
	if strings.TrimSpace(o.ID) == "" || strings.TrimSpace(o.Name) == "" {
		return fmt.Errorf("operator identity is required: %w", ErrConflict)
	}
	switch o.Role {
	case RoleDroneOperator, RoleTelemetryOperator, RoleQualityReviewer, RoleSafetySupervisor:
		return nil
	default:
		return fmt.Errorf("unknown operator role: %w", ErrConflict)
	}
}

func (o FleetOperator) CanRecordObservation() bool {
	if err := o.Validate(); err != nil {
		return false
	}
	switch o.Role {
	case RoleTelemetryOperator, RoleSafetySupervisor:
		return true
	default:
		return false
	}
}

func (o FleetOperator) Can(action string) bool {
	switch action {
	case "complete", "device_transfer":
		return o.Role == RoleDroneOperator || o.Role == RoleSafetySupervisor
	case "record_telemetry":
		return o.Role == RoleTelemetryOperator || o.Role == RoleSafetySupervisor
	case "review_telemetry":
		return o.Role == RoleQualityReviewer || o.Role == RoleSafetySupervisor
	case "close_mission", "archive":
		return o.Role == RoleSafetySupervisor
	default:
		return false
	}
}
