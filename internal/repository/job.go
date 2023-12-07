package repository

import (
	"context"
	"time"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository/dao"
)

type JobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	Stop(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
}

type PreemptCronJobRepository struct {
	dao dao.JobDAO
}

func (p *PreemptCronJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	return domain.Job{
		Cfg:      j.Cfg,
		Id:       j.Id,
		Name:     j.Name,
		Executor: j.Executor,
	}, nil
}

func (p *PreemptCronJobRepository) Release(ctx context.Context, id int64) error {
	return p.dao.Release(ctx, id)
}

func (p *PreemptCronJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return p.dao.UpdateUtime(ctx, id)
}

func (p *PreemptCronJobRepository) Stop(ctx context.Context, id int64) error {
	return p.dao.Stop(ctx, id)
}

func (p *PreemptCronJobRepository) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return p.dao.UpdateNextTime(ctx, id, next)
}
