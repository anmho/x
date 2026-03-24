package watch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anmhela/x/mcp/internal/tools"
)

type MailboxEventsHandler struct {
	store     *tools.CollabStore
	keepAlive time.Duration
}

type mailboxReadyEvent struct {
	ChannelID     string `json:"channel_id,omitempty"`
	AfterSequence int64  `json:"after_sequence"`
}

func NewMailboxEventsHandler(store *tools.CollabStore) http.Handler {
	return &MailboxEventsHandler{
		store:     store,
		keepAlive: 15 * time.Second,
	}
}

func (h *MailboxEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	channelID := strings.TrimSpace(r.URL.Query().Get("channel_id"))
	afterSequence, err := parseAfterSequence(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, err := parsePositiveInt(r.URL.Query().Get("replay_limit"), 200, 500)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	events, cancel := h.store.Subscribe(channelID)
	defer cancel()

	replay, err := h.readReplay(channelID, afterSequence, limit)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	lastSequence := afterSequence
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	if err := writeSSE(w, "ready", "", mailboxReadyEvent{
		ChannelID:     channelID,
		AfterSequence: afterSequence,
	}); err != nil {
		return
	}
	for _, message := range replay {
		if err := writeSSE(w, "message", strconv.FormatInt(message.Sequence, 10), message); err != nil {
			return
		}
		lastSequence = message.Sequence
	}
	flusher.Flush()

	heartbeat := time.NewTicker(h.keepAlive)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case message, ok := <-events:
			if !ok {
				return
			}
			if message.Sequence <= lastSequence {
				continue
			}
			if err := writeSSE(w, "message", strconv.FormatInt(message.Sequence, 10), message); err != nil {
				return
			}
			lastSequence = message.Sequence
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := fmt.Fprintf(w, ": keepalive %d\n\n", lastSequence); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *MailboxEventsHandler) readReplay(channelID string, afterSequence int64, limit int) ([]*tools.CollabMessage, error) {
	if channelID != "" {
		return h.store.ReadMessages(channelID, afterSequence, limit)
	}
	return h.store.ReadAllMessages(afterSequence, limit)
}

func parseAfterSequence(r *http.Request) (int64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("after_sequence"))
	if raw == "" {
		raw = strings.TrimSpace(r.Header.Get("Last-Event-ID"))
	}
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.New("after_sequence must be an integer")
	}
	if value < 0 {
		return 0, errors.New("after_sequence must be >= 0")
	}
	return value, nil
}

func parsePositiveInt(raw string, def, max int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("replay_limit must be an integer")
	}
	if value <= 0 {
		return 0, errors.New("replay_limit must be > 0")
	}
	if value > max {
		return max, nil
	}
	return value, nil
}

func writeSSE(w http.ResponseWriter, event, id string, payload any) error {
	if id != "" {
		if _, err := fmt.Fprintf(w, "id: %s\n", id); err != nil {
			return err
		}
	}
	if event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
			return err
		}
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return nil
}
