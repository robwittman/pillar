package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// builtinTool pairs a tool definition with its handler and a name for filtering.
type builtinTool struct {
	name    string
	tool    anthropic.ToolUnionParam
	handler ToolHandler
}

// builtinTools returns all available builtin tools. Registration is gated
// by the agent's ToolPermissions config.
func (r *Runner) builtinTools() []builtinTool {
	return []builtinTool{
		{
			name: "http_request",
			tool: anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        "http_request",
					Description: anthropic.String("Make an HTTP request. Use this to interact with REST APIs and web services."),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: map[string]interface{}{
							"method": map[string]interface{}{
								"type":        "string",
								"description": "HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)",
							},
							"url": map[string]interface{}{
								"type":        "string",
								"description": "Full URL to request",
							},
							"headers": map[string]interface{}{
								"type":        "object",
								"description": "HTTP headers as key-value pairs",
								"additionalProperties": map[string]interface{}{
									"type": "string",
								},
							},
							"body": map[string]interface{}{
								"type":        "string",
								"description": "Request body (for POST/PUT/PATCH)",
							},
						},
						Required: []string{"method", "url"},
					},
				},
			},
			handler: r.handleHTTPRequest,
		},
		{
			name: "get_attribute",
			tool: anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        "get_attribute",
					Description: anthropic.String("Read an agent attribute by namespace. Attributes contain credentials and configuration from external systems."),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: map[string]interface{}{
							"namespace": map[string]interface{}{
								"type":        "string",
								"description": "The attribute namespace (e.g., 'keycloak', 'redmine')",
							},
						},
						Required: []string{"namespace"},
					},
				},
			},
			handler: r.handleGetAttribute,
		},
		{
			name: "read_file",
			tool: anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        "read_file",
					Description: anthropic.String("Read the contents of a file from the local filesystem. Useful for reading configuration files, credentials, and system information (e.g., /etc/resolv.conf, /etc/hosts, service account tokens)."),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: map[string]interface{}{
							"path": map[string]interface{}{
								"type":        "string",
								"description": "Absolute path to the file to read",
							},
							"max_bytes": map[string]interface{}{
								"type":        "integer",
								"description": "Maximum bytes to read (default 65536)",
							},
						},
						Required: []string{"path"},
					},
				},
			},
			handler: r.handleReadFile,
		},
		{
			name: "dns_lookup",
			tool: anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        "dns_lookup",
					Description: anthropic.String("Perform DNS lookups. Supports forward lookups (hostname to IP), reverse lookups (IP to hostname), and SRV record queries. Useful for network discovery and service enumeration."),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: map[string]interface{}{
							"type": map[string]interface{}{
								"type":        "string",
								"description": "Lookup type: 'forward' (name→IPs), 'reverse' (IP→names), 'srv' (SRV records), 'txt' (TXT records), 'mx' (MX records), 'ns' (NS records)",
								"enum":        []string{"forward", "reverse", "srv", "txt", "mx", "ns"},
							},
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Hostname for forward/srv/txt/mx/ns lookup, or IP address for reverse lookup",
							},
						},
						Required: []string{"type", "name"},
					},
				},
			},
			handler: r.handleDNSLookup,
		},
		{
			name: "tcp_connect",
			tool: anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        "tcp_connect",
					Description: anthropic.String("Test TCP connectivity to a host:port. Returns whether the port is open, and captures any banner/greeting sent by the service. Useful for discovering non-HTTP services (SSH, SMTP, databases, IPMI, SNMP, etc)."),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: map[string]interface{}{
							"host": map[string]interface{}{
								"type":        "string",
								"description": "Hostname or IP address",
							},
							"port": map[string]interface{}{
								"type":        "integer",
								"description": "TCP port number",
							},
							"timeout_ms": map[string]interface{}{
								"type":        "integer",
								"description": "Connection timeout in milliseconds (default 3000)",
							},
						},
						Required: []string{"host", "port"},
					},
				},
			},
			handler: r.handleTCPConnect,
		},
	}
}

// toolAllowed checks whether a tool name is permitted by the agent's
// ToolPermissions. If allowed_tools is non-empty, only listed tools are
// permitted (allowlist). If denied_tools is non-empty, listed tools are
// excluded (denylist). If both are empty, all tools are allowed.
func (r *Runner) toolAllowed(name string) bool {
	perms := r.config.ToolPermissions
	if perms == nil {
		return true
	}

	if len(perms.AllowedTools) > 0 {
		for _, t := range perms.AllowedTools {
			if t == name {
				return true
			}
		}
		return false
	}

	for _, t := range perms.DeniedTools {
		if t == name {
			return false
		}
	}
	return true
}

// --- Tool handlers ---

func (r *Runner) handleReadFile(_ context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Path     string `json:"path"`
		MaxBytes int    `json:"max_bytes"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.MaxBytes <= 0 {
		params.MaxBytes = 65536
	}
	if params.MaxBytes > 1_000_000 {
		params.MaxBytes = 1_000_000
	}

	f, err := os.Open(params.Path)
	if err != nil {
		return fmt.Sprintf("error: %s", err), nil
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, int64(params.MaxBytes)))
	if err != nil {
		return fmt.Sprintf("error reading file: %s", err), nil
	}

	return string(data), nil
}

func (r *Runner) handleDNSLookup(_ context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	resolver := net.DefaultResolver
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lines []string

	switch params.Type {
	case "forward":
		addrs, err := resolver.LookupHost(ctx, params.Name)
		if err != nil {
			return fmt.Sprintf("lookup failed: %s", err), nil
		}
		lines = append(lines, fmt.Sprintf("Forward lookup: %s", params.Name))
		for _, addr := range addrs {
			lines = append(lines, fmt.Sprintf("  %s", addr))
		}

	case "reverse":
		names, err := resolver.LookupAddr(ctx, params.Name)
		if err != nil {
			return fmt.Sprintf("reverse lookup failed: %s", err), nil
		}
		lines = append(lines, fmt.Sprintf("Reverse lookup: %s", params.Name))
		for _, name := range names {
			lines = append(lines, fmt.Sprintf("  %s", name))
		}

	case "srv":
		// SRV lookups: name should be like _http._tcp.example.com or just a service name
		_, addrs, err := resolver.LookupSRV(ctx, "", "", params.Name)
		if err != nil {
			return fmt.Sprintf("SRV lookup failed: %s", err), nil
		}
		lines = append(lines, fmt.Sprintf("SRV lookup: %s", params.Name))
		for _, srv := range addrs {
			lines = append(lines, fmt.Sprintf("  %s:%d (priority=%d, weight=%d)", srv.Target, srv.Port, srv.Priority, srv.Weight))
		}

	case "txt":
		records, err := resolver.LookupTXT(ctx, params.Name)
		if err != nil {
			return fmt.Sprintf("TXT lookup failed: %s", err), nil
		}
		lines = append(lines, fmt.Sprintf("TXT lookup: %s", params.Name))
		for _, txt := range records {
			lines = append(lines, fmt.Sprintf("  %s", txt))
		}

	case "mx":
		records, err := resolver.LookupMX(ctx, params.Name)
		if err != nil {
			return fmt.Sprintf("MX lookup failed: %s", err), nil
		}
		lines = append(lines, fmt.Sprintf("MX lookup: %s", params.Name))
		for _, mx := range records {
			lines = append(lines, fmt.Sprintf("  %s (preference=%d)", mx.Host, mx.Pref))
		}

	case "ns":
		records, err := resolver.LookupNS(ctx, params.Name)
		if err != nil {
			return fmt.Sprintf("NS lookup failed: %s", err), nil
		}
		lines = append(lines, fmt.Sprintf("NS lookup: %s", params.Name))
		for _, ns := range records {
			lines = append(lines, fmt.Sprintf("  %s", ns.Host))
		}

	default:
		return fmt.Sprintf("unknown lookup type: %s", params.Type), nil
	}

	return strings.Join(lines, "\n"), nil
}

func (r *Runner) handleTCPConnect(_ context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Host      string `json:"host"`
		Port      int    `json:"port"`
		TimeoutMs int    `json:"timeout_ms"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if params.TimeoutMs <= 0 {
		params.TimeoutMs = 3000
	}
	if params.TimeoutMs > 10000 {
		params.TimeoutMs = 10000
	}

	addr := fmt.Sprintf("%s:%d", params.Host, params.Port)
	timeout := time.Duration(params.TimeoutMs) * time.Millisecond

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return fmt.Sprintf("CLOSED — %s:%d — %s", params.Host, params.Port, err), nil
	}
	defer conn.Close()

	// Try to read a banner (many services send a greeting)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	banner := make([]byte, 1024)
	n, _ := conn.Read(banner) // ignore error — not all services send banners

	result := fmt.Sprintf("OPEN — %s:%d", params.Host, params.Port)
	if n > 0 {
		bannerStr := strings.TrimSpace(string(banner[:n]))
		result += fmt.Sprintf("\nBanner: %s", bannerStr)
	}

	return result, nil
}
