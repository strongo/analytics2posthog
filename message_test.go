package analytics2posthog

import (
	"errors"
	"testing"
	"time"

	"github.com/strongo/analytics"
)

type dummyMessage struct {
	analytics.Message
	user analytics.UserContext
}

func (d dummyMessage) Event() string                    { return "dummy" }
func (d dummyMessage) User() analytics.UserContext      { return d.user }
func (d dummyMessage) Properties() analytics.Properties { return nil }
func (d dummyMessage) Category() string                 { return "dummy_cat" }

func TestCapture(t *testing.T) {
	t.Run("panic_on_nil", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic on nil message")
			}
		}()
		_, _ = capture(nil)
	})

	t.Run("error_on_empty_user_id", func(t *testing.T) {
		u := analytics.NewUserContext("")
		msg := analytics.NewEvent("test_event", "cat1", "act1")
		msg.SetUserContext(u)
		_, err := capture(msg)
		if err == nil {
			t.Fatal("expected error on empty user ID")
		}
	})

	t.Run("pageview", func(t *testing.T) {
		u := analytics.NewUserContext("user123")
		u.SetUserLanguage("en-US")
		u.SetUserAgent("Go-http-client/1.1")

		pv := analytics.NewPageview("example.com", "/test/path")
		pv.SetUserContext(u)
		pv.SetCategory("web")
		pv.SetTitle("Test Page")
		pv.SetURL("https://example.com/test/path")
		pv.Properties()["custom_key"] = "custom_val"

		cap, err := capture(pv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cap.DistinctId != "user123" {
			t.Fatalf("expected distinctId 'user123', got %q", cap.DistinctId)
		}
		if cap.Properties["title"] != "Test Page" {
			t.Fatalf("expected title 'Test Page', got %v", cap.Properties["title"])
		}
		if cap.Properties["$current_url"] != "https://example.com/test/path" {
			t.Fatalf("expected $current_url, got %v", cap.Properties["$current_url"])
		}
		if cap.Properties["$host"] != "example.com" {
			t.Fatalf("expected $host, got %v", cap.Properties["$host"])
		}
		if cap.Properties["$pathname"] != "/test/path" {
			t.Fatalf("expected $pathname, got %v", cap.Properties["$pathname"])
		}
		if cap.Properties["category"] != "web" {
			t.Fatalf("expected category, got %v", cap.Properties["category"])
		}
		if cap.Properties["user_lang"] != "en-US" {
			t.Fatalf("expected user_lang, got %v", cap.Properties["user_lang"])
		}
		if cap.Properties["user_agent"] != "Go-http-client/1.1" {
			t.Fatalf("expected user_agent, got %v", cap.Properties["user_agent"])
		}
		if cap.Properties["custom_key"] != "custom_val" {
			t.Fatalf("expected custom_key, got %v", cap.Properties["custom_key"])
		}
	})

	t.Run("event", func(t *testing.T) {
		u := analytics.NewUserContext("user123")
		ev := analytics.NewEvent("click", "buttons", "submit")
		ev.SetUserContext(u)
		ev.SetLabel("checkout_btn")
		ev.SetValue(42)
		ev.SetTitle("Submit Button")

		cap, err := capture(ev)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cap.Properties["action"] != "submit" {
			t.Fatalf("expected action 'submit', got %v", cap.Properties["action"])
		}
		if cap.Properties["label"] != "checkout_btn" {
			t.Fatalf("expected label 'checkout_btn', got %v", cap.Properties["label"])
		}
		if cap.Properties["value"] != uint(42) {
			t.Fatalf("expected value 42, got %v", cap.Properties["value"])
		}
		if cap.Properties["title"] != "Submit Button" {
			t.Fatalf("expected title 'Submit Button', got %v", cap.Properties["title"])
		}
	})

	t.Run("timing", func(t *testing.T) {
		u := analytics.NewUserContext("user123")
		tm := analytics.NewTiming("load", 150*time.Millisecond)
		tm.SetUserContext(u)

		cap, err := capture(tm)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cap.Properties["duration_ms"] != int64(150) {
			t.Fatalf("expected duration_ms 150, got %v", cap.Properties["duration_ms"])
		}
	})

	t.Run("unsupported_message_type", func(t *testing.T) {
		u := analytics.NewUserContext("user123")
		dummy := dummyMessage{user: u}
		_, err := capture(dummy)
		if !errors.Is(err, errUnsupportedMessageType) {
			t.Fatalf("expected errUnsupportedMessageType, got %v", err)
		}
	})
}
