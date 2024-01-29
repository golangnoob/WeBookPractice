package repository

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/sync/errgroup"

	"webooktrial/comment/domain"
	"webooktrial/comment/repository/dao"
	"webooktrial/pkg/logger"
)

type CommentRepository interface {
	// FindByBiz 根据 ID 倒序查找
	// 并且会返回每个评论的三条直接回复
	FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, comment domain.Comment) error
	CreateComment(ctx context.Context, comment domain.Comment) error
	// GetCommentByIds 获取单条评论 支持批量获取
	GetCommentByIds(ctx context.Context, ids []int64) ([]domain.Comment, error)
	GetMoreReplies(ctx context.Context, rid int64, maxId, limit int64) ([]domain.Comment, error)
}

type CachedCommentRepo struct {
	dao dao.CommentDAO
	l   logger.LoggerV1
}

func (c *CachedCommentRepo) FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error) {
	// 事实上，最新评论它的缓存效果不是很好
	// 在这里缓存第一页，缓存没有，就去找数据库
	// 也可以考虑定时刷新缓存
	// 拿到的就是顶级评论
	daoComments, err := c.dao.FindByBiz(ctx, biz, bizId, minId, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(daoComments))
	// 拿到前三条子评论
	// 按照 pid 来分组，取组内三条（这三条是按照 ID 降序排序）
	var eg errgroup.Group
	downgraded := ctx.Value("downgraded") == true
	for _, d := range daoComments {
		d := d
		cm := c.toDomain(d)
		res = append(res, cm)
		if downgraded {
			continue
		}
		eg.Go(func() error {
			rs, err := c.dao.FindRepliesByPid(ctx, d.Id, 0, 3)
			if err != nil {
				// 我们认为这是一个可以容忍的错误
				c.l.Error("查询子评论失败", logger.Error(err))
				return nil
			}
			for _, r := range rs {
				cm.Children = append(cm.Children, c.toDomain(r))
			}
			return nil
		})
	}
	return res, eg.Wait()
}

func (c *CachedCommentRepo) DeleteComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Delete(ctx, dao.Comment{
		Id: comment.Id,
	})
}

func (c *CachedCommentRepo) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Insert(ctx, c.toEntity(comment))
}

func (c *CachedCommentRepo) GetCommentByIds(ctx context.Context, ids []int64) ([]domain.Comment, error) {
	vals, err := c.dao.FindOneByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	comments := make([]domain.Comment, 0, len(vals))
	for _, v := range vals {
		comment := c.toDomain(v)
		comments = append(comments, comment)
	}
	return comments, nil
}

func (c *CachedCommentRepo) GetMoreReplies(ctx context.Context, rid int64, maxId, limit int64) ([]domain.Comment, error) {
	cs, err := c.dao.FindRepliesByRid(ctx, rid, maxId, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(cs))
	for _, cm := range cs {
		res = append(res, c.toDomain(cm))
	}
	return res, nil
}

func (c *CachedCommentRepo) toDomain(daoComment dao.Comment) domain.Comment {
	val := domain.Comment{
		Id: daoComment.Id,
		Commentator: domain.User{
			ID: daoComment.Uid,
		},
		Biz:     daoComment.Biz,
		BizID:   daoComment.BizId,
		Content: daoComment.Content,
		Ctime:   time.UnixMilli(daoComment.Ctime),
		Utime:   time.UnixMilli(daoComment.Utime),
	}
	if daoComment.PID.Valid {
		val.ParentComment = &domain.Comment{
			Id: daoComment.PID.Int64,
		}
	}
	if daoComment.RootID.Valid {
		val.RootComment = &domain.Comment{
			Id: daoComment.RootID.Int64,
		}
	}
	return val
}

func (c *CachedCommentRepo) toEntity(domainComment domain.Comment) dao.Comment {
	daoComment := dao.Comment{
		Id:      domainComment.Id,
		Uid:     domainComment.Commentator.ID,
		Biz:     domainComment.Biz,
		BizId:   domainComment.BizID,
		Content: domainComment.Content,
	}
	if domainComment.RootComment != nil {
		daoComment.RootID = sql.NullInt64{
			Valid: true,
			Int64: domainComment.RootComment.Id,
		}
	}
	if domainComment.ParentComment != nil {
		daoComment.PID = sql.NullInt64{
			Valid: true,
			Int64: domainComment.ParentComment.Id,
		}
	}
	daoComment.Ctime = time.Now().UnixMilli()
	daoComment.Utime = time.Now().UnixMilli()
	return daoComment
}

func NewCommentRepo(commentDAO dao.CommentDAO, l logger.LoggerV1) CommentRepository {
	return &CachedCommentRepo{
		dao: commentDAO,
		l:   l,
	}
}
