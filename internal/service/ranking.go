package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"

	intrv1 "webooktrial/api/proto/gen/intr/v1"
	"webooktrial/internal/domain"
	"webooktrial/internal/repository"
)

type RankingService interface {
	TopN(ctx context.Context) error
	//TopN(ctx context.Context, n int64) error
	//TopN(ctx context.Context, n int64) ([]domain.Article, error)
}

type BatchRankingService struct {
	artSvc    ArticleService
	intrSvc   intrv1.InteractiveServiceClient
	repo      repository.RankingRepository
	batchSize int
	n         int
	// scoreFunc 不能返回负数
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService,
	intrSvc intrv1.InteractiveServiceClient,
	repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		batchSize: 100,
		n:         100,
		repo:      repo,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(sec, 1.5)
		},
	}
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	// 在这里，存起来

	return b.repo.ReplaceTopN(ctx, arts)
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
		intrs, err := b.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz: "article",
			Ids: ids,
		})
		if err != nil {
			return nil, err
		}

		if len(intrs.Intrs) == 0 {
			return nil, errors.New("没有数据")
		}
		// 合并计算 score
		// 排序
		for _, art := range arts {
			intr := intrs.Intrs[art.Id]
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
					_ = topN.Enqueue(Score{art: art, score: score})
				} else {
					_ = topN.Enqueue(val)
				}
			}
		}
		// 判断是否还有下一批需要处理
		if len(arts) < b.batchSize || now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			// 当前批次为取满或者已经取到一周之前的数据，说明可以中断计算热榜
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
