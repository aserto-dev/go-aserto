package internal

import (
	"context"
	"strings"
)

func ValueOrEmpty(ctx context.Context, key any) string {
	identity, ok := ctx.Value(key).(string)
	if ok {
		return identity
	}

	return ""
}

func ToPolicyPath(path string) string {
	return strings.ReplaceAll(strings.Trim(path, "/"), "/", ".")
}
