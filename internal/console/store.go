package console

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"drone-operations-control/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Store struct {
	pool *db.Pool
}

func NewStore(pool *db.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Login(ctx context.Context, username, password string) (User, string, time.Time, error) {
	hash := sha256.Sum256([]byte(password))
	var user User
	var passwordHash string
	err := s.pool.QueryRow(ctx, `SELECT id,username,password_hash,real_name,phone,role,status FROM console_users WHERE username=$1`, strings.TrimSpace(username)).Scan(
		&user.ID, &user.Username, &passwordHash, &user.RealName, &user.Phone, &user.Role, &user.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) || passwordHash != hex.EncodeToString(hash[:]) {
		return User{}, "", time.Time{}, fmt.Errorf("用户名或密码错误")
	}
	if err != nil {
		return User{}, "", time.Time{}, fmt.Errorf("查询用户: %w", err)
	}
	if user.Status != 1 {
		return User{}, "", time.Time{}, fmt.Errorf("账号已被禁用")
	}
	token := uuid.NewString()
	expiresAt := time.Now().UTC().Add(12 * time.Hour)
	if _, err := s.pool.Exec(ctx, `INSERT INTO console_sessions(token,user_id,expires_at) VALUES($1,$2,$3)`, token, user.ID, expiresAt); err != nil {
		return User{}, "", time.Time{}, fmt.Errorf("创建会话: %w", err)
	}
	return user, token, expiresAt, nil
}

func (s *Store) SessionUser(ctx context.Context, token string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `SELECT u.id,u.username,u.real_name,u.phone,u.role,u.status
		FROM console_sessions s JOIN console_users u ON u.id=s.user_id
		WHERE s.token=$1 AND s.revoked_at IS NULL AND s.expires_at>now() AND u.status=1`, strings.TrimSpace(token)).Scan(
		&user.ID, &user.Username, &user.RealName, &user.Phone, &user.Role, &user.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, fmt.Errorf("会话无效或已过期")
	}
	return user, wrap("查询会话", err)
}

func (s *Store) RevokeSession(ctx context.Context, token string) error {
	result, err := s.pool.Exec(ctx, `UPDATE console_sessions SET revoked_at=now()
		WHERE token=$1 AND revoked_at IS NULL AND expires_at>now()`, strings.TrimSpace(token))
	if err != nil {
		return fmt.Errorf("撤销会话: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("会话无效或已过期")
	}
	return nil
}

func (s *Store) UserInfo(ctx context.Context, username string) (User, error) {
	if strings.TrimSpace(username) == "" {
		username = "admin"
	}
	var user User
	err := s.pool.QueryRow(ctx, `SELECT id,username,real_name,phone,role,status FROM console_users WHERE username=$1`, username).Scan(
		&user.ID, &user.Username, &user.RealName, &user.Phone, &user.Role, &user.Status,
	)
	return user, wrap("查询用户", err)
}

func (s *Store) Dashboard(ctx context.Context) (DashboardStats, error) {
	var stats DashboardStats
	err := s.pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM console_drones WHERE deleted_at IS NULL AND status=1),
		(SELECT count(*) FROM console_operators WHERE deleted_at IS NULL AND status=1),
		(SELECT count(*) FROM console_commands WHERE status=0),
		(SELECT count(*) FROM console_commands WHERE status=2)`).Scan(
		&stats.DroneCount, &stats.OperatorCount, &stats.PendingCommands, &stats.CompletedCommands,
	)
	return stats, wrap("统计首页数据", err)
}

func pageBounds(current, size int) (int, int, int) {
	if current < 1 {
		current = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	return current, size, (current - 1) * size
}

func wrap(action string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: 记录不存在", action)
	}
	return fmt.Errorf("%s: %w", action, err)
}

func (s *Store) WriteLog(ctx context.Context, operation, method, ip string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO console_logs(id,username,operation,method,ip) VALUES($1,'admin',$2,$3,$4)`, uuid.NewString(), operation, method, ip)
	return wrap("记录操作日志", err)
}

func parseDate(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return strings.TrimSpace(*value)
}

func formatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}
