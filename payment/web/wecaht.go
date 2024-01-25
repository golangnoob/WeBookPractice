package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"

	"webooktrial/payment/service/wechat"
	"webooktrial/pkg/ginx"
	"webooktrial/pkg/logger"
)

type WechatHandler struct {
	handler   *notify.Handler
	l         logger.LoggerV1
	nativeSvc *wechat.NativePaymentService
}

func NewWechatHandler(handler *notify.Handler, l logger.LoggerV1, nativeSvc *wechat.NativePaymentService) *WechatHandler {
	return &WechatHandler{handler: handler, l: l, nativeSvc: nativeSvc}
}

func (h *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.GET("/hello", func(context *gin.Context) {
		context.String(http.StatusOK, "我进来了")
	})
	server.Any("/pay/callback", ginx.Wrap(h.HandleNative))
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) (ginx.Result, error) {
	transaction := &payments.Transaction{}
	// 第一个返回值里面的内容暂时用不上
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		// 这里不可能触发对账，解密出错了，拿不到 BizTradeNO

		// 返回非 2xx 的响应
		// 就一个原因：有人伪造请求，有人在伪造微信支付的回调
		// 做好监控和告警
		// 大量进来这个分支，就说明有人搞事
		return ginx.Result{}, err
	}
	err = h.nativeSvc.HandleCallback(ctx, transaction)
	if err != nil {
		// 这里处理失败了，可以再次触发对账
		// 返回非 2xx 的响应
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
