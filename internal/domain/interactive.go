package domain

type Interactive struct {
	Biz   string
	BizId int64 `gorm:"comment:article_id"`

	ReadCnt    int64 `json:"read_cnt"`
	LikeCnt    int64 `json:"like_cnt"`
	CollectCnt int64 `json:"collect_cnt"`
	// 这个是当下这个资源，你有没有点赞或者收集
	// 你也可以考虑把这两个字段分离出去，作为一个单独的结构体
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

type Self struct {
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

type Collection struct {
	Name  string
	Uid   int64
	Items []Resource
}
