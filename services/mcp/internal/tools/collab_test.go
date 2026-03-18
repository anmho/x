package tools

import (
	"path/filepath"
	"testing"
)

func TestCollabStoreChannelLifecycle(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	channel, err := store.GetOrCreateChannel("task:alpha", "Alpha", []string{"agent-2", "agent-1"}, map[string]string{"task": "alpha"})
	if err != nil {
		t.Fatalf("GetOrCreateChannel(create): %v", err)
	}
	if channel.Key != "task:alpha" {
		t.Fatalf("unexpected key: %s", channel.Key)
	}
	if len(channel.Participants) != 2 || channel.Participants[0] != "agent-1" {
		t.Fatalf("participants not normalized: %#v", channel.Participants)
	}

	same, err := store.GetOrCreateChannel("task:alpha", "", []string{"agent-3"}, nil)
	if err != nil {
		t.Fatalf("GetOrCreateChannel(update): %v", err)
	}
	if same.ID != channel.ID {
		t.Fatalf("expected same channel id, got %s vs %s", same.ID, channel.ID)
	}

	found, err := store.ListChannels("agent-3", "", 10)
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}
	if len(found) != 1 || found[0].ID != channel.ID {
		t.Fatalf("expected channel filtered by participant, got %#v", found)
	}
}

func TestCollabStoreMessages(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	channel, err := store.GetOrCreateChannel("task:beta", "Beta", []string{"agent-a"}, nil)
	if err != nil {
		t.Fatalf("GetOrCreateChannel: %v", err)
	}

	if _, err := store.PostMessage(channel.ID, "agent-a", "message", "hello", map[string]string{"priority": "high"}); err != nil {
		t.Fatalf("PostMessage(1): %v", err)
	}
	if _, err := store.PostMessage(channel.ID, "agent-b", "status", "working", nil); err != nil {
		t.Fatalf("PostMessage(2): %v", err)
	}

	messages, err := store.ReadMessages(channel.ID, 0, 10)
	if err != nil {
		t.Fatalf("ReadMessages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Sequence >= messages[1].Sequence {
		t.Fatalf("expected ordered sequence values, got %#v", messages)
	}
	if messages[0].Body != "hello" || messages[1].Sender != "agent-b" {
		t.Fatalf("unexpected messages: %#v", messages)
	}

	replay, err := store.ReadMessages(channel.ID, messages[0].Sequence, 10)
	if err != nil {
		t.Fatalf("ReadMessages replay: %v", err)
	}
	if len(replay) != 1 || replay[0].Sender != "agent-b" {
		t.Fatalf("unexpected replay result: %#v", replay)
	}

	channels, err := store.ListChannels("agent-b", "", 10)
	if err != nil {
		t.Fatalf("ListChannels after post: %v", err)
	}
	if len(channels) != 1 || channels[0].MessageCount != 2 {
		t.Fatalf("unexpected channel summaries: %#v", channels)
	}
}

func TestGetChannelForAgent(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	channel, err := store.GetChannelForAgent("agent-mail")
	if err != nil {
		t.Fatalf("GetChannelForAgent first: %v", err)
	}
	same, err := store.GetChannelForAgent("agent-mail")
	if err != nil {
		t.Fatalf("GetChannelForAgent second: %v", err)
	}
	if channel.ID != same.ID {
		t.Fatalf("expected stable channel, got %s and %s", channel.ID, same.ID)
	}
	if channel.Key != "agent:agent-mail" {
		t.Fatalf("unexpected key: %s", channel.Key)
	}
}
