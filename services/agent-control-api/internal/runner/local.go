package runner

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LocalRunner executes agent runs by invoking the selected local agent runtime.
// Used for local development when CLOUD_RUN_JOB_NAME is not set.
type LocalRunner struct {
	timeout  time.Duration
	bus      *Bus
	lookPath func(string) (string, error)
	environ  func() []string
	onChunk  func(context.Context, uuid.UUID, string) error
}

func NewLocalRunner(bus *Bus, onChunk func(context.Context, uuid.UUID, string) error) *LocalRunner {
	return &LocalRunner{
		timeout:  10 * time.Minute,
		bus:      bus,
		lookPath: exec.LookPath,
		environ:  os.Environ,
		onChunk:  onChunk,
	}
}

// Run executes the prompt, streams lines to the bus, and returns the full output.
func (r *LocalRunner) Run(ctx context.Context, id uuid.UUID, spec ExecutionSpec) (output string, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	invocation, err := BuildInvocation(spec)
	if err != nil {
		return "", err
	}
	bin, err := r.lookPath(invocation.Command)
	if err != nil {
		bin = invocation.Command
	}
	cmd := exec.CommandContext(ctx, bin, invocation.Args...)
	cmd.Env = append(r.environ(), serializeEnv(invocation.Env)...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into the same pipe

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start claude: %w", err)
	}

	var sb strings.Builder
	var chunkErr error
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		sb.WriteString(line)
		if r.bus != nil {
			r.bus.Publish(Chunk{RunID: id, Output: line})
		}
		if r.onChunk != nil {
			if err := r.onChunk(ctx, id, line); err != nil && chunkErr == nil {
				chunkErr = err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return sb.String(), fmt.Errorf("scan output: %w", err)
	}

	runErr := cmd.Wait()
	full := strings.TrimRight(sb.String(), "\n")
	if chunkErr != nil || runErr != nil {
		return full, errors.Join(runErr, chunkErr)
	}
	return full, nil
}

func serializeEnv(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}

	pairs := make([]string, 0, len(env))
	for k, v := range env {
		if k == "" {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return pairs
}
