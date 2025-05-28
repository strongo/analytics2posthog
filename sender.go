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
	var client posthog.Client
	if clientID == "" || clientID == DefaultClientID {
		if client = s.defaultClient; client == nil {
			return
		}
	} else if client = s.clients[clientID]; client == nil {
		s.logger.Warningf(ctx, "could not find PostHog client with ID='%s'", clientID)
		if client = s.defaultClient; client == nil {
			return
		}
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
