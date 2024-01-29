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

// AddActivities 一次业务里面的相关账号的余额变动
func (a *AccountGORMDAO) AddActivities(ctx context.Context, activities ...AccountActivity) error {
	// 这里应该是一个事务
	// 同一个业务，牵涉到了多个账号，你必然是要求，要么全部成功，要么全部失败，不然就会出于中间状态
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 修改余额
		// 添加支付记录
		now := time.Now().UnixMilli()
		for _, act := range activities {
			// 一般在用户注册的时候就会创建好账号，但是我们并咩有，所以要兼容处理一下
			// 注意，系统账号是默认肯定存在的，一般是离线创建好的
			// 正常来说，你在一个平台注册的时候，
			// 后面的这些支撑系统，都会提前给你准备好账号
			err := tx.Create(&Account{
				Uid:      act.Uid,
				Account:  act.Account,
				Type:     act.AccountType,
				Balance:  act.Amount,
				Currency: act.Currency,
				Ctime:    now,
				Utime:    now,
			}).Clauses(clause.OnConflict{
				DoUpdates: clause.Assignments(map[string]any{
					// 记账，如果是减少呢？
					"balance": gorm.Expr("`balance` + ?", act.Amount),
					"utime":   now,
				}),
			}).Error
			if err != nil {
				return err
			}
		}
		return tx.Create(activities).Error
	})
}

func NewCreditGORMDAO(db *gorm.DB) AccountDAO {
	return &AccountGORMDAO{db: db}
}
