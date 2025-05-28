package analytics2posthog

import "context"

type Logger interface {
	Warningf(ctx context.Context, format string, args ...any)
	Errorf(ctx context.Context, format string, args ...any)
}
