package article

type Article struct {
	Id int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	// 标题的长度
	// 正常都不会超过这个长度
	Title   string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 作者
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	Utime    int64 `bson:"utime,omitempty"`
}

// PublishedArticle 衍生类型，偷个懒
type PublishedArticle Article

// PublishedArticleV1 s3 演示专属

type PublishedArticleV1 struct {
	Id       int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}

//func (u *Article) BeforeCreate(tx *gorm.DB) (err error) {
//	startTime := time.Now()
//	tx.Set("start_time", startTime)
//	slog.Default().Info("这是 BeforeCreate 钩子函数")
//	return nil
//}

//func (u *Article) AfterCreate(tx *gorm.DB) (err error) {
//	// 我要计算执行时间，我怎么拿到 before 里面的 startTime?
//	val, _ := tx.Get("start_time")
//	startTime, ok := val.(time.Time)
//	if !ok {
//		return nil
//	}
//	// 执行时间就出来了
//	duration := time.Since(startTime)
//	slog.Default().Info("这是 AfterCreate 钩子函数")
//	return nil
//}

//type model struct {
//}
//
//func (u model) BeforeSave(tx *gorm.DB) (err error) {
//	startTime := time.Now()
//	tx.Set("start_time", startTime)
//	slog.Default().Info("这是 BeforeCreate 钩子函数")
//	return nil
//}

//func (u model) AfterSave(tx *gorm.DB) (err error) {
//	// 我要计算执行时间，我怎么拿到 before 里面的 startTime?
//	val, _ := tx.Get("start_time")
//	startTime, ok := val.(time.Time)
//	if !ok {
//		return nil
//	}
//	// 执行时间就出来了
//	duration := time.Since(startTime)
//	slog.Default().Info("这是 AfterCreate 钩子函数")
//	return nil
//}
