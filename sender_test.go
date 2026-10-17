package analytics2posthog

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/posthog/posthog-go"
	"github.com/strongo/analytics"
)

type mockPosthogClient struct {
	posthog.Client
	enqueueFunc func(posthog.Message) error
}

func (m *mockPosthogClient) Enqueue(msg posthog.Message) error {
	if m.enqueueFunc != nil {
		return m.enqueueFunc(msg)
	}
	return nil
}

type recordingLogger struct {
	warnings []string
	errors   []string
}

func (r *recordingLogger) Warningf(_ context.Context, format string, args ...any) {
	r.warnings = append(r.warnings, fmt.Sprintf(format, args...))
}

func (r *recordingLogger) Errorf(_ context.Context, format string, args ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

func TestNewSender_Errors(t *testing.T) {
	t.Run("no_clients", func(t *testing.T) {
		_, err := NewSender(&recordingLogger{})
		if err == nil || err.Error() != "no clients provided" {
			t.Fatalf("expected 'no clients provided', got %v", err)
		}
	})

	t.Run("client_creation_error", func(t *testing.T) {
		origNew := newPosthogClient
		defer func() { newPosthogClient = origNew }()

		newPosthogClient = func(apiKey string, config posthog.Config) (posthog.Client, error) {
			return nil, errors.New("fail_posthog_client")
		}

		_, err := NewSender(&recordingLogger{}, Client{ID: "c1", ApiKey: "key"})
		if err == nil {
			t.Fatal("expected error creating posthog client")
		}
	})

	t.Run("duplicate_client", func(t *testing.T) {
		origNew := newPosthogClient
		defer func() { newPosthogClient = origNew }()

		newPosthogClient = func(apiKey string, config posthog.Config) (posthog.Client, error) {
			return &mockPosthogClient{}, nil
		}

		_, err := NewSender(&recordingLogger{},
			Client{ID: "c1", ApiKey: "key1"},
			Client{ID: "c1", ApiKey: "key2"},
		)
		if err == nil {
			t.Fatal("expected duplicate client error")
		}
	})

	t.Run("success_with_default_and_custom", func(t *testing.T) {
		origNew := newPosthogClient
		defer func() { newPosthogClient = origNew }()

		newPosthogClient = func(apiKey string, config posthog.Config) (posthog.Client, error) {
			return &mockPosthogClient{}, nil
		}

		s, err := NewSender(&recordingLogger{},
			Client{ID: DefaultClientID, ApiKey: "key_def"},
			Client{ID: "client_custom", ApiKey: "key_custom"},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s == nil {
			t.Fatal("expected non-nil sender")
		}
	})
}

type anEvent interface {
	analytics.Event
}

type mockEventWithClientID struct {
	anEvent
	clientID string
}

func (m mockEventWithClientID) GetApiClientID() string {
	return m.clientID
}

func TestQueueMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("no_default_client_returns", func(t *testing.T) {
		s := &sender{
			logger:  &recordingLogger{},
			clients: make(map[string]posthog.Client),
		}
		msg := analytics.NewEvent("e1", "cat1", "act1")
		// Should return cleanly without panic or logging
		s.QueueMessage(ctx, msg)
	})

	t.Run("unknown_client_no_default_client", func(t *testing.T) {
		logger := &recordingLogger{}
		s := &sender{
			logger:  logger,
			clients: make(map[string]posthog.Client),
		}
		msg := mockEventWithClientID{
			anEvent:  analytics.NewEvent("e1", "cat1", "act1"),
			clientID: "unknown_client",
		}
		s.QueueMessage(ctx, msg)
		if len(logger.warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d", len(logger.warnings))
		}
	})

	t.Run("unknown_client_fallback_to_default", func(t *testing.T) {
		logger := &recordingLogger{}
		enqueued := false
		defClient := &mockPosthogClient{
			enqueueFunc: func(msg posthog.Message) error {
				enqueued = true
				return nil
			},
		}
		s := &sender{
			logger:        logger,
			defaultClient: defClient,
			clients:       make(map[string]posthog.Client),
		}
		ev := analytics.NewEvent("e1", "cat1", "act1")
		ev.SetUserContext(analytics.NewUserContext("user1"))
		msg := mockEventWithClientID{
			anEvent:  ev,
			clientID: "unknown_client",
		}
		s.QueueMessage(ctx, msg)
		if len(logger.warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d", len(logger.warnings))
		}
		if !enqueued {
			t.Fatal("expected message to be enqueued to defaultClient")
		}
	})

	t.Run("capture_error_pageview", func(t *testing.T) {
		logger := &recordingLogger{}
		defClient := &mockPosthogClient{}
		s := &sender{
			logger:        logger,
			defaultClient: defClient,
			clients:       make(map[string]posthog.Client),
		}
		// Pageview with empty user ID causes capture error
		pv := analytics.NewPageview("host.com", "/home")
		pv.SetUserContext(analytics.NewUserContext(""))
		s.QueueMessage(ctx, pv)
		if len(logger.errors) != 1 {
			t.Fatalf("expected 1 error logged, got %d", len(logger.errors))
		}
	})

	t.Run("capture_error_event", func(t *testing.T) {
		logger := &recordingLogger{}
		defClient := &mockPosthogClient{}
		s := &sender{
			logger:        logger,
			defaultClient: defClient,
			clients:       make(map[string]posthog.Client),
		}
		// Event with empty user ID causes capture error
		ev := analytics.NewEvent("click", "cat", "act")
		ev.SetUserContext(analytics.NewUserContext(""))
		s.QueueMessage(ctx, ev)
		if len(logger.errors) != 1 {
			t.Fatalf("expected 1 error logged, got %d", len(logger.errors))
		}
	})

	t.Run("enqueue_error", func(t *testing.T) {
		logger := &recordingLogger{}
		clientErr := &mockPosthogClient{
			enqueueFunc: func(msg posthog.Message) error {
				return errors.New("fail_enqueue")
			},
		}
		s := &sender{
			logger:        logger,
			defaultClient: clientErr,
			clients:       map[string]posthog.Client{"custom": clientErr},
		}
		ev := analytics.NewEvent("e1", "cat1", "act1")
		ev.SetUserContext(analytics.NewUserContext("user1"))
		msg := mockEventWithClientID{
			anEvent:  ev,
			clientID: "custom",
		}
		s.QueueMessage(ctx, msg)
		if len(logger.errors) != 1 {
			t.Fatalf("expected 1 error logged, got %d", len(logger.errors))
		}
	})
}
