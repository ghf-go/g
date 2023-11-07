package g

// 管理后台账号登录验证中间件
func AdminCheckoutLoginMiddwleWare(g *GContext) {
	if g.GetAdminId() <= 0 {
		g.WebJsonFail(-1, "账号没有登录")
		return
	}
	g.Next()
}
