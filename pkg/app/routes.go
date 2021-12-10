package app

func (h *Handler) SetupRoutes() {
	auth := h.Engin.Group("/auth")
	{
		auth.POST("/local/new", h.LocalRegister)
		auth.GET("/naver", h.NaverLogin)
		auth.GET("/naver/callback", h.NaverCallback)
		auth.POST("/local", h.LocalLogin)
		auth.DELETE("", h.Logout)
		auth.POST("/re", h.RefreshToken)
		auth.GET("/mail", h.SendEmail)
		auth.POST("/code", h.VerifyCode)
		auth.GET("/nickname/available", h.Available)
	}
	users := h.Engin.Group("/users")
	{
		users.DELETE("")
	}
}
