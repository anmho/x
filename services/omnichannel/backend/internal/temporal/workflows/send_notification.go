package workflows

import (
	"time"

	"github.com/andrewho/omnichannel/internal/domain"
	"github.com/andrewho/omnichannel/internal/temporal/activities"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// SendNotificationInput represents input for the SendNotification workflow
type SendNotificationInput struct {
	NotificationID uuid.UUID
	Channel        domain.NotificationChannel
	TemplateID     *uuid.UUID
	RecipientEmail string
	Subject        string
	Body           string
	Variables      map[string]string
}

// SendNotificationWorkflow orchestrates sending a notification
func SendNotificationWorkflow(ctx workflow.Context, input SendNotificationInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting SendNotificationWorkflow", "notification_id", input.NotificationID, "channel", input.Channel)

	// Set activity options with retry policy
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 5,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Render template
	var renderOutput *activities.RenderOutput
	err := workflow.ExecuteActivity(ctx, "RenderTemplate", activities.RenderInput{
		TemplateID: input.TemplateID,
		Subject:    input.Subject,
		Body:       input.Body,
		Variables:  input.Variables,
	}).Get(ctx, &renderOutput)

	if err != nil {
		logger.Error("Failed to render template", "error", err)
		return err
	}

	logger.Info("Template rendered", "subject", renderOutput.Subject)

	if input.Channel == "" {
		input.Channel = domain.ChannelEmail
	}

	// Step 2: Route by channel
	switch input.Channel {
	case domain.ChannelEmail:
		err = workflow.ExecuteActivity(ctx, "SendEmail", activities.EmailInput{
			To:      input.RecipientEmail,
			Subject: renderOutput.Subject,
			Body:    renderOutput.Body,
		}).Get(ctx, nil)
	case domain.ChannelSMS:
		err = workflow.ExecuteActivity(ctx, "SendSMS", activities.SMSInput{
			To:      input.RecipientEmail,
			Message: renderOutput.Body,
		}).Get(ctx, nil)
	case domain.ChannelPush:
		err = workflow.ExecuteActivity(ctx, "SendPush", activities.PushInput{
			DeviceToken: input.RecipientEmail,
			Title:       renderOutput.Subject,
			Body:        renderOutput.Body,
		}).Get(ctx, nil)
	case domain.ChannelWebhook:
		err = workflow.ExecuteActivity(ctx, "SendWebhook", activities.WebhookInput{
			URL:  input.RecipientEmail,
			Body: renderOutput.Body,
		}).Get(ctx, nil)
	case domain.ChannelApp:
		err = workflow.ExecuteActivity(ctx, "SendApp", activities.AppInput{
			DeviceID: input.RecipientEmail,
			Channel:  string(domain.ChannelApp),
			Title:    renderOutput.Subject,
			Body:     renderOutput.Body,
		}).Get(ctx, nil)
	case domain.ChannelIMessage:
		err = workflow.ExecuteActivity(ctx, "SendApp", activities.AppInput{
			DeviceID: input.RecipientEmail,
			Channel:  string(domain.ChannelIMessage),
			Title:    renderOutput.Subject,
			Body:     renderOutput.Body,
		}).Get(ctx, nil)
	default:
		return temporal.NewNonRetryableApplicationError("unsupported channel", "ValidationError", nil)
	}

	if err != nil {
		logger.Error("Failed to send notification", "channel", input.Channel, "error", err)
		return err
	}

	logger.Info("Notification sent successfully", "notification_id", input.NotificationID, "channel", input.Channel)
	return nil
}
