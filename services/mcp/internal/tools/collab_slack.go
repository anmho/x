package tools

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const defaultSlackAPIURL = "https://slack.com/api"

var linearIssuePattern = regexp.MustCompile(`\b[A-Z][A-Z0-9]+-\d+\b`)

type SlackClient struct {
	botToken      string
	signingSecret string
	baseURL       string
	httpClient    *http.Client
}

type SlackBridge struct {
	store  *CollabStore
	slack  *SlackClient
	linear *LinearClient
}

func NewSlackClientFromEnv() *SlackClient {
	return &SlackClient{
		botToken:      strings.TrimSpace(os.Getenv("SLACK_BOT_TOKEN")),
		signingSecret: strings.TrimSpace(os.Getenv("SLACK_SIGNING_SECRET")),
		baseURL:       strings.TrimSpace(os.Getenv("SLACK_API_URL")),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func NewSlackBridge(store *CollabStore, slack *SlackClient, linear *LinearClient) *SlackBridge {
	if store == nil {
		return nil
	}
	return &SlackBridge{store: store, slack: slack, linear: linear}
}

func (c *SlackClient) Enabled() bool {
	return c != nil && c.botToken != ""
}

func (c *SlackClient) VerifyRequest(body []byte, header http.Header) error {
	if c == nil || c.signingSecret == "" {
		return nil
	}
	timestamp := strings.TrimSpace(header.Get("X-Slack-Request-Timestamp"))
	signature := strings.TrimSpace(header.Get("X-Slack-Signature"))
	if timestamp == "" || signature == "" {
		return fmt.Errorf("missing Slack signature headers")
	}
	base := "v0:" + timestamp + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(c.signingSecret))
	_, _ = mac.Write([]byte(base))
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("invalid Slack signature")
	}
	return nil
}

func (c *SlackClient) PostThreadReply(ctx context.Context, channelID, threadTS, text string) error {
	if !c.Enabled() {
		return nil
	}
	payload := map[string]string{
		"channel":   channelID,
		"text":      text,
		"thread_ts": threadTS,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = defaultSlackAPIURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/chat.postMessage", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.botToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	if !body.OK {
		if body.Error == "" {
			body.Error = fmt.Sprintf("status %d", resp.StatusCode)
		}
		return fmt.Errorf("slack post failed: %s", body.Error)
	}
	return nil
}

func (b *SlackBridge) HandleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request", http.StatusBadRequest)
		return
	}
	if b.slack != nil {
		if err := b.slack.VerifyRequest(raw, r.Header); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	var envelope struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		EventID   string `json:"event_id"`
		Event     struct {
			Type     string `json:"type"`
			Subtype  string `json:"subtype"`
			BotID    string `json:"bot_id"`
			User     string `json:"user"`
			Text     string `json:"text"`
			Channel  string `json:"channel"`
			ThreadTS string `json:"thread_ts"`
			TS       string `json:"ts"`
		} `json:"event"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if envelope.Type == "url_verification" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"challenge": envelope.Challenge})
		return
	}
	if envelope.Type != "event_callback" {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if envelope.Event.Type != "app_mention" && envelope.Event.Type != "message" {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if envelope.Event.Subtype != "" || envelope.Event.BotID != "" || envelope.Event.User == "" || strings.TrimSpace(envelope.Event.Text) == "" {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	channelID := strings.TrimSpace(envelope.Event.Channel)
	threadTS := strings.TrimSpace(envelope.Event.ThreadTS)
	if threadTS == "" {
		threadTS = strings.TrimSpace(envelope.Event.TS)
	}
	if channelID == "" || threadTS == "" {
		http.Error(w, "missing Slack channel or thread id", http.StatusBadRequest)
		return
	}

	channel, err := b.resolveChannel(r.Context(), channelID, threadTS, envelope.Event.Text, envelope.Event.User)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if envelope.EventID != "" {
		seen, err := b.store.HasMessageWithMetadata(channel.ID, "slack_event_id", envelope.EventID)
		if err == nil && seen {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	metadata := map[string]string{
		"origin":           "slack",
		"slack_channel_id": channelID,
		"slack_thread_ts":  threadTS,
	}
	if envelope.EventID != "" {
		metadata["slack_event_id"] = envelope.EventID
	}
	if issueID := extractLinearIssueID(envelope.Event.Text); issueID != "" {
		metadata["linear_issue_id"] = issueID
	}
	_, err = b.store.PostMessage(channel.ID, "human:"+envelope.Event.User, "message", strings.TrimSpace(envelope.Event.Text), metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (b *SlackBridge) MirrorMessage(ctx context.Context, message *CollabMessage) error {
	if b == nil || b.slack == nil || !b.slack.Enabled() || message == nil {
		return nil
	}
	if strings.EqualFold(message.Metadata["origin"], "slack") {
		return nil
	}
	channel, err := b.store.GetChannel(message.ChannelID)
	if err != nil {
		return err
	}
	ref := selectSlackRef(channel)
	if ref == nil {
		return nil
	}
	return b.slack.PostThreadReply(ctx, ref.ParentID, ref.ExternalID, formatSlackMirrorText(channel, message))
}

func (b *SlackBridge) resolveChannel(ctx context.Context, slackChannelID, slackThreadTS, text, slackUser string) (*CollabChannelSummary, error) {
	channel, err := b.store.GetChannelByExternalRef("slack", slackThreadTS, slackChannelID)
	if err == nil {
		return channel, nil
	}

	sender := "human:" + strings.TrimSpace(slackUser)
	issueID := extractLinearIssueID(text)
	metadata := map[string]string{
		"scope":      "task",
		"source":     "slack",
		"owner_mode": "collaborative",
	}
	if issueID != "" {
		issue, issueErr := b.resolveLinearIssue(ctx, issueID)
		if issueErr != nil {
			metadata["linear_lookup_error"] = issueErr.Error()
		}
		channel, err = b.store.GetOrCreateLinearIssueChannel(issue, []string{sender}, metadata)
		if err != nil {
			return nil, err
		}
	} else {
		channel, err = b.store.GetOrCreateChannel(
			fmt.Sprintf("slack:%s:%s", slackChannelID, slackThreadTS),
			truncateSlackTitle(strings.TrimSpace(text)),
			[]string{sender},
			metadata,
		)
		if err != nil {
			return nil, err
		}
	}
	return b.store.LinkExternalRef(channel.ID, &CollabExternalRef{
		Source:     "slack",
		ExternalID: slackThreadTS,
		ParentID:   slackChannelID,
		Title:      truncateSlackTitle(strings.TrimSpace(text)),
		Metadata: map[string]string{
			"kind": "thread",
		},
	})
}

func (b *SlackBridge) resolveLinearIssue(ctx context.Context, identifier string) (*LinearIssue, error) {
	issue := &LinearIssue{Identifier: strings.ToUpper(strings.TrimSpace(identifier))}
	if b == nil || b.linear == nil || !b.linear.Enabled() {
		return issue, nil
	}
	enriched, err := b.linear.GetIssue(ctx, identifier)
	if err != nil {
		return issue, err
	}
	return enriched, nil
}

func extractLinearIssueID(text string) string {
	match := linearIssuePattern.FindString(strings.ToUpper(text))
	return strings.TrimSpace(match)
}

func selectSlackRef(channel *CollabChannelSummary) *CollabExternalRef {
	if channel == nil {
		return nil
	}
	for _, ref := range channel.ExternalRefs {
		if ref != nil && strings.EqualFold(ref.Source, "slack") && ref.ParentID != "" && ref.ExternalID != "" {
			return ref
		}
	}
	return nil
}

func formatSlackMirrorText(channel *CollabChannelSummary, message *CollabMessage) string {
	header := "*" + strings.TrimSpace(message.Sender) + "*"
	if message.Kind != "" && message.Kind != "message" {
		header += " [" + message.Kind + "]"
	}
	lines := []string{header}
	if channel != nil {
		if issueID := channel.Metadata["linear_issue_id"]; issueID != "" {
			lines = append(lines, issueID+" · "+channel.Status)
		} else if channel.Status != "" {
			lines = append(lines, "status: "+channel.Status)
		}
	}
	lines = append(lines, strings.TrimSpace(message.Body))
	return strings.Join(lines, "\n")
}

func truncateSlackTitle(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "Slack collaboration thread"
	}
	if len(text) <= 80 {
		return text
	}
	return text[:77] + "..."
}
