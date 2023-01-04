// Package login provides the Dto related to the Piped authentication.
package login

// LoginRequestDto represents the request payload needed to authenticate using the Piped Api.
type LoginRequestDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
