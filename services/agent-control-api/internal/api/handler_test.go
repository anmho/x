package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/anmho/agent-control-api/internal/api"
	"github.com/anmho/agent-control-api/internal/domain"
	agentcontrolv1 "github.com/anmho/agent-control-api/internal/rpc/gen/agentcontrol/v1"
	agentcontrolv1connect "github.com/anmho/agent-control-api/internal/rpc/gen/agentcontrol/v1/agentcontrolv1connect"
	"github.com/anmho/agent-control-api/internal/runner"
	"github.com/anmho/agent-control-api/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type fakeLocalRunner struct {
	run func(context.Context, uuid.UUID, runner.ExecutionSpec) (string, error)
}

func (f fakeLocalRunner) Run(ctx context.Context, id uuid.UUID, spec runner.ExecutionSpec) (string, error) {
	return f.run(ctx, id, spec)
}

type testEnv struct {
	client agentcontrolv1connect.AgentControlServiceClient
	store  *store.RunStore
}

func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := "postgresql://postgres:postgres@127.0.0.1:54322/agent_control_test?sslmode=disable"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("no test database: %v", err)
	}
	if err := db.Ping(ctx); err != nil {
		t.Skipf("test database unreachable: %v", err)
	}
	t.Cleanup(db.Close)
	return db
}

func newTestEnv(t *testing.T, localRunner fakeLocalRunner) *testEnv {
	t.Helper()
	db := testDB(t)
	s := store.NewRunStore(db)
	if err := s.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := db.Exec(context.Background(), "TRUNCATE agent_run_events, agent_runs"); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	bus := runner.NewBus()
	h := api.NewHandler(zap.NewNop(), s, nil, localRunner, bus)
	server := httptest.NewServer(h.Routes())
	t.Cleanup(server.Close)

	client := agentcontrolv1connect.NewAgentControlServiceClient(http.DefaultClient, server.URL)
	return &testEnv{client: client, store: s}
}

func TestHealthz(t *testing.T) {
	bus := runner.NewBus()
	h := api.NewHandler(zap.NewNop(), nil, nil, fakeLocalRunner{
		run: func(context.Context, uuid.UUID, runner.ExecutionSpec) (string, error) { return "", nil },
	}, bus)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestCreateRunMissingPrompt(t *testing.T) {
	env := newTestEnv(t, fakeLocalRunner{
		run: func(context.Context, uuid.UUID, runner.ExecutionSpec) (string, error) { return "", nil },
	})
	_, err := env.client.CreateRun(context.Background(), connect.NewRequest(&agentcontrolv1.CreateRunRequest{}))
	if err == nil {
		t.Fatal("expected invalid argument error")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("want invalid argument, got %v", connect.CodeOf(err))
	}
}

func TestCreateGetAndListRuns(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	resourceTitle := "Linear issue"
	resourceMime := "text/markdown"
	resourceText := "# Plan\n- land bootstrap envelope\n- keep ConnectRPC"
	env := newTestEnv(t, fakeLocalRunner{
		run: func(ctx context.Context, _ uuid.UUID, _ runner.ExecutionSpec) (string, error) {
			select {
			case started <- struct{}{}:
			default:
			}
			select {
			case <-release:
				return "done", nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		},
	})
	t.Cleanup(func() { close(release) })

	createResp, err := env.client.CreateRun(context.Background(), connect.NewRequest(&agentcontrolv1.CreateRunRequest{
		Message: "unit test prompt",
		Resources: []*agentcontrolv1.Resource{
			{
				Uri:      "https://linear.app/anmho/issue/ANM-194",
				Title:    &resourceTitle,
				MimeType: &resourceMime,
				Text:     &resourceText,
			},
		},
	}))
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	run := createResp.Msg.GetRun()
	if run.GetId() == "" {
		t.Fatal("expected run id")
	}
	if run.GetMessage() != "unit test prompt" {
		t.Fatalf("want message round-trip, got %q", run.GetMessage())
	}
	if run.GetStatus() != agentcontrolv1.RunStatus_RUN_STATUS_RUNNING {
		t.Fatalf("want running, got %v", run.GetStatus())
	}

	<-started

	getResp, err := env.client.GetRun(context.Background(), connect.NewRequest(&agentcontrolv1.GetRunRequest{Id: run.GetId()}))
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if getResp.Msg.GetRun().GetId() != run.GetId() {
		t.Fatalf("id mismatch: %s != %s", getResp.Msg.GetRun().GetId(), run.GetId())
	}
	if got := getResp.Msg.GetRun(); got.GetMessage() != "unit test prompt" {
		t.Fatalf("unexpected run message: %+v", got)
	} else {
		if len(got.GetResources()) != 1 {
			t.Fatalf("message resources mismatch: %+v", got)
		}
		if got.GetResources()[0].GetText() != resourceText || got.GetResources()[0].GetMimeType() != resourceMime {
			t.Fatalf("resource mismatch: %+v", got.GetResources()[0])
		}
	}

	listResp, err := env.client.ListRuns(context.Background(), connect.NewRequest(&agentcontrolv1.ListRunsRequest{PageSize: 10}))
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if len(listResp.Msg.GetRuns()) == 0 {
		t.Fatal("expected at least one run")
	}
	if len(listResp.Msg.GetRuns()[0].GetResources()) != 1 {
		t.Fatal("expected resources on list runs")
	}
}

func TestPushRunEventInterruptAndReplan(t *testing.T) {
	firstStarted := make(chan struct{}, 1)
	secondPrompt := make(chan string, 1)
	var calls int

	env := newTestEnv(t, fakeLocalRunner{
		run: func(ctx context.Context, _ uuid.UUID, spec runner.ExecutionSpec) (string, error) {
			calls++
			if calls == 1 {
				firstStarted <- struct{}{}
				<-ctx.Done()
				return "interrupted", ctx.Err()
			}
			secondPrompt <- spec.Prompt
			return "replanned output", nil
		},
	})

	createResp, err := env.client.CreateRun(context.Background(), connect.NewRequest(&agentcontrolv1.CreateRunRequest{
		Message: "investigate agent infra",
	}))
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	runID := createResp.Msg.GetRun().GetId()
	<-firstStarted

	pushResp, err := env.client.PushRunEvent(context.Background(), connect.NewRequest(&agentcontrolv1.PushRunEventRequest{
		TargetRunIds: []string{runID},
		Message:      "relay this to the MCP workstream",
		Reason:       "new related task",
		Sender:       "control-plane",
		DeliveryMode: agentcontrolv1.PushDeliveryMode_PUSH_DELIVERY_MODE_INTERRUPT_AND_REPLAN,
	}))
	if err != nil {
		t.Fatalf("push run event: %v", err)
	}
	if len(pushResp.Msg.GetDeliveredRunIds()) != 1 || pushResp.Msg.GetDeliveredRunIds()[0] != runID {
		t.Fatalf("unexpected delivered ids: %v", pushResp.Msg.GetDeliveredRunIds())
	}

	select {
	case prompt := <-secondPrompt:
		if !strings.Contains(prompt, "investigate agent infra") {
			t.Fatalf("replanned prompt missing original prompt: %q", prompt)
		}
		if !strings.Contains(prompt, "relay this to the MCP workstream") {
			t.Fatalf("replanned prompt missing routed message: %q", prompt)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for replan")
	}

	runUUID, err := uuid.Parse(runID)
	if err != nil {
		t.Fatalf("parse run id: %v", err)
	}
	waitForRunStatus(t, env.store, runUUID, domain.StatusSucceeded)

	events, err := env.store.ListEvents(context.Background(), runUUID, 0, 10)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	if events[0].DeliveryMode != domain.PushDeliveryInterruptReplan {
		t.Fatalf("want interrupt/replan event, got %s", events[0].DeliveryMode)
	}
}

func TestWatchRunReplaysOutputAndControlEvents(t *testing.T) {
	env := newTestEnv(t, fakeLocalRunner{
		run: func(context.Context, uuid.UUID, runner.ExecutionSpec) (string, error) {
			return "final output", nil
		},
	})

	createResp, err := env.client.CreateRun(context.Background(), connect.NewRequest(&agentcontrolv1.CreateRunRequest{
		Message: "summarize the routing design",
	}))
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	runID := createResp.Msg.GetRun().GetId()
	runUUID, err := uuid.Parse(runID)
	if err != nil {
		t.Fatalf("parse run id: %v", err)
	}
	waitForRunStatus(t, env.store, runUUID, domain.StatusSucceeded)

	if _, err := env.client.PushRunEvent(context.Background(), connect.NewRequest(&agentcontrolv1.PushRunEventRequest{
		TargetRunIds: []string{runID},
		Message:      "follow-up note",
		Sender:       "control-plane",
		DeliveryMode: agentcontrolv1.PushDeliveryMode_PUSH_DELIVERY_MODE_APPEND_CONTEXT,
	})); err != nil {
		t.Fatalf("push event: %v", err)
	}

	stream, err := env.client.WatchRun(context.Background(), connect.NewRequest(&agentcontrolv1.WatchRunRequest{
		Id:           runID,
		ReplayOutput: true,
		ReplayEvents: true,
	}))
	if err != nil {
		t.Fatalf("watch run: %v", err)
	}

	var sawSnapshot, sawOutput, sawControl, sawCompleted bool
	for stream.Receive() {
		msg := stream.Msg()
		switch event := msg.Event.(type) {
		case *agentcontrolv1.WatchRunResponse_Snapshot:
			sawSnapshot = event.Snapshot.GetRun().GetId() == runID
		case *agentcontrolv1.WatchRunResponse_Output:
			sawOutput = strings.Contains(event.Output.GetChunk(), "final output")
		case *agentcontrolv1.WatchRunResponse_Control:
			sawControl = strings.Contains(event.Control.GetMessage(), "follow-up note")
		case *agentcontrolv1.WatchRunResponse_Completed:
			sawCompleted = event.Completed.GetStatus() == agentcontrolv1.RunStatus_RUN_STATUS_SUCCEEDED
		}
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream err: %v", err)
	}
	if !sawSnapshot || !sawOutput || !sawControl || !sawCompleted {
		t.Fatalf("missing replay event snapshot=%v output=%v control=%v completed=%v", sawSnapshot, sawOutput, sawControl, sawCompleted)
	}
}

func TestCancelRunMarksRunCanceledAndCompletesWatch(t *testing.T) {
	started := make(chan struct{}, 1)
	env := newTestEnv(t, fakeLocalRunner{
		run: func(ctx context.Context, _ uuid.UUID, _ runner.ExecutionSpec) (string, error) {
			select {
			case started <- struct{}{}:
			default:
			}
			<-ctx.Done()
			return "", ctx.Err()
		},
	})

	createResp, err := env.client.CreateRun(context.Background(), connect.NewRequest(&agentcontrolv1.CreateRunRequest{
		Message: "keep working on the request until canceled",
	}))
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	runID := createResp.Msg.GetRun().GetId()
	<-started

	cancelResp, err := env.client.CancelRun(context.Background(), connect.NewRequest(&agentcontrolv1.CancelRunRequest{
		Id: runID,
	}))
	if err != nil {
		t.Fatalf("cancel run: %v", err)
	}
	if cancelResp.Msg.GetRun().GetStatus() != agentcontrolv1.RunStatus_RUN_STATUS_CANCELED {
		t.Fatalf("want canceled response, got %v", cancelResp.Msg.GetRun().GetStatus())
	}

	runUUID, err := uuid.Parse(runID)
	if err != nil {
		t.Fatalf("parse run id: %v", err)
	}
	waitForRunStatus(t, env.store, runUUID, domain.StatusCanceled)

	getResp, err := env.client.GetRun(context.Background(), connect.NewRequest(&agentcontrolv1.GetRunRequest{Id: runID}))
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if getResp.Msg.GetRun().GetStatus() != agentcontrolv1.RunStatus_RUN_STATUS_CANCELED {
		t.Fatalf("want canceled get status, got %v", getResp.Msg.GetRun().GetStatus())
	}

	stream, err := env.client.WatchRun(context.Background(), connect.NewRequest(&agentcontrolv1.WatchRunRequest{
		Id:           runID,
		ReplayEvents: true,
	}))
	if err != nil {
		t.Fatalf("watch run: %v", err)
	}

	var sawControl, sawCompleted bool
	for stream.Receive() {
		msg := stream.Msg()
		switch event := msg.Event.(type) {
		case *agentcontrolv1.WatchRunResponse_Control:
			sawControl = strings.Contains(event.Control.GetMessage(), "run canceled")
		case *agentcontrolv1.WatchRunResponse_Completed:
			sawCompleted = event.Completed.GetStatus() == agentcontrolv1.RunStatus_RUN_STATUS_CANCELED
		}
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream err: %v", err)
	}
	if !sawControl || !sawCompleted {
		t.Fatalf("missing cancel replay control=%v completed=%v", sawControl, sawCompleted)
	}
}

func waitForRunStatus(t *testing.T, s *store.RunStore, id uuid.UUID, want domain.RunStatus) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		run, err := s.Get(context.Background(), id)
		if err == nil && run.Status == want {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	run, err := s.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("final get: %v", err)
	}
	t.Fatalf("timed out waiting for status %s, got %s", want, run.Status)
}

func TestPushRunEventInvalidID(t *testing.T) {
	env := newTestEnv(t, fakeLocalRunner{
		run: func(context.Context, uuid.UUID, runner.ExecutionSpec) (string, error) {
			return "", nil
		},
	})

	_, err := env.client.PushRunEvent(context.Background(), connect.NewRequest(&agentcontrolv1.PushRunEventRequest{
		TargetRunIds: []string{"not-a-uuid"},
		Message:      "hello",
	}))
	if err == nil {
		t.Fatal("expected invalid argument error")
	}
	if !errors.Is(connect.NewError(connect.CodeInvalidArgument, err), err) && connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("want invalid argument, got %v", connect.CodeOf(err))
	}
}
