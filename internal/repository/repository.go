package repository

import (
	"context"
	"time"

	"drone-operations-control/internal/domain"
)

type Page struct {
	Items  []domain.DroneTask `json:"items"`
	Total  int                `json:"total"`
	Offset int                `json:"offset"`
	Limit  int                `json:"limit"`
}

type DroneUnitInput struct {
	MissionID     string `json:"mission_id"`
	Code          string `json:"code"`
	RoomLabel     string `json:"room_label"`
	RequiredTasks int    `json:"required_tasks"`
}

type DroneTaskInput struct {
	MissionID   string    `json:"mission_id"`
	DroneUnitID string    `json:"drone_id"`
	TaskCode    string    `json:"task_code"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type HandoffInput struct {
	DroneTaskID string
	From        *string
	To          string
	Location    string
	RecordedAt  time.Time
	Note        string
}

type DroneMissionBatchInput struct {
	Code     string `json:"code"`
	Method   string `json:"method"`
	Capacity int    `json:"capacity"`
}

type ObservationInput struct {
	DroneTaskID         string    `json:"drone_task_id"`
	DroneMissionBatchID string    `json:"mission_batch_id"`
	RecorderID          string    `json:"recorded_by"`
	RiskScore           float64   `json:"risk_score"`
	Scale               string    `json:"scale"`
	AlertThreshold      float64   `json:"alert_threshold"`
	ObservedAt          time.Time `json:"observed_at"`
}

type InterventionInput struct {
	DroneTaskID string    `json:"drone_task_id"`
	Kind        string    `json:"kind"`
	Reason      string    `json:"reason"`
	DueAt       time.Time `json:"due_at"`
}

type AuditInput struct {
	RequestID       string
	FleetOperatorID *string
	ObjectType      string
	ObjectID        string
	Action          string
	Outcome         string
	Detail          []byte
}

type Repository interface {
	InTx(context.Context, func(Repository) error) error
	CreateMission(context.Context, *domain.Mission) error
	GetMission(context.Context, string) (domain.Mission, error)
	AdvanceMission(context.Context, string, domain.MissionStatus, int64) error
	CreateDroneUnit(context.Context, DroneUnitInput) (string, error)
	CreateDroneTask(context.Context, DroneTaskInput) (domain.DroneTask, error)
	GetDroneTask(context.Context, string) (domain.DroneTask, error)
	MoveDroneTask(context.Context, string, domain.DroneTaskStatus, int64, time.Time) error
	RecordHandoff(context.Context, HandoffInput) error
	CreateDroneMissionBatch(context.Context, DroneMissionBatchInput) (string, error)
	AttachDroneTasks(context.Context, string, []string) error
	CreateObservation(context.Context, ObservationInput) (string, error)
	ReviewObservationRecord(context.Context, string, bool, int64, time.Time) error
	CreateIntervention(context.Context, InterventionInput) (string, error)
	ListDroneTasks(context.Context, int, int, string, domain.DroneTaskStatus) (Page, error)
	DueInterventions(context.Context, time.Time, int) ([]InterventionInput, error)
	WriteAudit(context.Context, AuditInput) error
	Close() error
}
