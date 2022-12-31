// Package login provides the Dto related to the Piped authentication.
package login

// LoginResponseDto represents the request response corresponding to a valid authentication through the Piped Api.
type LoginResponseDto struct {
	Token string
}
