package validator

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"webooktrial/pkg/logger"
	"webooktrial/pkg/migrator"
	"webooktrial/pkg/migrator/events"
)

type Validator[T migrator.Entity] struct {
	// base 校验基准
	base *gorm.DB
	// target 校验目标
	target    *gorm.DB
	l         logger.LoggerV1
	p         events.Producer
	direction string
	batchSize int
	highLoad  *atomicx.Value[bool]

	// 在这里加字段，比如说，在查询 base 根据什么列来排序，在 target 的时候，根据什么列来查询数据
	// 最极端的情况，是这样

	utime int64
	// <=0 说明直接退出校验循环
	// > 0 真的 sleep
	sleepInterval time.Duration
	fromBase      func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB,
	l logger.LoggerV1, p events.Producer, direction string) *Validator[T] {
	highLoad := atomicx.NewValueOf[bool](false)
	go func() {
		// 在这里，去查询数据库的状态
		// 校验代码不太可能是性能瓶颈，性能瓶颈一般在数据库
		// 也可以结合本地的 CPU，内存负载来判定
	}()
	val := &Validator[T]{base: base, target: target,
		l: l, p: p,
		direction: direction,
		highLoad:  highLoad}
	val.fromBase = val.fullFromBase
	return val
}

func (v *Validator[T]) SleepInterval(i time.Duration) *Validator[T] {
	v.sleepInterval = i
	return v

}

func (v *Validator[T]) Utime(utime int64) *Validator[T] {
	v.utime = utime
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

// Validate 调用者可以通过 ctx 来控制校验程序退出
// utime 上面至少要有一个索引，并且 utime 必须是第一列
// <utime, col1, col2>, <utime> 这种可以
// <utime, id> 然后执行 SELECT * FROM xx WHERE utime > ? ORDER BY id
func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		v.validateBaseToTarget(ctx)
		return nil
	})

	eg.Go(func() error {
		v.validateTargetToBase(ctx)
		return nil
	})
	return eg.Wait()
}

// Validate 调用者可以通过 ctx 来控制校验程序退出
// utime 上面至少要有一个索引，并且 utime 必须是第一列
// <utime, col1, col2>, <utime> 这种可以
// <utime, id> 然后执行 SELECT * FROM xx WHERE utime > ? ORDER BY id
func (v *Validator[T]) validateBaseToTarget(ctx context.Context) {
	offset := 0
	for {
		if v.highLoad.Load() {
			// 挂起
		}
		src, err := v.fromBase(ctx, offset)
		switch err {
		case context.Canceled, context.DeadlineExceeded:
			return
		case nil:
			// 在 base 中查询到了数据，现在去 target 中查找对应的数据
			var dst T
			err = v.target.Where("id = ?", src.ID()).First(&dst).Error
			switch err {
			case context.Canceled, context.DeadlineExceeded:
				return
			case nil:
				// 在 target 中找到了对应数据，开始比较
				// 原则上可以使用反射来比较reflect.DeepEqual(src, dst)
				//var srcAny any = src
				//if c1, ok := srcAny.(interface {
				//	// 有没有自定义的比较逻辑
				//	CompareTo(c2 migrator.Entity) bool
				//}); ok {
				//	// 有，我就用它的
				//	if !c1.CompareTo(dst) {
				//
				//	}
				//} else {
				//	// 没有，我就用反射
				//	if !reflect.DeepEqual(src, dst) {
				//
				//	}
				//}
				if !src.CompareTo(dst) {
					// 不相等，上报给 kafka
					v.notify(ctx, src.ID(), events.InconsistentEventTypeNEQ)
				}
			case gorm.ErrRecordNotFound:
				// target 缺少数据
				v.notify(ctx, src.ID(), events.InconsistentEventTypeTargetMissing)
			default:
				// 这里，要不要汇报，数据不一致？
				// 你有两种做法：
				// 1. 我认为，大概率数据是一致的，我记录一下日志，下一条
				v.l.Error("查询 target 数据失败", logger.Error(err))
				// 2. 我认为，出于保险起见，我应该报数据不一致，试着去修一下
				// 如果真的不一致了，没事，修它
				// 如果假的不一致（也就是数据一致），也没事，就是多余修了一次
				// 不好用哪个 InconsistentType
			}
		case gorm.ErrRecordNotFound:
			// 没有数据了，全量校验结束
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue
		default:
			// 数据库错误
			v.l.Error("校验数据，查询 base 出错",
				logger.Error(err))
			// 同时支持全量校验和增量校验，这里就不能直接返回
			// 有些情况下，用户希望退出，有些情况下。用户希望继续
			// 当用户希望继续的时候，sleep 一下
		}
		offset++
	}
}

// 理论上来说，可以利用 count 来加速这个过程，
// 我举个例子，假如说你初始化目标表的数据是 昨天的 23:59:59 导出来的
// 那么你可以 COUNT(*) WHERE ctime < 今天的零点，count 如果相等，就说明没删除
// 这一步大多数情况下效果很好，尤其是那些软删除的。
// 如果 count 不一致，那么接下来，你理论上来说，还可以分段 count
// 比如说，我先 count 第一个月的数据，一旦有数据删除了，你还得一条条查出来

func (v *Validator[T]) validateTargetToBase(ctx context.Context) {
	offset := 0
	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		var dstTs []T
		err := v.target.WithContext(dbCtx).
			Where("utime > ?", v.utime).Select("id").
			Offset(offset).Limit(v.batchSize).
			Order("utime").Find(&dstTs).Error
		cancel()
		if len(dstTs) == 0 {
			// 没数据了。直接返回
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		switch err {
		case context.Canceled, context.DeadlineExceeded:
			// 超时或者被人取消了
			return
		// 正常来说，gorm 在 Find 方法接收的是切片的时候，不会返回 gorm.ErrRecordNotFound
		case gorm.ErrRecordNotFound:
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
			continue
		case nil:
			ids := slice.Map(dstTs, func(idx int, t T) int64 {
				return t.ID()
			})
			var srcTs []T
			err = v.base.Where("id IN ?", ids).Find(&srcTs).Error
			switch err {
			case context.Canceled, context.DeadlineExceeded:
				// 超时或者被人取消了
				return
			case gorm.ErrRecordNotFound:
				// 这一批次全部丢失
				v.notifyBaseMissing(ctx, ids)
			case nil:
				srcIds := slice.Map(srcTs, func(idx int, t T) int64 {
					return t.ID()
				})
				// 计算差集，即 src 中缺失的
				diff := slice.DiffSet(ids, srcIds)
				v.notifyBaseMissing(ctx, diff)
			default:
				// 记录日志
			}
		default:
			// 记录日志，continue 掉
			v.l.Error("查询target 失败", logger.Error(err))
		}
		offset += len(dstTs)
		if len(dstTs) < v.batchSize {
			if v.sleepInterval <= 0 {
				return
			}
			time.Sleep(v.sleepInterval)
		}
	}
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Offset(offset).Order("id").First(&src).Error
	return src, err
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Where("utime > ?", v.utime).
		Offset(offset).Order("utime ASC, id ASC").First(&src).Error
	return src, err
}

func (v *Validator[T]) notify(ctx context.Context, id int64, typ string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	err := v.p.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Direction: v.direction,
		Type:      typ,
	})
	cancel()
	if err != nil {
		// 可以重试，但是重试也会失败，记日志，告警，手动去修
		// 也可以直接忽略，下一轮回再找出来继续上报
		v.l.Error("发送数据不一致消息失败", logger.Error(err))
	}
}

// 通用写法，摆脱对 T 的依赖
//func (v *Validator[T]) intrFromBaseV1(ctx context.Context, offset int) (T, error) {
//	rows, err := v.base.WithContext(dbCtx).
//		// 最好不要取等号
//		Where("utime > ?", v.utime).
//		Offset(offset).
//		Order("utime ASC, id ASC").Rows()
//	cols, err := rows.Columns()
//	// 所有列的值
//	vals := make([]any, len(cols))
//	rows.Scan(vals...)
//	return vals
//}

func (v *Validator[T]) notifyBaseMissing(ctx context.Context, ids []int64) {
	for _, id := range ids {
		v.notify(ctx, id, events.InconsistentEventTypeBaseMissing)
	}
}
