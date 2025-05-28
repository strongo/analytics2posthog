package analytics2posthog

import "github.com/strongo/analytics"
import posthog "github.com/posthog/posthog-go"

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
	}
	return
}
