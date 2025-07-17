package analytics2posthog

import (
	"errors"
	"github.com/posthog/posthog-go"
	"github.com/strongo/analytics"
)

var errUnsupportedMessageType = errors.New("unsupported message type")

func capture(msg analytics.Message) (phc posthog.Capture, err error) {
	if msg == nil {
		panic("analytics2posthog.capture(nil)")
	}
	phc = posthog.Capture{
		DistinctId: msg.User().GetUserID(),
		Event:      msg.Event(),
	}
	if phc.DistinctId == "" {
		err = errors.New("msg.User().UserID is empty string")
		return
	}

	phc.Properties = mapCustomProperties(msg.Properties())

	if err = mapGenericProps(msg, phc.Properties); err != nil {
		return
	}
	return
}

func mapCustomProperties(properties analytics.Properties) (props posthog.Properties) {
	props = make(posthog.Properties, len(properties)+10)
	for k, v := range properties {
		props[k] = v
	}
	return
}

func mapGenericProps(msg analytics.Message, props posthog.Properties) (err error) {
	if category := msg.Category(); category != "" {
		props.Set("category", category)
	}
	if user := msg.User(); user != nil {
		if lang := user.GetUserLanguage(); lang != "" {
			props.Set("user_lang", lang)
		}
		if userAgent := user.GetUserAgent(); userAgent != "" {
			props.Set("user_agent", userAgent)
		}
	}
	switch m := msg.(type) {
	case analytics.Pageview:
		if title := m.Title(); title != "" {
			props.Set("title", title)
		}
		if url := m.URL(); url != "" {
			props.Set("$current_url", url)
		}
		if host := m.Host(); host != "" {
			props.Set("$host", host)
		}
		if path := m.Path(); path != "" {
			props.Set("$pathname", path)
		}
	case analytics.Event:
		if action := m.Action(); action != "" {
			props.Set("action", action)
		}
		if label := m.Label(); label != "" {
			props.Set("label", label)
		}
		if value := m.Value(); value != 0 {
			props.Set("value", value)
		}
		if title := m.Title(); title != "" {
			props.Set("title", title)
		}
	case analytics.Timing:
		props.Set("duration_ms", m.Duration().Milliseconds())
	default:
		return errUnsupportedMessageType
	}
	return nil
}
