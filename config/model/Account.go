package model

type Account struct {
	Username string `validate:"required"`
	Password string `validate:"required"`
}
