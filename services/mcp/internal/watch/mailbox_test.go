package watch

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anmhela/x/mcp/internal/tools"
)

func TestMailboxEventsHandlerReplayAndTail(t *testing.T) {
	store := tools.NewCollabStore(filepath.Join(t.TempDir(), "collab.json"))
	channel, err := store.GetOrCreateChannel("task:stream", "Stream", []string{"agent-a"}, nil)
	if err != nil {
		t.Fatalf("GetOrCreateChannel: %v", err)
	}
	replayMessage, err := store.PostMessage(channel.ID, "agent-a", "message", "replay", nil)
	if err != nil {
		t.Fatalf("PostMessage(replay): %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/mailbox/events?channel_id="+channel.ID+"&after_sequence=0", nil).WithContext(ctx)
	recorder := newStreamRecorder()
	done := make(chan struct{})
	go func() {
		NewMailboxEventsHandler(store).ServeHTTP(recorder, req)
		close(done)
	}()

	if err := recorder.waitFor("event: ready", 2*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := recorder.waitFor(`"body":"replay"`, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	output := recorder.String()
	if !strings.Contains(output, "id: 1") {
		t.Fatalf("expected replay id in stream, got %s", output)
	}
	if !strings.Contains(output, replayMessage.Body) {
		t.Fatalf("expected replay body in stream, got %s", output)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		_, _ = store.PostMessage(channel.ID, "agent-b", "message", "live", nil)
	}()

	if err := recorder.waitFor(`"body":"live"`, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	output = recorder.String()
	if !strings.Contains(output, "id: 2") {
		t.Fatalf("expected live id in stream, got %s", output)
	}
	if strings.Index(output, `"body":"replay"`) >= strings.Index(output, `"body":"live"`) {
		t.Fatalf("expected replay before live in stream, got %s", output)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not exit after cancel")
	}
}

type streamRecorder struct {
	mu     sync.Mutex
	header http.Header
	buf    bytes.Buffer
	flush  chan struct{}
}

func newStreamRecorder() *streamRecorder {
	return &streamRecorder{
		header: make(http.Header),
		flush:  make(chan struct{}, 32),
	}
}

func (s *streamRecorder) Header() http.Header {
	return s.header
}

func (s *streamRecorder) WriteHeader(_ int) {}

func (s *streamRecorder) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *streamRecorder) Flush() {
	select {
	case s.flush <- struct{}{}:
	default:
	}
}

func (s *streamRecorder) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func (s *streamRecorder) waitFor(fragment string, timeout time.Duration) error {
	deadline := time.After(timeout)
	for {
		if strings.Contains(s.String(), fragment) {
			return nil
		}
		select {
		case <-s.flush:
		case <-time.After(20 * time.Millisecond):
		case <-deadline:
			return context.DeadlineExceeded
		}
	}
}
