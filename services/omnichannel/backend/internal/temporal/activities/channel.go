package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
)

// SMSInput represents input for sending an SMS message.
type SMSInput struct {
	To      string
	Message string
}

// PushInput represents input for sending a push notification.
type PushInput struct {
	DeviceToken string
	Title       string
	Body        string
}

// AppInput represents input for delivering an app/emulator notification.
type AppInput struct {
	DeviceID string
	Channel  string
	Title    string
	Body     string
}

// WebhookInput represents input for delivering a webhook notification.
type WebhookInput struct {
	URL  string
	Body string
}

// ChannelActivities holds non-email channel test activities.
type ChannelActivities struct {
	appRelayURL string
	httpClient  *http.Client
}

// NewChannelActivities creates a new ChannelActivities instance.
func NewChannelActivities(appRelayURL string) *ChannelActivities {
	return &ChannelActivities{
		appRelayURL: strings.TrimSpace(appRelayURL),
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

// SendSMS logs and simulates SMS delivery for channel testing.
func (a *ChannelActivities) SendSMS(ctx context.Context, input SMSInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Simulated SMS delivery", "to", input.To)
	return nil
}

// SendPush logs and simulates push delivery for channel testing.
func (a *ChannelActivities) SendPush(ctx context.Context, input PushInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Simulated push delivery", "token", input.DeviceToken, "title", input.Title)
	return a.deliverToAppRelay(ctx, "push", input.DeviceToken, input.Title, input.Body)
}

// SendApp delivers app/emulator notifications through an optional relay endpoint.
func (a *ChannelActivities) SendApp(ctx context.Context, input AppInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("App delivery requested", "device_id", input.DeviceID, "title", input.Title)
	channel := strings.TrimSpace(input.Channel)
	if channel == "" {
		channel = "app"
	}
	return a.deliverToAppRelay(ctx, channel, input.DeviceID, input.Title, input.Body)
}

// SendWebhook logs and simulates webhook delivery for channel testing.
func (a *ChannelActivities) SendWebhook(ctx context.Context, input WebhookInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Simulated webhook delivery", "url", input.URL)
	return nil
}

func (a *ChannelActivities) deliverToAppRelay(
	ctx context.Context,
	channel string,
	destination string,
	title string,
	body string,
) error {
	if a.appRelayURL == "" {
		return nil
	}

	payload := map[string]string{
		"channel":     channel,
		"destination": destination,
		"title":       title,
		"body":        body,
		"sent_at":     time.Now().UTC().Format(time.RFC3339),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode app relay payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.appRelayURL, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("failed to create app relay request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to deliver to app relay: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("app relay returned %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	return nil
}
