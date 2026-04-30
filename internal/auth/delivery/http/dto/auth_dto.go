package dto

type LoginAndRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
