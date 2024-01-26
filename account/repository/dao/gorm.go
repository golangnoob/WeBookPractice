package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountGORMDAO struct {
	db *gorm.DB
}

func (a *AccountGORMDAO) AddActivities(ctx context.Context, activities ...AccountActivity) error {
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		for _, act := range activities {
			// 一般在用户注册的时候就会创建好账号，但是我们并咩有，所以要兼容处理一下
			// 注意，系统账号是默认肯定存在的，一般是离线创建好的
			err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.Assignments(map[string]interface{}{
					"balance": gorm.Expr("balance + ?", act.Amount),
					"utime":   now,
				}),
			}).Create(&Account{
				Uid:      act.Uid,
				Account:  act.Account,
				Type:     act.AccountType,
				Balance:  act.Amount,
				Currency: act.Currency,
				Ctime:    now,
				Utime:    now,
			}).Error
			if err != nil {
				return err
			}
		}
		return tx.Create(&activities).Error
	})
}

func NewCreditGORMDAO(db *gorm.DB) AccountDAO {
	return &AccountGORMDAO{db: db}
}
