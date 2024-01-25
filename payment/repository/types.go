package repository

import (
	"context"
	"time"

	"webooktrial/payment/domain"
)

//go:generate mockgen -source=types.go -package=repomocks -destination=mocks/payment.mock.go PaymentRepository
type PaymentRepository interface {
	AddPayment(ctx context.Context, pmt domain.Payment) error
	// UpdatePayment 这个设计有点差
	UpdatePayment(ctx context.Context, pmt domain.Payment) error
	FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error)
	GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error)
}
