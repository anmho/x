package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestSlackBridgeHandleEventsCreatesLinearChannel(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	linearClient := &LinearClient{
		apiKey:  "linear-test-key",
		baseURL: "https://linear.test/graphql",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			payload, _ := json.Marshal(map[string]any{
				"data": map[string]any{
					"issues": map[string]any{
						"nodes": []map[string]any{
							{
								"id":         "issue-1",
								"identifier": "ANM-190",
								"title":      "Bridge collaboration context",
								"url":        "https://linear.app/anmho/issue/ANM-190/test",
								"state": map[string]any{
									"name": "In Progress",
									"type": "started",
								},
								"team": map[string]any{
									"id":   "team-1",
									"key":  "ANM",
									"name": "Anmho",
								},
							},
						},
					},
				},
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(payload))),
			}, nil
		})},
	}

	bridge := NewSlackBridge(store, &SlackClient{}, linearClient)

	body := `{"type":"event_callback","event_id":"evt_1","event":{"type":"app_mention","user":"U123","text":"Please coordinate on ANM-190","channel":"C123","ts":"1742800000.100000"}}`
	req, _ := http.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	res := newTestResponseWriter()
	bridge.HandleEvents(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", res.Code, res.Body)
	}

	channel, err := store.GetChannelByExternalRef("slack", "1742800000.100000", "C123")
	if err != nil {
		t.Fatalf("GetChannelByExternalRef: %v", err)
	}
	if channel.Key != "linear:ANM-190" {
		t.Fatalf("unexpected channel key: %s", channel.Key)
	}
	if channel.Metadata["linear_issue_id"] != "ANM-190" {
		t.Fatalf("missing linear issue metadata: %#v", channel.Metadata)
	}

	messages, err := store.ReadMessages(channel.ID, 0, 10)
	if err != nil {
		t.Fatalf("ReadMessages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Metadata["origin"] != "slack" {
		t.Fatalf("expected slack-origin metadata: %#v", messages[0].Metadata)
	}

	res = newTestResponseWriter()
	req, _ = http.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	bridge.HandleEvents(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("unexpected duplicate status: %d body=%s", res.Code, res.Body)
	}
	messages, err = store.ReadMessages(channel.ID, 0, 10)
	if err != nil {
		t.Fatalf("ReadMessages after duplicate: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected duplicate event to be ignored, got %d messages", len(messages))
	}
}

func TestSlackBridgeMirrorMessagePostsThreadReply(t *testing.T) {
	store := NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))

	var posted struct {
		Channel  string `json:"channel"`
		ThreadTS string `json:"thread_ts"`
		Text     string `json:"text"`
	}
	slackClient := &SlackClient{
		botToken: "xoxb-test",
		baseURL:  "https://slack.test/api",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/api/chat.postMessage" {
				t.Fatalf("unexpected slack path: %s", r.URL.Path)
			}
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode slack post: %v", err)
			}
			payload, _ := json.Marshal(map[string]any{"ok": true})
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(payload))),
			}, nil
		})},
	}

	channel, err := store.GetOrCreateLinearIssueChannel(&LinearIssue{
		Identifier: "ANM-190",
		Title:      "Bridge collaboration context",
	}, []string{"agent-a"}, nil)
	if err != nil {
		t.Fatalf("GetOrCreateLinearIssueChannel: %v", err)
	}
	if _, err := store.LinkExternalRef(channel.ID, &CollabExternalRef{
		Source:     "slack",
		ExternalID: "1742800000.100000",
		ParentID:   "C123",
	}); err != nil {
		t.Fatalf("LinkExternalRef: %v", err)
	}
	message, err := store.PostMessage(channel.ID, "agent-a", "status", "Implementation started", map[string]string{"origin": "agent"})
	if err != nil {
		t.Fatalf("PostMessage: %v", err)
	}

	bridge := NewSlackBridge(store, slackClient, nil)
	if err := bridge.MirrorMessage(context.Background(), message); err != nil {
		t.Fatalf("MirrorMessage: %v", err)
	}
	if posted.Channel != "C123" || posted.ThreadTS != "1742800000.100000" {
		t.Fatalf("unexpected slack routing payload: %#v", posted)
	}
	if !strings.Contains(posted.Text, "agent-a") || !strings.Contains(posted.Text, "Implementation started") {
		t.Fatalf("unexpected slack text: %s", posted.Text)
	}
}

type testResponseWriter struct {
	HeaderMap http.Header
	Code      int
	Body      string
}

func newTestResponseWriter() *testResponseWriter {
	return &testResponseWriter{HeaderMap: make(http.Header)}
}

func (w *testResponseWriter) Header() http.Header {
	return w.HeaderMap
}

func (w *testResponseWriter) Write(data []byte) (int, error) {
	w.Body += string(data)
	if w.Code == 0 {
		w.Code = http.StatusOK
	}
	return len(data), nil
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
}
