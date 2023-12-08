package dao

import (
	"gorm.io/gorm"

	"webooktrial/internal/repository/dao/article"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&article.Article{},
		&SMSMsg{},
		&article.PublishedArticle{},
		&Job{})
}
