package ioc

import (
	"webooktrial/internal/service/sms"
	"webooktrial/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
