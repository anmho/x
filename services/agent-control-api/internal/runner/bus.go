package runner

import (
	"sync"

	"github.com/anmho/agent-control-api/internal/domain"
	"github.com/google/uuid"
)

// Chunk is a piece of output from a run.
type Chunk struct {
	RunID        uuid.UUID
	Output       string
	ControlEvent *domain.AgentRunEvent
	Status       domain.RunStatus
	Done         bool // true on the final chunk (success or failure)
	Err          error
}

// Bus is a simple in-memory pub/sub for run output chunks.
// Subscribers receive chunks via a channel until the run completes.
type Bus struct {
	mu   sync.Mutex
	subs map[uuid.UUID][]chan Chunk
}

func NewBus() *Bus {
	return &Bus{subs: make(map[uuid.UUID][]chan Chunk)}
}

// Subscribe returns a channel that receives chunks for the given run.
// The channel is closed after the Done chunk is sent.
func (b *Bus) Subscribe(id uuid.UUID) <-chan Chunk {
	ch := make(chan Chunk, 64)
	b.mu.Lock()
	b.subs[id] = append(b.subs[id], ch)
	b.mu.Unlock()
	return ch
}

// Publish sends a chunk to all subscribers of the run.
func (b *Bus) Publish(c Chunk) {
	b.mu.Lock()
	chans := b.subs[c.RunID]
	b.mu.Unlock()
	for _, ch := range chans {
		select {
		case ch <- c:
		default:
		}
		if c.Done {
			close(ch)
		}
	}
	if c.Done {
		b.mu.Lock()
		delete(b.subs, c.RunID)
		b.mu.Unlock()
	}
}
