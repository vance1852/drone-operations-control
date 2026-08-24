package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"drone-operations-control/internal/db"
	"drone-operations-control/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Postgres struct {
	pool *db.Pool
}

func NewPostgres(pool *db.Pool) *Postgres { return &Postgres{pool: pool} }

func (p *Postgres) InTx(ctx context.Context, fn func(Repository) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := fn(&transaction{tx: tx}); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (p *Postgres) Close() error { p.pool.Close(); return nil }

func (p *Postgres) CreateMission(ctx context.Context, mission *domain.Mission) error {
	return createMission(ctx, p.pool, mission)
}

func (p *Postgres) GetMission(ctx context.Context, id string) (domain.Mission, error) {
	return getMission(ctx, p.pool, id)
}

func (p *Postgres) AdvanceMission(ctx context.Context, id string, status domain.MissionStatus, version int64) error {
	return advanceMission(ctx, p.pool, id, status, version)
}

func (p *Postgres) CreateDroneUnit(ctx context.Context, in DroneUnitInput) (string, error) {
	return createDroneUnit(ctx, p.pool, in)
}
func (p *Postgres) CreateDroneTask(ctx context.Context, in DroneTaskInput) (domain.DroneTask, error) {
	return createDroneTask(ctx, p.pool, in)
}
func (p *Postgres) GetDroneTask(ctx context.Context, id string) (domain.DroneTask, error) {
	return getDroneTask(ctx, p.pool, id)
}
func (p *Postgres) MoveDroneTask(ctx context.Context, id string, status domain.DroneTaskStatus, version int64, now time.Time) error {
	return moveDroneTask(ctx, p.pool, id, status, version, now)
}
func (p *Postgres) RecordHandoff(ctx context.Context, in HandoffInput) error {
	return recordHandoff(ctx, p.pool, in)
}
func (p *Postgres) CreateDroneMissionBatch(ctx context.Context, in DroneMissionBatchInput) (string, error) {
	return createDroneMissionBatch(ctx, p.pool, in)
}
func (p *Postgres) AttachDroneTasks(ctx context.Context, missionBatchID string, taskIDs []string) error {
	return attachDroneTasks(ctx, p.pool, missionBatchID, taskIDs)
}
func (p *Postgres) CreateObservation(ctx context.Context, in ObservationInput) (string, error) {
	return createObservation(ctx, p.pool, in)
}
func (p *Postgres) ReviewObservationRecord(ctx context.Context, id string, accepted bool, version int64, now time.Time) error {
	return reviewObservationRecord(ctx, p.pool, id, accepted, version, now)
}
func (p *Postgres) CreateIntervention(ctx context.Context, in InterventionInput) (string, error) {
	return createIntervention(ctx, p.pool, in)
}
func (p *Postgres) ListDroneTasks(ctx context.Context, offset, limit int, missionID string, status domain.DroneTaskStatus) (Page, error) {
	return listDroneTasks(ctx, p.pool, offset, limit, missionID, status)
}
func (p *Postgres) DueInterventions(ctx context.Context, before time.Time, limit int) ([]InterventionInput, error) {
	return dueInterventions(ctx, p.pool, before, limit)
}
func (p *Postgres) WriteAudit(ctx context.Context, in AuditInput) error {
	return writeAudit(ctx, p.pool, in)
}

type transaction struct{ tx pgx.Tx }

func (t *transaction) InTx(_ context.Context, _ func(Repository) error) error {
	return errors.New("nested transaction")
}
func (t *transaction) Close() error { return nil }
func (t *transaction) CreateMission(ctx context.Context, p *domain.Mission) error {
	return createMission(ctx, t.tx, p)
}
func (t *transaction) GetMission(ctx context.Context, id string) (domain.Mission, error) {
	return getMission(ctx, t.tx, id)
}
func (t *transaction) AdvanceMission(ctx context.Context, id string, s domain.MissionStatus, v int64) error {
	return advanceMission(ctx, t.tx, id, s, v)
}
func (t *transaction) CreateDroneUnit(ctx context.Context, in DroneUnitInput) (string, error) {
	return createDroneUnit(ctx, t.tx, in)
}
func (t *transaction) CreateDroneTask(ctx context.Context, in DroneTaskInput) (domain.DroneTask, error) {
	return createDroneTask(ctx, t.tx, in)
}
func (t *transaction) GetDroneTask(ctx context.Context, id string) (domain.DroneTask, error) {
	return getDroneTask(ctx, t.tx, id)
}
func (t *transaction) MoveDroneTask(ctx context.Context, id string, s domain.DroneTaskStatus, v int64, now time.Time) error {
	return moveDroneTask(ctx, t.tx, id, s, v, now)
}
func (t *transaction) RecordHandoff(ctx context.Context, in HandoffInput) error {
	return recordHandoff(ctx, t.tx, in)
}
func (t *transaction) CreateDroneMissionBatch(ctx context.Context, in DroneMissionBatchInput) (string, error) {
	return createDroneMissionBatch(ctx, t.tx, in)
}
func (t *transaction) AttachDroneTasks(ctx context.Context, id string, ids []string) error {
	return attachDroneTasks(ctx, t.tx, id, ids)
}
func (t *transaction) CreateObservation(ctx context.Context, in ObservationInput) (string, error) {
	return createObservation(ctx, t.tx, in)
}
func (t *transaction) ReviewObservationRecord(ctx context.Context, id string, accepted bool, v int64, now time.Time) error {
	return reviewObservationRecord(ctx, t.tx, id, accepted, v, now)
}
func (t *transaction) CreateIntervention(ctx context.Context, in InterventionInput) (string, error) {
	return createIntervention(ctx, t.tx, in)
}
func (t *transaction) ListDroneTasks(ctx context.Context, offset, limit int, missionID string, status domain.DroneTaskStatus) (Page, error) {
	return listDroneTasks(ctx, t.tx, offset, limit, missionID, status)
}
func (t *transaction) DueInterventions(ctx context.Context, before time.Time, limit int) ([]InterventionInput, error) {
	return dueInterventions(ctx, t.tx, before, limit)
}
func (t *transaction) WriteAudit(ctx context.Context, in AuditInput) error {
	return writeAudit(ctx, t.tx, in)
}

type sqler interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func createMission(ctx context.Context, q sqler, mission *domain.Mission) error {
	_, err := q.Exec(ctx, `INSERT INTO drone_missions(id,code,name,status,timezone,starts_at,ends_at,version,created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, mission.ID, mission.Code, mission.Name, mission.Status, mission.Timezone, mission.StartsAt, mission.EndsAt, mission.Version, mission.CreatedBy)
	return wrapWrite(err)
}

func getMission(ctx context.Context, q sqler, id string) (domain.Mission, error) {
	var p domain.Mission
	err := q.QueryRow(ctx, `SELECT id,code,name,status,timezone,starts_at,ends_at,version,created_by FROM drone_missions WHERE id=$1`, id).Scan(&p.ID, &p.Code, &p.Name, &p.Status, &p.Timezone, &p.StartsAt, &p.EndsAt, &p.Version, &p.CreatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Mission{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Mission{}, fmt.Errorf("get mission: %w", err)
	}
	return p, nil
}

func advanceMission(ctx context.Context, q sqler, id string, status domain.MissionStatus, version int64) error {
	result, err := q.Exec(ctx, `UPDATE drone_missions SET status=$1,version=version+1 WHERE id=$2 AND version=$3`, status, id, version)
	if err != nil {
		return fmt.Errorf("advance mission: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func createDroneUnit(ctx context.Context, q sqler, in DroneUnitInput) (string, error) {
	id := uuid.NewString()
	_, err := q.Exec(ctx, `INSERT INTO mission_drones(id,mission_id,code,room_label,required_tasks) VALUES ($1,$2,$3,$4,$5)`, id, in.MissionID, in.Code, in.RoomLabel, in.RequiredTasks)
	return wrapID(err, id)
}

func createDroneTask(ctx context.Context, q sqler, in DroneTaskInput) (domain.DroneTask, error) {
	s := domain.DroneTask{ID: uuid.NewString(), MissionID: in.MissionID, DroneUnitID: in.DroneUnitID, TaskCode: in.TaskCode, Status: domain.DroneTaskQueued, ExpiresAt: in.ExpiresAt, Version: 1}
	_, err := q.Exec(ctx, `INSERT INTO drone_tasks(id,mission_id,drone_id,task_code,status,expires_at,version) VALUES ($1,$2,$3,$4,$5,$6,$7)`, s.ID, s.MissionID, s.DroneUnitID, s.TaskCode, s.Status, s.ExpiresAt, s.Version)
	return wrapDroneTask(err, s)
}

func getDroneTask(ctx context.Context, q sqler, id string) (domain.DroneTask, error) {
	var s domain.DroneTask
	err := q.QueryRow(ctx, `SELECT id,mission_id,drone_id,task_code,status,completed_at,accepted_at,expires_at,version FROM drone_tasks WHERE id=$1`, id).Scan(&s.ID, &s.MissionID, &s.DroneUnitID, &s.TaskCode, &s.Status, &s.CompletedAt, &s.AcceptedAt, &s.ExpiresAt, &s.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DroneTask{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DroneTask{}, fmt.Errorf("get task: %w", err)
	}
	return s, nil
}

func moveDroneTask(ctx context.Context, q sqler, id string, status domain.DroneTaskStatus, version int64, now time.Time) error {
	if status == domain.DroneTaskCompleted {
		result, err := q.Exec(ctx, `WITH eligible_drone AS (
			SELECT id FROM mission_drones WHERE id=(SELECT drone_id FROM drone_tasks WHERE id=$3) AND completed_tasks < required_tasks FOR UPDATE
		), updated AS (
			UPDATE drone_tasks SET status=$1,version=version+1,completed_at=$2 WHERE id=$3 AND version=$4 AND EXISTS (SELECT 1 FROM eligible_drone) RETURNING drone_id
		)
		UPDATE mission_drones SET completed_tasks=completed_tasks+1 WHERE id IN (SELECT drone_id FROM updated)`, status, now, id, version)
		if err != nil {
			return fmt.Errorf("complete task: %w", err)
		}
		if result.RowsAffected() != 1 {
			return domain.ErrConflict
		}
		return nil
	}
	result, err := q.Exec(ctx, `UPDATE drone_tasks SET status=$1,version=version+1,completed_at=CASE WHEN $1='completed' THEN $2 ELSE completed_at END,accepted_at=CASE WHEN $1='accepted' THEN $2 ELSE accepted_at END WHERE id=$3 AND version=$4`, status, now, id, version)
	if err != nil {
		return fmt.Errorf("move task: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func recordHandoff(ctx context.Context, q sqler, in HandoffInput) error {
	_, err := q.Exec(ctx, `INSERT INTO device_transfer_events(id,drone_task_id,from_operator,to_operator,location,recorded_at,note) VALUES ($1,$2,$3,$4,$5,$6,$7)`, uuid.NewString(), in.DroneTaskID, in.From, in.To, in.Location, in.RecordedAt, in.Note)
	return wrapWrite(err)
}

func createDroneMissionBatch(ctx context.Context, q sqler, in DroneMissionBatchInput) (string, error) {
	id := uuid.NewString()
	_, err := q.Exec(ctx, `INSERT INTO mission_batches(id,code,status,method,capacity) VALUES ($1,$2,'queued',$3,$4)`, id, in.Code, in.Method, in.Capacity)
	return wrapID(err, id)
}

func attachDroneTasks(ctx context.Context, q sqler, missionBatchID string, taskIDs []string) error {
	for _, taskID := range taskIDs {
		if _, err := q.Exec(ctx, `INSERT INTO mission_batch_tasks(mission_batch_id,drone_task_id) VALUES ($1,$2)`, missionBatchID, taskID); err != nil {
			return wrapWrite(err)
		}
		result, err := q.Exec(ctx, `UPDATE drone_tasks SET status='in_progress',version=version+1 WHERE id=$1 AND status='accepted' AND expires_at >= now()`, taskID)
		if err != nil {
			return wrapWrite(err)
		}
		if result.RowsAffected() != 1 {
			return fmt.Errorf("task %s is not eligible for drone-round execution: %w", taskID, domain.ErrInvalidTransition)
		}
	}
	return nil
}

func createObservation(ctx context.Context, q sqler, in ObservationInput) (string, error) {
	id := uuid.NewString()
	_, err := q.Exec(ctx, `INSERT INTO telemetry_events(id,drone_task_id,mission_batch_id,recorded_by,status,risk_score,scale,alert_threshold,observed_at) VALUES ($1,$2,$3,$4,'pending',$5,$6,$7,$8)`, id, in.DroneTaskID, in.DroneMissionBatchID, in.RecorderID, in.RiskScore, in.Scale, in.AlertThreshold, in.ObservedAt)
	return wrapID(err, id)
}

func reviewObservationRecord(ctx context.Context, q sqler, id string, accepted bool, version int64, now time.Time) error {
	status := domain.ObservationVerified
	if !accepted {
		status = domain.ObservationRejected
	}
	result, err := q.Exec(ctx, `UPDATE telemetry_events SET status=$1,reviewed_at=$2,version=version+1 WHERE id=$3 AND status='pending' AND version=$4`, status, now, id, version)
	if err != nil {
		return fmt.Errorf("review telemetry: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func createIntervention(ctx context.Context, q sqler, in InterventionInput) (string, error) {
	id := uuid.NewString()
	_, err := q.Exec(ctx, `INSERT INTO safety_alerts(id,drone_task_id,kind,status,reason,due_at) VALUES ($1,$2,$3,'open',$4,$5)`, id, in.DroneTaskID, in.Kind, in.Reason, in.DueAt)
	return wrapID(err, id)
}

func listDroneTasks(ctx context.Context, q sqler, offset, limit int, missionID string, status domain.DroneTaskStatus) (Page, error) {
	page := Page{Offset: offset, Limit: limit, Items: make([]domain.DroneTask, 0)}
	args := []any{limit, offset}
	where := "WHERE TRUE"
	if missionID != "" {
		args = append(args, missionID)
		where += fmt.Sprintf(" AND mission_id=$%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		where += fmt.Sprintf(" AND status=$%d", len(args))
	}
	query := fmt.Sprintf(`SELECT id,mission_id,drone_id,task_code,status,completed_at,accepted_at,expires_at,version FROM drone_tasks %s ORDER BY created_at DESC LIMIT $1 OFFSET $2`, where)
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return Page{}, fmt.Errorf("list drone_tasks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s domain.DroneTask
		if err := rows.Scan(&s.ID, &s.MissionID, &s.DroneUnitID, &s.TaskCode, &s.Status, &s.CompletedAt, &s.AcceptedAt, &s.ExpiresAt, &s.Version); err != nil {
			return Page{}, fmt.Errorf("scan task: %w", err)
		}
		page.Items = append(page.Items, s)
	}
	if err := rows.Err(); err != nil {
		return Page{}, fmt.Errorf("list task rows: %w", err)
	}
	countWhere := "WHERE TRUE"
	countArgs := make([]any, 0, len(args)-2)
	if missionID != "" {
		countArgs = append(countArgs, missionID)
		countWhere += fmt.Sprintf(" AND mission_id=$%d", len(countArgs))
	}
	if status != "" {
		countArgs = append(countArgs, status)
		countWhere += fmt.Sprintf(" AND status=$%d", len(countArgs))
	}
	countQuery := fmt.Sprintf("SELECT count(*) FROM drone_tasks %s", countWhere)
	if err := q.QueryRow(ctx, countQuery, countArgs...).Scan(&page.Total); err != nil {
		return Page{}, fmt.Errorf("count drone_tasks: %w", err)
	}
	return page, nil
}

func dueInterventions(ctx context.Context, q sqler, before time.Time, limit int) ([]InterventionInput, error) {
	rows, err := q.Query(ctx, `SELECT drone_task_id,kind,reason,due_at FROM safety_alerts WHERE status IN ('open','in_progress') AND due_at <= $1 ORDER BY due_at LIMIT $2`, before, limit)
	if err != nil {
		return nil, fmt.Errorf("due safety_alerts: %w", err)
	}
	defer rows.Close()
	out := make([]InterventionInput, 0)
	for rows.Next() {
		var item InterventionInput
		if err := rows.Scan(&item.DroneTaskID, &item.Kind, &item.Reason, &item.DueAt); err != nil {
			return nil, fmt.Errorf("scan safety_alert: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func writeAudit(ctx context.Context, q sqler, in AuditInput) error {
	detail := in.Detail
	if len(detail) == 0 {
		detail = []byte(`{}`)
	}
	if !json.Valid(detail) {
		return fmt.Errorf("invalid audit detail: %w", domain.ErrConflict)
	}
	_, err := q.Exec(ctx, `INSERT INTO audit_events(id,request_id,operator_id,object_type,object_id,action,outcome,detail) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, uuid.NewString(), in.RequestID, in.FleetOperatorID, in.ObjectType, in.ObjectID, in.Action, in.Outcome, detail)
	return wrapWrite(err)
}

func wrapWrite(err error) error {
	if err == nil {
		return nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "22003", "22P02", "23502", "23503", "23505", "23514":
			return fmt.Errorf("repository write: %w: %w", err, domain.ErrConflict)
		}
	}
	return fmt.Errorf("repository write: %w", err)
}
func wrapID(err error, id string) (string, error) {
	if err != nil {
		return "", wrapWrite(err)
	}
	return id, nil
}
func wrapDroneTask(err error, s domain.DroneTask) (domain.DroneTask, error) {
	if err != nil {
		return domain.DroneTask{}, wrapWrite(err)
	}
	return s, nil
}
