package service

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"webooktrial/internal/domain"
	events "webooktrial/internal/events/article"
	"webooktrial/internal/repository/article"
	"webooktrial/pkg/logger"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx *gin.Context, art domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	// ListPub 只会取 start 七天内的数据
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
	GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error)
}

type ArticleCoreService struct {
	repo article.ArticleRepository

	// V1 依靠两个不同的 repository 来解决这种跨表，或者跨库的问题
	author   article.ArticleAuthorRepository
	reader   article.ArticleReaderRepository
	l        logger.LoggerV1
	producer events.Producer
	ch       chan readInfo
}

func (a *ArticleCoreService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	//TODO implement me
	panic("implement me")
}

type readInfo struct {
	uid int64
	aid int64
}

func (a *ArticleCoreService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.List(ctx, uid, offset, limit)
}

func (a *ArticleCoreService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetByID(ctx, id)

}

func (a *ArticleCoreService) GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	// 另一个选项，在这里组装 Author，调用 UserService
	art, err := a.repo.GetPublishedById(ctx, id)
	if err == nil {
		go func() {
			// 生产者也可以通过改批量来提高性能
			er := a.producer.ProduceReadEvent(
				// 即便你的消费者要用 art 的里面的数据，
				// 让它去查询，你不要在 event 里面带
				ctx, events.ReadEvent{
					Uid: uid,
					Aid: id,
				})
			if er != nil {
				a.l.Error("发送读者阅读事件失败",
					logger.Int64("Uid", uid),
					logger.Int64("Aid", id))
			}
		}()
	}
	go func() {
		// 改批量的做法
		a.ch <- readInfo{
			aid: id,
			uid: uid,
		}

	}()
	return art, err
}

func NewArticleService(repo article.ArticleRepository,
	l logger.LoggerV1,
	producer events.Producer) ArticleService {
	return &ArticleCoreService{
		repo:     repo,
		l:        l,
		producer: producer,
		//ch: make(chan readInfo, 10),
	}
}

func NewArticleServiceV1(author article.ArticleAuthorRepository,
	reader article.ArticleReaderRepository, l logger.LoggerV1) ArticleService {
	return &ArticleCoreService{
		author: author,
		reader: reader,
		l:      l,
	}
}

func NewArticleServiceV2(repo article.ArticleRepository,
	l logger.LoggerV1,
	producer events.Producer) ArticleService {
	ch := make(chan readInfo, 10)
	go func() {
		for {
			uids := make([]int64, 0, 10)
			aids := make([]int64, 0, 10)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			for i := 0; i < 10; i++ {
				select {
				case info, ok := <-ch:
					if !ok {
						cancel()
						return
					}
					uids = append(uids, info.uid)
					aids = append(aids, info.aid)
				case <-ctx.Done():
					break
				}
			}
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			producer.ProduceReadEventV1(ctx, events.ReadEventV1{
				Uids: uids,
				Aids: aids,
			})
			cancel()
		}
	}()
	return &ArticleCoreService{
		repo:     repo,
		producer: producer,
		l:        l,
		ch:       ch,
	}
}

func (a *ArticleCoreService) Withdraw(ctx *gin.Context, art domain.Article) error {
	return a.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (a *ArticleCoreService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *ArticleCoreService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = a.author.Update(ctx, art)
	} else {
		id, err = a.author.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second * time.Duration(i))
		id, err = a.reader.Save(ctx, art)
		if err == nil {
			break
		}
		a.l.Error("部分失败，保存到线上库失败",
			logger.Int64("art_id", art.Id),
			logger.Error(err))
	}
	if err != nil {
		a.l.Error("部分失败，重试彻底失败",
			logger.Int64("art_id", art.Id),
			logger.Error(err))
		// 接入你的告警系统，手工处理一下
		// 走异步，我直接保存到本地文件
		// 走 Canal
		// 打 MQ
	}
	return id, err
}

func (a *ArticleCoreService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func (a *ArticleCoreService) update(ctx context.Context, art domain.Article) error {
	// 只要你不更新 author_id
	// 但是性能比较差
	//artInDB := a.repo.FindById(ctx, art.Id)
	//if art.Author.Id != artInDB.Author.Id {
	//	return errors.New("更新别人的数据")
	//}
	return a.repo.Update(ctx, art)
}
