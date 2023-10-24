package service

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"webooktrial/internal/domain"
	"webooktrial/internal/repository/article"
	"webooktrial/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx *gin.Context, art domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

type ArticleCoreService struct {
	repo article.ArticleRepository

	// V1 依靠两个不同的 repository 来解决这种跨表，或者跨库的问题
	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository
	l      logger.LoggerV1
}

func (a *ArticleCoreService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.List(ctx, uid, offset, limit)
}

func (a *ArticleCoreService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetByID(ctx, id)

}

func (a *ArticleCoreService) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	// 另一个选项，在这里组装 Author，调用 UserService
	return a.repo.GetPublishedById(ctx, id)

}

func NewArticleService(repo article.ArticleRepository) ArticleService {
	return &ArticleCoreService{
		repo: repo,
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
