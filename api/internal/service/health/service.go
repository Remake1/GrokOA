package health

import "context"

type statusRepository interface {
	IsReady(ctx context.Context) error
}

type Service struct {
	repo statusRepository
}

func NewService(repo statusRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Live(ctx context.Context) error {
	_ = ctx

	return nil
}

func (s *Service) Ready(ctx context.Context) error {
	return s.repo.IsReady(ctx)
}
