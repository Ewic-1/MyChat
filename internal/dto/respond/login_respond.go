package respond

type LoginRespond struct {
	Uuid      string `json:"uuid"`
	Nickname  string `json:"nickname"`
	Telephone string `json:"telephone"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	Gender    int8   `json:"gender"`
	Birthday  string `json:"birthday"`
	Signature string `json:"signature"`
	CreatedAt string `json:"created_at"`
	IsAdmin   int8   `json:"is_admin"`
	Status    int8   `json:"status"`
	// Token 为 access token；RefreshToken 用于续期。
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	// 过期时间使用 Unix 秒时间戳，便于前端直接比较。
	AccessExpiresAt  int64 `json:"access_expires_at"`
	RefreshExpiresAt int64 `json:"refresh_expires_at"`
}
