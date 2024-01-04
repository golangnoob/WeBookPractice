package fixer

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"webooktrial/pkg/migrator"
)

type OverrideFixer[T migrator.Entity] struct {
	// 因为本身其实这个不涉及什么领域对象，
	// 这里操作的不是 migrator 本身的领域对象
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

func NewOverrideFixer[T migrator.Entity](base *gorm.DB,
	target *gorm.DB) (*OverrideFixer[T], error) {
	var t T
	rows, err := base.Model(&t).Limit(1).Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: columns,
	}, nil
}

func (o *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	var src T
	// 找出数据
	err := o.base.WithContext(ctx).Where("id = ?", id).First(&src).Error
	switch err {
	// 找到了数据
	case nil:
		return o.target.WithContext(ctx).Clauses(&clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(o.columns),
		}).Create(&src).Error
	case gorm.ErrRecordNotFound:
		return o.target.Delete("id = ?", id).Error
	default:
		return err
	}
}
