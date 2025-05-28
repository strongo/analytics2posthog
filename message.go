package analytics2posthog

import (
	"errors"
	"github.com/strongo/analytics"
)
import posthog "github.com/posthog/posthog-go"

var errUnsupportedMessageType = errors.New("unsupported message type")

func capture(msg analytics.Message) (phm posthog.Message, err error) {
	switch m := msg.(type) {
	case analytics.Pageview:
		props := make(map[string]any)
		if title := m.Title(); title != "" {
			props["title"] = title
		}
		phm = posthog.Capture{
			Event:      msg.Event(),
			Properties: props,
		}
	case analytics.Timing:
		phm = posthog.Capture{
			Event: msg.Event(),
			Properties: map[string]any{
				"duration_ms": m.Duration().Milliseconds(),
			},
		}
	default:
		err = errUnsupportedMessageType
	}
	return
}
