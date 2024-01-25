package domain

import "github.com/wechatpay-apiv3/wechatpay-go/services/payments"

type Amount struct {
	// Currency 货币种类
	Currency string
	// 这里遵循微信的做法，就用 int64 来记录分数。
	// 那么对于不同的货币来说，这个字段的含义就不同。
	// 比如说一些货币没有分，只有整数。
	Total int64
}

type Payment struct {
	Amt Amount
	// BizTradeNO 代表业务，业务方决定怎么生成，
	BizTradeNO string
	// Description 订单本身的描述
	Description string
	Status      PaymentStatus
	// TxnID 第三方返回的 ID
	TxnID string
}

type WePayment struct {
	Payment
}

type PaymentStatus uint8

func (s PaymentStatus) AsUint8() uint8 {
	return uint8(s)
}

const (
	PaymentStatusUnknown = iota
	PaymentStatusInit
	PaymentStatusSuccess
	PaymentStatusFailed
	PaymentStatusRefund

	//PaymentStatusRefundFail
	//PaymentStatusRefundSuccess
	// PaymentStatusRecoup
	// PaymentStatusRecoupFailed
	// PaymentStatusRecoupSuccess
)

type Txn = payments.Transaction
