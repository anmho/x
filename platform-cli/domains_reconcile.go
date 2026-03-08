package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var resolveDomainProvider = providerForDomain

func providerForDomain(provider string) (domainProvider, error) {
	switch normalizeDomainProvider(provider) {
	case "cloudflare":
		return newCloudflareDNSProviderFromEnv()
	case "vercel":
		return newVercelDNSProviderFromEnv()
	default:
		return nil, fmt.Errorf("unsupported domain provider %q", provider)
	}
}

func reconcileDomainSet(projectName string, zones []domainZoneSpec, dryRun bool, prune bool) error {
	for _, zone := range zones {
		zoneName := strings.TrimSpace(zone.Name)
		if zoneName == "" {
			return errors.New("domain zone name is required")
		}
		providerName := normalizeDomainProvider(zone.Provider)
		if providerName == "" {
			return fmt.Errorf("domain zone %q has empty provider", zoneName)
		}

		state := strings.ToLower(strings.TrimSpace(zone.DesiredState))
		if state == "" {
			state = "present"
		}

		provider, err := resolveDomainProvider(providerName)
		if err != nil {
			return err
		}

		fmt.Printf("control-plane: domains project=%s zone=%s provider=%s desired=%s\n", projectName, zoneName, providerName, state)

		existing, err := provider.ListRecords(context.Background(), zone)
		if err != nil {
			return fmt.Errorf("domains list records failed for zone=%s provider=%s: %w", zoneName, providerName, err)
		}

		if state == "absent" {
			if !prune {
				fmt.Printf("control-plane: domains skip zone %s deletion (use --prune or destroy)\n", zoneName)
				continue
			}
			for _, record := range existing {
				if dryRun {
					fmt.Printf("control-plane: domains would delete record %s %s -> %s (%s)\n", record.Type, record.Name, record.Content, zoneName)
					continue
				}
				if err := provider.DeleteRecord(context.Background(), zone, record.ID); err != nil {
					return fmt.Errorf("domains delete record %s failed: %w", record.ID, err)
				}
				fmt.Printf("control-plane: domains deleted record %s %s -> %s (%s)\n", record.Type, record.Name, record.Content, zoneName)
			}
			continue
		}

		for _, desired := range zone.Records {
			recordState := strings.ToLower(strings.TrimSpace(desired.DesiredState))
			if recordState == "" {
				recordState = "present"
			}

			match, found := findMatchingRecord(zone, desired, existing)

			if recordState == "absent" {
				if !prune {
					fmt.Printf("control-plane: domains skip record deletion %s %s (use --prune or destroy)\n", desired.Type, desired.Name)
					continue
				}
				if !found {
					fmt.Printf("control-plane: domains record already absent %s %s (%s)\n", desired.Type, desired.Name, zoneName)
					continue
				}
				if dryRun {
					fmt.Printf("control-plane: domains would delete record %s %s -> %s (%s)\n", match.Type, match.Name, match.Content, zoneName)
					continue
				}
				if err := provider.DeleteRecord(context.Background(), zone, match.ID); err != nil {
					return fmt.Errorf("domains delete record %s failed: %w", match.ID, err)
				}
				fmt.Printf("control-plane: domains deleted record %s %s -> %s (%s)\n", match.Type, match.Name, match.Content, zoneName)
				continue
			}

			input := domainRecordInput{
				Type:    strings.ToUpper(strings.TrimSpace(desired.Type)),
				Name:    strings.TrimSpace(desired.Name),
				Content: strings.TrimSpace(desired.Content),
				TTL:     desired.TTL,
				Proxied: desired.Proxied,
			}

			if !found {
				if dryRun {
					fmt.Printf("control-plane: domains would create record %s %s -> %s (%s)\n", input.Type, input.Name, input.Content, zoneName)
					continue
				}
				created, err := provider.CreateRecord(context.Background(), zone, input)
				if err != nil {
					return fmt.Errorf("domains create record %s %s failed: %w", input.Type, input.Name, err)
				}
				fmt.Printf("control-plane: domains created record %s %s -> %s (%s)\n", created.Type, created.Name, created.Content, zoneName)
				continue
			}

			if !recordNeedsUpdate(match, input) {
				fmt.Printf("control-plane: domains no-op record %s %s -> %s (%s)\n", match.Type, match.Name, match.Content, zoneName)
				continue
			}

			if dryRun {
				fmt.Printf("control-plane: domains would update record %s %s -> %s (%s)\n", input.Type, input.Name, input.Content, zoneName)
				continue
			}
			updated, err := provider.UpdateRecord(context.Background(), zone, match.ID, input)
			if err != nil {
				return fmt.Errorf("domains update record %s failed: %w", match.ID, err)
			}
			fmt.Printf("control-plane: domains updated record %s %s -> %s (%s)\n", updated.Type, updated.Name, updated.Content, zoneName)
		}
	}
	return nil
}

func findMatchingRecord(zone domainZoneSpec, desired domainRecordSpec, existing []domainRecord) (domainRecord, bool) {
	if desiredID := strings.TrimSpace(desired.ID); desiredID != "" {
		for _, candidate := range existing {
			if strings.TrimSpace(candidate.ID) == desiredID {
				return candidate, true
			}
		}
	}

	desiredType := strings.ToUpper(strings.TrimSpace(desired.Type))
	desiredName := strings.TrimSpace(desired.Name)
	zoneName := strings.TrimSpace(zone.Name)
	for _, candidate := range existing {
		if strings.ToUpper(strings.TrimSpace(candidate.Type)) != desiredType {
			continue
		}
		if !recordNameMatches(candidate.Name, desiredName, zoneName) {
			continue
		}
		return candidate, true
	}
	return domainRecord{}, false
}

func recordNeedsUpdate(existing domainRecord, desired domainRecordInput) bool {
	if strings.ToUpper(strings.TrimSpace(existing.Type)) != strings.ToUpper(strings.TrimSpace(desired.Type)) {
		return true
	}
	if strings.TrimSpace(existing.Content) != strings.TrimSpace(desired.Content) {
		return true
	}
	if desired.TTL > 0 && existing.TTL != desired.TTL {
		return true
	}
	if desired.Proxied != nil {
		if existing.Proxied == nil {
			return true
		}
		return *existing.Proxied != *desired.Proxied
	}
	return false
}

func recordNameMatches(existing string, desired string, zoneName string) bool {
	existing = strings.TrimSuffix(strings.TrimSpace(existing), ".")
	desired = strings.TrimSuffix(strings.TrimSpace(desired), ".")
	zoneName = strings.TrimSuffix(strings.TrimSpace(zoneName), ".")
	if existing == desired {
		return true
	}
	if zoneName == "" {
		return false
	}
	if existing == desired+"."+zoneName {
		return true
	}
	if desired == "@" && existing == zoneName {
		return true
	}
	return false
}
