package cqrs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuery represents a sample query for testing.
type TestQuery struct {
	Input int
}

// ErrQueryFailed is a sample error for testing error flows.
var ErrQueryFailed = errors.New("query failed")

func TestApplyQueryDecorators(t *testing.T) {
	t.Run("single decorator", func(t *testing.T) {
		called := false
		base := func(ctx context.Context, q *TestQuery) (int, error) {
			return 0, nil
		}

		decor := func(next cqrs.QueryHandlerFunc[TestQuery, int]) cqrs.QueryHandlerFunc[TestQuery, int] {
			return func(ctx context.Context, q *TestQuery) (int, error) {
				called = true
				return next(ctx, q)
			}
		}

		h := cqrs.ApplyQueryDecorators(base, decor)
		res, err := h(context.Background(), &TestQuery{Input: 42})

		require.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, 0, res)
	})

	t.Run("multiple decorators in order", func(t *testing.T) {
		order := []string{}
		base := func(ctx context.Context, q *TestQuery) (int, error) {
			order = append(order, "base")
			return q.Input, nil
		}

		first := func(next cqrs.QueryHandlerFunc[TestQuery, int]) cqrs.QueryHandlerFunc[TestQuery, int] {
			return func(ctx context.Context, q *TestQuery) (int, error) {
				order = append(order, "first")
				return next(ctx, q)
			}
		}

		second := func(next cqrs.QueryHandlerFunc[TestQuery, int]) cqrs.QueryHandlerFunc[TestQuery, int] {
			return func(ctx context.Context, q *TestQuery) (int, error) {
				order = append(order, "second")
				return next(ctx, q)
			}
		}

		h := cqrs.ApplyQueryDecorators(base, first, second)
		res, err := h(context.Background(), &TestQuery{Input: 5})

		require.NoError(t, err)
		assert.Equal(t, 5, res)
		assert.Equal(t, []string{"first", "second", "base"}, order)
	})

	t.Run("error handling in decorators", func(t *testing.T) {
		called := false
		base := func(ctx context.Context, q *TestQuery) (int, error) {
			return 0, ErrQueryFailed
		}

		decor := func(next cqrs.QueryHandlerFunc[TestQuery, int]) cqrs.QueryHandlerFunc[TestQuery, int] {
			return func(ctx context.Context, q *TestQuery) (int, error) {
				res, err := next(ctx, q)
				if err != nil {
					called = true
				}
				return res, err
			}
		}

		h := cqrs.ApplyQueryDecorators(base, decor)
		res, err := h(context.Background(), &TestQuery{Input: 0})

		require.Error(t, err)
		assert.Equal(t, ErrQueryFailed, err)
		assert.True(t, called)
		assert.Equal(t, 0, res)
	})

	t.Run("decorator modifying query", func(t *testing.T) {
		var got int
		base := func(ctx context.Context, q *TestQuery) (int, error) {
			got = q.Input
			return q.Input, nil
		}

		decor := func(next cqrs.QueryHandlerFunc[TestQuery, int]) cqrs.QueryHandlerFunc[TestQuery, int] {
			return func(ctx context.Context, q *TestQuery) (int, error) {
				q.Input += 3
				return next(ctx, q)
			}
		}

		h := cqrs.ApplyQueryDecorators(base, decor)
		res, err := h(context.Background(), &TestQuery{Input: 2})

		require.NoError(t, err)
		assert.Equal(t, 5, res)
		assert.Equal(t, 5, got)
	})
}
