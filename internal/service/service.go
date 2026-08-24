package service

import (
	"context"
	"fmt"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
	"github.com/google/uuid"
)

type Clock func() time.Time

type Service struct {
	repo repository.Repository
	now  Clock
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) WithClock(clock Clock) *Service {
	if clock != nil {
		s.now = clock
	}
	return s
}

type RequestMeta struct {
	RequestID       string
	FleetOperatorID *string
}

type CreateMissionRequest struct {
	Code       string
	Name       string
	Timezone   string
	StartsAt   time.Time
	EndsAt     time.Time
	CreatedBy  string
	DroneUnits []repository.DroneUnitInput
}

type CreateMissionResponse struct {
	Mission      domain.Mission `json:"mission"`
	DroneUnitIDs []string       `json:"drone_ids"`
}

func (s *Service) CreateMission(ctx context.Context, meta RequestMeta, in CreateMissionRequest) (CreateMissionResponse, error) {
	if in.Code == "" || in.Name == "" || in.CreatedBy == "" || len(in.DroneUnits) == 0 {
		return CreateMissionResponse{}, fmt.Errorf("code, name, creator and drones are required: %w", domain.ErrConflict)
	}
	if err := domain.ValidateBusinessCode(in.Code); err != nil {
		return CreateMissionResponse{}, err
	}
	if err := validateCode(in.Name, "mission name"); err != nil {
		return CreateMissionResponse{}, err
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil {
		return CreateMissionResponse{}, fmt.Errorf("invalid mission timezone: %w", domain.ErrConflict)
	}
	seenDroneUnits := make(map[string]struct{}, len(in.DroneUnits))
	for _, drone := range in.DroneUnits {
		if err := (domain.DroneUnit{Code: drone.Code, RoomLabel: drone.RoomLabel, RequiredTasks: drone.RequiredTasks}).Validate(); err != nil {
			return CreateMissionResponse{}, err
		}
		if err := domain.ValidateBusinessCode(drone.Code); err != nil {
			return CreateMissionResponse{}, err
		}
		if _, exists := seenDroneUnits[drone.Code]; exists {
			return CreateMissionResponse{}, fmt.Errorf("duplicate drone code %s: %w", drone.Code, domain.ErrConflict)
		}
		seenDroneUnits[drone.Code] = struct{}{}
	}
	mission := domain.Mission{ID: uuid.NewString(), Code: in.Code, Name: in.Name, Status: domain.MissionDraft, Timezone: in.Timezone, StartsAt: in.StartsAt, EndsAt: in.EndsAt, Version: 1, CreatedBy: in.CreatedBy}
	if err := mission.ValidateWindow(s.now()); err != nil {
		return CreateMissionResponse{}, err
	}
	response := CreateMissionResponse{Mission: mission, DroneUnitIDs: make([]string, 0, len(in.DroneUnits))}
	err := s.runCreateMissionTransaction(ctx, func(tx repository.Repository) error {
		if err := tx.CreateMission(ctx, &mission); err != nil {
			return err
		}
		for _, drone := range in.DroneUnits {
			drone.MissionID = mission.ID
			id, err := tx.CreateDroneUnit(ctx, drone)
			if err != nil {
				return err
			}
			response.DroneUnitIDs = append(response.DroneUnitIDs, id)
		}
		return tx.WriteAudit(ctx, audit(meta, "drone_mission", mission.ID, "create", "success", nil))
	})
	if err != nil {
		return CreateMissionResponse{}, err
	}
	return response, nil
}

func (s *Service) ScheduleMission(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.advanceMission(ctx, meta, id, domain.MissionScheduled, version, "schedule")
}

func (s *Service) ActivateMission(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.advanceMission(ctx, meta, id, domain.MissionActive, version, "activate")
}

func (s *Service) CloseMission(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.advanceMission(ctx, meta, id, domain.MissionClosed, version, "close")
}

func (s *Service) advanceMission(ctx context.Context, meta RequestMeta, id string, next domain.MissionStatus, version int64, action string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		mission, err := tx.GetMission(ctx, id)
		if err != nil {
			return err
		}
		if !mission.Status.CanMoveTo(next) {
			return fmt.Errorf("mission %s: %w", id, domain.ErrInvalidTransition)
		}
		if err := tx.AdvanceMission(ctx, id, next, version); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "drone_mission", id, action, "success", nil))
	})
}

func (s *Service) CreateDroneTask(ctx context.Context, meta RequestMeta, in repository.DroneTaskInput) (domain.DroneTask, error) {
	request := domain.DroneTaskRequest{MissionID: in.MissionID, DroneUnitID: in.DroneUnitID, TaskCode: in.TaskCode, ExpiresAt: in.ExpiresAt}
	if err := request.Validate(s.now()); err != nil {
		return domain.DroneTask{}, err
	}
	if err := domain.ValidateBusinessCode(in.TaskCode); err != nil {
		return domain.DroneTask{}, err
	}
	var task domain.DroneTask
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		placement, ok := tx.(interface {
			ValidateMissionDroneUnit(context.Context, string, string) error
		})
		if !ok {
			return fmt.Errorf("task placement repository unavailable")
		}
		if err := placement.ValidateMissionDroneUnit(ctx, in.MissionID, in.DroneUnitID); err != nil {
			return err
		}
		var err error
		task, err = tx.CreateDroneTask(ctx, in)
		if err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "drone_task", task.ID, "create", "success", nil))
	})
	return task, err
}

func (s *Service) CompleteDroneTask(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.moveDroneTask(ctx, meta, id, version, domain.DroneTaskCompleted, "complete")
}

func (s *Service) HandoffDroneTask(ctx context.Context, meta RequestMeta, in repository.HandoffInput, version int64) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		if err := tx.MoveDroneTask(ctx, in.DroneTaskID, domain.DroneTaskHandoffPending, version, in.RecordedAt); err != nil {
			return err
		}
		if err := tx.RecordHandoff(ctx, in); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "drone_task", in.DroneTaskID, "device_transfer", "success", nil))
	})
}

func (s *Service) AcceptDroneTask(ctx context.Context, meta RequestMeta, in repository.HandoffInput, version int64) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		if err := tx.MoveDroneTask(ctx, in.DroneTaskID, domain.DroneTaskAccepted, version, in.RecordedAt); err != nil {
			return err
		}
		if err := tx.RecordHandoff(ctx, in); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "drone_task", in.DroneTaskID, "accept", "success", nil))
	})
}

func (s *Service) CreateDroneMissionBatch(ctx context.Context, meta RequestMeta, in repository.DroneMissionBatchInput, taskIDs []string) (string, error) {
	if len(taskIDs) == 0 || len(taskIDs) > in.Capacity {
		return "", domain.ErrCapacityExceeded
	}
	if err := (domain.DroneMissionBatch{Code: in.Code, Method: in.Method, Capacity: in.Capacity, DroneTasks: taskIDs}).Validate(); err != nil {
		return "", err
	}
	if err := domain.ValidateBusinessCode(in.Code); err != nil {
		return "", err
	}
	if err := validateIDs(taskIDs); err != nil {
		return "", err
	}
	var id string
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		var err error
		id, err = tx.CreateDroneMissionBatch(ctx, in)
		if err != nil {
			return err
		}
		if err := tx.AttachDroneTasks(ctx, id, append([]string(nil), taskIDs...)); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "mission_batch", id, "create", "success", nil))
	})
	return id, err
}

func (s *Service) SubmitObservation(ctx context.Context, meta RequestMeta, in repository.ObservationInput) (string, error) {
	if in.DroneTaskID == "" || in.DroneMissionBatchID == "" || in.RecorderID == "" || in.ObservedAt.IsZero() {
		return "", fmt.Errorf("telemetry identifiers and observed_at are required: %w", domain.ErrConflict)
	}
	if err := domain.ValidateObservation(in.RiskScore, in.AlertThreshold, in.Scale); err != nil {
		return "", err
	}
	var id string
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		target, ok := tx.(interface {
			ValidateObservationTarget(context.Context, string, string) error
			GetFleetOperator(context.Context, string) (domain.FleetOperator, error)
		})
		if !ok {
			return fmt.Errorf("telemetry target repository unavailable")
		}
		if err := target.ValidateObservationTarget(ctx, in.DroneTaskID, in.DroneMissionBatchID); err != nil {
			return err
		}
		recorder, err := target.GetFleetOperator(ctx, in.RecorderID)
		if err != nil {
			return err
		}
		if !recorder.CanRecordObservation() {
			return fmt.Errorf("operator role %s cannot record telemetry_events: %w", recorder.Role, domain.ErrConflict)
		}
		id, err = tx.CreateObservation(ctx, in)
		if err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "telemetry", id, "record", "success", nil))
	})
	return id, err
}

func (s *Service) ReviewObservation(ctx context.Context, meta RequestMeta, telemetryID, taskID string, accepted bool, telemetryVersion, taskVersion int64) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		telemetry, ok := tx.(interface {
			ObservationTaskID(context.Context, string) (string, error)
		})
		if !ok {
			return fmt.Errorf("telemetry query repository unavailable")
		}
		actualDroneTaskID, err := telemetry.ObservationTaskID(ctx, telemetryID)
		if err != nil {
			return err
		}
		if actualDroneTaskID != taskID {
			return fmt.Errorf("telemetry and task do not match: %w", domain.ErrConflict)
		}
		if err := tx.ReviewObservationRecord(ctx, telemetryID, accepted, telemetryVersion, s.now()); err != nil {
			return err
		}
		next := domain.DroneTaskVerified
		if !accepted {
			next = domain.DroneTaskRejected
		}
		if err := tx.MoveDroneTask(ctx, taskID, next, taskVersion, s.now()); err != nil {
			return err
		}
		if !accepted {
			safety_alertID, err := tx.CreateIntervention(ctx, repository.InterventionInput{DroneTaskID: taskID, Kind: "reassess", Reason: "risk score exceeded the alert threshold", DueAt: s.now().Add(72 * time.Hour)})
			if err != nil {
				return err
			}
			if err := tx.WriteAudit(ctx, audit(meta, "safety_alert", safety_alertID, "open", "success", nil)); err != nil {
				return err
			}
		}
		return tx.WriteAudit(ctx, audit(meta, "telemetry", telemetryID, "review", "success", nil))
	})
}

func (s *Service) ArchiveDroneTask(ctx context.Context, meta RequestMeta, taskID string, version int64) error {
	return s.moveDroneTask(ctx, meta, taskID, version, domain.DroneTaskArchived, "archive")
}

func (s *Service) ListDroneTasks(ctx context.Context, offset, limit int, missionID string, status domain.DroneTaskStatus) (repository.Page, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.ListDroneTasks(ctx, offset, limit, missionID, status)
}

func (s *Service) DueInterventions(ctx context.Context, before time.Time, limit int) ([]repository.InterventionInput, error) {
	return s.repo.DueInterventions(ctx, before, limit)
}

func (s *Service) moveDroneTask(ctx context.Context, meta RequestMeta, id string, version int64, next domain.DroneTaskStatus, action string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		task, err := tx.GetDroneTask(ctx, id)
		if err != nil {
			return err
		}
		if next == domain.DroneTaskCompleted {
			mission, err := tx.GetMission(ctx, task.MissionID)
			if err != nil {
				return err
			}
			if !mission.CanExecuteAt(s.now()) {
				return fmt.Errorf("mission is not active for task execution: %w", domain.ErrInvalidTransition)
			}
		}
		updated, err := task.Move(next, s.now())
		if err != nil {
			return err
		}
		if err := tx.MoveDroneTask(ctx, id, updated.Status, version, s.now()); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "drone_task", id, action, "success", nil))
	})
}

func audit(meta RequestMeta, objectType, objectID, action, outcome string, detail []byte) repository.AuditInput {
	return repository.AuditInput{RequestID: meta.RequestID, FleetOperatorID: meta.FleetOperatorID, ObjectType: objectType, ObjectID: objectID, Action: action, Outcome: outcome, Detail: detail}
}
