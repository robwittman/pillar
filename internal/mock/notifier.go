package mock

type AgentNotifier struct {
	NotifyDirectiveFn func(agentID string, directiveType string, payload string) error
}

func (m *AgentNotifier) NotifyDirective(agentID string, directiveType string, payload string) error {
	if m.NotifyDirectiveFn != nil {
		return m.NotifyDirectiveFn(agentID, directiveType, payload)
	}
	return nil
}
