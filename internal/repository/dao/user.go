package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{
		db: db,
	}
}

func (dao *UserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *UserDao) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			//
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *UserDao) Update(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	err := dao.db.Model(&User{}).Where(&User{ID: u.ID}).Updates(u).Error
	return err
}

func (dao *UserDao) Select(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Model(&User{}).Where(&User{ID: id}).First(&u).Error
	return u, err
}

// User 直接对应数据库表结构
// 有些人叫做 entity，有些人叫做 model，有些人叫做 PO(persistent object)
type User struct {
	ID int64 `gorm:"primaryKey,autoIncrement"`
	// 全部用户唯一
	Email    string `gorm:"unique"`
	Password string
	// 往这里面加

	// 用户详细信息
	Nickname string
	Birthday string
	Describe string

	// 毫秒级
	Ctime int64
	Utime int64
}
