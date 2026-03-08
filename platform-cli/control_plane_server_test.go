package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestControlPlaneServerListDomains(t *testing.T) {
	handler := newControlPlaneHTTPServer(&controlPlaneConfig{
		Projects: []controlPlaneProject{
			{
				Name: "cloud-console",
				Domains: []domainZoneSpec{
					{Name: "anmhela.com", Provider: "cloudflare", ZoneID: "zone-1"},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/domains", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Domains []domainZone `json:"domains"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if len(body.Domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(body.Domains))
	}
}

func TestControlPlaneServerRecordCRUD(t *testing.T) {
	fake := &fakeDomainProvider{
		existing: []domainRecord{
			{ID: "rec-1", Type: "CNAME", Name: "c.anmhela.com", Content: "cname.vercel-dns.com", TTL: 300},
		},
	}
	withFakeProvider(t, fake)

	handler := newControlPlaneHTTPServer(&controlPlaneConfig{
		Projects: []controlPlaneProject{
			{
				Name: "cloud-console",
				Domains: []domainZoneSpec{
					{Name: "anmhela.com", Provider: "cloudflare", ZoneID: "zone-1"},
				},
			},
		},
	})

	getReq := httptest.NewRequest(http.MethodGet, "/v1/domains/anmhela.com/records?provider=cloudflare", nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET expected 200, got %d body=%s", getRec.Code, getRec.Body.String())
	}

	createBody := []byte(`{"type":"CNAME","name":"c","content":"cname.vercel-dns.com","ttl":300}`)
	createReq := httptest.NewRequest(http.MethodPost, "/v1/domains/anmhela.com/records?provider=cloudflare", bytes.NewReader(createBody))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("POST expected 201, got %d body=%s", createRec.Code, createRec.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/domains/anmhela.com/records/rec-1?provider=cloudflare", nil)
	deleteRec := httptest.NewRecorder()
	handler.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("DELETE expected 200, got %d body=%s", deleteRec.Code, deleteRec.Body.String())
	}

	if fake.createCalls != 1 || fake.deleteCalls != 1 {
		t.Fatalf("expected create=1 delete=1, got create=%d delete=%d", fake.createCalls, fake.deleteCalls)
	}
}

func TestControlPlaneServerReconcileEndpoint(t *testing.T) {
	old := runControlPlaneReconcile
	runControlPlaneReconcile = func(cfg *controlPlaneConfig, selectedProject string, dryRun bool, prune bool) error {
		if selectedProject != "cloud-console" {
			t.Fatalf("unexpected project: %s", selectedProject)
		}
		if !dryRun {
			t.Fatalf("expected dryRun=true")
		}
		if !prune {
			t.Fatalf("expected prune=true")
		}
		return nil
	}
	t.Cleanup(func() { runControlPlaneReconcile = old })

	handler := newControlPlaneHTTPServer(&controlPlaneConfig{})
	body := []byte(`{"project":"cloud-console","dry_run":true,"prune":true}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/domains/reconcile", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
