package domain

type Permission string

const (
	PermissionMissionWrite      Permission = "mission:write"
	PermissionDroneTaskComplete Permission = "drone_task:complete"
	PermissionObservationRecord Permission = "telemetry:record"
	PermissionObservationReview Permission = "telemetry:review"
	PermissionInterventionClose Permission = "safety_alert:close"
)

func (o FleetOperator) Permissions() []Permission {
	switch o.Role {
	case RoleDroneOperator:
		return []Permission{PermissionDroneTaskComplete}
	case RoleTelemetryOperator:
		return []Permission{PermissionObservationRecord}
	case RoleQualityReviewer:
		return []Permission{PermissionObservationReview}
	case RoleSafetySupervisor:
		return []Permission{PermissionMissionWrite, PermissionDroneTaskComplete, PermissionObservationRecord, PermissionObservationReview, PermissionInterventionClose}
	default:
		return nil
	}
}

func (o FleetOperator) Has(permission Permission) bool {
	for _, item := range o.Permissions() {
		if item == permission {
			return true
		}
	}
	return false
}
