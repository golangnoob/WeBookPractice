package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"webooktrial/interactive/repository"
	"webooktrial/pkg/logger"
	"webooktrial/pkg/saramax"
)

type InteractiveReadEventBatchConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.LoggerV1) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{client: client, repo: repo, l: l}
}

func (r *InteractiveReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewBatchHandler[ReadEvent](r.l, r.Consume))
		if er != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (r *InteractiveReadEventBatchConsumer) Consume(msgs []*sarama.ConsumerMessage, ts []ReadEvent) error {
	aids := make([]int64, 0, len(ts))
	bizs := make([]string, 0, len(ts))
	for _, evt := range ts {
		aids = append(aids, evt.Aid)
		bizs = append(bizs, "article")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := r.repo.BatchIncrReadCnt(ctx, bizs, aids)
	if err != nil {
		r.l.Error("批量增加阅读计数失败",
			logger.Field{Key: "ids", Value: aids},
			logger.Error(err))
	}
	return nil
}
