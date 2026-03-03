package respond

type TokenRespond struct {
	// 刷新接口返回的新 token 对。
	Token            string `json:"token"`
	RefreshToken     string `json:"refresh_token"`
	AccessExpiresAt  int64  `json:"access_expires_at"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
}
