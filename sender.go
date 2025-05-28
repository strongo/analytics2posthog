package analytics2posthog

import (
	"context"
	"fmt"
	"github.com/posthog/posthog-go"
	"github.com/strongo/analytics"
	"github.com/strongo/logus"
)

var _ analytics.Sender = (*sender)(nil)

func NewSender(configs map[string]posthog.Config) (analytics.Sender, error) {
	phSender := sender{
		clients: make(map[string]posthog.Client, len(configs)),
	}
	for clientID, config := range configs {
		client, err := posthog.NewWithConfig("YOUR_PROJECT_API_KEY", config)
		if err != nil {
			return nil, fmt.Errorf("error creating posthog client for ClientID=%s: %v", clientID, err)
		}
		phSender.clients[clientID] = client
	}
	return &phSender, nil
}

var _ analytics.Sender = (*sender)(nil)

type sender struct {
	clients map[string]posthog.Client
}

func (s *sender) QueueMessage(ctx context.Context, message analytics.Message) {
	clientID := message.GetApiClientID()
	client := s.clients[clientID]
	m, err := capture(message)
	if err != nil {
		logus.Errorf(ctx, "capture error: %v", err)
	}
	if err = client.Enqueue(m); err != nil {
		logus.Errorf(ctx, "posthog enqueue error: %v", err)
	}
}
