package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/anmho/agent-control-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RunStore manages persistence for AgentRun records.
type RunStore struct {
	db *pgxpool.Pool
}

func NewRunStore(db *pgxpool.Pool) *RunStore {
	return &RunStore{db: db}
}

const createTable = `
CREATE TABLE IF NOT EXISTS agent_runs (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  prompt       TEXT        NOT NULL,
  bootstrap_context JSONB,
  runtime_config JSONB     NOT NULL DEFAULT '{"provider":"claude","sandbox_mode":"workspace-write","approval_policy":"never"}'::jsonb,
  status       TEXT        NOT NULL DEFAULT 'PENDING',
  job_name     TEXT,
  output       TEXT,
  started_at   TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

const createEventsTable = `
CREATE TABLE IF NOT EXISTS agent_run_events (
  id            UUID        PRIMARY KEY,
  sequence      BIGSERIAL   NOT NULL,
  run_id        UUID        NOT NULL REFERENCES agent_runs(id) ON DELETE CASCADE,
  delivery_mode TEXT        NOT NULL,
  message       TEXT        NOT NULL,
  reason        TEXT,
  sender        TEXT,
  metadata      JSONB,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

const addOutputColumn = `
ALTER TABLE agent_runs ADD COLUMN IF NOT EXISTS output TEXT`

const addRuntimeConfigColumn = `
ALTER TABLE agent_runs ADD COLUMN IF NOT EXISTS runtime_config JSONB NOT NULL DEFAULT '{"provider":"claude","sandbox_mode":"workspace-write","approval_policy":"never"}'::jsonb`

const addBootstrapContextColumn = `
ALTER TABLE agent_runs ADD COLUMN IF NOT EXISTS bootstrap_context JSONB`

func (s *RunStore) Migrate(ctx context.Context) error {
	if _, err := s.db.Exec(ctx, createTable); err != nil {
		return err
	}
	if _, err := s.db.Exec(ctx, createEventsTable); err != nil {
		return err
	}
	if _, err := s.db.Exec(ctx, addOutputColumn); err != nil {
		return err
	}
	if _, err := s.db.Exec(ctx, addRuntimeConfigColumn); err != nil {
		return err
	}
	_, err := s.db.Exec(ctx, addBootstrapContextColumn)
	return err
}

type storedInputContext struct {
	Resources []domain.Resource `json:"resources,omitempty"`
}

func (s *RunStore) Create(ctx context.Context, message string, runtime domain.RuntimeConfig, resources []domain.Resource) (*domain.AgentRun, error) {
	runtime = runtime.Normalized()
	runtimeJSON, err := json.Marshal(runtime)
	if err != nil {
		return nil, err
	}
	contextJSON, err := json.Marshal(storedInputContext{
		Resources: cloneResources(resources),
	})
	if err != nil {
		return nil, err
	}

	run := &domain.AgentRun{
		ID:        uuid.New(),
		Message:   message,
		Runtime:   runtime,
		Resources: cloneResources(resources),
		Status:    domain.StatusPending,
		CreatedAt: time.Now().UTC(),
	}
	_, err = s.db.Exec(ctx,
		`INSERT INTO agent_runs (id, prompt, bootstrap_context, runtime_config, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		run.ID, run.Message, contextJSON, runtimeJSON, run.Status, run.CreatedAt,
	)
	return run, err
}

func (s *RunStore) Get(ctx context.Context, id uuid.UUID) (*domain.AgentRun, error) {
	row := s.db.QueryRow(ctx,
		`SELECT id, prompt, bootstrap_context, runtime_config, status, job_name, output, started_at, completed_at, created_at
		 FROM agent_runs WHERE id = $1`, id)
	return scanRun(row)
}

func (s *RunStore) List(ctx context.Context, limit int) ([]*domain.AgentRun, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, prompt, bootstrap_context, runtime_config, status, job_name, output, started_at, completed_at, created_at
		 FROM agent_runs ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*domain.AgentRun
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (s *RunStore) ListActive(ctx context.Context, limit int) ([]*domain.AgentRun, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, prompt, bootstrap_context, runtime_config, status, job_name, output, started_at, completed_at, created_at
		 FROM agent_runs
		 WHERE status IN ('PENDING', 'RUNNING')
		 ORDER BY created_at DESC
		 LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*domain.AgentRun
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (s *RunStore) AppendEvent(ctx context.Context, event *domain.AgentRunEvent) (*domain.AgentRunEvent, error) {
	if event == nil {
		return nil, errors.New("event is required")
	}
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	event.DeliveryMode = event.DeliveryMode.Normalized()
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRow(ctx,
		`INSERT INTO agent_run_events (id, run_id, delivery_mode, message, reason, sender, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING sequence`,
		event.ID, event.RunID, event.DeliveryMode, event.Message, event.Reason, nullIfEmpty(event.Sender), metadataJSON, event.CreatedAt,
	)
	if err := row.Scan(&event.Sequence); err != nil {
		return nil, err
	}
	return cloneEvent(event), nil
}

func (s *RunStore) ListEvents(ctx context.Context, runID uuid.UUID, afterSequence int64, limit int) ([]*domain.AgentRunEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, run_id, sequence, delivery_mode, message, reason, sender, metadata, created_at
		 FROM agent_run_events
		 WHERE run_id = $1 AND sequence > $2
		 ORDER BY sequence ASC
		 LIMIT $3`,
		runID, afterSequence, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.AgentRunEvent
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *RunStore) SetOutput(ctx context.Context, id uuid.UUID, output string, status domain.RunStatus) error {
	t := time.Now().UTC()
	_, err := s.db.Exec(ctx,
		`UPDATE agent_runs SET output = $1, status = $2, completed_at = $3 WHERE id = $4`,
		output, status, t, id)
	return err
}

func (s *RunStore) AppendOutput(ctx context.Context, id uuid.UUID, chunk string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE agent_runs
		 SET output = COALESCE(output, '') || $1,
		     started_at = COALESCE(started_at, NOW()),
		     status = 'RUNNING'
		 WHERE id = $2`,
		chunk, id)
	return err
}

func (s *RunStore) SetJobName(ctx context.Context, id uuid.UUID, jobName string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE agent_runs SET job_name = $1, status = 'RUNNING', started_at = NOW() WHERE id = $2`,
		jobName, id)
	return err
}

func (s *RunStore) SetStatus(ctx context.Context, id uuid.UUID, status domain.RunStatus) error {
	var completedAt *time.Time
	if status == domain.StatusSucceeded || status == domain.StatusFailed || status == domain.StatusCanceled {
		t := time.Now().UTC()
		completedAt = &t
	}
	if status == domain.StatusRunning {
		_, err := s.db.Exec(ctx,
			`UPDATE agent_runs
			 SET status = $1,
			     started_at = COALESCE(started_at, NOW()),
			     completed_at = NULL
			 WHERE id = $2`,
			status, id)
		return err
	}
	_, err := s.db.Exec(ctx,
		`UPDATE agent_runs SET status = $1, completed_at = $2 WHERE id = $3`,
		status, completedAt, id)
	return err
}

type runScanner interface {
	Scan(dest ...any) error
}

func scanRun(scanner runScanner) (*domain.AgentRun, error) {
	var run domain.AgentRun
	var contextJSON, runtimeJSON []byte
	var jobName, output *string
	if err := scanner.Scan(&run.ID, &run.Message, &contextJSON, &runtimeJSON, &run.Status, &jobName, &output,
		&run.StartedAt, &run.CompletedAt, &run.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(runtimeJSON) > 0 {
		if err := json.Unmarshal(runtimeJSON, &run.Runtime); err != nil {
			return nil, err
		}
	}
	if len(contextJSON) > 0 && string(contextJSON) != "null" {
		var inputContext storedInputContext
		if err := json.Unmarshal(contextJSON, &inputContext); err != nil {
			return nil, err
		}
		run.Resources = cloneResources(inputContext.Resources)
	}
	run.Runtime = run.Runtime.Normalized()
	if jobName != nil {
		run.JobName = *jobName
	}
	if output != nil {
		run.Output = *output
	}
	return &run, nil
}

func scanEvent(scanner runScanner) (*domain.AgentRunEvent, error) {
	var event domain.AgentRunEvent
	var reason, sender *string
	var metadataJSON []byte
	if err := scanner.Scan(&event.ID, &event.RunID, &event.Sequence, &event.DeliveryMode, &event.Message, &reason, &sender, &metadataJSON, &event.CreatedAt); err != nil {
		return nil, err
	}
	event.DeliveryMode = event.DeliveryMode.Normalized()
	if reason != nil {
		event.Reason = *reason
	}
	if sender != nil {
		event.Sender = *sender
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, err
		}
	}
	return &event, nil
}

func cloneEvent(in *domain.AgentRunEvent) *domain.AgentRunEvent {
	if in == nil {
		return nil
	}
	out := *in
	if in.Metadata != nil {
		out.Metadata = make(map[string]string, len(in.Metadata))
		for k, v := range in.Metadata {
			out.Metadata[k] = v
		}
	}
	return &out
}

func nullIfEmpty(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func cloneResources(in []domain.Resource) []domain.Resource {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.Resource, len(in))
	copy(out, in)
	return out
}
