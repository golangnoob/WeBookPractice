package dao

import (
	"context"
	"database/sql"
	"time"

	"webooktrial/payment/domain"
)

type PaymentDAO interface {
	Insert(ctx context.Context, pmt Payment) error
	UpdateTxnIDAndStatus(ctx context.Context, bizTradeNO string, txnID string, status domain.PaymentStatus) error
	FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error)
	GetPayment(ctx context.Context, bizTradeNO string) (Payment, error)
}

type Payment struct {
	Id          int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Amt         int64
	Currency    string
	Description string `gorm:"description"`
	// 后续可以考虑增加字段，来标记是用的是微信支付亦或是支付宝支付
	// Type uint8 // 微信支付或者支付宝支付
	// 也可以考虑提供一个巨大的 BLOB 字段，
	// 来存储和支付有关的其它字段
	// ExtraData

	// WepayExtraData
	// AliPayExtraData
	// UnionPayExtraData

	// 业务方传过来的
	BizTradeNO string `gorm:"column:biz_trade_no;type:varchar(256);unique"`

	// 第三方支付平台的事务 ID 唯一的
	TxnID sql.NullString `gorm:"column:txn_id;type:varchar(128);unique"`

	Status uint8
	Utime  int64
	Ctime  int64
}

// WechatPaymentExt 微信支付独有的
type WechatPaymentExt struct {
}

// AliPaymentExt 支付宝支付独有的
type AliPaymentExt struct {
}
