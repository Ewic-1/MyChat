package request

type LoginRequest struct {
	Telephone string `json:"telephone" binding:"required"`
	Password  string `json:"password" binding:"required"`
}
