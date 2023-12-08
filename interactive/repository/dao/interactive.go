package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate mockgen -source=./interactive.go -package=daomocks -destination=mocks/interactive.mock.go InteractiveDAO
var ErrDataNotFound = gorm.ErrRecordNotFound

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error)
	DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, aids []int64) error
	GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func (G *GORMInteractiveDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	var res []Interactive
	err := G.db.WithContext(ctx).Where("biz = ? AND id IN ?", biz, ids).Find(&res).Error
	return res, err
}

// BatchIncrReadCnt bizs 和 aids 的长度必须相等
func (G *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, aids []int64) error {
	return G.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 让调用者保证两者是相等的
		for i := 0; i < len(bizs); i++ {
			err := G.incrReadCnt(tx, bizs[i], aids[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// IncrReadCnt 是一个插入或者更新语义
func (G *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return G.incrReadCnt(G.db.WithContext(ctx), biz, bizId)
}

func (G *GORMInteractiveDAO) incrReadCnt(tx *gorm.DB, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return tx.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"read_cnt": gorm.Expr("`read_cnt`+1"),
			"utime":    now,
		}),
	}).Create(&Interactive{
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
		Biz:     biz,
		BizId:   bizId,
	}).Error
}

func (G *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	// 一把梭
	// 同时记录点赞，以及更新点赞计数
	// 首先你需要一张表来记录，谁点给什么资源点了赞
	now := time.Now().UnixMilli()
	err := G.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先准备插入点赞记录
		// 有没有可能已经点赞过了？
		// 我要不要校验一下，这里必须是没有点赞过
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Uid:    uid,
			Ctime:  now,
			Utime:  now,
			Biz:    biz,
			BizId:  bizId,
			Status: 1,
		}).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			// MySQL 不写
			//Columns:
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("like_cnt + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
			Biz:     biz,
			BizId:   bizId,
		}).Error
	})
	return err
}

func (G *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := G.db.WithContext(ctx).
		Where("biz=? AND biz_id = ? AND uid = ? AND status = ?",
			biz, bizId, uid, 1).First(&res).Error
	return res, err
}

func (G *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	now := time.Now().UnixMilli()
	// 控制事务超时
	err := G.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 两个操作
		// 一个是软删除点赞记录
		// 一个是减点赞数量
		err := tx.Model(&UserLikeBiz{}).
			Where("biz=? AND biz_id = ? AND uid = ?", biz, bizId, uid).
			Updates(map[string]any{
				"status": 0,
				"utime":  now,
			}).Error
		if err != nil {
			return err
		}
		return G.db.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("`like_cnt`-1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
			Biz:     biz,
			BizId:   bizId,
		}).Error
	})
	return err
}

func (G *GORMInteractiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var res Interactive
	err := G.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ?", biz, bizId).
		First(&res).Error
	return res, err
}

func (G *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Utime = now
	cb.Ctime = now
	return G.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 插入收藏项目
		err := G.db.WithContext(ctx).Create(&cb).Error
		if err != nil {
			return err
		}
		// 这边就是更新数量
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt`+1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
			Biz:        cb.Biz,
			BizId:      cb.BizId,
		}).Error
	})
}

func (G *GORMInteractiveDAO) GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := G.db.WithContext(ctx).
		Where("biz=? AND biz_id = ? AND uid = ?", biz, bizId, uid).First(&res).Error
	return res, err
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}
}

// Interactive 正常来说，一张主表和与它有关联关系的表会共用一个DAO，
// 所以我们就用一个 DAO 来操作
// 假如说我要查找点赞数量前 100 的，
// SELECT * FROM
// (SELECT biz, biz_id, COUNT(*) as cnt FROM `interactives` GROUP BY biz, biz_id) ORDER BY cnt LIMIT 100
// 实时查找，性能贼差，上面这个语句，就是全表扫描，
// 高性能，我不要求准确性
// 面试标准答案：用 zset
// 但是，面试标准答案不够有特色，烂大街了
// 你可以考虑别的方案
// 1. 定时计算
// 1.1 定时计算 + 本地缓存
// 2. 优化版的 zset，定时筛选 + 实时 zset 计算
// 还要别的方案你们也可以考虑
type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 业务标识符
	// 同一个资源，在这里应该只有一行
	// 也就是说我要在 bizId 和 biz 上创建联合唯一索引
	// 1. bizId, biz。优先选择这个，因为 bizId 的区分度更高
	// 2. biz, bizId。如果有 WHERE biz = xx 这种查询条件（不带 bizId）的，就只能这种
	//
	// 联合索引的列的顺序：查询条件，区分度
	// 这个名字无所谓
	BizId int64 `gorm:"uniqueIndex:biz_type_id"`
	// 我这里biz 用的是 string，有些公司枚举使用的是 int 类型
	// 0-article
	// 1- xxx
	// 默认是 BLOB/TEXT 类型
	Biz string `gorm:"uniqueIndex:biz_type_id;type:varchar(128)"`
	// 这个是阅读计数
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

//// InteractiveV1 对写更友好
//// Interactive 对读更加友好
//type InteractiveV1 struct {
//	Id    int64 `gorm:"primaryKey,autoIncrement"`
//	BizId int64
//	Biz   string
//	// 这个是阅读计数
//	Cnt int64
//	// 阅读数/点赞数/收藏数
//	CntType string
//	Ctime   int64
//	Utime   int64
//}

// UserLikeBiz 命名无能，用户点赞的某个东西
type UserLikeBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`

	// 我在前端展示的时候，
	// WHERE uid = ? AND biz_id = ? AND biz = ?
	// 来判定你有没有点赞
	// 这里，联合顺序应该是什么？

	// 要分场景
	// 1. 如果你的场景是，用户要看看自己点赞过那些，那么 Uid 在前
	// WHERE uid =?
	// 2. 如果你的场景是，我的点赞数量，需要通过这里来比较/纠正
	// biz_id 和 biz 在前
	// select count(*) where biz = ? and biz_id = ?
	Biz   string `gorm:"uniqueIndex:biz_type_id_uid;type:varchar(128)"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`

	// 谁的操作
	Uid int64 `gorm:"uniqueIndex:biz_type_id_uid"`

	Ctime int64
	Utime int64
	// 如果这样设计，那么，取消点赞的时候，怎么办？
	// 我删了这个数据
	// 你就软删除
	// 这个状态是存储状态，纯纯用于软删除的，业务层面上没有感知
	// 0-代表删除，1 代表有效
	Status uint8

	// 有效/无效
	//Type string
}

// Collection 收藏夹
type Collection struct {
	Id   int64  `gorm:"primaryKey,autoIncrement"`
	Name string `gorm:"type=varchar(1024)"`
	Uid  int64  `gorm:""`

	Ctime int64
	Utime int64
}

// UserCollectionBiz 收藏的东西
type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 收藏夹 ID
	// 作为关联关系中的外键，我们这里需要索引
	Cid   int64  `gorm:"index"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	// 这算是一个冗余，因为正常来说，
	// 只需要在 Collection 中维持住 Uid 就可以
	Uid   int64 `gorm:"uniqueIndex:biz_type_id_uid"`
	Ctime int64
	Utime int64
}

//// 假如说我有一个需求，需要查询到收藏夹的信息，和收藏夹里面的资源
//// SELECT c.id as cid , c.name as cname, uc.biz_id as biz_id, uc.biz as biz
//// FROM `collection` as c JOIN `user_collection_biz` as uc
//// ON c.id = uc.cid
//// WHERE c.id IN (1,2,3)
//
//type CollectionItem struct {
//	Cid   int64
//	Cname string
//	BizId int64
//	Biz   string
//}
//
//func (G *GORMInteractiveDAO) GetItems() ([]CollectionItem, error) {
//	// 不记得构造 JOIN 查询
//	var items []CollectionItem
//	err := G.db.Raw("", 1, 2, 3).Find(&items).Error
//	return items, err
//}
