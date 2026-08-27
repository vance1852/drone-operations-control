package console

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Store) DronePage(ctx context.Context, current, size int, keyword string) (Page[Drone], error) {
	current, size, offset := pageBounds(current, size)
	keyword = "%" + strings.TrimSpace(keyword) + "%"
	page := Page[Drone]{Records: make([]Drone, 0), Current: current, Size: size}
	rows, err := s.pool.Query(ctx, `SELECT id,name,model_class,commissioned_on,serial_number,endpoint,home_zone,owner_name,owner_phone,telemetry_status,status
		FROM console_drones WHERE deleted_at IS NULL AND (name ILIKE $1 OR serial_number ILIKE $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, keyword, size, offset)
	if err != nil {
		return page, fmt.Errorf("查询无人机设备列表: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Drone
		var date *time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.ModelClass, &date, &item.SerialNumber, &item.Endpoint, &item.HomeZone, &item.OwnerName, &item.OwnerPhone, &item.TelemetryStatus, &item.Status); err != nil {
			return page, fmt.Errorf("读取无人机设备信息: %w", err)
		}
		item.CommissionedOn = formatDate(date)
		page.Records = append(page.Records, item)
	}
	if err := rows.Err(); err != nil {
		return page, fmt.Errorf("读取无人机设备列表: %w", err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM console_drones WHERE deleted_at IS NULL AND (name ILIKE $1 OR serial_number ILIKE $1)`, keyword).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("统计无人机设备数量: %w", err)
	}
	return page, nil
}

func (s *Store) DroneList(ctx context.Context) ([]Drone, error) {
	page, err := s.DronePage(ctx, 1, 100, "")
	return page.Records, err
}

func (s *Store) DroneByID(ctx context.Context, id string) (Drone, error) {
	var item Drone
	var date *time.Time
	err := s.pool.QueryRow(ctx, `SELECT id,name,model_class,commissioned_on,serial_number,endpoint,home_zone,owner_name,owner_phone,telemetry_status,status
		FROM console_drones WHERE id=$1 AND deleted_at IS NULL`, id).Scan(
		&item.ID, &item.Name, &item.ModelClass, &date, &item.SerialNumber, &item.Endpoint, &item.HomeZone, &item.OwnerName, &item.OwnerPhone, &item.TelemetryStatus, &item.Status,
	)
	item.CommissionedOn = formatDate(date)
	return item, wrap("查询无人机设备", err)
}

func (s *Store) CreateDrone(ctx context.Context, item Drone) (Drone, error) {
	item.ID = uuid.NewString()
	_, err := s.pool.Exec(ctx, `INSERT INTO console_drones(id,name,model_class,commissioned_on,serial_number,endpoint,home_zone,owner_name,owner_phone,telemetry_status,status)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, item.ID, strings.TrimSpace(item.Name), item.ModelClass, parseDate(item.CommissionedOn), item.SerialNumber, item.Endpoint, item.HomeZone, item.OwnerName, item.OwnerPhone, item.TelemetryStatus, item.Status)
	return item, wrap("新增无人机设备", err)
}

func (s *Store) UpdateDrone(ctx context.Context, item Drone) error {
	tag, err := s.pool.Exec(ctx, `UPDATE console_drones SET name=$1,model_class=$2,commissioned_on=$3,serial_number=$4,endpoint=$5,home_zone=$6,owner_name=$7,owner_phone=$8,telemetry_status=$9,status=$10,updated_at=now()
		WHERE id=$11 AND deleted_at IS NULL`, strings.TrimSpace(item.Name), item.ModelClass, parseDate(item.CommissionedOn), item.SerialNumber, item.Endpoint, item.HomeZone, item.OwnerName, item.OwnerPhone, item.TelemetryStatus, item.Status, item.ID)
	if err == nil && tag.RowsAffected() != 1 {
		return fmt.Errorf("更新无人机设备: 记录不存在")
	}
	return wrap("更新无人机设备", err)
}

func (s *Store) DeleteDrone(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE console_drones SET deleted_at=now(),updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err == nil && tag.RowsAffected() != 1 {
		return fmt.Errorf("删除无人机设备: 记录不存在")
	}
	return wrap("删除无人机设备", err)
}
