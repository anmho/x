package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultCloudflareAPIBaseURL = "https://api.cloudflare.com/client/v4"

type cloudflareDNSProvider struct {
	baseURL string
	token   string
	client  *http.Client
}

func newCloudflareDNSProviderFromEnv() (*cloudflareDNSProvider, error) {
	token := strings.TrimSpace(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if token == "" {
		return nil, errors.New("missing CLOUDFLARE_API_TOKEN")
	}
	baseURL := strings.TrimRight(getEnv("CLOUDFLARE_API_BASE_URL", defaultCloudflareAPIBaseURL), "/")
	return &cloudflareDNSProvider{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}, nil
}

func newCloudflareDNSProvider(baseURL string, token string, client *http.Client) *cloudflareDNSProvider {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &cloudflareDNSProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   strings.TrimSpace(token),
		client:  client,
	}
}

func (p *cloudflareDNSProvider) Name() string {
	return "cloudflare"
}

func (p *cloudflareDNSProvider) ListZones(ctx context.Context) ([]domainZone, error) {
	u := p.baseURL + "/zones?per_page=100"
	var payload cloudflareResponse[[]cloudflareZone]
	if err := p.doJSON(ctx, http.MethodGet, u, nil, &payload); err != nil {
		return nil, err
	}

	out := make([]domainZone, 0, len(payload.Result))
	for _, zone := range payload.Result {
		out = append(out, domainZone{
			Name:     zone.Name,
			Provider: p.Name(),
			ZoneID:   zone.ID,
		})
	}
	return out, nil
}

func (p *cloudflareDNSProvider) ListRecords(ctx context.Context, zone domainZoneSpec) ([]domainRecord, error) {
	zoneID, err := p.resolveZoneID(ctx, zone)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s/zones/%s/dns_records?per_page=200", p.baseURL, url.PathEscape(zoneID))
	var payload cloudflareResponse[[]cloudflareDNSRecord]
	if err := p.doJSON(ctx, http.MethodGet, u, nil, &payload); err != nil {
		return nil, err
	}

	out := make([]domainRecord, 0, len(payload.Result))
	for _, record := range payload.Result {
		out = append(out, domainRecord{
			ID:       record.ID,
			Type:     record.Type,
			Name:     record.Name,
			Content:  record.Content,
			TTL:      record.TTL,
			Proxied:  record.Proxied,
			Provider: p.Name(),
			Metadata: map[string]string{"zone_id": zoneID},
		})
	}
	return out, nil
}

func (p *cloudflareDNSProvider) CreateRecord(ctx context.Context, zone domainZoneSpec, input domainRecordInput) (domainRecord, error) {
	zoneID, err := p.resolveZoneID(ctx, zone)
	if err != nil {
		return domainRecord{}, err
	}
	body := map[string]any{
		"type":    strings.ToUpper(strings.TrimSpace(input.Type)),
		"name":    strings.TrimSpace(input.Name),
		"content": strings.TrimSpace(input.Content),
	}
	if input.TTL > 0 {
		body["ttl"] = input.TTL
	}
	if input.Proxied != nil {
		body["proxied"] = *input.Proxied
	}

	u := fmt.Sprintf("%s/zones/%s/dns_records", p.baseURL, url.PathEscape(zoneID))
	var payload cloudflareResponse[cloudflareDNSRecord]
	if err := p.doJSON(ctx, http.MethodPost, u, body, &payload); err != nil {
		return domainRecord{}, err
	}

	return domainRecord{
		ID:       payload.Result.ID,
		Type:     payload.Result.Type,
		Name:     payload.Result.Name,
		Content:  payload.Result.Content,
		TTL:      payload.Result.TTL,
		Proxied:  payload.Result.Proxied,
		Provider: p.Name(),
		Metadata: map[string]string{"zone_id": zoneID},
	}, nil
}

func (p *cloudflareDNSProvider) UpdateRecord(ctx context.Context, zone domainZoneSpec, recordID string, input domainRecordInput) (domainRecord, error) {
	if strings.TrimSpace(recordID) == "" {
		return domainRecord{}, errors.New("record id is required")
	}
	zoneID, err := p.resolveZoneID(ctx, zone)
	if err != nil {
		return domainRecord{}, err
	}
	body := map[string]any{
		"type":    strings.ToUpper(strings.TrimSpace(input.Type)),
		"name":    strings.TrimSpace(input.Name),
		"content": strings.TrimSpace(input.Content),
	}
	if input.TTL > 0 {
		body["ttl"] = input.TTL
	}
	if input.Proxied != nil {
		body["proxied"] = *input.Proxied
	}

	u := fmt.Sprintf("%s/zones/%s/dns_records/%s", p.baseURL, url.PathEscape(zoneID), url.PathEscape(strings.TrimSpace(recordID)))
	var payload cloudflareResponse[cloudflareDNSRecord]
	if err := p.doJSON(ctx, http.MethodPut, u, body, &payload); err != nil {
		return domainRecord{}, err
	}

	return domainRecord{
		ID:       payload.Result.ID,
		Type:     payload.Result.Type,
		Name:     payload.Result.Name,
		Content:  payload.Result.Content,
		TTL:      payload.Result.TTL,
		Proxied:  payload.Result.Proxied,
		Provider: p.Name(),
		Metadata: map[string]string{"zone_id": zoneID},
	}, nil
}

func (p *cloudflareDNSProvider) DeleteRecord(ctx context.Context, zone domainZoneSpec, recordID string) error {
	if strings.TrimSpace(recordID) == "" {
		return errors.New("record id is required")
	}
	zoneID, err := p.resolveZoneID(ctx, zone)
	if err != nil {
		return err
	}

	u := fmt.Sprintf("%s/zones/%s/dns_records/%s", p.baseURL, url.PathEscape(zoneID), url.PathEscape(strings.TrimSpace(recordID)))
	var payload cloudflareResponse[cloudflareDeleteResult]
	if err := p.doJSON(ctx, http.MethodDelete, u, nil, &payload); err != nil {
		return err
	}
	return nil
}

func (p *cloudflareDNSProvider) resolveZoneID(ctx context.Context, zone domainZoneSpec) (string, error) {
	zoneID := strings.TrimSpace(zone.ZoneID)
	if zoneID != "" {
		return zoneID, nil
	}
	zoneName := strings.TrimSpace(zone.Name)
	if zoneName == "" {
		return "", errors.New("zone name is required when zone_id is missing")
	}
	u := fmt.Sprintf("%s/zones?name=%s&per_page=1", p.baseURL, url.QueryEscape(zoneName))
	var payload cloudflareResponse[[]cloudflareZone]
	if err := p.doJSON(ctx, http.MethodGet, u, nil, &payload); err != nil {
		return "", err
	}
	if len(payload.Result) == 0 || strings.TrimSpace(payload.Result[0].ID) == "" {
		return "", fmt.Errorf("zone %q not found in cloudflare", zoneName)
	}
	return strings.TrimSpace(payload.Result[0].ID), nil
}

func (p *cloudflareDNSProvider) doJSON(ctx context.Context, method string, endpoint string, body any, out any) error {
	if strings.TrimSpace(p.token) == "" {
		return errors.New("cloudflare token is empty")
	}
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cloudflare %s %s failed: status=%d body=%s", method, endpoint, resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return err
	}

	switch typed := out.(type) {
	case *cloudflareResponse[[]cloudflareZone]:
		if !typed.Success {
			return errors.New("cloudflare API returned success=false")
		}
	case *cloudflareResponse[[]cloudflareDNSRecord]:
		if !typed.Success {
			return errors.New("cloudflare API returned success=false")
		}
	case *cloudflareResponse[cloudflareDNSRecord]:
		if !typed.Success {
			return errors.New("cloudflare API returned success=false")
		}
	case *cloudflareResponse[cloudflareDeleteResult]:
		if !typed.Success {
			return errors.New("cloudflare API returned success=false")
		}
	}
	return nil
}

type cloudflareResponse[T any] struct {
	Success bool `json:"success"`
	Result  T    `json:"result"`
}

type cloudflareZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cloudflareDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied *bool  `json:"proxied,omitempty"`
}

type cloudflareDeleteResult struct {
	ID string `json:"id"`
}
