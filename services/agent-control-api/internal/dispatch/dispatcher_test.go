package dispatch

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/anmho/agent-control-api/internal/domain"
	"go.uber.org/zap"
)

func TestHandleMessageExplicitTargets(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "collab.json")
	writeSnapshot(t, storePath, &collabSnapshot{})

	var delivered DeliveryRequest
	dispatcher := &Dispatcher{
		log: zap.NewNop(),
		cfg: Config{Mailbox: storePath},
		deliver: func(_ context.Context, req DeliveryRequest) ([]string, error) {
			delivered = req
			return append([]string(nil), req.TargetRunIDs...), nil
		},
	}

	handled, err := dispatcher.HandleMessage(context.Background(), &mailboxMessage{
		ID:        "msg-1",
		Sequence:  7,
		ChannelID: "channel-1",
		Sender:    "control-plane",
		Kind:      "handoff",
		Body:      "reroute this run",
		Metadata: map[string]string{
			"target_run_ids": "run-1, run-2, run-1",
			"delivery_mode":  "interrupt_and_replan",
			"reason":         "mailbox",
		},
	})
	if err != nil {
		t.Fatalf("HandleMessage: %v", err)
	}
	if !handled {
		t.Fatal("expected handled=true")
	}
	if !reflect.DeepEqual(delivered.TargetRunIDs, []string{"run-1", "run-2"}) {
		t.Fatalf("unexpected target runs: %#v", delivered.TargetRunIDs)
	}
	if delivered.DeliveryMode != domain.PushDeliveryInterruptReplan {
		t.Fatalf("unexpected delivery mode: %s", delivered.DeliveryMode)
	}
	if delivered.Metadata["mailbox_sequence"] != "7" || delivered.Metadata["mailbox_message_id"] != "msg-1" {
		t.Fatalf("missing mailbox metadata: %#v", delivered.Metadata)
	}
}

func TestBuildDeliveryRequestResolvesAgentMailbox(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "collab.json")
	writeSnapshot(t, storePath, &collabSnapshot{
		Agents: []struct {
			ID       string            `json:"id"`
			Metadata map[string]string `json:"metadata,omitempty"`
		}{
			{ID: "agent-a", Metadata: map[string]string{"run_id": "run-123"}},
		},
		Channels: []struct {
			ID  string `json:"id"`
			Key string `json:"key,omitempty"`
		}{
			{ID: "channel-agent-a", Key: "agent:agent-a"},
		},
	})

	dispatcher := &Dispatcher{
		log: zap.NewNop(),
		cfg: Config{Mailbox: storePath},
	}

	req, ok, err := dispatcher.buildDeliveryRequest(&mailboxMessage{
		ID:        "msg-2",
		Sequence:  11,
		ChannelID: "channel-agent-a",
		Sender:    "system",
		Kind:      "message",
		Body:      "take over this run",
	})
	if err != nil {
		t.Fatalf("buildDeliveryRequest: %v", err)
	}
	if !ok {
		t.Fatal("expected request to resolve")
	}
	if !reflect.DeepEqual(req.TargetRunIDs, []string{"run-123"}) {
		t.Fatalf("unexpected target runs: %#v", req.TargetRunIDs)
	}
	if got := req.Metadata["target_agent_id"]; got != "agent-a" {
		t.Fatalf("expected target_agent_id agent-a, got %q", got)
	}
	if req.DeliveryMode != domain.PushDeliveryAppendContext {
		t.Fatalf("unexpected default delivery mode: %s", req.DeliveryMode)
	}
}

func TestHandleMessageSkipsUnroutableMailboxMessage(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "collab.json")
	writeSnapshot(t, storePath, &collabSnapshot{
		Channels: []struct {
			ID  string `json:"id"`
			Key string `json:"key,omitempty"`
		}{
			{ID: "channel-generic", Key: "task:generic"},
		},
	})

	called := false
	dispatcher := &Dispatcher{
		log: zap.NewNop(),
		cfg: Config{Mailbox: storePath},
		deliver: func(_ context.Context, req DeliveryRequest) ([]string, error) {
			called = true
			return nil, nil
		},
	}

	handled, err := dispatcher.HandleMessage(context.Background(), &mailboxMessage{
		ID:        "msg-3",
		Sequence:  12,
		ChannelID: "channel-generic",
		Sender:    "system",
		Kind:      "message",
		Body:      "observer-only note",
	})
	if err != nil {
		t.Fatalf("HandleMessage: %v", err)
	}
	if handled {
		t.Fatal("expected handled=false")
	}
	if called {
		t.Fatal("deliver should not be called for unroutable message")
	}
}

func writeSnapshot(t *testing.T, path string, snapshot *collabSnapshot) {
	t.Helper()
	raw, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir snapshot dir: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
}
