package health

import "context"

type StatusRepository struct{}

func NewStatusRepository() *StatusRepository {
	return &StatusRepository{}
}

func (r *StatusRepository) IsReady(ctx context.Context) error {
	// This method is where external dependency checks should be done.
	_ = ctx

	return nil
}
