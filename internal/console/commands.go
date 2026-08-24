package console

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *Store) CommandPage(ctx context.Context, current, size int) (Page[Command], error) {
	current, size, offset := pageBounds(current, size)
	page := Page[Command]{Records: make([]Command, 0), Current: current, Size: size}
	rows, err := s.pool.Query(ctx, `SELECT o.id,o.command_no,o.drone_id,e.name,o.capability_id,svc.name,o.operator_id,coalesce(w.name,''),o.appointment_time,o.status,o.remark,o.version
		FROM console_commands o JOIN console_drones e ON e.id=o.drone_id JOIN console_capabilities svc ON svc.id=o.capability_id
		LEFT JOIN console_operators w ON w.id=o.operator_id ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`, size, offset)
	if err != nil {
		return page, fmt.Errorf("查询订单列表: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Command
		if err := rows.Scan(&item.ID, &item.CommandNo, &item.DroneID, &item.DroneName, &item.CapabilityID, &item.CapabilityName, &item.OperatorID, &item.OperatorName, &item.AppointmentTime, &item.Status, &item.Remark, &item.Version); err != nil {
			return page, fmt.Errorf("读取订单信息: %w", err)
		}
		page.Records = append(page.Records, item)
	}
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM console_commands`).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("统计订单数量: %w", err)
	}
	return page, rows.Err()
}

func (s *Store) CreateCommand(ctx context.Context, item Command) (Command, error) {
	item.ID = uuid.NewString()
	item.CommandNo = fmt.Sprintf("ROBOT%s", time.Now().UTC().Format("20060102150405.000000"))
	item.Status = 0
	item.Version = 1
	_, err := s.pool.Exec(ctx, `INSERT INTO console_commands(id,command_no,drone_id,capability_id,operator_id,appointment_time,status,remark,version)
		VALUES($1,$2,$3,$4,$5,$6,0,$7,1)`, item.ID, item.CommandNo, item.DroneID, item.CapabilityID, item.OperatorID, item.AppointmentTime, item.Remark)
	return item, wrap("创建订单", err)
}

func (s *Store) UpdateCommandStatus(ctx context.Context, id string, next int) error {
	allowed := map[int][]int{1: {0}, 2: {1}, 3: {0}}
	from, ok := allowed[next]
	if !ok {
		return fmt.Errorf("不支持的订单状态")
	}
	tag, err := s.pool.Exec(ctx, `UPDATE console_commands SET status=$1,version=version+1,updated_at=now() WHERE id=$2 AND status=ANY($3)`, next, id, from)
	if err == nil && tag.RowsAffected() != 1 {
		return fmt.Errorf("订单状态已变化，请刷新后重试")
	}
	return wrap("更新订单状态", err)
}
