package plugin

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// Process manages a single plugin binary process.
type Process struct {
	name       string
	path       string
	socketPath string
	cmd        *exec.Cmd
	logger     *slog.Logger
	mu         sync.Mutex
}

// NewProcess creates a new plugin process handle.
func NewProcess(name, binaryPath string, logger *slog.Logger) *Process {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("pillar-plugin-%s.sock", name))
	return &Process{
		name:       name,
		path:       binaryPath,
		socketPath: socketPath,
		logger:     logger,
	}
}

// SocketPath returns the unix socket path for this plugin.
func (p *Process) SocketPath() string {
	return p.socketPath
}

// Start launches the plugin binary.
func (p *Process) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Remove stale socket
	os.Remove(p.socketPath)

	p.cmd = exec.Command(p.path)
	// Minimal environment — only what the plugin needs to operate.
	// Secrets and config are delivered via the Configure RPC, not env vars.
	p.cmd.Env = []string{
		fmt.Sprintf("PILLAR_PLUGIN_SOCKET=%s", p.socketPath),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin %s: %w", p.name, err)
	}

	p.logger.Info("plugin process started", "name", p.name, "pid", p.cmd.Process.Pid)
	return nil
}

// Stop kills the plugin process and cleans up.
func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	p.logger.Info("stopping plugin process", "name", p.name, "pid", p.cmd.Process.Pid)

	if err := p.cmd.Process.Signal(os.Interrupt); err != nil {
		// If interrupt fails, kill
		_ = p.cmd.Process.Kill()
	}

	err := p.cmd.Wait()
	os.Remove(p.socketPath)
	p.cmd = nil
	return err
}

// Running returns whether the plugin process is alive.
func (p *Process) Running() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}

	// Check if process is still running
	if p.cmd.ProcessState != nil {
		return false
	}
	return true
}
