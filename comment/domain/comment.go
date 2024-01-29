package domain

import "time"

type Comment struct {
	Id int64 `json:"id"`
	// 评论者
	Commentator User `json:"user"`
	// 评论对象类型
	Biz string `json:"biz"`
	// 评论对象ID
	BizID int64 `json:"bizId"`
	// 根评论
	RootComment *Comment `json:"rootComment"`
	// 父级评论
	ParentComment *Comment `json:"parentComment"`
	// 评论内容
	Content  string
	Children []Comment `json:"children"`
	Ctime    time.Time
	Utime    time.Time
}

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
