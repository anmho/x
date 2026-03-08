package main

import (
	"context"
	"testing"
)

type fakeDomainProvider struct {
	existing    []domainRecord
	created     []domainRecordInput
	updated     []domainRecordInput
	updatedIDs  []string
	deletedIDs  []string
	listCalls   int
	createCalls int
	updateCalls int
	deleteCalls int
}

func (f *fakeDomainProvider) Name() string { return "fake" }

func (f *fakeDomainProvider) ListZones(_ context.Context) ([]domainZone, error) { return nil, nil }

func (f *fakeDomainProvider) ListRecords(_ context.Context, _ domainZoneSpec) ([]domainRecord, error) {
	f.listCalls++
	out := make([]domainRecord, len(f.existing))
	copy(out, f.existing)
	return out, nil
}

func (f *fakeDomainProvider) CreateRecord(_ context.Context, _ domainZoneSpec, input domainRecordInput) (domainRecord, error) {
	f.createCalls++
	f.created = append(f.created, input)
	return domainRecord{
		ID:      "created-1",
		Type:    input.Type,
		Name:    input.Name,
		Content: input.Content,
		TTL:     input.TTL,
	}, nil
}

func (f *fakeDomainProvider) UpdateRecord(_ context.Context, _ domainZoneSpec, recordID string, input domainRecordInput) (domainRecord, error) {
	f.updateCalls++
	f.updatedIDs = append(f.updatedIDs, recordID)
	f.updated = append(f.updated, input)
	return domainRecord{
		ID:      recordID,
		Type:    input.Type,
		Name:    input.Name,
		Content: input.Content,
		TTL:     input.TTL,
	}, nil
}

func (f *fakeDomainProvider) DeleteRecord(_ context.Context, _ domainZoneSpec, recordID string) error {
	f.deleteCalls++
	f.deletedIDs = append(f.deletedIDs, recordID)
	return nil
}

func withFakeProvider(t *testing.T, provider domainProvider) {
	t.Helper()
	previous := resolveDomainProvider
	resolveDomainProvider = func(_ string) (domainProvider, error) { return provider, nil }
	t.Cleanup(func() {
		resolveDomainProvider = previous
	})
}

func TestReconcileDomainSetCreate(t *testing.T) {
	fake := &fakeDomainProvider{existing: []domainRecord{}}
	withFakeProvider(t, fake)

	err := reconcileDomainSet("cloud-console", []domainZoneSpec{
		{
			Name:     "anmhela.com",
			Provider: "cloudflare",
			Records: []domainRecordSpec{
				{Type: "CNAME", Name: "c", Content: "cname.vercel-dns.com", TTL: 300},
			},
		},
	}, false, false)
	if err != nil {
		t.Fatalf("reconcileDomainSet error: %v", err)
	}
	if fake.createCalls != 1 {
		t.Fatalf("expected 1 create call, got %d", fake.createCalls)
	}
}

func TestReconcileDomainSetNoOp(t *testing.T) {
	fake := &fakeDomainProvider{
		existing: []domainRecord{
			{ID: "rec-1", Type: "CNAME", Name: "c.anmhela.com", Content: "cname.vercel-dns.com", TTL: 300},
		},
	}
	withFakeProvider(t, fake)

	err := reconcileDomainSet("cloud-console", []domainZoneSpec{
		{
			Name:     "anmhela.com",
			Provider: "cloudflare",
			Records: []domainRecordSpec{
				{Type: "CNAME", Name: "c", Content: "cname.vercel-dns.com", TTL: 300},
			},
		},
	}, false, false)
	if err != nil {
		t.Fatalf("reconcileDomainSet error: %v", err)
	}
	if fake.createCalls != 0 || fake.updateCalls != 0 || fake.deleteCalls != 0 {
		t.Fatalf("expected no writes; got create=%d update=%d delete=%d", fake.createCalls, fake.updateCalls, fake.deleteCalls)
	}
}

func TestReconcileDomainSetUpdateOnDrift(t *testing.T) {
	fake := &fakeDomainProvider{
		existing: []domainRecord{
			{ID: "rec-1", Type: "CNAME", Name: "c.anmhela.com", Content: "old-target.example.com", TTL: 300},
		},
	}
	withFakeProvider(t, fake)

	err := reconcileDomainSet("cloud-console", []domainZoneSpec{
		{
			Name:     "anmhela.com",
			Provider: "cloudflare",
			Records: []domainRecordSpec{
				{Type: "CNAME", Name: "c", Content: "cname.vercel-dns.com", TTL: 300},
			},
		},
	}, false, false)
	if err != nil {
		t.Fatalf("reconcileDomainSet error: %v", err)
	}
	if fake.updateCalls != 1 {
		t.Fatalf("expected 1 update call, got %d", fake.updateCalls)
	}
	if len(fake.updatedIDs) != 1 || fake.updatedIDs[0] != "rec-1" {
		t.Fatalf("unexpected updated ids: %+v", fake.updatedIDs)
	}
}

func TestReconcileDomainSetDeleteWhenPrune(t *testing.T) {
	fake := &fakeDomainProvider{
		existing: []domainRecord{
			{ID: "rec-1", Type: "CNAME", Name: "c.anmhela.com", Content: "cname.vercel-dns.com", TTL: 300},
		},
	}
	withFakeProvider(t, fake)

	err := reconcileDomainSet("cloud-console", []domainZoneSpec{
		{
			Name:     "anmhela.com",
			Provider: "cloudflare",
			Records: []domainRecordSpec{
				{Type: "CNAME", Name: "c", DesiredState: "absent"},
			},
		},
	}, false, true)
	if err != nil {
		t.Fatalf("reconcileDomainSet error: %v", err)
	}
	if fake.deleteCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", fake.deleteCalls)
	}
	if len(fake.deletedIDs) != 1 || fake.deletedIDs[0] != "rec-1" {
		t.Fatalf("unexpected deleted ids: %+v", fake.deletedIDs)
	}
}
