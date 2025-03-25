package middlewares

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// RecovererErrorHandlerFunc is a function that handles recovered panics.
// It takes the http.ResponseWriter, *http.Request, and the recovered value.
type RecovererErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err any)

// DefaultRecovererErrorHandler is the default handler for recovered panics.
// It logs the panic and returns a HTTP 500 Internal Server Error.
func DefaultRecovererErrorHandler(w http.ResponseWriter, r *http.Request, err any) {
	// Log the panic with backtrace
	slog.ErrorContext(r.Context(), "panic recovered", "error", err, "stacktrace", string(debug.Stack()))

	// Write the default error response
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Recoverer is a middleware that recovers from panics, logs the panic (and a
// backtrace), and returns a HTTP 500 (Internal Server Error) status if
// possible. Recoverer prints a request ConfirmationID if one is provided.
//
// If a custom error handler is provided, it will be called to handle the error.
// Otherwise, the default error handler will be used.
func Recoverer(next http.Handler) http.Handler {
	return RecovererWithHandler(next, nil)
}

// RecovererWithHandler is a middleware that recovers from panics and calls the
// provided error handler function to handle the error. If no handler is provided,
// the default handler is used which logs the panic and returns a HTTP 500 status.
func RecovererWithHandler(next http.Handler, handler RecovererErrorHandlerFunc) http.Handler {
	// Use default handler if none provided
	if handler == nil {
		handler = DefaultRecovererErrorHandler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Check for http.ErrAbortHandler which shouldn't be recovered
				if typedErr, ok := err.(error); ok && errors.Is(typedErr, http.ErrAbortHandler) {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(typedErr)
				}

				// Call the error handler
				handler(w, r, err)
			}
		}()

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
