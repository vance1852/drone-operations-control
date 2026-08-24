package console

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Store) TelemetryPage(ctx context.Context, current, size int, droneID string) (Page[TelemetryRecord], error) {
	current, size, offset := pageBounds(current, size)
	page := Page[TelemetryRecord]{Records: make([]TelemetryRecord, 0), Current: current, Size: size}
	where := ""
	args := []any{size, offset}
	if droneID != "" {
		where = "WHERE h.drone_id=$3"
		args = append(args, droneID)
	}
	rows, err := s.pool.Query(ctx, `SELECT h.id,h.drone_id,e.name,h.battery_level,h.motor_temperature,h.network_latency_ms,h.localization_error,h.joint_load,h.remark,h.recorded_at
		FROM console_telemetry_records h JOIN console_drones e ON e.id=h.drone_id `+where+` ORDER BY h.recorded_at DESC LIMIT $1 OFFSET $2`, args...)
	if err != nil {
		return page, fmt.Errorf("查询健康记录: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item TelemetryRecord
		if err := rows.Scan(&item.ID, &item.DroneID, &item.DroneName, &item.BatteryLevel, &item.MotorTemperature, &item.NetworkLatencyMS, &item.LocalizationError, &item.JointLoad, &item.Remark, &item.RecordTime); err != nil {
			return page, fmt.Errorf("读取健康记录: %w", err)
		}
		page.Records = append(page.Records, item)
	}
	countQuery := `SELECT count(*) FROM console_telemetry_records`
	countArgs := []any{}
	if droneID != "" {
		countQuery += ` WHERE drone_id=$1`
		countArgs = append(countArgs, droneID)
	}
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("统计健康记录: %w", err)
	}
	return page, rows.Err()
}

func (s *Store) CreateTelemetry(ctx context.Context, item TelemetryRecord) (TelemetryRecord, error) {
	item.ID = uuid.NewString()
	if item.RecordTime.IsZero() {
		if err := s.pool.QueryRow(ctx, `INSERT INTO console_telemetry_records(id,drone_id,battery_level,motor_temperature,network_latency_ms,localization_error,joint_load,remark)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING recorded_at`, item.ID, item.DroneID, item.BatteryLevel, item.MotorTemperature, item.NetworkLatencyMS, item.LocalizationError, item.JointLoad, item.Remark).Scan(&item.RecordTime); err != nil {
			return TelemetryRecord{}, wrap("新增健康记录", err)
		}
	} else {
		_, err := s.pool.Exec(ctx, `INSERT INTO console_telemetry_records(id,drone_id,battery_level,motor_temperature,network_latency_ms,localization_error,joint_load,remark,recorded_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, item.ID, item.DroneID, item.BatteryLevel, item.MotorTemperature, item.NetworkLatencyMS, item.LocalizationError, item.JointLoad, item.Remark, item.RecordTime)
		if err != nil {
			return TelemetryRecord{}, wrap("新增健康记录", err)
		}
	}
	return item, nil
}

func (s *Store) LogPage(ctx context.Context, current, size int) (Page[Log], error) {
	current, size, offset := pageBounds(current, size)
	page := Page[Log]{Records: make([]Log, 0), Current: current, Size: size}
	rows, err := s.pool.Query(ctx, `SELECT id,username,operation,method,ip,created_at FROM console_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, size, offset)
	if err != nil {
		return page, fmt.Errorf("查询日志: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Log
		if err := rows.Scan(&item.ID, &item.Username, &item.Operation, &item.Method, &item.IP, &item.CreateTime); err != nil {
			return page, fmt.Errorf("读取日志: %w", err)
		}
		page.Records = append(page.Records, item)
	}
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM console_logs`).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("统计日志: %w", err)
	}
	return page, rows.Err()
}
