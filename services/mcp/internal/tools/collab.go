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
	path       string
	mu         sync.Mutex
	nextSubID  int
	subscribed map[int]collabSubscription
}

type collabData struct {
	NextSequence int64            `json:"next_sequence"`
	Agents       []*CollabAgent   `json:"agents,omitempty"`
	Channels     []*CollabChannel `json:"channels"`
}

type CollabChannel struct {
	ID           string               `json:"id"`
	Key          string               `json:"key,omitempty"`
	Title        string               `json:"title,omitempty"`
	Status       string               `json:"status,omitempty"`
	Participants []string             `json:"participants,omitempty"`
	Metadata     map[string]string    `json:"metadata,omitempty"`
	ExternalRefs []*CollabExternalRef `json:"external_refs,omitempty"`
	Messages     []*CollabMessage     `json:"messages,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

type CollabChannelSummary struct {
	ID            string               `json:"id"`
	Key           string               `json:"key,omitempty"`
	Title         string               `json:"title,omitempty"`
	Status        string               `json:"status,omitempty"`
	Participants  []string             `json:"participants,omitempty"`
	Metadata      map[string]string    `json:"metadata,omitempty"`
	ExternalRefs  []*CollabExternalRef `json:"external_refs,omitempty"`
	MessageCount  int                  `json:"message_count"`
	LastMessageAt *time.Time           `json:"last_message_at,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type CollabExternalRef struct {
	Source     string            `json:"source"`
	ExternalID string            `json:"external_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Title      string            `json:"title,omitempty"`
	URL        string            `json:"url,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
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

type CollabAgent struct {
	ID               string            `json:"id"`
	Aliases          []string          `json:"aliases,omitempty"`
	Capabilities     []string          `json:"capabilities,omitempty"`
	Topics           []string          `json:"topics,omitempty"`
	CurrentFocus     []string          `json:"current_focus,omitempty"`
	ActiveChannelIDs []string          `json:"active_channel_ids,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	LastSeenAt       time.Time         `json:"last_seen_at"`
}

type CollabRouteRequest struct {
	Sender          string
	Kind            string
	Body            string
	ChannelID       string
	Topic           string
	Query           string
	TargetAgentIDs  []string
	ExcludeAgentIDs []string
	Limit           int
	DryRun          bool
	Metadata        map[string]string
	DeliveryMode    string
}

type CollabRouteDelivery struct {
	TargetAgentID   string            `json:"target_agent_id"`
	TargetChannelID string            `json:"target_channel_id"`
	MessageID       string            `json:"message_id,omitempty"`
	Reason          string            `json:"reason"`
	Score           int               `json:"score"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type CollabRouteResult struct {
	RouteID     string                 `json:"route_id"`
	DryRun      bool                   `json:"dry_run"`
	Deliveries  []*CollabRouteDelivery `json:"deliveries"`
	Skipped     []string               `json:"skipped,omitempty"`
	ResolvedBy  []string               `json:"resolved_by,omitempty"`
	DeliveredAt time.Time              `json:"delivered_at"`
}

type collabSubscription struct {
	channelID string
	ch        chan *CollabMessage
}

func NewCollabStore(path string) *CollabStore {
	if path == "" {
		path = defaultCollabStorePath()
	}
	return &CollabStore{
		path:       path,
		subscribed: map[int]collabSubscription{},
	}
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

func (s *CollabStore) GetChannel(channelID string) (*CollabChannelSummary, error) {
	channelID = strings.TrimSpace(channelID)
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
	return summarizeChannel(channel), nil
}

func (s *CollabStore) GetChannelByExternalRef(source, externalID, parentID string) (*CollabChannelSummary, error) {
	source = strings.TrimSpace(strings.ToLower(source))
	externalID = strings.TrimSpace(externalID)
	parentID = strings.TrimSpace(parentID)
	if source == "" {
		return nil, errors.New("source is required")
	}
	if externalID == "" {
		return nil, errors.New("external_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	channel := findChannelByExternalRef(data.Channels, source, externalID, parentID)
	if channel == nil {
		return nil, fmt.Errorf("channel not found for source=%q external_id=%q", source, externalID)
	}
	return summarizeChannel(channel), nil
}

func (s *CollabStore) LinkExternalRef(channelID string, ref *CollabExternalRef) (*CollabChannelSummary, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return nil, errors.New("channel_id is required")
	}
	ref = normalizeExternalRef(ref)
	if ref == nil {
		return nil, errors.New("valid external ref is required")
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
	upsertExternalRef(channel, ref)
	channel.UpdatedAt = time.Now().UTC()
	if err := s.saveLocked(data); err != nil {
		return nil, err
	}
	return summarizeChannel(channel), nil
}

func (s *CollabStore) MarkChannelStatus(channelID, status string, metadata map[string]string) (*CollabChannelSummary, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return nil, errors.New("channel_id is required")
	}
	status = normalizeChannelStatus(status)
	if status == "" {
		return nil, errors.New("status is required")
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
	channel.Status = status
	if len(metadata) > 0 {
		if channel.Metadata == nil {
			channel.Metadata = map[string]string{}
		}
		for k, v := range metadata {
			channel.Metadata[k] = v
		}
	}
	channel.UpdatedAt = time.Now().UTC()
	if err := s.saveLocked(data); err != nil {
		return nil, err
	}
	return summarizeChannel(channel), nil
}

func (s *CollabStore) GetOrCreateLinearIssueChannel(issue *LinearIssue, participants []string, metadata map[string]string) (*CollabChannelSummary, error) {
	if issue == nil {
		return nil, errors.New("issue is required")
	}
	identifier := strings.ToUpper(strings.TrimSpace(issue.Identifier))
	if identifier == "" {
		return nil, errors.New("issue identifier is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	title := identifier
	if strings.TrimSpace(issue.Title) != "" {
		title = identifier + " · " + strings.TrimSpace(issue.Title)
	}
	key := "linear:" + identifier
	participants = normalizeStrings(participants)
	metadata = cloneMap(metadata)
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["linear_issue_id"] = identifier
	if issue.URL != "" {
		metadata["linear_issue_url"] = issue.URL
	}
	if issue.StateName != "" {
		metadata["linear_issue_status"] = issue.StateName
	}
	if issue.TeamKey != "" {
		metadata["linear_team_key"] = issue.TeamKey
	}
	ref := &CollabExternalRef{
		Source:     "linear",
		ExternalID: identifier,
		Title:      issue.Title,
		URL:        issue.URL,
		Metadata: map[string]string{
			"state_name": issue.StateName,
			"state_type": issue.StateType,
			"team_key":   issue.TeamKey,
			"team_name":  issue.TeamName,
		},
	}

	channel := findChannelByKey(data.Channels, key)
	if channel == nil {
		channel = findChannelByExternalRef(data.Channels, "linear", identifier, "")
	}
	if channel != nil {
		if title != "" {
			channel.Title = title
		}
		channel.Participants = normalizeStrings(append(channel.Participants, participants...))
		if channel.Metadata == nil {
			channel.Metadata = map[string]string{}
		}
		for k, v := range metadata {
			channel.Metadata[k] = v
		}
		channel.Key = key
		upsertExternalRef(channel, ref)
		channel.UpdatedAt = now
		if err := s.saveLocked(data); err != nil {
			return nil, err
		}
		return summarizeChannel(channel), nil
	}

	channel = &CollabChannel{
		ID:           newCollabID(),
		Key:          key,
		Title:        title,
		Status:       defaultChannelStatus,
		Participants: participants,
		Metadata:     metadata,
		ExternalRefs: []*CollabExternalRef{normalizeExternalRef(ref)},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	data.Channels = append(data.Channels, channel)
	if err := s.saveLocked(data); err != nil {
		return nil, err
	}
	return summarizeChannel(channel), nil
}

func (s *CollabStore) HasMessageWithMetadata(channelID, key, value string) (bool, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return false, errors.New("channel_id is required")
	}
	if strings.TrimSpace(key) == "" {
		return false, errors.New("metadata key is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return false, err
	}
	channel := findChannelByID(data.Channels, channelID)
	if channel == nil {
		return false, fmt.Errorf("channel %q not found", channelID)
	}
	for _, message := range channel.Messages {
		if message.Metadata[key] == value {
			return true, nil
		}
	}
	return false, nil
}

func (s *CollabStore) SetAgentFocus(agentID string, aliases, topics, capabilities []string, activeChannelID string, metadata map[string]string) (*CollabAgent, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return nil, errors.New("agent_id is required")
	}

	s.mu.Lock()
	data, err := s.loadLocked()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	now := time.Now().UTC()
	agent := findAgentByID(data.Agents, agentID)
	if agent == nil {
		agent = &CollabAgent{ID: agentID}
		data.Agents = append(data.Agents, agent)
	}
	if len(aliases) > 0 {
		agent.Aliases = normalizeStrings(append(agent.Aliases, aliases...))
	}
	if len(topics) > 0 {
		topics = normalizeStrings(topics)
		agent.Topics = normalizeStrings(append(agent.Topics, topics...))
		agent.CurrentFocus = topics
	}
	if len(capabilities) > 0 {
		agent.Capabilities = normalizeStrings(append(agent.Capabilities, capabilities...))
	}
	if activeChannelID != "" {
		agent.ActiveChannelIDs = normalizeStrings(append(agent.ActiveChannelIDs, activeChannelID))
	}
	if len(metadata) > 0 {
		if agent.Metadata == nil {
			agent.Metadata = map[string]string{}
		}
		for k, v := range metadata {
			agent.Metadata[k] = v
		}
	}
	agent.LastSeenAt = now
	if err := s.saveLocked(data); err != nil {
		s.mu.Unlock()
		return nil, err
	}
	s.mu.Unlock()
	return cloneAgent(agent), nil
}

func (s *CollabStore) ListAgents(query string, limit int) ([]*CollabAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query = strings.TrimSpace(strings.ToLower(query))
	var agents []*CollabAgent
	for _, agent := range data.Agents {
		if query != "" && !agentMatchesQuery(agent, query) {
			continue
		}
		agents = append(agents, cloneAgent(agent))
	}
	sort.Slice(agents, func(i, j int) bool {
		if !agents[i].LastSeenAt.Equal(agents[j].LastSeenAt) {
			return agents[i].LastSeenAt.After(agents[j].LastSeenAt)
		}
		return agents[i].ID < agents[j].ID
	})
	if len(agents) > limit {
		agents = agents[:limit]
	}
	return agents, nil
}

func (s *CollabStore) RouteMessage(req CollabRouteRequest) (*CollabRouteResult, error) {
	if strings.TrimSpace(req.Sender) == "" {
		return nil, errors.New("sender is required")
	}
	if strings.TrimSpace(req.Body) == "" {
		return nil, errors.New("body is required")
	}
	if req.Kind == "" {
		req.Kind = "message"
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 25
	}
	if req.DeliveryMode == "" {
		req.DeliveryMode = "routed"
	}

	s.mu.Lock()
	data, err := s.loadLocked()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	candidates := resolveAgentCandidates(data, req)
	result := &CollabRouteResult{
		RouteID:     newCollabID(),
		DryRun:      req.DryRun,
		Deliveries:  make([]*CollabRouteDelivery, 0, len(candidates)),
		Skipped:     []string{},
		ResolvedBy:  []string{},
		DeliveredAt: time.Now().UTC(),
	}
	if len(candidates) == 0 {
		s.mu.Unlock()
		return result, nil
	}

	resolvedBy := map[string]struct{}{}
	for _, candidate := range candidates {
		if _, ok := resolvedBy[candidate.reason]; candidate.reason != "" && !ok {
			resolvedBy[candidate.reason] = struct{}{}
			result.ResolvedBy = append(result.ResolvedBy, candidate.reason)
		}
	}

	published := make([]*CollabMessage, 0, len(candidates))
	for _, candidate := range candidates {
		channel := ensureAgentMailbox(data, candidate.agent.ID)
		metadata := cloneMap(req.Metadata)
		if metadata == nil {
			metadata = map[string]string{}
		}
		metadata["route_id"] = result.RouteID
		metadata["target_agent_id"] = candidate.agent.ID
		metadata["delivery_mode"] = req.DeliveryMode
		metadata["reason"] = candidate.reason
		metadata["score"] = fmt.Sprintf("%d", candidate.score)
		if req.Topic != "" {
			metadata["topic"] = req.Topic
		}
		if req.ChannelID != "" {
			metadata["source_channel_id"] = req.ChannelID
		}
		delivery := &CollabRouteDelivery{
			TargetAgentID:   candidate.agent.ID,
			TargetChannelID: channel.ID,
			Reason:          candidate.reason,
			Score:           candidate.score,
			Metadata:        cloneMap(metadata),
		}
		if !req.DryRun {
			message, err := appendMessageLocked(data, channel, req.Sender, req.Kind, req.Body, metadata)
			if err != nil {
				s.mu.Unlock()
				return nil, err
			}
			delivery.MessageID = message.ID
			published = append(published, cloneMessage(message))
			candidate.agent.LastSeenAt = result.DeliveredAt
		}
		result.Deliveries = append(result.Deliveries, delivery)
	}
	if !req.DryRun {
		if err := s.saveLocked(data); err != nil {
			s.mu.Unlock()
			return nil, err
		}
	}
	s.mu.Unlock()
	if !req.DryRun {
		s.publish(published...)
	}
	return result, nil
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
		Status:       defaultChannelStatus,
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

func (s *CollabStore) Subscribe(channelID string) (<-chan *CollabMessage, func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextSubID++
	id := s.nextSubID
	ch := make(chan *CollabMessage, 64)
	s.subscribed[id] = collabSubscription{
		channelID: strings.TrimSpace(channelID),
		ch:        ch,
	}
	return ch, func() {
		s.unsubscribe(id)
	}
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

	data, err := s.loadLocked()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	channel := findChannelByID(data.Channels, channelID)
	if channel == nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("channel %q not found", channelID)
	}

	message, err := appendMessageLocked(data, channel, sender, kind, body, metadata)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	if err := s.saveLocked(data); err != nil {
		s.mu.Unlock()
		return nil, err
	}
	s.mu.Unlock()
	s.publish(message)
	return message, nil
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

func (s *CollabStore) ReadAllMessages(afterSequence int64, limit int) ([]*CollabMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	out := make([]*CollabMessage, 0, limit)
	for _, channel := range data.Channels {
		for _, message := range channel.Messages {
			if message.Sequence <= afterSequence {
				continue
			}
			out = append(out, cloneMessage(message))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Sequence < out[j].Sequence
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

type routeCandidate struct {
	agent  *CollabAgent
	score  int
	reason string
}

func (s *CollabStore) loadLocked() (*collabData, error) {
	data := &collabData{Agents: []*CollabAgent{}, Channels: []*CollabChannel{}}
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
	if data.Agents == nil {
		data.Agents = []*CollabAgent{}
	}
	for _, channel := range data.Channels {
		channel.Status = normalizeChannelStatus(channel.Status)
		var refs []*CollabExternalRef
		for _, ref := range channel.ExternalRefs {
			if normalized := normalizeExternalRef(ref); normalized != nil {
				refs = append(refs, normalized)
			}
		}
		channel.ExternalRefs = refs
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

func appendMessageLocked(data *collabData, channel *CollabChannel, sender, kind, body string, metadata map[string]string) (*CollabMessage, error) {
	message := &CollabMessage{
		ID:        newCollabID(),
		Sequence:  data.NextSequence + 1,
		ChannelID: channel.ID,
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
	return cloneMessage(message), nil
}

func ensureAgentMailbox(data *collabData, agentID string) *CollabChannel {
	key := "agent:" + agentID
	for _, channel := range data.Channels {
		if channel.Key == key {
			channel.Participants = normalizeStrings(append(channel.Participants, agentID))
			channel.UpdatedAt = time.Now().UTC()
			return channel
		}
	}
	now := time.Now().UTC()
	channel := &CollabChannel{
		ID:           newCollabID(),
		Key:          key,
		Title:        "Mailbox for " + agentID,
		Status:       defaultChannelStatus,
		Participants: []string{agentID},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	data.Channels = append(data.Channels, channel)
	return channel
}

func (s *CollabStore) unsubscribe(id int) {
	s.mu.Lock()
	sub, ok := s.subscribed[id]
	if ok {
		delete(s.subscribed, id)
	}
	s.mu.Unlock()
	if ok {
		close(sub.ch)
	}
}

func (s *CollabStore) publish(messages ...*CollabMessage) {
	if len(messages) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	closeSet := map[int]collabSubscription{}
	for _, message := range messages {
		if message == nil {
			continue
		}
		for id, sub := range s.subscribed {
			if sub.channelID != "" && sub.channelID != message.ChannelID {
				continue
			}
			select {
			case sub.ch <- cloneMessage(message):
			default:
				closeSet[id] = sub
			}
		}
	}
	for id, sub := range closeSet {
		delete(s.subscribed, id)
		close(sub.ch)
	}
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
		Status:       normalizeChannelStatus(channel.Status),
		Participants: append([]string(nil), channel.Participants...),
		Metadata:     cloneMap(channel.Metadata),
		ExternalRefs: cloneExternalRefs(channel.ExternalRefs),
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

func findChannelByKey(channels []*CollabChannel, key string) *CollabChannel {
	for _, channel := range channels {
		if channel.Key == key {
			return channel
		}
	}
	return nil
}

func findChannelByExternalRef(channels []*CollabChannel, source, externalID, parentID string) *CollabChannel {
	source = strings.TrimSpace(strings.ToLower(source))
	externalID = strings.TrimSpace(externalID)
	parentID = strings.TrimSpace(parentID)
	for _, channel := range channels {
		for _, ref := range channel.ExternalRefs {
			if ref == nil {
				continue
			}
			if !strings.EqualFold(ref.Source, source) {
				continue
			}
			if ref.ExternalID != externalID {
				continue
			}
			if parentID != "" && ref.ParentID != parentID {
				continue
			}
			return channel
		}
	}
	return nil
}

func findMessageByID(channels []*CollabChannel, id string) *CollabMessage {
	for _, channel := range channels {
		for _, message := range channel.Messages {
			if message.ID == id {
				return message
			}
		}
	}
	return nil
}

func findAgentByID(agents []*CollabAgent, id string) *CollabAgent {
	for _, agent := range agents {
		if agent.ID == id {
			return agent
		}
	}
	return nil
}

func agentMatchesQuery(agent *CollabAgent, query string) bool {
	if agent == nil {
		return false
	}
	for _, candidate := range collectAgentTokens(agent) {
		if strings.Contains(candidate, query) {
			return true
		}
	}
	return false
}

func resolveAgentCandidates(data *collabData, req CollabRouteRequest) []routeCandidate {
	excluded := make(map[string]struct{})
	for _, agentID := range normalizeStrings(req.ExcludeAgentIDs) {
		excluded[agentID] = struct{}{}
	}
	explicit := make(map[string]struct{})
	for _, agentID := range normalizeStrings(req.TargetAgentIDs) {
		explicit[agentID] = struct{}{}
	}

	channelParticipants := map[string]struct{}{}
	if req.ChannelID != "" {
		if channel := findChannelByID(data.Channels, req.ChannelID); channel != nil {
			for _, participant := range channel.Participants {
				channelParticipants[participant] = struct{}{}
			}
		}
	}

	queryTokens := tokenize(req.Query)
	topic := strings.TrimSpace(strings.ToLower(req.Topic))
	out := make([]routeCandidate, 0, len(data.Agents))
	for _, agent := range data.Agents {
		if _, denied := excluded[agent.ID]; denied {
			continue
		}
		score := 0
		reasons := []string{}
		_, explicitTarget := explicit[agent.ID]
		if explicitTarget {
			score += 100
			reasons = append(reasons, "explicit_agent")
		}
		if _, inChannel := channelParticipants[agent.ID]; req.ChannelID != "" && inChannel {
			score += 50
			reasons = append(reasons, "source_channel_participant")
		}
		if topic != "" {
			if containsFold(agent.CurrentFocus, topic) {
				score += 60
				reasons = append(reasons, "current_focus")
			}
			if containsFold(agent.Topics, topic) {
				score += 40
				reasons = append(reasons, "agent_topic")
			}
			if containsFold(agent.Capabilities, topic) {
				score += 25
				reasons = append(reasons, "capability")
			}
		}
		if len(queryTokens) > 0 {
			tokenSet := make(map[string]struct{})
			for _, token := range collectAgentTokens(agent) {
				tokenSet[token] = struct{}{}
			}
			matches := 0
			for _, token := range queryTokens {
				if _, ok := tokenSet[token]; ok {
					matches++
				}
			}
			if matches > 0 {
				score += 15 * matches
				reasons = append(reasons, "query_match")
			}
		}
		if !explicitTarget && agent.ID == req.Sender {
			score -= 20
		}
		if score <= 0 {
			continue
		}
		out = append(out, routeCandidate{
			agent:  agent,
			score:  score,
			reason: strings.Join(reasons, "+"),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].score != out[j].score {
			return out[i].score > out[j].score
		}
		return out[i].agent.ID < out[j].agent.ID
	})
	if len(out) > req.Limit {
		out = out[:req.Limit]
	}
	return out
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
	if strings.Contains(strings.ToLower(channel.Status), query) {
		return true
	}
	for k, v := range channel.Metadata {
		if strings.Contains(strings.ToLower(k), query) || strings.Contains(strings.ToLower(v), query) {
			return true
		}
	}
	for _, ref := range channel.ExternalRefs {
		if ref == nil {
			continue
		}
		if strings.Contains(strings.ToLower(ref.Source), query) ||
			strings.Contains(strings.ToLower(ref.ExternalID), query) ||
			strings.Contains(strings.ToLower(ref.ParentID), query) ||
			strings.Contains(strings.ToLower(ref.Title), query) ||
			strings.Contains(strings.ToLower(ref.URL), query) {
			return true
		}
		for k, v := range ref.Metadata {
			if strings.Contains(strings.ToLower(k), query) || strings.Contains(strings.ToLower(v), query) {
				return true
			}
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

func cloneExternalRefs(in []*CollabExternalRef) []*CollabExternalRef {
	if len(in) == 0 {
		return nil
	}
	out := make([]*CollabExternalRef, 0, len(in))
	for _, ref := range in {
		if ref == nil {
			continue
		}
		copyRef := *ref
		copyRef.Metadata = cloneMap(ref.Metadata)
		out = append(out, &copyRef)
	}
	return out
}

func cloneAgent(in *CollabAgent) *CollabAgent {
	if in == nil {
		return nil
	}
	out := *in
	out.Aliases = append([]string(nil), in.Aliases...)
	out.Capabilities = append([]string(nil), in.Capabilities...)
	out.Topics = append([]string(nil), in.Topics...)
	out.CurrentFocus = append([]string(nil), in.CurrentFocus...)
	out.ActiveChannelIDs = append([]string(nil), in.ActiveChannelIDs...)
	out.Metadata = cloneMap(in.Metadata)
	return &out
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

func tokenize(raw string) []string {
	parts := strings.FieldsFunc(strings.ToLower(raw), func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return false
		case r >= '0' && r <= '9':
			return false
		default:
			return true
		}
	})
	return normalizeStrings(parts)
}

func collectAgentTokens(agent *CollabAgent) []string {
	if agent == nil {
		return nil
	}
	values := []string{agent.ID}
	values = append(values, agent.Aliases...)
	values = append(values, agent.Capabilities...)
	values = append(values, agent.Topics...)
	values = append(values, agent.CurrentFocus...)
	for k, v := range agent.Metadata {
		values = append(values, k, v)
	}
	var tokens []string
	for _, value := range values {
		tokens = append(tokens, tokenize(value)...)
	}
	return normalizeStrings(tokens)
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

const defaultChannelStatus = "open"

func normalizeChannelStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", defaultChannelStatus:
		return defaultChannelStatus
	case "blocked":
		return "blocked"
	case "waiting_approval", "waiting-approval":
		return "waiting_approval"
	case "done":
		return "done"
	default:
		return ""
	}
}

func normalizeExternalRef(ref *CollabExternalRef) *CollabExternalRef {
	if ref == nil {
		return nil
	}
	source := strings.TrimSpace(strings.ToLower(ref.Source))
	externalID := strings.TrimSpace(ref.ExternalID)
	if source == "" || externalID == "" {
		return nil
	}
	return &CollabExternalRef{
		Source:     source,
		ExternalID: externalID,
		ParentID:   strings.TrimSpace(ref.ParentID),
		Title:      strings.TrimSpace(ref.Title),
		URL:        strings.TrimSpace(ref.URL),
		Metadata:   cloneMap(ref.Metadata),
	}
}

func upsertExternalRef(channel *CollabChannel, ref *CollabExternalRef) {
	ref = normalizeExternalRef(ref)
	if channel == nil || ref == nil {
		return
	}
	for i, existing := range channel.ExternalRefs {
		if existing == nil {
			continue
		}
		if strings.EqualFold(existing.Source, ref.Source) && existing.ExternalID == ref.ExternalID && existing.ParentID == ref.ParentID {
			channel.ExternalRefs[i] = ref
			return
		}
	}
	channel.ExternalRefs = append(channel.ExternalRefs, ref)
	sort.Slice(channel.ExternalRefs, func(i, j int) bool {
		left := channel.ExternalRefs[i]
		right := channel.ExternalRefs[j]
		if left.Source != right.Source {
			return left.Source < right.Source
		}
		if left.ParentID != right.ParentID {
			return left.ParentID < right.ParentID
		}
		return left.ExternalID < right.ExternalID
	})
}

func newCollabID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("collab-%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}
