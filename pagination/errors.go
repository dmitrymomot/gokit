package pagination

import "errors"

// Predefined errors for the pagination package.
var (
	ErrInvalidLimit       = errors.New("invalid limit value")
	ErrInvalidOffset      = errors.New("invalid offset value")
	ErrInvalidCursor      = errors.New("invalid cursor value")
	ErrInvalidPage        = errors.New("invalid page number")
	ErrInvalidSize        = errors.New("invalid page size")
	ErrMissingBaseURL     = errors.New("base URL is required for link generation")
	ErrNegativeTotalItems = errors.New("total items cannot be negative")
	ErrInvalidQueryParam  = errors.New("invalid query parameter")
)
