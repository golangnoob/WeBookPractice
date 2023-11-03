package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

//go:generate mockgen -source=./user.go -package=daomocks -destination=mocks/user.mock.go UserDAO
var (
	ErrUserDuplicate = errors.New("邮箱冲突或手机号冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDAO interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechat(ctx context.Context, openId string) (User, error)
	Insert(ctx context.Context, u User) error
	Update(ctx context.Context, u User) error
	Select(ctx context.Context, id int64) (User, error)
}

type GormUserDao struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDao{
		db: db,
	}
}

func (dao *GormUserDao) FindByWechat(ctx context.Context, openId string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openId).First(&u).Error
	//err := dao.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return u, err
}

func (dao *GormUserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		err = fmt.Errorf("findByEmail fail, err:%w", err)
	}

	return u, err
}

func (dao *GormUserDao) FindByPhone(ctx context.Context, Phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", Phone).Error
	return u, err
}

func (dao *GormUserDao) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("`id` = ?", id).First(&u).Error
	return u, err
}

func (dao *GormUserDao) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突 or 手机号码冲突
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *GormUserDao) Update(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	err := dao.db.WithContext(ctx).Model(&User{}).Where(&User{ID: u.ID}).Updates(u).Error
	return err
}

func (dao *GormUserDao) Select(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Model(&User{}).Where(&User{ID: id}).First(&u).Error
	return u, err
}

// User 直接对应数据库表结构
// 有些人叫做 entity，有些人叫做 model，有些人叫做 PO(persistent object)
type User struct {
	ID int64 `gorm:"primaryKey,autoIncrement"`
	// 全部用户唯一
	Email sql.NullString `gorm:"unique"`

	// 唯一索引允许有多个空值
	// 但是不能有多个 ""
	Phone sql.NullString `gorm:"unique"`
	// 最大问题就是，你要解引用
	// 你要判空
	//Phone *string

	Password string
	// 往这里面加

	// 用户详细信息
	Nickname string
	Birthday string
	Describe string
	// 微信的字段
	WeChatUnionId sql.NullString
	WeChatOpenId  sql.NullString

	// 毫秒级 Ctime 创建时间， Utime更新时间
	Ctime int64
	Utime int64
}
