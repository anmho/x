package dispatch

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anmho/agent-control-api/internal/domain"
	"go.uber.org/zap"
)

type DeliveryRequest struct {
	TargetRunIDs []string
	Query        string
	Limit        int
	Message      string
	Reason       string
	Sender       string
	DeliveryMode domain.PushDeliveryMode
	Metadata     map[string]string
}

type DeliverFunc func(context.Context, DeliveryRequest) ([]string, error)

type Config struct {
	EventsURL string
	APIKey    string
	Mailbox   string
	StateFile string
}

type Dispatcher struct {
	log        *zap.Logger
	cfg        Config
	httpClient *http.Client
	deliver    DeliverFunc
}

type mailboxMessage struct {
	ID        string            `json:"id"`
	Sequence  int64             `json:"sequence"`
	ChannelID string            `json:"channel_id"`
	Sender    string            `json:"sender"`
	Kind      string            `json:"kind"`
	Body      string            `json:"body"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type collabSnapshot struct {
	Agents []struct {
		ID       string            `json:"id"`
		Metadata map[string]string `json:"metadata,omitempty"`
	} `json:"agents,omitempty"`
	Channels []struct {
		ID  string `json:"id"`
		Key string `json:"key,omitempty"`
	} `json:"channels,omitempty"`
}

type state struct {
	LastSequence int64 `json:"last_sequence"`
}

type sseEvent struct {
	Event string
	ID    string
	Data  string
}

func NewFromEnv(log *zap.Logger, deliver DeliverFunc) *Dispatcher {
	eventsURL := strings.TrimSpace(os.Getenv("MCP_MAILBOX_EVENTS_URL"))
	if eventsURL == "" || deliver == nil {
		return nil
	}
	mailbox := strings.TrimSpace(os.Getenv("MCP_MAILBOX_FILE"))
	if mailbox == "" {
		mailbox = strings.TrimSpace(os.Getenv("MCP_COLLAB_STORE"))
	}
	if mailbox == "" {
		mailbox = defaultMailboxStorePath()
	}
	stateFile := strings.TrimSpace(os.Getenv("MCP_DISPATCH_STATE_FILE"))
	if stateFile == "" {
		stateFile = defaultStatePath()
	}
	return &Dispatcher{
		log: log,
		cfg: Config{
			EventsURL: eventsURL,
			APIKey:    strings.TrimSpace(os.Getenv("MCP_API_KEY")),
			Mailbox:   mailbox,
			StateFile: stateFile,
		},
		httpClient: &http.Client{Timeout: 0},
		deliver:    deliver,
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	if d == nil {
		return
	}
	go d.run(ctx)
}

func (d *Dispatcher) HandleMessage(ctx context.Context, msg *mailboxMessage) (bool, error) {
	req, ok, err := d.buildDeliveryRequest(msg)
	if err != nil || !ok {
		return ok, err
	}
	delivered, err := d.deliver(ctx, req)
	if err != nil {
		return true, err
	}
	d.log.Info("dispatched mailbox message",
		zap.Int64("sequence", msg.Sequence),
		zap.String("channel_id", msg.ChannelID),
		zap.String("message_id", msg.ID),
		zap.Int("delivered", len(delivered)),
	)
	return true, nil
}

func (d *Dispatcher) run(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		lastSequence, err := d.loadState()
		if err != nil {
			d.log.Warn("load dispatcher state", zap.Error(err))
		}
		err = d.consume(ctx, lastSequence)
		if ctx.Err() != nil {
			return
		}
		d.log.Warn("dispatcher stream ended", zap.Error(err), zap.Duration("retry_in", backoff))
		time.Sleep(backoff)
		if backoff < 10*time.Second {
			backoff *= 2
		}
	}
}

func (d *Dispatcher) consume(ctx context.Context, afterSequence int64) error {
	reqURL, err := url.Parse(d.cfg.EventsURL)
	if err != nil {
		return err
	}
	query := reqURL.Query()
	query.Set("after_sequence", strconv.FormatInt(afterSequence, 10))
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return err
	}
	if d.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+d.cfg.APIKey)
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("mailbox stream failed: status %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	lastSequence := afterSequence
	for {
		event, err := readSSEEvent(reader)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		if event.Event != "message" || strings.TrimSpace(event.Data) == "" {
			continue
		}
		var msg mailboxMessage
		if err := json.Unmarshal([]byte(event.Data), &msg); err != nil {
			return err
		}
		if event.ID != "" {
			if sequence, err := strconv.ParseInt(event.ID, 10, 64); err == nil {
				msg.Sequence = sequence
			}
		}
		if msg.Sequence <= lastSequence {
			continue
		}
		if _, err := d.HandleMessage(ctx, &msg); err != nil {
			return err
		}
		lastSequence = msg.Sequence
		if err := d.saveState(lastSequence); err != nil {
			return err
		}
	}
}

func (d *Dispatcher) buildDeliveryRequest(msg *mailboxMessage) (DeliveryRequest, bool, error) {
	if msg == nil || strings.TrimSpace(msg.Body) == "" {
		return DeliveryRequest{}, false, nil
	}
	snapshot, err := d.loadSnapshot()
	if err != nil {
		return DeliveryRequest{}, false, err
	}

	metadata := cloneMap(msg.Metadata)
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["mailbox_message_id"] = msg.ID
	metadata["mailbox_sequence"] = strconv.FormatInt(msg.Sequence, 10)
	metadata["mailbox_channel_id"] = msg.ChannelID

	req := DeliveryRequest{
		TargetRunIDs: parseCSV(metadata["target_run_ids"]),
		Query:        strings.TrimSpace(metadata["dispatch_query"]),
		Limit:        25,
		Message:      strings.TrimSpace(msg.Body),
		Reason:       fallback(strings.TrimSpace(metadata["reason"]), strings.TrimSpace(msg.Kind), "mailbox_dispatch"),
		Sender:       fallback(strings.TrimSpace(msg.Sender), "mailbox-dispatcher"),
		DeliveryMode: parseDeliveryMode(metadata["delivery_mode"]),
		Metadata:     metadata,
	}
	if targetRunID := strings.TrimSpace(metadata["target_run_id"]); targetRunID != "" {
		req.TargetRunIDs = append(req.TargetRunIDs, targetRunID)
	}
	if limitRaw := strings.TrimSpace(metadata["dispatch_limit"]); limitRaw != "" {
		if limit, err := strconv.Atoi(limitRaw); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	if len(req.TargetRunIDs) == 0 && req.Query == "" {
		targetAgentID := strings.TrimSpace(metadata["target_agent_id"])
		if targetAgentID == "" {
			targetAgentID = snapshot.channelAgent(msg.ChannelID)
		}
		if targetAgentID != "" {
			metadata["target_agent_id"] = targetAgentID
			if runID := snapshot.agentRunID(targetAgentID); runID != "" {
				req.TargetRunIDs = []string{runID}
			}
		}
	}
	req.TargetRunIDs = dedupeStrings(req.TargetRunIDs)
	if len(req.TargetRunIDs) == 0 && req.Query == "" {
		return DeliveryRequest{}, false, nil
	}
	return req, true, nil
}

func (d *Dispatcher) loadSnapshot() (*collabSnapshot, error) {
	raw, err := os.ReadFile(d.cfg.Mailbox)
	if err != nil {
		return nil, err
	}
	var snapshot collabSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (d *Dispatcher) loadState() (int64, error) {
	raw, err := os.ReadFile(d.cfg.StateFile)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var state state
	if err := json.Unmarshal(raw, &state); err != nil {
		return 0, err
	}
	return state.LastSequence, nil
}

func (d *Dispatcher) saveState(sequence int64) error {
	if err := os.MkdirAll(filepath.Dir(d.cfg.StateFile), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(state{LastSequence: sequence}, "", "  ")
	if err != nil {
		return err
	}
	tmp := d.cfg.StateFile + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, d.cfg.StateFile)
}

func (s *collabSnapshot) agentRunID(agentID string) string {
	for _, agent := range s.Agents {
		if agent.ID == agentID {
			return strings.TrimSpace(agent.Metadata["run_id"])
		}
	}
	return ""
}

func (s *collabSnapshot) channelAgent(channelID string) string {
	for _, channel := range s.Channels {
		if channel.ID == channelID && strings.HasPrefix(channel.Key, "agent:") {
			return strings.TrimPrefix(channel.Key, "agent:")
		}
	}
	return ""
}

func readSSEEvent(reader *bufio.Reader) (sseEvent, error) {
	var event sseEvent
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return event, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if event.Event != "" || event.ID != "" || event.Data != "" {
				return event, nil
			}
			continue
		}
		switch {
		case strings.HasPrefix(line, "event: "):
			event.Event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "id: "):
			event.ID = strings.TrimPrefix(line, "id: ")
		case strings.HasPrefix(line, "data: "):
			payload := strings.TrimPrefix(line, "data: ")
			if event.Data == "" {
				event.Data = payload
			} else {
				event.Data += "\n" + payload
			}
		}
	}
}

func parseDeliveryMode(raw string) domain.PushDeliveryMode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "interrupt_and_replan", "interrupt-and-replan":
		return domain.PushDeliveryInterruptReplan
	default:
		return domain.PushDeliveryAppendContext
	}
}

func parseCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	return dedupeStrings(parts)
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func cloneMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func fallback(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func defaultMailboxStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "x-mcp", "collab.json")
	}
	return filepath.Join(home, ".x-mcp", "collab.json")
}

func defaultStatePath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "x-agent-control", "dispatcher.json")
	}
	return filepath.Join(home, ".x-agent-control", "dispatcher.json")
}
