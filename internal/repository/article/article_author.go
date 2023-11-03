package article

import (
	"context"

	"webooktrial/internal/domain"
)

// ArticleAuthorRepository 演示在 service 层面上分流
//
//go:generate mockgen -source=./article_author.go -package=repomocks -destination=mocks/article_author.mock.go ArticleAuthorRepository
type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}
