package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCloudflareListZones(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("expected auth header, got %q", got)
		}
		if r.URL.Path != "/zones" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("per_page") != "100" {
			t.Fatalf("unexpected per_page query: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"result": []map[string]any{
				{"id": "zone-1", "name": "example.com"},
			},
		})
	}))
	defer server.Close()

	p := newCloudflareDNSProvider(server.URL, "test-token", server.Client())
	zones, err := p.ListZones(context.Background())
	if err != nil {
		t.Fatalf("ListZones error: %v", err)
	}
	if len(zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(zones))
	}
	if zones[0].ZoneID != "zone-1" || zones[0].Name != "example.com" {
		t.Fatalf("unexpected zone: %+v", zones[0])
	}
}

func TestCloudflareCreateRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/zones/zone-1/dns_records" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if !strings.Contains(string(body), "\"name\":\"c\"") {
			t.Fatalf("expected name in payload: %s", string(body))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"result": map[string]any{
				"id":      "record-1",
				"type":    "CNAME",
				"name":    "c",
				"content": "cname.vercel-dns.com",
				"ttl":     300,
				"proxied": false,
			},
		})
	}))
	defer server.Close()

	p := newCloudflareDNSProvider(server.URL, "test-token", server.Client())
	proxied := false
	record, err := p.CreateRecord(context.Background(), domainZoneSpec{
		Name:   "example.com",
		ZoneID: "zone-1",
	}, domainRecordInput{
		Type:    "CNAME",
		Name:    "c",
		Content: "cname.vercel-dns.com",
		TTL:     300,
		Proxied: &proxied,
	})
	if err != nil {
		t.Fatalf("CreateRecord error: %v", err)
	}
	if record.ID != "record-1" || record.Type != "CNAME" {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func TestCloudflareDeleteRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/zones/zone-1/dns_records/record-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"result":  map[string]any{"id": "record-1"},
		})
	}))
	defer server.Close()

	p := newCloudflareDNSProvider(server.URL, "test-token", server.Client())
	err := p.DeleteRecord(context.Background(), domainZoneSpec{
		Name:   "example.com",
		ZoneID: "zone-1",
	}, "record-1")
	if err != nil {
		t.Fatalf("DeleteRecord error: %v", err)
	}
}
