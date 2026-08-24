package service

import (
	"context"
	"time"

	"drone-operations-control/internal/repository"
)

type transactionPolicyRunner interface {
	RunWithPolicy(context.Context, repository.TransactionPolicy, func(repository.Repository) error) error
}

func (s *Service) runCreateMissionTransaction(ctx context.Context, operation func(repository.Repository) error) error {
	runner, ok := s.repo.(transactionPolicyRunner)
	if !ok {
		return s.repo.InTx(ctx, operation)
	}
	return runner.RunWithPolicy(ctx, repository.TransactionPolicy{
		Timeout:       5 * time.Second,
		Serializable:  true,
		RetryAttempts: 3,
	}, operation)
}
