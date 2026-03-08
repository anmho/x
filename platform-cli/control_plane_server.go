package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultControlPlaneAPIAddr = ":8091"

var runControlPlaneReconcile = reconcileControlPlane

type domainReconcileRequest struct {
	Project string `json:"project,omitempty"`
	DryRun  *bool  `json:"dry_run,omitempty"`
	Prune   bool   `json:"prune,omitempty"`
}

type domainRecordRequest struct {
	Provider string `json:"provider,omitempty"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl,omitempty"`
	Proxied  *bool  `json:"proxied,omitempty"`
}

type controlPlaneHTTPServer struct {
	cfg *controlPlaneConfig
}

func runControlPlaneServe(args []string) error {
	fs := flag.NewFlagSet("control-plane serve", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	addr := fs.String("addr", getEnv("PLATFORM_CONTROL_PLANE_ADDR", defaultControlPlaneAPIAddr), "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, _, err := loadControlPlaneConfig(true)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              strings.TrimSpace(*addr),
		Handler:           newControlPlaneHTTPServer(cfg),
		ReadHeaderTimeout: 10 * time.Second,
	}
	fmt.Printf("control-plane: api listening on %s\n", srv.Addr)
	return srv.ListenAndServe()
}

func newControlPlaneHTTPServer(cfg *controlPlaneConfig) http.Handler {
	s := &controlPlaneHTTPServer{cfg: cfg}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/v1/domains", s.handleDomains)
	mux.HandleFunc("/v1/domains/", s.handleDomainSubroutes)
	return withControlPlaneCORS(mux)
}

func withControlPlaneCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *controlPlaneHTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeControlPlaneJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeControlPlaneJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *controlPlaneHTTPServer) handleDomains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeControlPlaneJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	projectFilter := strings.TrimSpace(r.URL.Query().Get("project"))
	zones := []domainZone{}
	for _, project := range s.cfg.Projects {
		if projectFilter != "" && project.Name != projectFilter {
			continue
		}
		for _, zone := range project.Domains {
			zones = append(zones, domainZone{
				Name:     strings.TrimSpace(zone.Name),
				Provider: normalizeDomainProvider(zone.Provider),
				ZoneID:   strings.TrimSpace(zone.ZoneID),
				Project:  project.Name,
			})
		}
	}
	writeControlPlaneJSON(w, http.StatusOK, map[string]any{"domains": zones})
}

func (s *controlPlaneHTTPServer) handleDomainSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/domains/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeControlPlaneJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	if path == "reconcile" {
		if r.Method != http.MethodPost {
			writeControlPlaneJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		s.handleDomainReconcile(w, r)
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "records" {
		writeControlPlaneJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	zoneName := strings.TrimSpace(parts[0])
	providerName := normalizeDomainProvider(r.URL.Query().Get("provider"))
	if providerName == "" {
		providerName = normalizeDomainProvider(r.Header.Get("X-Domain-Provider"))
	}
	if providerName == "" {
		writeControlPlaneJSON(w, http.StatusBadRequest, map[string]string{"error": "provider is required (query param: provider)"})
		return
	}

	zone, err := s.findZone(zoneName, providerName)
	if err != nil {
		writeControlPlaneJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	provider, err := resolveDomainProvider(providerName)
	if err != nil {
		writeControlPlaneJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	switch {
	case len(parts) == 2 && r.Method == http.MethodGet:
		records, err := provider.ListRecords(r.Context(), zone)
		if err != nil {
			writeControlPlaneJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeControlPlaneJSON(w, http.StatusOK, map[string]any{"records": records})
	case len(parts) == 2 && r.Method == http.MethodPost:
		var req domainRecordRequest
		if err := decodeControlPlaneJSON(r, &req); err != nil {
			writeControlPlaneJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		created, err := provider.CreateRecord(r.Context(), zone, domainRecordInput{
			Type:    req.Type,
			Name:    req.Name,
			Content: req.Content,
			TTL:     req.TTL,
			Proxied: req.Proxied,
		})
		if err != nil {
			writeControlPlaneJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeControlPlaneJSON(w, http.StatusCreated, map[string]any{"record": created})
	case len(parts) == 3 && r.Method == http.MethodPatch:
		recordID := strings.TrimSpace(parts[2])
		var req domainRecordRequest
		if err := decodeControlPlaneJSON(r, &req); err != nil {
			writeControlPlaneJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		updated, err := provider.UpdateRecord(r.Context(), zone, recordID, domainRecordInput{
			Type:    req.Type,
			Name:    req.Name,
			Content: req.Content,
			TTL:     req.TTL,
			Proxied: req.Proxied,
		})
		if err != nil {
			writeControlPlaneJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeControlPlaneJSON(w, http.StatusOK, map[string]any{"record": updated})
	case len(parts) == 3 && r.Method == http.MethodDelete:
		recordID := strings.TrimSpace(parts[2])
		if recordID == "" {
			writeControlPlaneJSON(w, http.StatusBadRequest, map[string]string{"error": "record id is required"})
			return
		}
		if err := provider.DeleteRecord(r.Context(), zone, recordID); err != nil {
			writeControlPlaneJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeControlPlaneJSON(w, http.StatusOK, map[string]any{"deleted": true})
	default:
		writeControlPlaneJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (s *controlPlaneHTTPServer) handleDomainReconcile(w http.ResponseWriter, r *http.Request) {
	var req domainReconcileRequest
	if err := decodeControlPlaneJSON(r, &req); err != nil {
		writeControlPlaneJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	dryRun := false
	if req.DryRun != nil {
		dryRun = *req.DryRun
	}
	if err := runControlPlaneReconcile(s.cfg, strings.TrimSpace(req.Project), dryRun, req.Prune); err != nil {
		writeControlPlaneJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeControlPlaneJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"dry_run": dryRun,
		"prune":   req.Prune,
		"project": strings.TrimSpace(req.Project),
	})
}

func (s *controlPlaneHTTPServer) findZone(zoneName string, providerName string) (domainZoneSpec, error) {
	zoneName = strings.TrimSuffix(strings.TrimSpace(zoneName), ".")
	for _, project := range s.cfg.Projects {
		for _, zone := range project.Domains {
			if normalizeDomainProvider(zone.Provider) != providerName {
				continue
			}
			if strings.TrimSuffix(strings.TrimSpace(zone.Name), ".") == zoneName {
				return zone, nil
			}
		}
	}
	return domainZoneSpec{}, fmt.Errorf("domain zone %q (%s) not found", zoneName, providerName)
}

func decodeControlPlaneJSON(r *http.Request, out any) error {
	if out == nil {
		return errors.New("output is nil")
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	return nil
}

func writeControlPlaneJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
