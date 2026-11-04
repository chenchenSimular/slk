package main

import (
	"context"
	"errors"
	"testing"
)

type fakeMuter struct {
	channelID string
	muted     bool
	err       error
	hasDL     bool
}

func (f *fakeMuter) SetChannelMuted(ctx context.Context, channelID string, muted bool) error {
	f.channelID, f.muted = channelID, muted
	_, f.hasDL = ctx.Deadline()
	return f.err
}

func TestSetMutedReportsTheResult(t *testing.T) {
	f := &fakeMuter{}
	msg := setMuted(context.Background(), f, "T1", "C1", true)
	if f.channelID != "C1" || !f.muted || !f.hasDL {
		t.Errorf("call = %+v, want C1 muted with a deadline", f)
	}
	if msg.TeamID != "T1" || msg.ChannelID != "C1" || !msg.Muted || msg.Err != nil {
		t.Errorf("msg = %+v", msg)
	}

	f.err = errors.New("users.prefs.setNotifications: invalid_arguments")
	if msg := setMuted(context.Background(), f, "T1", "C1", false); !errors.Is(msg.Err, f.err) || msg.Muted {
		t.Errorf("msg = %+v, want the client's error", msg)
	}
}

func TestSetMutedWithoutWorkspaceFails(t *testing.T) {
	if msg := setMuted(context.Background(), nil, "T9", "C1", true); msg.Err == nil {
		t.Error("no error for a workspace that is gone")
	}
}
