package ioc

import (
	"os"

	"webooktrial/internal/service/oauth2/wechat"
	"webooktrial/internal/web"
)

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_ID ")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewService(appId, appKey)
}

func NewWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{
		Secure: false,
	}
}
