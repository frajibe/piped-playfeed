package login

type LoginRequestDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
