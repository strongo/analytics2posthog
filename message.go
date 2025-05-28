package analytics2posthog

import (
	"errors"
	"github.com/posthog/posthog-go"
	"github.com/strongo/analytics"
)

var errUnsupportedMessageType = errors.New("unsupported message type")

func capture(msg analytics.Message) (phm posthog.Message, err error) {
	distinctID := msg.User().UserID
	if distinctID == "" {
		err = errors.New("msg.User().UserID is empty string")
		return
	}
	properties := msg.Properties()
	props := make(posthog.Properties, len(properties)+1)
	for k, v := range properties {
		props[k] = v
	}
	switch m := msg.(type) {
	case analytics.Pageview:
		if title := m.Title(); title != "" {
			props.Set("title", title)
		}
	case analytics.Timing:
		props.Set("duration_ms", m.Duration().Milliseconds())
	default:
		err = errUnsupportedMessageType
		return
	}
	phm = posthog.Capture{
		Event:      msg.Event(),
		DistinctId: distinctID,
		Properties: props,
	}
	return
}
