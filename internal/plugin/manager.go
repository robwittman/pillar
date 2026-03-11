package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/robwittman/pillar/internal/config"
	"github.com/robwittman/pillar/internal/plugin/resolver"
)

const socketTimeout = 15 * time.Second

// ManagedPlugin holds a running plugin process and its gRPC client.
type ManagedPlugin struct {
	Name    string
	Process *Process
	Client  *Client
	Config  map[string]string
}

// ManagerOption configures a Manager.
type ManagerOption func(*Manager)

// WithResolver sets the resolver used to download remote plugins.
func WithResolver(r resolver.Resolver) ManagerOption {
	return func(m *Manager) {
		m.resolver = r
	}
}

// Manager starts, configures, and manages plugin processes.
type Manager struct {
	plugins  map[string]*ManagedPlugin
	resolver resolver.Resolver
	logger   *slog.Logger
	mu       sync.RWMutex
}

// NewManager creates a new plugin manager.
func NewManager(logger *slog.Logger, opts ...ManagerOption) *Manager {
	m := &Manager{
		plugins: make(map[string]*ManagedPlugin),
		logger:  logger,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// StartAll launches and configures all plugins from config.
func (m *Manager) StartAll(configs []config.PluginConfig) error {
	for _, cfg := range configs {
		if err := m.startPlugin(cfg); err != nil {
			return fmt.Errorf("failed to start plugin %s: %w", cfg.Name, err)
		}
	}
	return nil
}

func (m *Manager) startPlugin(cfg config.PluginConfig) error {
	binaryPath := cfg.Path

	// Resolve remote plugins when no local path is specified
	if binaryPath == "" && cfg.Source != "" {
		if m.resolver == nil {
			return fmt.Errorf("plugin %s has source but no resolver configured", cfg.Name)
		}
		version := cfg.Version
		if version == "" {
			version = "latest"
		}
		res, err := m.resolver.Resolve(context.Background(), cfg.Source, version)
		if err != nil {
			return fmt.Errorf("failed to resolve plugin %s: %w", cfg.Name, err)
		}
		binaryPath = res.BinaryPath
		m.logger.Info("resolved plugin", "name", cfg.Name, "version", res.Version, "path", binaryPath)
	}

	if binaryPath == "" {
		return fmt.Errorf("plugin %s has no path or source configured", cfg.Name)
	}

	proc := NewProcess(cfg.Name, binaryPath, m.logger)

	if err := proc.Start(); err != nil {
		return err
	}

	// Wait for the plugin to start listening
	if err := WaitForSocket(proc.SocketPath(), socketTimeout); err != nil {
		proc.Stop()
		return err
	}

	client, err := NewClient(proc.SocketPath())
	if err != nil {
		proc.Stop()
		return err
	}

	// Configure the plugin
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expandedConfig := expandEnvVars(cfg.Config)
	if err := client.Configure(ctx, expandedConfig); err != nil {
		client.Close()
		proc.Stop()
		return err
	}

	mp := &ManagedPlugin{
		Name:    cfg.Name,
		Process: proc,
		Client:  client,
		Config:  expandedConfig,
	}

	m.mu.Lock()
	m.plugins[cfg.Name] = mp
	m.mu.Unlock()

	m.logger.Info("plugin started and configured", "name", cfg.Name)
	return nil
}

// Plugins returns all managed plugins.
func (m *Manager) Plugins() []*ManagedPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ManagedPlugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		result = append(result, p)
	}
	return result
}

// StopAll gracefully stops all plugins.
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, mp := range m.plugins {
		if err := mp.Client.Close(); err != nil {
			m.logger.Warn("failed to close plugin client", "name", name, "error", err)
		}
		if err := mp.Process.Stop(); err != nil {
			m.logger.Warn("failed to stop plugin process", "name", name, "error", err)
		}
	}
	m.plugins = make(map[string]*ManagedPlugin)
	m.logger.Info("all plugins stopped")
}

// expandEnvVars replaces ${VAR} references in config values with the
// corresponding environment variable. Unknown variables expand to empty string.
func expandEnvVars(config map[string]string) map[string]string {
	expanded := make(map[string]string, len(config))
	for k, v := range config {
		expanded[k] = os.Expand(v, os.Getenv)
	}
	return expanded
}
