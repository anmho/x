package tools

import (
	"path/filepath"
	"testing"
	"time"
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

func TestCollabStoreLinearIssueAndExternalRefs(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	channel, err := store.GetOrCreateLinearIssueChannel(&LinearIssue{
		Identifier: "ANM-190",
		Title:      "Slack and Linear bridge",
		URL:        "https://linear.app/anmho/issue/ANM-190/test",
		StateName:  "In Progress",
		TeamKey:    "ANM",
		TeamName:   "Anmho",
	}, []string{"agent-a"}, map[string]string{"scope": "task"})
	if err != nil {
		t.Fatalf("GetOrCreateLinearIssueChannel: %v", err)
	}
	if channel.Key != "linear:ANM-190" {
		t.Fatalf("unexpected key: %s", channel.Key)
	}
	if channel.Metadata["linear_issue_id"] != "ANM-190" {
		t.Fatalf("missing linear metadata: %#v", channel.Metadata)
	}
	if len(channel.ExternalRefs) != 1 || channel.ExternalRefs[0].Source != "linear" {
		t.Fatalf("expected linear external ref: %#v", channel.ExternalRefs)
	}

	channel, err = store.LinkExternalRef(channel.ID, &CollabExternalRef{
		Source:     "slack",
		ExternalID: "1742800000.100000",
		ParentID:   "C123",
		Title:      "Bridge thread",
	})
	if err != nil {
		t.Fatalf("LinkExternalRef: %v", err)
	}
	if len(channel.ExternalRefs) != 2 {
		t.Fatalf("expected 2 external refs, got %#v", channel.ExternalRefs)
	}

	resolved, err := store.GetChannelByExternalRef("slack", "1742800000.100000", "C123")
	if err != nil {
		t.Fatalf("GetChannelByExternalRef: %v", err)
	}
	if resolved.ID != channel.ID {
		t.Fatalf("expected same channel, got %s vs %s", resolved.ID, channel.ID)
	}

	resolved, err = store.MarkChannelStatus(channel.ID, "waiting_approval", map[string]string{"status_reason": "waiting on human"})
	if err != nil {
		t.Fatalf("MarkChannelStatus: %v", err)
	}
	if resolved.Status != "waiting_approval" {
		t.Fatalf("unexpected status: %s", resolved.Status)
	}
}

func TestCollabStoreGetOrCreateExternalChannel(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	channel, err := store.GetOrCreateExternalChannel(
		"github",
		"anmho/x#212",
		"pull_request",
		"anmho/x#212 · external adapters",
		"https://github.com/anmho/x/pull/212",
		[]string{"agent-a"},
		map[string]string{"scope": "review"},
		map[string]string{"kind": "pull_request"},
	)
	if err != nil {
		t.Fatalf("GetOrCreateExternalChannel(create): %v", err)
	}
	if channel.Key != "github:anmho/x#212" {
		t.Fatalf("unexpected key: %s", channel.Key)
	}
	if len(channel.ExternalRefs) != 1 || channel.ExternalRefs[0].Source != "github" {
		t.Fatalf("expected github external ref, got %#v", channel.ExternalRefs)
	}

	same, err := store.GetOrCreateExternalChannel(
		"github",
		"anmho/x#212",
		"pull_request",
		"updated title",
		"",
		[]string{"agent-b"},
		map[string]string{"owner": "platform"},
		nil,
	)
	if err != nil {
		t.Fatalf("GetOrCreateExternalChannel(update): %v", err)
	}
	if same.ID != channel.ID {
		t.Fatalf("expected same channel id, got %s and %s", same.ID, channel.ID)
	}
	if len(same.Participants) != 2 {
		t.Fatalf("expected participant merge, got %#v", same.Participants)
	}
	if same.Metadata["external_source"] != "github" || same.Metadata["owner"] != "platform" {
		t.Fatalf("unexpected metadata: %#v", same.Metadata)
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

func TestCollabStoreReadAllMessagesOrdered(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	alpha, err := store.GetOrCreateChannel("task:alpha", "Alpha", []string{"agent-a"}, nil)
	if err != nil {
		t.Fatalf("GetOrCreateChannel(alpha): %v", err)
	}
	beta, err := store.GetOrCreateChannel("task:beta", "Beta", []string{"agent-b"}, nil)
	if err != nil {
		t.Fatalf("GetOrCreateChannel(beta): %v", err)
	}

	first, err := store.PostMessage(alpha.ID, "agent-a", "message", "first", nil)
	if err != nil {
		t.Fatalf("PostMessage(first): %v", err)
	}
	second, err := store.PostMessage(beta.ID, "agent-b", "message", "second", nil)
	if err != nil {
		t.Fatalf("PostMessage(second): %v", err)
	}
	third, err := store.PostMessage(alpha.ID, "agent-c", "message", "third", nil)
	if err != nil {
		t.Fatalf("PostMessage(third): %v", err)
	}

	all, err := store.ReadAllMessages(0, 10)
	if err != nil {
		t.Fatalf("ReadAllMessages: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(all))
	}
	if all[0].Sequence != first.Sequence || all[1].Sequence != second.Sequence || all[2].Sequence != third.Sequence {
		t.Fatalf("unexpected sequence order: %#v", all)
	}

	replay, err := store.ReadAllMessages(first.Sequence, 10)
	if err != nil {
		t.Fatalf("ReadAllMessages replay: %v", err)
	}
	if len(replay) != 2 || replay[0].Sequence != second.Sequence || replay[1].Sequence != third.Sequence {
		t.Fatalf("unexpected replay messages: %#v", replay)
	}
}

func TestCollabStoreSubscribeFanout(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	channel, err := store.GetOrCreateChannel("task:stream", "Stream", []string{"agent-a"}, nil)
	if err != nil {
		t.Fatalf("GetOrCreateChannel: %v", err)
	}

	subOne, cancelOne := store.Subscribe(channel.ID)
	defer cancelOne()
	subTwo, cancelTwo := store.Subscribe(channel.ID)
	defer cancelTwo()

	expected, err := store.PostMessage(channel.ID, "agent-a", "message", "watch me", nil)
	if err != nil {
		t.Fatalf("PostMessage: %v", err)
	}

	assertMessage := func(name string, ch <-chan *CollabMessage) {
		t.Helper()
		select {
		case message := <-ch:
			if message.Sequence != expected.Sequence || message.Body != expected.Body {
				t.Fatalf("%s got unexpected message: %#v", name, message)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("%s did not receive message", name)
		}
	}

	assertMessage("subscriber one", subOne)
	assertMessage("subscriber two", subTwo)
}

func TestCollabStoreSetAgentFocusAndListAgents(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	agent, err := store.SetAgentFocus(
		"agent-a",
		[]string{"alpha"},
		[]string{"dispatch"},
		[]string{"interrupt"},
		"",
		map[string]string{"run_id": "run-123"},
	)
	if err != nil {
		t.Fatalf("SetAgentFocus: %v", err)
	}
	if agent.Metadata["run_id"] != "run-123" {
		t.Fatalf("expected run_id metadata, got %#v", agent.Metadata)
	}

	agents, err := store.ListAgents("dispatch", 10)
	if err != nil {
		t.Fatalf("ListAgents: %v", err)
	}
	if len(agents) != 1 || agents[0].ID != "agent-a" {
		t.Fatalf("unexpected agents: %#v", agents)
	}
}

func TestCollabStoreRouteMessage(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	if _, err := store.SetAgentFocus("agent-a", nil, []string{"routing"}, []string{"interrupt"}, "", map[string]string{"run_id": "run-a"}); err != nil {
		t.Fatalf("SetAgentFocus(agent-a): %v", err)
	}
	if _, err := store.SetAgentFocus("agent-b", nil, []string{"storage"}, []string{"append"}, "", map[string]string{"run_id": "run-b"}); err != nil {
		t.Fatalf("SetAgentFocus(agent-b): %v", err)
	}

	result, err := store.RouteMessage(CollabRouteRequest{
		Sender:         "control-plane",
		Body:           "reroute this task",
		Topic:          "routing",
		TargetAgentIDs: []string{"agent-a"},
		DeliveryMode:   "interrupt_and_replan",
		Metadata:       map[string]string{"reason": "dispatcher"},
	})
	if err != nil {
		t.Fatalf("RouteMessage: %v", err)
	}
	if len(result.Deliveries) != 1 {
		t.Fatalf("expected 1 delivery, got %#v", result.Deliveries)
	}
	if result.Deliveries[0].TargetAgentID != "agent-a" {
		t.Fatalf("unexpected target: %#v", result.Deliveries[0])
	}

	channel, err := store.GetChannelForAgent("agent-a")
	if err != nil {
		t.Fatalf("GetChannelForAgent: %v", err)
	}
	messages, err := store.ReadMessages(channel.ID, 0, 10)
	if err != nil {
		t.Fatalf("ReadMessages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected routed mailbox message, got %#v", messages)
	}
	if messages[0].Metadata["target_agent_id"] != "agent-a" {
		t.Fatalf("expected target_agent_id metadata, got %#v", messages[0].Metadata)
	}
	if messages[0].Metadata["delivery_mode"] != "interrupt_and_replan" {
		t.Fatalf("expected delivery_mode metadata, got %#v", messages[0].Metadata)
	}
}
