package grpc

import (
	"context"
	"math"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	commentv1 "webooktrial/api/proto/gen/comment/v1"
	"webooktrial/comment/domain"
	"webooktrial/comment/service"
)

type CommentServiceServer struct {
	commentv1.UnimplementedCommentServiceServer
	svc service.CommentService
}

func (c *CommentServiceServer) Register(server grpc.ServiceRegistrar) {
	commentv1.RegisterCommentServiceServer(server, c)
}

func NewCommentServiceServer(svc service.CommentService) *CommentServiceServer {
	return &CommentServiceServer{svc: svc}
}

func (c *CommentServiceServer) GetCommentList(ctx context.Context, req *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	// 可以在这里判断是否触发了限流或者降级，如果触发，直接返回
	minID := req.MinId
	// 第一次查询，这边我们认为用户没有传
	if minID <= 0 {
		minID = math.MaxInt64
	}
	domainComments, err := c.svc.
		GetCommentList(ctx,
			req.GetBiz(),
			req.GetBizId(),
			req.GetMinId(),
			req.GetLimit())
	if err != nil {
		return nil, err
	}
	return &commentv1.CommentListResponse{
		Comments: c.toDTO(domainComments),
	}, nil
}

func (c *CommentServiceServer) DeleteComment(ctx context.Context, req *commentv1.DeleteCommentRequest) (*commentv1.DeleteCommentResponse, error) {
	err := c.svc.DeleteComment(ctx, req.Id)
	return &commentv1.DeleteCommentResponse{}, err
}

func (c *CommentServiceServer) CreateComment(ctx context.Context, req *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	// 可以在这里判断是否触发了限流或者降级，如果触发，则将消息丢进kafka后返回
	err := c.svc.CreateComment(ctx, convertToDomain(req.GetComment()))
	return &commentv1.CreateCommentResponse{}, err
}

func (c *CommentServiceServer) GetMoreReplies(ctx context.Context, req *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	cs, err := c.svc.GetMoreReplies(ctx, req.Rid, req.MaxId, req.Limit)
	if err != nil {
		return nil, err
	}
	return &commentv1.GetMoreRepliesResponse{
		Replies: c.toDTO(cs),
	}, nil
}

func (c *CommentServiceServer) toDTO(domainComments []domain.Comment) []*commentv1.Comment {
	rpcComments := make([]*commentv1.Comment, 0, len(domainComments))
	for _, domainComment := range domainComments {
		rpcComment := &commentv1.Comment{
			Id:      domainComment.Id,
			Uid:     domainComment.Commentator.ID,
			Biz:     domainComment.Biz,
			Bizid:   domainComment.BizID,
			Content: domainComment.Content,
			Ctime:   timestamppb.New(domainComment.Ctime),
			Utime:   timestamppb.New(domainComment.Utime),
		}
		if domainComment.RootComment != nil {
			rpcComment.RootComment = &commentv1.Comment{
				Id: domainComment.RootComment.Id,
			}
		}
		if domainComment.ParentComment != nil {
			rpcComment.ParentComment = &commentv1.Comment{
				Id: domainComment.ParentComment.Id,
			}
		}
		rpcComments = append(rpcComments, rpcComment)
	}
	rpcCommentMap := make(map[int64]*commentv1.Comment, len(rpcComments))
	for _, rpcComment := range rpcComments {
		rpcCommentMap[rpcComment.Id] = rpcComment
	}
	for _, domainComment := range domainComments {
		rpcComment := rpcCommentMap[domainComment.Id]
		if domainComment.RootComment != nil {
			val, ok := rpcCommentMap[domainComment.RootComment.Id]
			if ok {
				rpcComment.RootComment = val
			}
		}
		if domainComment.ParentComment != nil {
			val, ok := rpcCommentMap[domainComment.ParentComment.Id]
			if ok {
				rpcComment.ParentComment = val
			}
		}
	}
	return rpcComments
}

func convertToDomain(comment *commentv1.Comment) domain.Comment {
	domainComment := domain.Comment{
		Id:      comment.Id,
		Biz:     comment.Biz,
		BizID:   comment.Bizid,
		Content: comment.Content,
		Commentator: domain.User{
			ID: comment.Uid,
		},
	}
	if comment.GetParentComment() != nil {
		domainComment.ParentComment = &domain.Comment{
			Id: comment.GetParentComment().GetId(),
		}
	}
	if comment.GetRootComment() != nil {
		domainComment.RootComment = &domain.Comment{
			Id: comment.GetRootComment().GetId(),
		}
	}
	return domainComment
}
