package tools

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

type CollabStore struct {
	path string
	mu   sync.Mutex
}

type collabData struct {
	NextSequence int64            `json:"next_sequence"`
	Channels     []*CollabChannel `json:"channels"`
}

type CollabChannel struct {
	ID           string            `json:"id"`
	Key          string            `json:"key,omitempty"`
	Title        string            `json:"title,omitempty"`
	Participants []string          `json:"participants,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Messages     []*CollabMessage  `json:"messages,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type CollabChannelSummary struct {
	ID            string            `json:"id"`
	Key           string            `json:"key,omitempty"`
	Title         string            `json:"title,omitempty"`
	Participants  []string          `json:"participants,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	MessageCount  int               `json:"message_count"`
	LastMessageAt *time.Time        `json:"last_message_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type CollabMessage struct {
	ID        string            `json:"id"`
	Sequence  int64             `json:"sequence"`
	ChannelID string            `json:"channel_id"`
	Sender    string            `json:"sender"`
	Kind      string            `json:"kind"`
	Body      string            `json:"body"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

func NewCollabStore(path string) *CollabStore {
	if path == "" {
		path = defaultCollabStorePath()
	}
	return &CollabStore{path: path}
}

func defaultCollabStorePath() string {
	if path := os.Getenv("MCP_MAILBOX_FILE"); path != "" {
		return path
	}
	if path := os.Getenv("MCP_COLLAB_STORE"); path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "x-mcp", "collab.json")
	}
	return filepath.Join(home, ".x-mcp", "collab.json")
}

func (s *CollabStore) GetChannelForAgent(agentID string) (*CollabChannelSummary, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return nil, errors.New("agent_id is required")
	}
	return s.GetOrCreateChannel("agent:"+agentID, "Mailbox for "+agentID, []string{agentID}, nil)
}

func (s *CollabStore) GetOrCreateChannel(key, title string, participants []string, metadata map[string]string) (*CollabChannelSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	participants = normalizeStrings(participants)
	metadata = cloneMap(metadata)
	if key != "" {
		for _, channel := range data.Channels {
			if channel.Key != key {
				continue
			}
			if title != "" {
				channel.Title = title
			}
			channel.Participants = normalizeStrings(append(channel.Participants, participants...))
			if len(metadata) > 0 {
				if channel.Metadata == nil {
					channel.Metadata = map[string]string{}
				}
				for k, v := range metadata {
					channel.Metadata[k] = v
				}
			}
			channel.UpdatedAt = now
			if err := s.saveLocked(data); err != nil {
				return nil, err
			}
			return summarizeChannel(channel), nil
		}
	}

	channel := &CollabChannel{
		ID:           newCollabID(),
		Key:          key,
		Title:        title,
		Participants: participants,
		Metadata:     metadata,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	data.Channels = append(data.Channels, channel)
	if err := s.saveLocked(data); err != nil {
		return nil, err
	}
	return summarizeChannel(channel), nil
}

func (s *CollabStore) ListChannels(participant, query string, limit int) ([]*CollabChannelSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query = strings.ToLower(strings.TrimSpace(query))

	var out []*CollabChannelSummary
	for _, channel := range data.Channels {
		if participant != "" && !slices.Contains(channel.Participants, participant) {
			continue
		}
		if query != "" && !channelMatchesQuery(channel, query) {
			continue
		}
		out = append(out, summarizeChannel(channel))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *CollabStore) PostMessage(channelID, sender, kind, body string, metadata map[string]string) (*CollabMessage, error) {
	if channelID == "" {
		return nil, errors.New("channel_id is required")
	}
	if sender == "" {
		return nil, errors.New("sender is required")
	}
	if body == "" {
		return nil, errors.New("body is required")
	}
	if kind == "" {
		kind = "message"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	channel := findChannelByID(data.Channels, channelID)
	if channel == nil {
		return nil, fmt.Errorf("channel %q not found", channelID)
	}

	message := &CollabMessage{
		ID:        newCollabID(),
		Sequence:  data.NextSequence + 1,
		ChannelID: channelID,
		Sender:    sender,
		Kind:      kind,
		Body:      body,
		Metadata:  cloneMap(metadata),
		CreatedAt: time.Now().UTC(),
	}
	data.NextSequence = message.Sequence
	channel.Messages = append(channel.Messages, message)
	channel.UpdatedAt = message.CreatedAt
	if !slices.Contains(channel.Participants, sender) {
		channel.Participants = normalizeStrings(append(channel.Participants, sender))
	}
	if err := s.saveLocked(data); err != nil {
		return nil, err
	}
	return cloneMessage(message), nil
}

func (s *CollabStore) ReadMessages(channelID string, afterSequence int64, limit int) ([]*CollabMessage, error) {
	if channelID == "" {
		return nil, errors.New("channel_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	channel := findChannelByID(data.Channels, channelID)
	if channel == nil {
		return nil, fmt.Errorf("channel %q not found", channelID)
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	out := make([]*CollabMessage, 0, limit)
	for _, message := range channel.Messages {
		if message.Sequence <= afterSequence {
			continue
		}
		out = append(out, cloneMessage(message))
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func (s *CollabStore) loadLocked() (*collabData, error) {
	data := &collabData{Channels: []*CollabChannel{}}
	raw, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return data, nil
	}
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return data, nil
	}
	if err := json.Unmarshal(raw, data); err != nil {
		return nil, err
	}
	if data.Channels == nil {
		data.Channels = []*CollabChannel{}
	}
	var maxSequence int64
	for _, channel := range data.Channels {
		for _, message := range channel.Messages {
			if message.Sequence == 0 {
				maxSequence++
				message.Sequence = maxSequence
				continue
			}
			if message.Sequence > maxSequence {
				maxSequence = message.Sequence
			}
		}
	}
	if data.NextSequence < maxSequence {
		data.NextSequence = maxSequence
	}
	return data, nil
}

func (s *CollabStore) saveLocked(data *collabData) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func summarizeChannel(channel *CollabChannel) *CollabChannelSummary {
	summary := &CollabChannelSummary{
		ID:           channel.ID,
		Key:          channel.Key,
		Title:        channel.Title,
		Participants: append([]string(nil), channel.Participants...),
		Metadata:     cloneMap(channel.Metadata),
		MessageCount: len(channel.Messages),
		CreatedAt:    channel.CreatedAt,
		UpdatedAt:    channel.UpdatedAt,
	}
	if n := len(channel.Messages); n > 0 {
		last := channel.Messages[n-1].CreatedAt
		summary.LastMessageAt = &last
	}
	return summary
}

func findChannelByID(channels []*CollabChannel, id string) *CollabChannel {
	for _, channel := range channels {
		if channel.ID == id {
			return channel
		}
	}
	return nil
}

func channelMatchesQuery(channel *CollabChannel, query string) bool {
	if strings.Contains(strings.ToLower(channel.ID), query) ||
		strings.Contains(strings.ToLower(channel.Key), query) ||
		strings.Contains(strings.ToLower(channel.Title), query) {
		return true
	}
	for _, participant := range channel.Participants {
		if strings.Contains(strings.ToLower(participant), query) {
			return true
		}
	}
	return false
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

func cloneMessage(in *CollabMessage) *CollabMessage {
	if in == nil {
		return nil
	}
	return &CollabMessage{
		ID:        in.ID,
		Sequence:  in.Sequence,
		ChannelID: in.ChannelID,
		Sender:    in.Sender,
		Kind:      in.Kind,
		Body:      in.Body,
		Metadata:  cloneMap(in.Metadata),
		CreatedAt: in.CreatedAt,
	}
}

func normalizeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
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
	sort.Strings(out)
	return out
}

func newCollabID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("collab-%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}
