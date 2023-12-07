package article

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewGormArticleDao(db *gorm.DB) ArticleDao {
	return &GormArticleDao{
		db: db,
	}
}

type GormArticleDao struct {
	db *gorm.DB
}

func (g *GormArticleDao) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	var res []Article
	err := g.db.WithContext(ctx).
		Where("utime < ?", start.UnixMilli()).
		Order("utime DESC").Offset(offset).Limit(limit).Error
	return res, err
}

func (g *GormArticleDao) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	// SELECT * FROM XXX WHERE XX order by aaa
	// 在设计 order by 语句的时候，要注意让 order by 中的数据命中索引
	// SQL 优化的案例：早期的时候，
	// 我们的 order by 没有命中索引的，内存排序非常慢
	// 你的工作就是优化了这个查询，加进去了索引
	// author_id => author_id, utime 的联合索引
	var arts []Article
	err := g.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", author).
		Offset(offset).
		Limit(limit).
		// 升序排序。 utime ASC
		// 混合排序
		// ctime ASC, utime desc
		Order("utime DESC").
		//Order(clause.OrderBy{Columns: []clause.OrderByColumn{
		//	{Column: clause.Column{Name: "utime"}, Desc: true},
		//	{Column: clause.Column{Name: "ctime"}, Desc: false},
		//}}).
		Find(arts).Error
	return arts, err
}

func (g *GormArticleDao) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := g.db.WithContext(ctx).Model(&Article{}).
		Where("id = ?", id).
		First(&art).Error
	return art, err
}

func (g *GormArticleDao) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pub PublishedArticle
	err := g.db.WithContext(ctx).Model(&PublishedArticle{}).
		Where("id = ?", id).
		First(&pub).Error
	return pub, err
}

func (g *GormArticleDao) SyncStatus(ctx context.Context, author, id int64, status uint8) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", id, author).
			Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			// 要么 ID 是错的，要么作者不对
			// 后者情况下，你就要小心，可能有人在搞你的系统
			// 没必要再用 ID 搜索数据库来区分这两种情况
			// 用 prometheus 打点，只要频繁出现，你就告警，然后手工介入排查
			return ErrPossibleIncorrectAuthor
		}
		res = tx.Model(&PublishedArticle{}).
			Where("id = ? AND author_id = ?", id, author).
			Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		return nil
	})
}

func (g *GormArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	// 先操作制作库（此时应该是表），后操作线上库（此时应该是表）
	tx := g.db.WithContext(ctx).Begin()
	now := time.Now().UnixMilli()
	defer tx.Rollback()
	txDao := NewGormArticleDao(tx)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = txDao.Insert(ctx, art)
	} else {
		err = txDao.UpdateById(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	publishArt := PublishedArticle(art)
	publishArt.Utime = now
	publishArt.Ctime = now
	err = g.db.Clauses(clause.OnConflict{
		// SQL 2003 标准
		// INSERT AAAA ON CONFLICT(BBB) DO NOTHING
		// INSERT AAAA ON CONFLICT(BBB) DO UPDATES CCC WHERE DDD

		// 哪些列冲突
		//Columns: []clause.Column{clause.Column{Name: "id"}},
		// 意思是数据冲突，啥也不干
		// DoNothing:
		// 数据冲突了，并且符合 WHERE 条件的就会执行 DO UPDATES
		// Where:

		// MySQL 只需要关心这里
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		}),
	}).Create(&art).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, tx.Error
}

func (g *GormArticleDao) SyncClosure(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		now := time.Now().UnixMilli()
		txDao := NewGormArticleDao(tx)
		if id == 0 {
			id, err = txDao.Insert(ctx, art)
		} else {
			err = txDao.UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		publishArt := art
		publishArt.Utime = now
		publishArt.Ctime = now
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   art.Title,
				"content": art.Content,
				"status":  art.Status,
				"utime":   now,
			}),
		}).Create(&publishArt).Error
	})
	return id, err
}

func (g *GormArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := g.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (g *GormArticleDao) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	// 依赖 gorm 忽略零值的特性，会用主键进行更新
	// 可读性很差
	res := g.db.WithContext(ctx).Model(&art).
		Where("id=? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   art.Utime,
		})
	// 你要不要检查真的更新了没？
	// res.RowsAffected // 更新行数
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		//dangerousDBOp.Count(1)
		// 补充一点日志
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d",
			art.Id, art.AuthorId)
	}
	return nil
}
