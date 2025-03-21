package google

import "errors"

var (
	ErrFailedToGetProfile   = errors.New("failed to get profile")
	ErrInvalidState         = errors.New("invalid oauth state")
	ErrFailedToExchangeCode = errors.New("failed to exchange code")
	ErrAccountNotVerified   = errors.New("account is not verified")
	ErrFailedToSaveSession  = errors.New("failed to save session")
)
