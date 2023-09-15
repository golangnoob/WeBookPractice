package domain

import "time"

// User 领域对象， 是 DDD 中的entity
// BO(business object)
type User struct {
	Id       int64
	Email    string
	Phone    string
	Password string
	Nickname string
	Birthday string
	AboutMe  string

	// 不要组合，万一你将来可能还有 DingDingInfo，里面有同名字段 UnionID
	WechatInfo WechatInfo
	Ctime      time.Time
}
