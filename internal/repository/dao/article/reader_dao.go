package article

import (
	"context"

	"gorm.io/gorm"
)

type ReaderDao interface {
	Upsert(ctx context.Context, art Article) error
	UpsertV2(ctx context.Context, art PublishedArticle) error
}

func NewReaderDao(db *gorm.DB) ReaderDao {
	panic("implement me")
}
