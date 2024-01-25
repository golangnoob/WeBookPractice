package wechat

import (
	"context"

	"webooktrial/payment/domain"
)

type PaymentService interface {
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}
