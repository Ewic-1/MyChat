package request

type RefreshTokenRequest struct {
	// RefreshToken 由登录/刷新接口下发。
	RefreshToken string `json:"refresh_token" binding:"required"`
}
