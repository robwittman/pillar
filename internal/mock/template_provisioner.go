package mock

import "context"

type TemplateProvisioner struct {
	ProvisionForAgentFn func(ctx context.Context, agentID string, labels map[string]string) error
}

func (m *TemplateProvisioner) ProvisionForAgent(ctx context.Context, agentID string, labels map[string]string) error {
	return m.ProvisionForAgentFn(ctx, agentID, labels)
}
