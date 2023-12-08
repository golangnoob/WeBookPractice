package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"webooktrial/interactive/repository"
	"webooktrial/pkg/logger"
	"webooktrial/pkg/saramax"
)

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewInteractiveReadEventConsumer(
	client sarama.Client,
	repo repository.InteractiveRepository,
	l logger.LoggerV1) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (r *InteractiveReadEventConsumer) Start() error {
	// 在这里，上报 prometheus 就可以
	cg, err := sarama.NewConsumerGroupFromClient("interactive",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewHandler[ReadEvent](r.l, r.Consume))
		if er != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (r *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.IncrReadCnt(ctx, "article", t.Aid)
}
