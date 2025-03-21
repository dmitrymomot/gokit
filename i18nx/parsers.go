package i18nx

import "context"

type Parser interface {
	Parse(ctx context.Context, content string) (map[string]map[string]any, error)
}
