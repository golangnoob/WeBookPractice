package grpc

import (
	"context"

	"google.golang.org/grpc"

	pmtv1 "webooktrial/api/proto/gen/payment/v1"
	"webooktrial/payment/domain"
	"webooktrial/payment/service/wechat"
)

type WechatServiceServer struct {
	pmtv1.UnimplementedWechatPaymentServiceServer
	svc *wechat.NativePaymentService
}

func NewWechatServiceServer(svc *wechat.NativePaymentService) *WechatServiceServer {
	return &WechatServiceServer{svc: svc}
}

func (s *WechatServiceServer) Register(server *grpc.Server) {
	pmtv1.RegisterWechatPaymentServiceServer(server, s)
}

func (s *WechatServiceServer) GetPayment(ctx context.Context, req *pmtv1.GetPaymentRequest) (*pmtv1.GetPaymentResponse, error) {
	p, err := s.svc.GetPayment(ctx, req.GetBizTradeNo())
	if err != nil {
		return nil, err
	}
	return &pmtv1.GetPaymentResponse{
		Status: pmtv1.PaymentStatus(p.Status),
	}, nil
}

// 根据 type 来分发
//func (s *WechatServiceServer) NativePrePay(ctx context.Context, request *pmtv1.PrePayRequest) (*pmtv1.NativePrePayResponse, error) {
//	switch request.Type {
//	case "native":
//		return s.svc.Prepay()
//	case "jsapi":
//		// 掉另外一个方法
//	}
//}

func (s *WechatServiceServer) NativePrepay(ctx context.Context, req *pmtv1.PrepayRequest) (*pmtv1.NativePrepayResponse, error) {
	codeURL, err := s.svc.Prepay(ctx, domain.Payment{
		Amt: domain.Amount{
			Currency: req.Amt.Currency,
			Total:    req.Amt.Total,
		},
		BizTradeNO:  req.BizTradeNo,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return &pmtv1.NativePrepayResponse{
		CodeUrl: codeURL,
	}, nil
}
