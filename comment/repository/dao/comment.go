package dao

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = gorm.ErrRecordNotFound

//go:generate mockgen -source=./comment.go -package=daomocks -destination=mocks/comment.mock.go CommentDAO
type CommentDAO interface {
	Insert(ctx context.Context, c Comment) error
	// FindByBiz 只查找一级评论
	FindByBiz(ctx context.Context, biz string,
		bizId, minId, limit int64) ([]Comment, error)
	// FindCommentList Comment的ID为0 获取一级评论，如果不为0获取对应的评论，和其评论的所有回复
	FindCommentList(ctx context.Context, c Comment) ([]Comment, error)
	FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error)
	// Delete 删除本节点和其对应的子节点
	Delete(ctx context.Context, u Comment) error
	FindOneByIds(ctx context.Context, Ids []int64) ([]Comment, error)
	FindRepliesByRid(ctx context.Context, rid int64, Id int64, limit int64) ([]Comment, error)
}

type GORMCommentDAO struct {
	db *gorm.DB
}

func NewGORMCommentDAO(db *gorm.DB) CommentDAO {
	return &GORMCommentDAO{db: db}
}

func (g *GORMCommentDAO) Insert(ctx context.Context, c Comment) error {
	return g.db.WithContext(ctx).Create(&c).Error
}

func (g *GORMCommentDAO) FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]Comment, error) {
	var res []Comment
	err := g.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND id < ? AND pid IS NULL", biz, bizId, minId).
		Limit(int(limit)).
		Find(&res).Error
	return res, err
}

func (g *GORMCommentDAO) FindCommentList(ctx context.Context, c Comment) ([]Comment, error) {
	var res []Comment
	builder := g.db.WithContext(ctx)
	if c.Id == 0 {
		builder = builder.
			Where("biz=?", c.Biz).
			Where("biz_ID=?", c.BizId).
			Where("root_ID is null")
	} else {
		builder = builder.Where("root_ID=? or id =?", c.Id, c.Id)
	}
	err := builder.Find(&res).Error
	return res, err
}

func (g *GORMCommentDAO) FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error) {
	var res []Comment
	err := g.db.WithContext(ctx).Where("pid = ?", pid).
		Order("ID DESC").
		Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (g *GORMCommentDAO) Delete(ctx context.Context, u Comment) error {
	// 数据库帮你级联删除了，不需要担忧并发问题
	// 假如 4 已经删了，按照外键的约束，如果你插入一个 pid=4 的行，你是插不进去的
	return g.db.WithContext(ctx).Delete(&Comment{
		Id: u.Id,
	}).Error
}

func (g *GORMCommentDAO) FindOneByIds(ctx context.Context, Ids []int64) ([]Comment, error) {
	var res []Comment
	err := g.db.WithContext(ctx).
		Where("id in ?", Ids).
		First(&res).
		Error
	return res, err
}

func (g *GORMCommentDAO) FindRepliesByRid(ctx context.Context, rid int64, id int64, limit int64) ([]Comment, error) {
	var res []Comment
	err := g.db.WithContext(ctx).
		Where("root_ID = ? AND id > ?", rid, id).
		Order("ID DESC").
		Limit(int(limit)).Find(&res).Error
	return res, err
}

type Comment struct {
	// 代表你评论本体
	Id int64
	// 发表评论的人
	// 要不要在这个列创建索引？
	// 取决于有没有 WHERE uID = ? 的查询
	Uid int64
	// 这个代表的是你评论的对象是什么？
	// 比如说代表某个帖子，代表某个视频，代表某个图片
	Biz   string `gorm:"index:biz_type_id"`
	BizId int64  `gorm:"index:biz_type_if"`

	// 用 NULL 来表达没有父亲
	// 你可以考虑用 -1 来代表没有父亲
	// 索引是如何处理 NULL 的？？？
	// NULL 的取值非常多

	PID sql.NullInt64 `gorm:"index"`
	// 外键指向的也是同一张表
	ParentComment *Comment `gorm:"ForeignKey:PID;AssociationForeignKey:ID;constraint:OnDelete:CASCADE"`

	// 引入 RootID 这个设计
	// 顶级评论的 ID
	// 主要是为了加载整棵评论的回复组成树
	RootID sql.NullInt64 `gorm:"index:root_ID_ctime"`
	Ctime  int64         `gorm:"index:root_ID_ctime"`

	// 评论的内容
	Content string

	Utime int64
}

func (*Comment) TableName() string {
	return "comments"
}
