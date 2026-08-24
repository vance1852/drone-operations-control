package console

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *Store) CapabilityPage(ctx context.Context, current, size int, keyword string) (Page[CapabilityItem], error) {
	current, size, offset := pageBounds(current, size)
	keyword = "%" + strings.TrimSpace(keyword) + "%"
	page := Page[CapabilityItem]{Records: make([]CapabilityItem, 0), Current: current, Size: size}
	rows, err := s.pool.Query(ctx, `SELECT id,name,description,price,duration,status FROM console_capabilities
		WHERE deleted_at IS NULL AND name ILIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, keyword, size, offset)
	if err != nil {
		return page, fmt.Errorf("查询服务列表: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item CapabilityItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Duration, &item.Status); err != nil {
			return page, fmt.Errorf("读取服务信息: %w", err)
		}
		page.Records = append(page.Records, item)
	}
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM console_capabilities WHERE deleted_at IS NULL AND name ILIKE $1`, keyword).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("统计服务数量: %w", err)
	}
	return page, rows.Err()
}

func (s *Store) CapabilityList(ctx context.Context) ([]CapabilityItem, error) {
	page, err := s.CapabilityPage(ctx, 1, 100, "")
	return page.Records, err
}

func (s *Store) CreateCapability(ctx context.Context, item CapabilityItem) (CapabilityItem, error) {
	item.ID = uuid.NewString()
	_, err := s.pool.Exec(ctx, `INSERT INTO console_capabilities(id,name,description,price,duration,status) VALUES($1,$2,$3,$4,$5,$6)`, item.ID, strings.TrimSpace(item.Name), item.Description, item.Price, item.Duration, item.Status)
	return item, wrap("新增服务", err)
}

func (s *Store) UpdateCapability(ctx context.Context, item CapabilityItem) error {
	tag, err := s.pool.Exec(ctx, `UPDATE console_capabilities SET name=$1,description=$2,price=$3,duration=$4,status=$5,updated_at=now() WHERE id=$6 AND deleted_at IS NULL`, strings.TrimSpace(item.Name), item.Description, item.Price, item.Duration, item.Status, item.ID)
	if err == nil && tag.RowsAffected() != 1 {
		return fmt.Errorf("更新服务: 记录不存在")
	}
	return wrap("更新服务", err)
}

func (s *Store) DeleteCapability(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE console_capabilities SET deleted_at=now(),updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err == nil && tag.RowsAffected() != 1 {
		return fmt.Errorf("删除服务: 记录不存在")
	}
	return wrap("删除服务", err)
}
