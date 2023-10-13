package dao

import (
	"context"
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm"
)

type SMSMsg struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 1 失败未重试 2 重试成功 3 重试失败
	Status       int
	Ctime        int64
	Utime        int64
	Biz          string
	PhoneNumbers PhoneNums `json:"phone_numbers"`
	Args         Args      `json:"args"`
}

type PhoneNums []string
type Args []string

func (p *PhoneNums) Scan(value any) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, p)
}

func (p *PhoneNums) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (a *Args) Scan(value any) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, a)
}

func (a *Args) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type SMSDaoInterface interface {
	Insert(ctx context.Context, s SMSMsg) error
	Select(ctx context.Context, status int) ([]SMSMsg, error)
	Update(ctx context.Context, id int64, status int) error
}

type SMSDao struct {
	db *gorm.DB
}

func NewSMSDao(db *gorm.DB) SMSDaoInterface {
	return &SMSDao{
		db: db,
	}
}

func (S *SMSDao) Insert(ctx context.Context, s SMSMsg) error {
	//TODO implement me
	panic("implement me")
}

func (S *SMSDao) Select(ctx context.Context, status int) ([]SMSMsg, error) {
	//TODO implement me
	panic("implement me")
}

func (S *SMSDao) Update(ctx context.Context, id int64, status int) error {
	//TODO implement me
	panic("implement me")
}
