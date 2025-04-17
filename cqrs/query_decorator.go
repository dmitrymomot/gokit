package cqrs

import (
	"context"
)

// QueryHandlerFunc is a function type that handles queries of type Q and returns a result R.
type QueryHandlerFunc[Q any, R any] func(ctx context.Context, query *Q) (R, error)

// QueryDecorator is a function that wraps a QueryHandlerFunc to add additional behavior.
type QueryDecorator[Q any, R any] func(QueryHandlerFunc[Q, R]) QueryHandlerFunc[Q, R]

// ApplyQueryDecorators applies multiple decorators to a QueryHandlerFunc.
// Decorators are applied in reverse order so that the first decorator
// in the list wraps the handler outermost.
func ApplyQueryDecorators[Q any, R any](
	handler QueryHandlerFunc[Q, R],
	decorators ...QueryDecorator[Q, R],
) QueryHandlerFunc[Q, R] {
	result := handler
	for i := len(decorators) - 1; i >= 0; i-- {
		result = decorators[i](result)
	}
	return result
}
