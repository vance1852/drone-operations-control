package service

import (
	"context"
	"errors"
	"testing"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

type operatorTransactionRepository struct {
	repository.Repository
	operators map[string]domain.FleetOperator
	auditErr  error
}

func (r *operatorTransactionRepository) InTx(ctx context.Context, fn func(repository.Repository) error) error {
	pending := make(map[string]domain.FleetOperator, len(r.operators))
	for id, operator := range r.operators {
		pending[id] = operator
	}
	tx := &operatorTransactionRepository{operators: pending, auditErr: r.auditErr}
	if err := fn(tx); err != nil {
		return err
	}
	r.operators = pending
	return nil
}

func (r *operatorTransactionRepository) CreateFleetOperatorOutsideTransaction(_ context.Context, operator domain.FleetOperator) error {
	r.operators[operator.ID] = operator
	return nil
}

func (r *operatorTransactionRepository) CreateFleetOperator(_ context.Context, operator domain.FleetOperator) error {
	r.operators[operator.ID] = operator
	return nil
}

func (r *operatorTransactionRepository) WriteAudit(context.Context, repository.AuditInput) error {
	return r.auditErr
}

func TestFleetOperatorRegistrationRollsBackWhenAuditFails(t *testing.T) {
	repo := &operatorTransactionRepository{operators: map[string]domain.FleetOperator{}, auditErr: errors.New("audit rejected")}
	_, err := New(repo).RegisterFleetOperator(t.Context(), RequestMeta{RequestID: "operator-create"}, "Rollback FleetOperator", domain.RoleSafetySupervisor)
	if err == nil {
		t.Fatal("operator registration succeeded despite audit failure")
	}
	if len(repo.operators) != 0 {
		t.Fatalf("persisted operators=%d", len(repo.operators))
	}
}
