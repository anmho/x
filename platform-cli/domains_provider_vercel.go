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

const defaultVercelAPIBaseURL = "https://api.vercel.com"

type vercelDNSProvider struct {
	baseURL string
	token   string
	client  *http.Client
}

func newVercelDNSProviderFromEnv() (*vercelDNSProvider, error) {
	token := strings.TrimSpace(os.Getenv("VERCEL_API_TOKEN"))
	if token == "" {
		return nil, errors.New("missing VERCEL_API_TOKEN")
	}
	baseURL := strings.TrimRight(getEnv("VERCEL_API_BASE_URL", defaultVercelAPIBaseURL), "/")
	return &vercelDNSProvider{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}, nil
}

func newVercelDNSProvider(baseURL string, token string, client *http.Client) *vercelDNSProvider {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &vercelDNSProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   strings.TrimSpace(token),
		client:  client,
	}
}

func (p *vercelDNSProvider) Name() string {
	return "vercel"
}

func (p *vercelDNSProvider) ListZones(ctx context.Context) ([]domainZone, error) {
	u := p.baseURL + "/v5/domains?limit=100"
	var payload vercelDomainsResponse
	if err := p.doJSON(ctx, http.MethodGet, u, nil, &payload); err != nil {
		return nil, err
	}

	out := make([]domainZone, 0, len(payload.Domains))
	for _, zone := range payload.Domains {
		out = append(out, domainZone{
			Name:     zone.Name,
			Provider: p.Name(),
			ZoneID:   zone.Name,
		})
	}
	return out, nil
}

func (p *vercelDNSProvider) ListRecords(ctx context.Context, zone domainZoneSpec) ([]domainRecord, error) {
	zoneName, err := p.resolveZoneName(zone)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/v4/domains/%s/records", p.baseURL, url.PathEscape(zoneName))
	var payload vercelRecordsResponse
	if err := p.doJSON(ctx, http.MethodGet, u, nil, &payload); err != nil {
		return nil, err
	}
	out := make([]domainRecord, 0, len(payload.Records))
	for _, record := range payload.Records {
		out = append(out, domainRecord{
			ID:       record.ID,
			Type:     strings.ToUpper(record.Type),
			Name:     record.Name,
			Content:  record.Value,
			TTL:      record.TTL,
			Provider: p.Name(),
			Metadata: map[string]string{"zone_id": zoneName},
		})
	}
	return out, nil
}

func (p *vercelDNSProvider) CreateRecord(ctx context.Context, zone domainZoneSpec, input domainRecordInput) (domainRecord, error) {
	zoneName, err := p.resolveZoneName(zone)
	if err != nil {
		return domainRecord{}, err
	}
	body := map[string]any{
		"type":  strings.ToUpper(strings.TrimSpace(input.Type)),
		"name":  strings.TrimSpace(input.Name),
		"value": strings.TrimSpace(input.Content),
	}
	if input.TTL > 0 {
		body["ttl"] = input.TTL
	}

	u := fmt.Sprintf("%s/v2/domains/%s/records", p.baseURL, url.PathEscape(zoneName))
	var payload vercelRecordResponse
	if err := p.doJSON(ctx, http.MethodPost, u, body, &payload); err != nil {
		return domainRecord{}, err
	}

	return domainRecord{
		ID:       payload.Record.ID,
		Type:     strings.ToUpper(payload.Record.Type),
		Name:     payload.Record.Name,
		Content:  payload.Record.Value,
		TTL:      payload.Record.TTL,
		Provider: p.Name(),
		Metadata: map[string]string{"zone_id": zoneName},
	}, nil
}

func (p *vercelDNSProvider) UpdateRecord(ctx context.Context, zone domainZoneSpec, recordID string, input domainRecordInput) (domainRecord, error) {
	zoneName, err := p.resolveZoneName(zone)
	if err != nil {
		return domainRecord{}, err
	}
	recordID = strings.TrimSpace(recordID)
	if recordID == "" {
		return domainRecord{}, errors.New("record id is required")
	}
	body := map[string]any{
		"type":  strings.ToUpper(strings.TrimSpace(input.Type)),
		"name":  strings.TrimSpace(input.Name),
		"value": strings.TrimSpace(input.Content),
	}
	if input.TTL > 0 {
		body["ttl"] = input.TTL
	}

	u := fmt.Sprintf("%s/v2/domains/%s/records/%s", p.baseURL, url.PathEscape(zoneName), url.PathEscape(recordID))
	var payload vercelRecordResponse
	if err := p.doJSON(ctx, http.MethodPatch, u, body, &payload); err != nil {
		return domainRecord{}, err
	}

	return domainRecord{
		ID:       payload.Record.ID,
		Type:     strings.ToUpper(payload.Record.Type),
		Name:     payload.Record.Name,
		Content:  payload.Record.Value,
		TTL:      payload.Record.TTL,
		Provider: p.Name(),
		Metadata: map[string]string{"zone_id": zoneName},
	}, nil
}

func (p *vercelDNSProvider) DeleteRecord(ctx context.Context, zone domainZoneSpec, recordID string) error {
	zoneName, err := p.resolveZoneName(zone)
	if err != nil {
		return err
	}
	recordID = strings.TrimSpace(recordID)
	if recordID == "" {
		return errors.New("record id is required")
	}
	u := fmt.Sprintf("%s/v2/domains/%s/records/%s", p.baseURL, url.PathEscape(zoneName), url.PathEscape(recordID))
	return p.doJSON(ctx, http.MethodDelete, u, nil, nil)
}

func (p *vercelDNSProvider) resolveZoneName(zone domainZoneSpec) (string, error) {
	name := strings.TrimSpace(zone.Name)
	if name == "" {
		return "", errors.New("zone name is required")
	}
	return name, nil
}

func (p *vercelDNSProvider) doJSON(ctx context.Context, method string, endpoint string, body any, out any) error {
	if strings.TrimSpace(p.token) == "" {
		return errors.New("vercel token is empty")
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
		return fmt.Errorf("vercel %s %s failed: status=%d body=%s", method, endpoint, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

type vercelDomainsResponse struct {
	Domains []vercelDomain `json:"domains"`
}

type vercelDomain struct {
	Name string `json:"name"`
}

type vercelRecordsResponse struct {
	Records []vercelRecord `json:"records"`
}

type vercelRecordResponse struct {
	Record vercelRecord `json:"record"`
}

type vercelRecord struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}
