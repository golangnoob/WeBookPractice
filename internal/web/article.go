package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	intrv1 "webooktrial/api/proto/gen/intr/v1"
	rewardv1 "webooktrial/api/proto/gen/reward/v1"
	"webooktrial/internal/domain"
	"webooktrial/internal/service"
	ijwt "webooktrial/internal/web/jwt"
	"webooktrial/pkg/ginx"
	"webooktrial/pkg/logger"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc       service.ArticleService
	l         logger.LoggerV1
	rewardSvc rewardv1.RewardServiceClient
	intrSvc   intrv1.InteractiveServiceClient
	biz       string
}

func NewArticleHandler(svc service.ArticleService,
	l logger.LoggerV1,
	intrSvc intrv1.InteractiveServiceClient) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		l:       l,
		biz:     "article",
		intrSvc: intrSvc,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	// 修改
	//g.PUT("/")
	// 新增
	//g.POST("/")
	// g.DELETE("/a_id")

	g.POST("/edit", h.Edit)
	g.POST("/withdraw", h.Withdraw)
	g.POST("/publish", h.Publish)
	// 创作者的查询接口
	// 这个是获取数据的接口，理论上来说（遵循 RESTful 规范），应该是用 GET 方法
	// GET localhost/articles => List 接口
	g.POST("/list",
		ginx.WrapBodyAndToken[ListReq, ijwt.UserClaims](h.List))
	g.GET("/detail/:id", ginx.WrapToken[ijwt.UserClaims](h.Detail))

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail, func(ctx *gin.Context) {
		// 增加阅读计数。
		//go func() {
		//	// 开一个 goroutine，异步去执行
		//	er := a.intrSvc.IncrReadCnt(ctx, a.biz, art.Id)
		//	if er != nil {
		//		a.l.Error("增加阅读计数失败",
		//			logger.Int64("aid", art.Id),
		//			logger.Error(err))
		//	}
		//}()
	})
	// 点赞是这个接口，取消点赞也是这个接口
	// RESTful 风格
	//pub.POST("/like/:id", ginx.WrapBodyAndToken[LikeReq,
	//	ijwt.UserClaims](h.Like))
	pub.POST("/like", ginx.WrapBodyAndToken[LikeReq,
		ijwt.UserClaims](h.Like))
	//pub.POST("/cancel_like", ginx.WrapBodyAndToken[LikeReq,
	//	ijwt.UserClaims](h.Like))
	pub.POST("/reward", ginx.WrapBodyAndToken[RewardReq,
		ijwt.UserClaims](h.reward))
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		_, err = h.intrSvc.Like(ctx, &intrv1.LikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
		})
	} else {
		_, err = h.intrSvc.CancelLike(ctx, &intrv1.CancelLikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
		})
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}
	// 检测输入，跳过这一步
	// 调用 svc 的代码
	id, err := h.svc.Save(ctx, req.toDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 打日志？
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}
	id, err := h.svc.Publish(ctx, req.toDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 打日志？
		h.l.Error("发表帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	uc, ok := c.(ijwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}
	err := h.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 打日志？
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	res, err := h.svc.List(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res, func(idx int, src domain.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				Status:   src.Status.ToUint8(),
				// 这个列表请求，不需要返回内容
				//Content: src.Content,
				// 这个是创作者看自己的文章列表，也不需要这个字段
				//Author: src.Author
				Ctime: src.Ctime.Format(time.DateTime),
				Utime: src.Utime.Format(time.DateTime),
			}
		}),
	}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("获得文章信息失败", logger.Error(err))
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	// 这是不借助数据库查询来判定的方法
	if art.Author.Id != uc.Uid {
		//ctx.JSON(http.StatusOK)
		// 如果公司有风控系统，这个时候就要上报这种非法访问的用户了。
		//a.l.Error("非法访问文章，创作者 ID 不匹配",
		//	logger.Int64("uid", usr.Id))
		return ginx.Result{
			Code: 4,
			// 也不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", uc.Uid)
	}
	return ginx.Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			//Author: art.Author
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		h.l.Error("前端输入的 ID 不对", logger.Error(err))
		return
	}
	uc := ctx.MustGet("claims").(ijwt.UserClaims)
	var eg errgroup.Group
	var art domain.Article
	eg.Go(func() error {
		art, err = h.svc.GetPublishedById(ctx, id, uc.Uid)
		return err
	})

	var getResp *intrv1.GetResponse
	eg.Go(func() error {
		// 要在这里获得这篇文章的计数
		// 这个地方可以容忍错误
		getResp, err = h.intrSvc.Get(ctx, &intrv1.GetRequest{
			Biz:   h.biz,
			BizId: id,
			Uid:   uc.Uid,
		})
		// 这种是容错的写法
		//if err != nil {
		//	// 记录日志
		//}
		//return nil
		return err
	})

	// 在这儿等，要保证前面两个
	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 增加阅读计数。
	go func() {
		// 你都异步了，怎么还说有巨大的压力呢？
		// 即使是异步执行最后还是会访问数据库
		// 开一个 goroutine，异步去执行
		_, er := h.intrSvc.IncrReadCnt(ctx, &intrv1.IncrReadCntRequest{
			Biz:   h.biz,
			BizId: id,
		})
		if er != nil {
			h.l.Error("增加阅读计数失败",
				logger.Int64("aid", art.Id),
				logger.Error(err))
		}
	}()

	// ctx.Set("art", art)
	intr := getResp.Intr
	// 这个功能是不是可以让前端，主动发一个 HTTP 请求，来增加一个计数？
	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			Author:     art.Author.Name,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			LikeCnt:    intr.LikeCnt,
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
		},
	})
}

func (h *ArticleHandler) reward(ctx *gin.Context, req RewardReq,
	uc ijwt.UserClaims) (ginx.Result, error) {
	art, err := h.svc.GetPublishedById(ctx, req.Id, uc.Uid)
	if err != nil {
		return ginx.Result{}, err
	}
	// 我要在这里实现打赏
	// 拿到一个打赏的二维码
	// 我不是直接调用支付，而是调用打赏
	// 打赏什么东西，谁打赏，打赏多少钱？
	resp, err := h.rewardSvc.PreReward(ctx, &rewardv1.PreRewardRequest{
		Biz:   "article",
		BizId: req.Id,
		Uid:   uc.Uid,
		Amt:   req.Amount,
		// 创作者是谁？
		TargetUid: art.Author.Id,
		// 这个地方用作者呢？还是用标题呢？
		// 作者写得好
		BizName: art.Title,
	})
	return ginx.Result{
		Data: map[string]any{
			"codeURL": resp.CodeUrl,
			// 代表的是这一次的打赏
			"rid": resp.Rid,
		},
	}, nil
}
