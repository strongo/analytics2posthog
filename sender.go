package analytics2posthog

import (
	"context"
	"errors"
	"fmt"
	"github.com/posthog/posthog-go"
	"github.com/strongo/analytics"
)

const DefaultClientID = "default"

var _ analytics.Sender = (*sender)(nil)

type Client struct {
	ID string
	//
	ApiKey string
	Config posthog.Config
}

func NewSender(logger Logger, clients ...Client) (analytics.Sender, error) {
	if len(clients) == 0 {
		return nil, errors.New("no clients provided")
	}
	phSender := sender{
		logger:  logger,
		clients: make(map[string]posthog.Client, len(clients)),
	}
	for i, client := range clients {
		phClient, err := posthog.NewWithConfig(client.ApiKey, client.Config)
		if err != nil {
			return nil, fmt.Errorf("error creating posthog phClient #%d, ID=%s: %w", i+1, client.ID, err)
		}
		if client.ID == DefaultClientID {
			phSender.defaultClient = phClient
		}
		if _, ok := phSender.clients[client.ID]; ok {
			return nil, fmt.Errorf("duplicate posthog phClient #%d, ID=%s", i+1, client.ID)
		}
		phSender.clients[client.ID] = phClient
	}
	return &phSender, nil
}

var _ analytics.Sender = (*sender)(nil)

type sender struct {
	logger        Logger
	defaultClient posthog.Client
	clients       map[string]posthog.Client
}

func (s *sender) QueueMessage(ctx context.Context, message analytics.Message) {
	clientID := message.GetApiClientID()
	client := s.clients[clientID]
	if client == nil {
		client = s.defaultClient
	}
	if client == nil {
		s.logger.Warningf(ctx, "Could not find PostHog client with ID=%s", clientID)
		return
	}
	m, err := capture(message)
	if err != nil {
		s.logger.Errorf(ctx, "capture error: %v", err)
		return
	}
	if err = client.Enqueue(m); err != nil {
		s.logger.Errorf(ctx, "posthog enqueue error: %v", err)
	}
}
