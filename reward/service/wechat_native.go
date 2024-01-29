package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	accountv1 "webooktrial/api/proto/gen/account/v1"
	pmtv1 "webooktrial/api/proto/gen/payment/v1"
	"webooktrial/pkg/logger"
	"webooktrial/reward/domain"
	"webooktrial/reward/repository"
)

type WechatNativeRewardService struct {
	client pmtv1.WechatPaymentServiceClient
	repo   repository.RewardRepository
	l      logger.LoggerV1
	acli   accountv1.AccountServiceClient
}

func (w *WechatNativeRewardService) PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	// 可以考虑缓存我的二维码，一旦我发现支付成功了，我就清除我的二维码
	cu, err := w.repo.GetCachedCodeURL(ctx, r)
	if err == nil {
		return cu, err
	}
	r.Status = domain.RewardStatusInit
	rid, err := w.repo.CreateReward(ctx, r)
	if err != nil {
		return domain.CodeURL{}, err
	}
	// 也可以在这里记录分账信息
	resp, err := w.client.NativePrepay(ctx, &pmtv1.PrepayRequest{
		Amt: &pmtv1.Amount{
			Total:    r.Amt,
			Currency: "CNY",
		},
		BizTradeNo:  fmt.Sprintf("reward-%d", rid),
		Description: fmt.Sprintf("打赏-%s", r.Target.BizName),
	})
	if err != nil {
		return domain.CodeURL{}, err
	}
	cu = domain.CodeURL{
		Rid: rid,
		URL: resp.CodeUrl,
	}
	// 这一步可以异步进行
	err1 := w.repo.CachedCodeURL(ctx, cu, r)
	if err1 != nil {
		// 记录日志
	}
	return cu, nil
}

func (w *WechatNativeRewardService) GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error) {
	// 快路径
	r, err := w.repo.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	if r.Uid != uid {
		// 非法查询
		return domain.Reward{}, errors.New("查询的打赏记录和打赏人对不上")
	}
	if r.Completed() || ctx.Value("limited") == "true" {
		return r, nil
	}
	// 这个时候，考虑到支付到查询结果，我们搞一个慢路径
	// 你有可能支付了，但是我 reward 本身没有收到通知
	// 我直接查询 payment，
	// 只能解决，支付收到了，但是 reward 没收到
	// 降级状态，限流状态，熔断状态，不要走慢路径
	resp, err := w.client.GetPayment(ctx, &pmtv1.GetPaymentRequest{
		BizTradeNo: w.bizTradeNO(r.Id),
	})
	if err != nil {
		// 这边我们直接返回从数据库查询的数据
		w.l.Error("慢路径查询支付结果失败",
			logger.Int64("rid", r.Id), logger.Error(err))
		return r, nil
	}
	// 更新状态
	switch resp.Status {
	case pmtv1.PaymentStatus_PaymentStatusFailed:
		r.Status = domain.RewardStatusFailed
	case pmtv1.PaymentStatus_PaymentStatusInit:
		r.Status = domain.RewardStatusInit
	case pmtv1.PaymentStatus_PaymentStatusSuccess:
		r.Status = domain.RewardStatusPayed
	case pmtv1.PaymentStatus_PaymentStatusRefund:
		// 理论上来说不可能出现这个，直接设置为失败
		r.Status = domain.RewardStatusFailed
	}
	err = w.repo.UpdateStatus(ctx, rid, r.Status)
	if err != nil {
		w.l.Error("更新本地打赏状态失败",
			logger.Int64("rid", r.Id), logger.Error(err))
		return r, nil
	}
	return r, nil
}

func (w *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error {
	rid := w.toRid(bizTradeNO)
	err := w.repo.UpdateStatus(ctx, rid, status)
	if err != nil {
		return err
	}
	// 完成了支付，准备入账
	if status == domain.RewardStatusPayed {
		r, err := w.repo.GetReward(ctx, rid)
		if err != nil {
			return err
		}
		// webook 抽成
		weAmt := int64(float64(r.Amt) * 0.1)
		_, err = w.acli.Credit(ctx, &accountv1.CreditRequest{
			Biz:   "reward",
			BizId: rid,
			Items: []*accountv1.CreditItem{
				{
					AccountType: accountv1.AccountType_AccountTypeSystem,
					// 虽然可能为 0，但是也要记录出来
					Amt:      weAmt,
					Currency: "CNY",
				},
				{
					Account:     r.Uid,
					Uid:         r.Uid,
					AccountType: accountv1.AccountType_AccountTypeReward,
					Amt:         r.Amt - weAmt,
					Currency:    "CNY",
				},
			},
		})
		if err != nil {
			w.l.Error("入账失败了，快来修数据啊！！！",
				logger.String("biz_trade_no", bizTradeNO),
				logger.Error(err))
			// 做好监控和告警，这里
			// 引入自动修复功能
			return err
		}
	}
	return nil
}

func (w *WechatNativeRewardService) bizTradeNO(rid int64) string {
	return fmt.Sprintf("reward-%d", rid)
}

func (w *WechatNativeRewardService) toRid(tradeNO string) int64 {
	ridStr := strings.Split(tradeNO, "-")
	val, _ := strconv.ParseInt(ridStr[1], 10, 64)
	return val
}

func NewWechatNativeRewardService(client pmtv1.WechatPaymentServiceClient, repo repository.RewardRepository, l logger.LoggerV1, acli accountv1.AccountServiceClient) RewardService {
	return &WechatNativeRewardService{client: client, repo: repo, l: l, acli: acli}
}
