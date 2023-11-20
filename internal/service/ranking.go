package service

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"

	"webooktrial/internal/domain"
)

type RankingService interface {
	TopN(ctx context.Context) error
	//TopN(ctx context.Context, n int64) error
	//TopN(ctx context.Context, n int64) ([]domain.Article, error)
}

type BatchRankingService struct {
	artSvc    ArticleService
	intrSvc   InteractiveService
	batchSize int
	n         int
	// scoreFunc 不能返回负数
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc InteractiveService) *BatchRankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			return float64(likeCnt-1) / math.Pow(float64(likeCnt+2), 1.5)
		},
	}
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	// 在这里，存起来
	log.Println(arts)
	return nil
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	// 只取七天以内的数据
	now := time.Now()
	// 先拿一批数据
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	// 这里可以用非并发安全
	topN := queue.NewConcurrentPriorityQueue[Score](b.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})
	for {
		arts, err := b.artSvc.ListPub(ctx, now, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts,
			func(idx int, src domain.Article) int64 {
				return src.Id
			})
		intrs, err := b.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		// 合并计算 score
		// 排序
		for _, art := range arts {
			intr := intrs[art.Id]
			//if !ok {
			//	// 都没有，肯定不可能是热榜
			//	continue
			//}
			score := b.scoreFunc(art.Utime, intr.LikeCnt)
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})
			if errors.Is(err, queue.ErrOutOfCapacity) {
				val, _ := topN.Dequeue()
				if val.score < score {
					err = topN.Enqueue(Score{art: art, score: score})
				} else {
					_ = topN.Enqueue(val)
				}
			}
		}
		// 判断是否还有下一批需要处理
		if len(arts) < b.batchSize {
			break
		}
		// 更新 offset
		offset = offset + len(arts)
	}
	// 最后得出结果
	res := make([]domain.Article, b.n)
	for i := b.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			// 说明已经取完
			break
		}
		res[i] = val.art
	}
	return res, nil
}
