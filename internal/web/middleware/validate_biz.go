package middleware

//func Build() gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		// order id/order sn
//		bizId := ctx.GetHeader("biz_id")
//		// order
//		biz := ctx.GetHeader("biz")
//		uc := ctx.MustGet("user").(jwt.UserClaims)
//      单体应用就是数据库，
//      微服务呢？调用微服务 - 做客户端缓存
//		validate(biz, bizId, uc.Uid)
//	}
//}
