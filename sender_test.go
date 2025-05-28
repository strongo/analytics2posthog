package analytics2posthog

import (
	"context"
	"github.com/posthog/posthog-go"
	"testing"
)

func TestNewSender(t *testing.T) {
	type args struct {
		clients []Client
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default_client",
			args: args{
				clients: []Client{
					{
						ID:     DefaultClientID,
						ApiKey: "some-api-key",
						Config: posthog.Config{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSender(NoOpLogger{}, tt.args.clients...)
			if err != nil {
				t.Errorf("NewSender() returned unexpected error = %v", err)
			}
			if got == nil {
				t.Error("NewSender() got is nil")
			}
		})
	}
}

var _ Logger = (*NoOpLogger)(nil)

type NoOpLogger struct{}

func (n NoOpLogger) Warningf(_ context.Context, _ string, _ ...any) {
}

func (n NoOpLogger) Errorf(_ context.Context, _ string, _ ...any) {
}
