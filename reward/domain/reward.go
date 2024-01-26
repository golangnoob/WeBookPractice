package domain

type Target struct {
	// Biz 视频 文章 还是什么
	Biz   string
	BizId int64
	// BizName 打赏的东西叫什么
	BizName string

	// 打赏的目标用户
	Uid int64
}

type Reward struct {
	Id     int64
	Uid    int64
	Target Target
	Amt    int64
	Status RewardStatus
}

// Completed 是否已经完成
// 目前来说，也就是是否处理了支付回调
func (r Reward) Completed() bool {
	return r.Status == RewardStatusFailed || r.Status == RewardStatusPayed
}

type RewardStatus uint8

func (r RewardStatus) AsUint8() uint8 {
	return uint8(r)
}

const (
	RewardStatusUnknown = iota
	RewardStatusInit
	RewardStatusPayed
	RewardStatusFailed
)

// 垃圾设计
type CodeURL struct {
	Rid int64
	URL string
}
