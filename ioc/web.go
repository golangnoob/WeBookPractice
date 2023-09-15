package ioc

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"webooktrial/internal/web"
	ijwt "webooktrial/internal/web/jwt"
	"webooktrial/internal/web/middleware"
	"webooktrial/pkg/ginx/middlewares/ratelimit"
	ratelimit2 "webooktrial/pkg/ratelimit"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler,
	oauth2WechatHandler *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oauth2WechatHandler.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		IgnorePathsHdl(jwtHdl),
		ratelimit.NewBuilder(ratelimit2.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 允许前端获取 Header 中 key 为"x-jwt-token"和"x-refresh-token"的val
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
		// 是否允许你带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 你的开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	})
}

func IgnorePathsHdl(jwtHdl ijwt.Handler) gin.HandlerFunc {
	return middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/oauth2/wechat/authurl").
		IgnorePaths("/oauth2/wechat/callback").
		IgnorePaths("/users/login_sms").
		IgnorePaths("users/refresh_token").Build()
}
