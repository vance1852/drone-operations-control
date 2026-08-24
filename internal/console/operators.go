package console

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *Store) OperatorPage(ctx context.Context, current, size int, keyword string) (Page[Operator], error) {
	current, size, offset := pageBounds(current, size)
	keyword = "%" + strings.TrimSpace(keyword) + "%"
	page := Page[Operator]{Records: make([]Operator, 0), Current: current, Size: size}
	rows, err := s.pool.Query(ctx, `SELECT id,name,gender,phone,skills,status,created_at FROM console_operators
		WHERE deleted_at IS NULL AND (name ILIKE $1 OR phone ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`, keyword, size, offset)
	if err != nil {
		return page, fmt.Errorf("查询飞行调度员列表: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Operator
		if err := rows.Scan(&item.ID, &item.Name, &item.Gender, &item.Phone, &item.Skills, &item.Status, &item.CreateTime); err != nil {
			return page, fmt.Errorf("读取飞行调度员信息: %w", err)
		}
		page.Records = append(page.Records, item)
	}
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM console_operators WHERE deleted_at IS NULL AND (name ILIKE $1 OR phone ILIKE $1)`, keyword).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("统计飞行调度员数量: %w", err)
	}
	return page, rows.Err()
}

func (s *Store) OperatorList(ctx context.Context) ([]Operator, error) {
	page, err := s.OperatorPage(ctx, 1, 100, "")
	return page.Records, err
}

func (s *Store) CreateOperator(ctx context.Context, item Operator) (Operator, error) {
	item.ID = uuid.NewString()
	err := s.pool.QueryRow(ctx, `INSERT INTO console_operators(id,name,gender,phone,skills,status) VALUES($1,$2,$3,$4,$5,$6) RETURNING created_at`, item.ID, strings.TrimSpace(item.Name), item.Gender, item.Phone, item.Skills, item.Status).Scan(&item.CreateTime)
	return item, wrap("新增飞行调度员", err)
}

func (s *Store) UpdateOperator(ctx context.Context, item Operator) error {
	tag, err := s.pool.Exec(ctx, `UPDATE console_operators SET name=$1,gender=$2,phone=$3,skills=$4,status=$5,updated_at=now() WHERE id=$6 AND deleted_at IS NULL`, strings.TrimSpace(item.Name), item.Gender, item.Phone, item.Skills, item.Status, item.ID)
	if err == nil && tag.RowsAffected() != 1 {
		return fmt.Errorf("更新飞行调度员: 记录不存在")
	}
	return wrap("更新飞行调度员", err)
}

func (s *Store) DeleteOperator(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE console_operators SET deleted_at=now(),updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err == nil && tag.RowsAffected() != 1 {
		return fmt.Errorf("删除飞行调度员: 记录不存在")
	}
	return wrap("删除飞行调度员", err)
}
