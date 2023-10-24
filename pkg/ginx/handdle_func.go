package ginx

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	ijwt "webooktrial/internal/web/jwt"
)

func WrapReq[T any](fn func(ctx *gin.Context, req T, uc ijwt.UserClaims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 顺便把 userClaims 也取出来
		var req T
		if err := ctx.Bind(&req); err != nil {
			fmt.Println("Bind 失败:", err.Error())
			ctx.JSON(http.StatusOK, Result{
				Code: 5,
				Msg:  "参数不合法",
				Data: nil,
			})
			return
		}
		c, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := c.(ijwt.UserClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := fn(ctx, req, uc)
		if err != nil {
			Typ := reflect.TypeOf(fn)
			FnName := Typ.Name()
			fmt.Printf("函数调用失败%s\n, 错误信息%v:", FnName, err)
			ctx.JSON(http.StatusOK, res)
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

type Result struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
