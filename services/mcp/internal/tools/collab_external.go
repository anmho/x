package tools

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *CollabStore) GetOrCreateExternalChannel(source, externalID, parentID, title, url string, participants []string, metadata, refMetadata map[string]string) (*CollabChannelSummary, error) {
	source = strings.TrimSpace(strings.ToLower(source))
	externalID = strings.TrimSpace(externalID)
	parentID = strings.TrimSpace(parentID)
	title = strings.TrimSpace(title)
	url = strings.TrimSpace(url)
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

	now := time.Now().UTC()
	key := source + ":" + externalID
	participants = normalizeStrings(participants)
	metadata = cloneMap(metadata)
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["external_source"] = source
	metadata["external_id"] = externalID
	if parentID != "" {
		metadata["external_parent_id"] = parentID
	}

	ref := &CollabExternalRef{
		Source:     source,
		ExternalID: externalID,
		ParentID:   parentID,
		Title:      title,
		URL:        url,
		Metadata:   cloneMap(refMetadata),
	}

	channel := findChannelByKey(data.Channels, key)
	if channel == nil {
		channel = findChannelByExternalRef(data.Channels, source, externalID, parentID)
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
		if channel.Key == "" {
			channel.Key = key
		}
		upsertExternalRef(channel, ref)
		channel.UpdatedAt = now
		if err := s.saveLocked(data); err != nil {
			return nil, err
		}
		return summarizeChannel(channel), nil
	}

	if title == "" {
		title = fmt.Sprintf("%s · %s", source, externalID)
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
