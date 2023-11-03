package article

import (
	"context"

	"webooktrial/internal/domain"
)

//go:generate mockgen -source=./article_reader.go -package=repomocks -destination=mocks/article_reader.mock.go ArticleReaderRepository
type ArticleReaderRepository interface {
	// Save 有就更新，没有就新建，即 upsert 的语义
	Save(ctx context.Context, art domain.Article) (int64, error)
}
