package runner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/accursedgalaxy/forge/internal/provider"
)

// SpawnOptions configures a claude-cli invocation.
type SpawnOptions struct {
	// SessionID is the Forge session ID (used for process tracking, not passed to claude).
	SessionID string
	// Prompt is the text prompt to pass to claude via stdin.
	Prompt string
	// WorkDir is the working directory for the claude-cli process.
	// Should be the project's local_path. Defaults to the current working directory if empty.
	WorkDir string
	// ClaudeResumeID is the claude session ID for --resume. Empty means new session.
	ClaudeResumeID string
	// AllowedTools is the list of allowed tool names for read-only planning sessions.
	// Nil or empty means --dangerously-skip-permissions is used instead.
	AllowedTools []string
}

// spawn starts a claude-cli process and returns the os.Process and a channel of events.
// The event channel is closed when the process exits.
func spawn(ctx context.Context, binPath string, opts SpawnOptions) (*os.Process, <-chan provider.Event, error) {
	args := buildArgs(opts)

	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Env = filteredEnv()
	cmd.Stdin = bytes.NewBufferString(opts.Prompt)
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("runner: stdout pipe: %w", err)
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("runner: start claude-cli: %w", err)
	}

	events := make(chan provider.Event, 64)

	go func() {
		defer close(events)

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 256*1024), 256*1024)

		for scanner.Scan() {
			line := scanner.Bytes()
			evt, ok := ParseLine(line)
			if !ok {
				continue
			}
			select {
			case events <- evt:
			case <-ctx.Done():
				return
			}
		}

		if err := scanner.Err(); err != nil {
			slog.Warn("runner: scanner error", "session_id", opts.SessionID, "err", err)
		}

		// Reap the process; log any stderr output for debugging.
		_ = cmd.Wait()
		if stderrBuf.Len() > 0 {
			slog.Warn("runner: claude stderr", "session_id", opts.SessionID, "stderr", stderrBuf.String())
		}
	}()

	return cmd.Process, events, nil
}

// buildArgs constructs the argument list for the claude-cli invocation.
func buildArgs(opts SpawnOptions) []string {
	args := []string{}

	if opts.ClaudeResumeID != "" {
		args = append(args, "--resume", opts.ClaudeResumeID)
	}

	// --print mode: prompt is the positional argument at the end
	args = append(args, "--print")
	args = append(args, "--output-format", "stream-json")
	args = append(args, "--include-partial-messages")
	args = append(args, "--verbose")

	if len(opts.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(opts.AllowedTools, ","))
	} else {
		args = append(args, "--dangerously-skip-permissions")
	}

	return args
}

// filteredEnv returns os.Environ() with CLAUDECODE stripped to prevent the
// claude-cli subprocess from detecting it's running inside another Claude session.
func filteredEnv() []string {
	env := os.Environ()
	filtered := make([]string, 0, len(env))
	for _, kv := range env {
		if !strings.HasPrefix(kv, "CLAUDECODE=") {
			filtered = append(filtered, kv)
		}
	}
	return filtered
}
