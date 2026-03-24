package api

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/anmho/agent-control-api/internal/dispatch"
	"github.com/anmho/agent-control-api/internal/domain"
	agentcontrolv1 "github.com/anmho/agent-control-api/internal/rpc/gen/agentcontrol/v1"
	agentcontrolv1connect "github.com/anmho/agent-control-api/internal/rpc/gen/agentcontrol/v1/agentcontrolv1connect"
	"github.com/anmho/agent-control-api/internal/runner"
	"github.com/anmho/agent-control-api/internal/store"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	log         *zap.Logger
	store       *store.RunStore
	cloudRunner *runner.CloudRunClient
	localRunner interface {
		Run(context.Context, uuid.UUID, runner.ExecutionSpec) (string, error)
	}
	bus       *runner.Bus
	activeMu  sync.Mutex
	activeRun map[uuid.UUID]*activeRunState
	agentcontrolv1connect.UnimplementedAgentControlServiceHandler
}

type activeRunState struct {
	spec        runner.ExecutionSpec
	cancel      context.CancelFunc
	queuedEvent *domain.AgentRunEvent
	canceled    bool
}

type deliveryRequest struct {
	TargetRunIDs []string
	Query        string
	Limit        int
	Message      string
	Reason       string
	Sender       string
	DeliveryMode domain.PushDeliveryMode
	Metadata     map[string]string
}

func NewHandler(log *zap.Logger, s *store.RunStore, cloud *runner.CloudRunClient, local interface {
	Run(context.Context, uuid.UUID, runner.ExecutionSpec) (string, error)
}, bus *runner.Bus) *Handler {
	return &Handler{
		log:         log,
		store:       s,
		cloudRunner: cloud,
		localRunner: local,
		bus:         bus,
		activeRun:   map[uuid.UUID]*activeRunState{},
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	path, serviceHandler := agentcontrolv1connect.NewAgentControlServiceHandler(h)
	mux.Handle(path, serviceHandler)
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	return withCORS(mux)
}

func (h *Handler) CreateRun(
	ctx context.Context,
	req *connect.Request[agentcontrolv1.CreateRunRequest],
) (*connect.Response[agentcontrolv1.CreateRunResponse], error) {
	runtime := runtimeFromProto(req.Msg.Runtime)
	createReq := domain.CreateRunRequest{
		Message:   req.Msg.GetMessage(),
		Runtime:   runtime,
		Resources: resourcesFromProto(req.Msg.GetResources()),
	}
	if err := createReq.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	spec := runner.ExecutionSpec{
		Prompt:  createReq.Message,
		Runtime: createReq.Runtime,
	}

	run, err := h.store.Create(ctx, createReq.Message, createReq.Runtime, createReq.Resources)
	if err != nil {
		h.log.Error("store create run", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if h.cloudRunner != nil {
		jobName, err := h.cloudRunner.Dispatch(ctx, spec)
		if err != nil {
			h.log.Error("dispatch cloud run job", zap.Error(err), zap.String("run_id", run.ID.String()))
			_ = h.store.SetStatus(ctx, run.ID, domain.StatusFailed)
			run.Status = domain.StatusFailed
		} else {
			_ = h.store.SetJobName(ctx, run.ID, jobName)
			run.JobName = jobName
			run.Status = domain.StatusRunning
		}
	} else {
		run.Status = domain.StatusRunning
		_ = h.store.SetStatus(ctx, run.ID, domain.StatusRunning)
		go h.execLocal(run.ID, spec)
	}

	return connect.NewResponse(&agentcontrolv1.CreateRunResponse{
		Run: runToProto(run, false),
	}), nil
}

func (h *Handler) CancelRun(
	ctx context.Context,
	req *connect.Request[agentcontrolv1.CancelRunRequest],
) (*connect.Response[agentcontrolv1.CancelRunResponse], error) {
	runID, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	run, err := h.store.Get(ctx, runID)
	if err != nil {
		return nil, toConnectError(err)
	}
	if run.Status == domain.StatusSucceeded || run.Status == domain.StatusFailed || run.Status == domain.StatusCanceled {
		return connect.NewResponse(&agentcontrolv1.CancelRunResponse{Run: runToProto(run, false)}), nil
	}

	event, err := h.store.AppendEvent(ctx, &domain.AgentRunEvent{
		RunID:        runID,
		DeliveryMode: domain.PushDeliveryAppendContext,
		Message:      "run canceled",
		Reason:       "cancel_requested",
		Sender:       "control-plane",
	})
	if err != nil {
		h.log.Error("append cancel event", zap.Error(err), zap.String("run_id", runID.String()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if h.bus != nil {
		h.bus.Publish(runner.Chunk{RunID: runID, ControlEvent: event})
	}

	h.cancelActiveRun(runID)
	if err := h.store.SetStatus(ctx, runID, domain.StatusCanceled); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	now := time.Now().UTC()
	run.Status = domain.StatusCanceled
	run.CompletedAt = &now
	if h.bus != nil {
		h.bus.Publish(runner.Chunk{RunID: runID, Status: domain.StatusCanceled, Done: true})
	}
	return connect.NewResponse(&agentcontrolv1.CancelRunResponse{Run: runToProto(run, false)}), nil
}

func (h *Handler) GetRun(
	ctx context.Context,
	req *connect.Request[agentcontrolv1.GetRunRequest],
) (*connect.Response[agentcontrolv1.GetRunResponse], error) {
	run, err := h.fetchRun(ctx, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentcontrolv1.GetRunResponse{
		Run: runToProto(run, false),
	}), nil
}

func (h *Handler) ListRuns(
	ctx context.Context,
	req *connect.Request[agentcontrolv1.ListRunsRequest],
) (*connect.Response[agentcontrolv1.ListRunsResponse], error) {
	pageSize := int(req.Msg.GetPageSize())
	if pageSize == 0 {
		pageSize = 20
	}
	runs, err := h.store.List(ctx, pageSize)
	if err != nil {
		h.log.Error("list runs", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	resp := &agentcontrolv1.ListRunsResponse{}
	for _, run := range runs {
		resp.Runs = append(resp.Runs, runToProto(run, false))
	}
	return connect.NewResponse(resp), nil
}

func (h *Handler) WatchRun(
	ctx context.Context,
	req *connect.Request[agentcontrolv1.WatchRunRequest],
	stream *connect.ServerStream[agentcontrolv1.WatchRunResponse],
) error {
	run, err := h.fetchRun(ctx, req.Msg.GetId())
	if err != nil {
		return toConnectError(err)
	}

	if err := stream.Send(&agentcontrolv1.WatchRunResponse{
		Event: &agentcontrolv1.WatchRunResponse_Snapshot{
			Snapshot: &agentcontrolv1.AgentRunSnapshot{Run: runToProto(run, false)},
		},
	}); err != nil {
		return err
	}

	if req.Msg.GetReplayOutput() && run.Output != "" {
		if err := stream.Send(&agentcontrolv1.WatchRunResponse{
			Event: &agentcontrolv1.WatchRunResponse_Output{
				Output: &agentcontrolv1.OutputChunk{Chunk: run.Output},
			},
		}); err != nil {
			return err
		}
	}
	if req.Msg.GetReplayEvents() {
		runID, err := uuid.Parse(req.Msg.GetId())
		if err != nil {
			return connect.NewError(connect.CodeInvalidArgument, err)
		}
		events, err := h.store.ListEvents(ctx, runID, 0, 200)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		for _, event := range events {
			if err := stream.Send(&agentcontrolv1.WatchRunResponse{
				Event: &agentcontrolv1.WatchRunResponse_Control{
					Control: controlEventToProto(event),
				},
			}); err != nil {
				return err
			}
		}
	}

	if isTerminalStatus(run.Status) {
		return stream.Send(&agentcontrolv1.WatchRunResponse{
			Event: &agentcontrolv1.WatchRunResponse_Completed{
				Completed: &agentcontrolv1.RunCompleted{Status: statusToProto(run.Status)},
			},
		})
	}

	if h.bus == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("streaming unavailable"))
	}

	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	ch := h.bus.Subscribe(id)
	for {
		select {
		case <-ctx.Done():
			return nil
		case chunk, ok := <-ch:
			if !ok {
				return nil
			}
			if chunk.Done {
				status := chunk.Status
				if status == "" {
					status = domain.StatusSucceeded
					if chunk.Err != nil {
						status = domain.StatusFailed
					}
				}
				return stream.Send(&agentcontrolv1.WatchRunResponse{
					Event: &agentcontrolv1.WatchRunResponse_Completed{
						Completed: &agentcontrolv1.RunCompleted{Status: statusToProto(status)},
					},
				})
			}
			if chunk.ControlEvent != nil {
				if err := stream.Send(&agentcontrolv1.WatchRunResponse{
					Event: &agentcontrolv1.WatchRunResponse_Control{
						Control: controlEventToProto(chunk.ControlEvent),
					},
				}); err != nil {
					return err
				}
				continue
			}
			if err := stream.Send(&agentcontrolv1.WatchRunResponse{
				Event: &agentcontrolv1.WatchRunResponse_Output{
					Output: &agentcontrolv1.OutputChunk{Chunk: chunk.Output},
				},
			}); err != nil {
				return err
			}
		}
	}
}

func (h *Handler) PushRunEvent(
	ctx context.Context,
	req *connect.Request[agentcontrolv1.PushRunEventRequest],
) (*connect.Response[agentcontrolv1.PushRunEventResponse], error) {
	delivered, err := h.deliverRunEvent(ctx, deliveryRequest{
		TargetRunIDs: req.Msg.GetTargetRunIds(),
		Query:        req.Msg.GetQuery(),
		Limit:        int(req.Msg.GetLimit()),
		Message:      req.Msg.GetMessage(),
		Reason:       req.Msg.GetReason(),
		Sender:       req.Msg.GetSender(),
		DeliveryMode: deliveryModeFromProto(req.Msg.GetDeliveryMode()).Normalized(),
		Metadata:     req.Msg.GetMetadata(),
	})
	if err != nil {
		if connectErr, ok := err.(*connect.Error); ok {
			return nil, connectErr
		}
		if isInvalidUUID(err) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&agentcontrolv1.PushRunEventResponse{
		DeliveredRunIds: delivered,
	}), nil
}

func (h *Handler) DeliverMailboxMessage(ctx context.Context, req dispatch.DeliveryRequest) ([]string, error) {
	return h.deliverRunEvent(ctx, deliveryRequest{
		TargetRunIDs: req.TargetRunIDs,
		Query:        req.Query,
		Limit:        req.Limit,
		Message:      req.Message,
		Reason:       req.Reason,
		Sender:       req.Sender,
		DeliveryMode: req.DeliveryMode,
		Metadata:     req.Metadata,
	})
}

func (h *Handler) execLocal(id uuid.UUID, spec runner.ExecutionSpec) {
	for {
		ctx, cancel := context.WithCancel(context.Background())
		h.setActiveRun(id, spec, cancel)

		output, err := h.localRunner.Run(ctx, id, spec)
		state, nextSpec, restart := h.finishActiveRun(id, ctx.Err())
		if restart {
			spec = nextSpec
			continue
		}

		status := domain.StatusSucceeded
		if err != nil {
			if errors.Is(err, context.Canceled) && state != nil && state.canceled {
				status = domain.StatusCanceled
			} else {
				h.log.Warn("local run failed", zap.String("run_id", id.String()), zap.Error(err))
				status = domain.StatusFailed
			}
		}
		if h.bus != nil {
			h.bus.Publish(runner.Chunk{RunID: id, Status: status, Done: true, Err: err})
		}
		if storeErr := h.store.SetOutput(context.Background(), id, output, status); storeErr != nil {
			h.log.Error("store set output", zap.Error(storeErr))
		}
		return
	}
}

func (h *Handler) fetchRun(ctx context.Context, id string) (*domain.AgentRun, error) {
	runID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	run, err := h.store.Get(ctx, runID)
	if err != nil {
		return nil, err
	}
	if h.cloudRunner != nil && run.Status == domain.StatusRunning && run.JobName != "" {
		status, err := h.cloudRunner.ExecutionStatus(ctx, run.JobName)
		if err != nil {
			h.log.Warn("sync execution status", zap.Error(err))
		} else if status != string(run.Status) {
			_ = h.store.SetStatus(ctx, run.ID, domain.RunStatus(status))
			run.Status = domain.RunStatus(status)
		}
	}
	return run, nil
}

func toConnectError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case isInvalidUUID(err):
		return connect.NewError(connect.CodeInvalidArgument, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Connect-Protocol-Version,Connect-Timeout-Ms,X-User-Agent,Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) setActiveRun(id uuid.UUID, spec runner.ExecutionSpec, cancel context.CancelFunc) {
	h.activeMu.Lock()
	defer h.activeMu.Unlock()
	h.activeRun[id] = &activeRunState{spec: spec, cancel: cancel}
}

func (h *Handler) finishActiveRun(id uuid.UUID, runErr error) (*activeRunState, runner.ExecutionSpec, bool) {
	h.activeMu.Lock()
	defer h.activeMu.Unlock()
	state := h.activeRun[id]
	delete(h.activeRun, id)
	if state == nil || state.queuedEvent == nil {
		return state, runner.ExecutionSpec{}, false
	}
	if !errors.Is(runErr, context.Canceled) {
		return state, runner.ExecutionSpec{}, false
	}
	nextSpec := state.spec
	nextSpec.Prompt = appendRoutedContext(nextSpec.Prompt, state.queuedEvent)
	return state, nextSpec, true
}

func (h *Handler) cancelActiveRun(id uuid.UUID) {
	h.activeMu.Lock()
	state := h.activeRun[id]
	if state != nil {
		state.canceled = true
		cancel := state.cancel
		h.activeMu.Unlock()
		cancel()
		return
	}
	h.activeMu.Unlock()
}

func (h *Handler) queueInterruptReplan(id uuid.UUID, event *domain.AgentRunEvent) {
	h.activeMu.Lock()
	state := h.activeRun[id]
	if state != nil {
		state.queuedEvent = event
		cancel := state.cancel
		h.activeMu.Unlock()
		cancel()
		return
	}
	h.activeMu.Unlock()
}

func isTerminalStatus(status domain.RunStatus) bool {
	return status == domain.StatusSucceeded || status == domain.StatusFailed || status == domain.StatusCanceled
}

func (h *Handler) deliverRunEvent(ctx context.Context, req deliveryRequest) ([]string, error) {
	if strings.TrimSpace(req.Message) == "" {
		return nil, errors.New("message is required")
	}

	targets, err := h.resolveTargetRunIDs(ctx, req.TargetRunIDs, req.Query, req.Limit)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, nil
	}

	mode := req.DeliveryMode.Normalized()
	delivered := make([]string, 0, len(targets))
	for _, runID := range targets {
		event, err := h.store.AppendEvent(ctx, &domain.AgentRunEvent{
			RunID:        runID,
			DeliveryMode: mode,
			Message:      req.Message,
			Reason:       req.Reason,
			Sender:       req.Sender,
			Metadata:     cloneMetadata(req.Metadata),
		})
		if err != nil {
			h.log.Error("append run event", zap.Error(err), zap.String("run_id", runID.String()))
			return nil, err
		}
		if h.bus != nil {
			h.bus.Publish(runner.Chunk{RunID: runID, ControlEvent: event})
		}
		if mode == domain.PushDeliveryInterruptReplan {
			h.queueInterruptReplan(runID, event)
		}
		delivered = append(delivered, runID.String())
	}
	return delivered, nil
}

func (h *Handler) resolveTargetRunIDs(ctx context.Context, explicitIDs []string, query string, limit int) ([]uuid.UUID, error) {
	seen := map[uuid.UUID]struct{}{}
	var ids []uuid.UUID
	for _, raw := range explicitIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return ids, nil
	}
	runs, err := h.store.ListActive(ctx, max(limit, 50))
	if err != nil {
		return nil, err
	}
	type scoredRun struct {
		id    uuid.UUID
		score int
	}
	var scored []scoredRun
	for _, run := range runs {
		score := scoreRunForQuery(query, run)
		if score == 0 {
			continue
		}
		scored = append(scored, scoredRun{id: run.ID, score: score})
	}
	slices.SortFunc(scored, func(a, b scoredRun) int {
		if a.score != b.score {
			return b.score - a.score
		}
		return strings.Compare(a.id.String(), b.id.String())
	})
	for _, candidate := range scored {
		if _, ok := seen[candidate.id]; ok {
			continue
		}
		seen[candidate.id] = struct{}{}
		ids = append(ids, candidate.id)
		if limit > 0 && len(ids) >= limit {
			break
		}
	}
	return ids, nil
}

func scoreRunForQuery(query string, run *domain.AgentRun) int {
	if run == nil {
		return 0
	}
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery == "" {
		return 0
	}
	normalizedMessage := strings.ToLower(run.Message)
	score := 0
	if strings.Contains(normalizedMessage, normalizedQuery) {
		score += 100
	}
	queryTokens := tokenize(normalizedQuery)
	messageTokens := tokenize(normalizedMessage)
	for _, token := range queryTokens {
		if token == "" {
			continue
		}
		if slices.Contains(messageTokens, token) {
			score += 20
		}
	}
	return score
}

func tokenize(raw string) []string {
	return strings.FieldsFunc(strings.ToLower(raw), func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return false
		case r >= '0' && r <= '9':
			return false
		default:
			return true
		}
	})
}

func appendRoutedContext(prompt string, event *domain.AgentRunEvent) string {
	if event == nil {
		return prompt
	}
	var b strings.Builder
	b.WriteString(prompt)
	b.WriteString("\n\n[Control-plane routed message")
	if event.Sender != "" {
		b.WriteString(" from ")
		b.WriteString(event.Sender)
	}
	b.WriteString("]\n")
	if event.Reason != "" {
		b.WriteString("Reason: ")
		b.WriteString(event.Reason)
		b.WriteString("\n")
	}
	b.WriteString(event.Message)
	return b.String()
}

func cloneMetadata(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func isInvalidUUID(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid uuid")
}
