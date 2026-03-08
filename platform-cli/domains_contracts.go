package main

import (
	"context"
	"strings"
)

type domainZoneSpec struct {
	Name         string             `json:"name"`
	Provider     string             `json:"provider"`
	ZoneID       string             `json:"zone_id,omitempty"`
	DesiredState string             `json:"desired_state,omitempty"` // present|absent
	Project      string             `json:"project,omitempty"`
	Records      []domainRecordSpec `json:"records,omitempty"`
}

type domainRecordSpec struct {
	ID             string            `json:"id,omitempty"`
	Type           string            `json:"type"`
	Name           string            `json:"name"`
	Content        string            `json:"content"`
	TTL            int               `json:"ttl,omitempty"`
	Proxied        *bool             `json:"proxied,omitempty"`
	DesiredState   string            `json:"desired_state,omitempty"` // present|absent
	DeploymentLink *domainDeployLink `json:"deployment_link,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type domainDeployLink struct {
	Project      string `json:"project,omitempty"`
	DeploymentID string `json:"deployment_id,omitempty"`
	Host         string `json:"host,omitempty"`
}

type domainZone struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	ZoneID   string `json:"zone_id,omitempty"`
	Project  string `json:"project,omitempty"`
}

type domainRecord struct {
	ID       string            `json:"id,omitempty"`
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Content  string            `json:"content"`
	TTL      int               `json:"ttl,omitempty"`
	Proxied  *bool             `json:"proxied,omitempty"`
	Provider string            `json:"provider,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type domainRecordInput struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl,omitempty"`
	Proxied *bool  `json:"proxied,omitempty"`
}

// domainProvider is the provider adapter contract used by the unified control plane.
type domainProvider interface {
	Name() string
	ListZones(ctx context.Context) ([]domainZone, error)
	ListRecords(ctx context.Context, zone domainZoneSpec) ([]domainRecord, error)
	CreateRecord(ctx context.Context, zone domainZoneSpec, input domainRecordInput) (domainRecord, error)
	UpdateRecord(ctx context.Context, zone domainZoneSpec, recordID string, input domainRecordInput) (domainRecord, error)
	DeleteRecord(ctx context.Context, zone domainZoneSpec, recordID string) error
}

func normalizeDomainProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "cloudflare", "cf":
		return "cloudflare"
	case "vercel":
		return "vercel"
	default:
		return strings.ToLower(strings.TrimSpace(provider))
	}
}
