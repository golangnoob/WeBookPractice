package service

import (
	"context"
	"fmt"
	"math/rand"

	"webooktrial/internal/repository"
	"webooktrial/internal/service/sms"
)

const codeTplId = "1777556"

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
}

// Send 发验证码，我需要什么参数？
func (svc *CodeService) Send(ctx context.Context, biz, phone string) error {
	// 生成一个验证码
	code := svc.generateCode()
	// 存储到 redis 中
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 存储成功，然后发送出去
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	//if err != nil {
	// 这个地方怎么办？
	// 这意味着，Redis 有这个验证码，但是不好意思，
	// 我能不能删掉这个验证码？
	// 你这个 err 可能是超时的 err，你都不知道，发出了没
	// 在这里重试
	// 要重试的话，初始化的时候，传入一个自己就会重试的 smsSvc
	//}
	return err
}

func (svc *CodeService) Verify(ctx context.Context, biz string,
	phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	// 六位数，num 在 0, 999999 之间，包含 0 和 999999
	num := rand.Intn(1000000)
	// 不够六位的，加上前导 0
	// 000001
	return fmt.Sprintf("%6d", num)
}
