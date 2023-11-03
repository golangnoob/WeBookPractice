package service

import (
	"context"
	"fmt"
	"math/rand"

	"webooktrial/internal/repository"
	"webooktrial/internal/service/sms"
)

//go:generate mockgen -source=./code.go -package=svcmocks -destination=mocks/code.mock.go CodeService
const codeTplId = "1777556"

var (
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
)

type CodeService interface {
	Send(ctx context.Context,
		// 区别业务场景
		biz string, phone string) error
	Verify(ctx context.Context, biz string,
		phone string, inputCode string) (bool, error)
}

type CodeSCService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
	//tplId string
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &CodeSCService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

// Send 发验证码，我需要什么参数？
func (svc *CodeSCService) Send(ctx context.Context, biz, phone string) error {
	// 生成一个验证码
	code := svc.generateCode()
	// 存储到 redis 中
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 存储成功，然后发送出去
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	if err != nil {
		err = fmt.Errorf("发送短信出现异常 %w", err)
	}
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

func (svc *CodeSCService) Verify(ctx context.Context, biz string,
	phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeSCService) generateCode() string {
	// 六位数，num 在 0, 999999 之间，包含 0 和 999999
	num := rand.Intn(1000000)
	// 不够六位的，加上前导 0
	// 000001
	return fmt.Sprintf("%06d", num)
}
