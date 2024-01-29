package service

import (
	"context"

	"webooktrial/comment/domain"
	"webooktrial/comment/repository"
)

type CommentService interface {
	// GetCommentList Comment的id为0 获取一级评论
	// 按照 ID 升序排序
	GetCommentList(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error)
	// DeleteComment 删除评论，删除本评论及其子评论
	DeleteComment(ctx context.Context, id int64) error
	// CreateComment 创建评论
	CreateComment(ctx context.Context, comment domain.Comment) error
	GetMoreReplies(ctx context.Context, rid int64, maxId, limit int64) ([]domain.Comment, error)
}

type commentService struct {
	repo repository.CommentRepository
}

func NewCommentService(repo repository.CommentRepository) CommentService {
	return &commentService{repo: repo}
}

func (c *commentService) GetCommentList(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error) {
	list, err := c.repo.FindByBiz(ctx, biz, bizId, minId, limit)
	return list, err
}

func (c *commentService) DeleteComment(ctx context.Context, id int64) error {
	return c.DeleteComment(ctx, id)
}

func (c *commentService) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.repo.CreateComment(ctx, comment)
}

func (c *commentService) GetMoreReplies(ctx context.Context, rid int64, maxId int64, limit int64) ([]domain.Comment, error) {
	return c.repo.GetMoreReplies(ctx, rid, maxId, limit)
}
