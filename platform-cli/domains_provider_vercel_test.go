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

func TestVercelListZones(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("expected auth header, got %q", got)
		}
		if r.URL.Path != "/v5/domains" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"domains": []map[string]any{
				{"name": "anmhela.com"},
			},
		})
	}))
	defer server.Close()

	p := newVercelDNSProvider(server.URL, "test-token", server.Client())
	zones, err := p.ListZones(context.Background())
	if err != nil {
		t.Fatalf("ListZones error: %v", err)
	}
	if len(zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(zones))
	}
	if zones[0].Name != "anmhela.com" || zones[0].ZoneID != "anmhela.com" {
		t.Fatalf("unexpected zone: %+v", zones[0])
	}
}

func TestVercelCreateRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v2/domains/anmhela.com/records" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if !strings.Contains(string(body), "\"name\":\"c\"") {
			t.Fatalf("expected name in payload: %s", string(body))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"record": map[string]any{
				"id":    "rec_1",
				"type":  "CNAME",
				"name":  "c",
				"value": "cname.vercel-dns.com",
				"ttl":   300,
			},
		})
	}))
	defer server.Close()

	p := newVercelDNSProvider(server.URL, "test-token", server.Client())
	record, err := p.CreateRecord(context.Background(), domainZoneSpec{
		Name: "anmhela.com",
	}, domainRecordInput{
		Type:    "CNAME",
		Name:    "c",
		Content: "cname.vercel-dns.com",
		TTL:     300,
	})
	if err != nil {
		t.Fatalf("CreateRecord error: %v", err)
	}
	if record.ID != "rec_1" || record.Content != "cname.vercel-dns.com" {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func TestVercelDeleteRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v2/domains/anmhela.com/records/rec_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	p := newVercelDNSProvider(server.URL, "test-token", server.Client())
	err := p.DeleteRecord(context.Background(), domainZoneSpec{Name: "anmhela.com"}, "rec_1")
	if err != nil {
		t.Fatalf("DeleteRecord error: %v", err)
	}
}
