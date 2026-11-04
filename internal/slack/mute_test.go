package slackclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/slack-go/slack"
)

// The form is the web client's multi-pref users.prefs.setNotifications
// shape, verified by hand on 2026-11-04 against a real workspace (an
// xoxc token): it muted a channel, and users.prefs.get's
// all_notifications_prefs then listed it as muted. The single-pref form
// (name=muted, value=true, channel_id=…) was answered invalid_arguments.
func TestSetChannelMuted_PostsWebClientForm(t *testing.T) {
	for _, muted := range []bool{true, false} {
		c, got, _ := deleteAttachmentServer(t, http.StatusOK, `{"ok":true}`)
		if err := c.SetChannelMuted(context.Background(), "C1", muted); err != nil {
			t.Fatalf("SetChannelMuted(%v): %v", muted, err)
		}
		if got.URL.Path != "/api/users.prefs.setNotifications" {
			t.Errorf("path = %q", got.URL.Path)
		}
		for field, want := range map[string]string{
			"global":      "false",
			"channel_ids": "C1",
			"sync":        "false",
			"token":       "xoxc-test",
			"_x_reason":   "prefs-store/setMultiChannelNotificationOverride",
		} {
			if v := got.PostForm.Get(field); v != want {
				t.Errorf("muted=%v: form %s = %q, want %q", muted, field, v, want)
			}
		}
		var prefs []map[string]string
		if err := json.Unmarshal([]byte(got.PostForm.Get("prefs")), &prefs); err != nil {
			t.Fatalf("prefs not JSON: %v", err)
		}
		want := "false"
		if muted {
			want = "true"
		}
		if len(prefs) != 1 || prefs[0]["name"] != "muted" || prefs[0]["value"] != want {
			t.Errorf("muted=%v: prefs = %v", muted, prefs)
		}
		for _, legacy := range []string{"name", "value", "channel_id"} {
			if got.PostForm.Has(legacy) {
				t.Errorf("legacy single-pref field %q sent", legacy)
			}
		}
	}
}

func TestSetChannelMuted_RefusalIsTyped(t *testing.T) {
	c, _, _ := deleteAttachmentServer(t, http.StatusOK, `{"ok":false,"error":"invalid_arguments"}`)
	err := c.SetChannelMuted(context.Background(), "C1", true)
	var rej *MuteRejectedError
	if !errors.As(err, &rej) || rej.Code != "invalid_arguments" {
		t.Fatalf("err = %v, want *MuteRejectedError{invalid_arguments}", err)
	}
}

func TestSetChannelMuted_RateLimitIsSlackError(t *testing.T) {
	c, _, calls := deleteAttachmentServer(t, http.StatusTooManyRequests, ``)
	err := c.SetChannelMuted(context.Background(), "C1", true)
	var rl *slack.RateLimitedError
	if !errors.As(err, &rl) {
		t.Fatalf("err = %v, want *slack.RateLimitedError", err)
	}
	if *calls != 1 {
		t.Errorf("%d calls, want 1 (no retry)", *calls)
	}
}

func TestSetChannelMuted_EmptyChannelSendsNothing(t *testing.T) {
	c, _, calls := deleteAttachmentServer(t, http.StatusOK, `{"ok":true}`)
	if err := c.SetChannelMuted(context.Background(), "", true); err == nil {
		t.Fatal("no error for an empty channel")
	}
	if *calls != 0 {
		t.Errorf("%d calls, want 0", *calls)
	}
}
