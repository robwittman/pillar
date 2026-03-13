export type AgentStatus = 'pending' | 'running' | 'stopped' | 'error'
export type WebhookStatus = 'active' | 'inactive'

export interface Agent {
  id: string
  name: string
  status: AgentStatus
  metadata: Record<string, string>
  labels: Record<string, string>
  created_at: string
  updated_at: string
}

export interface AgentStatusInfo {
  agent_id: string
  status: string
  online: boolean
}

export interface ModelParams {
  temperature: number
  top_p: number
  max_tokens: number
}

export interface MCPServerConfig {
  name: string
  transport_type: 'stdio' | 'sse'
  command?: string
  args?: string[]
  url?: string
  headers?: Record<string, string>
  env?: Record<string, string>
}

export interface ToolPermissions {
  allowed_tools?: string[]
  denied_tools?: string[]
}

export interface EscalationRule {
  name: string
  condition: string
  action: 'pause' | 'notify' | 'stop'
  message?: string
}

export interface AgentConfig {
  agent_id: string
  model_provider: string
  model_id: string
  system_prompt: string
  model_params: ModelParams
  api_credential_ref: string
  mcp_servers: MCPServerConfig[]
  tool_permissions: ToolPermissions
  max_iterations: number
  token_budget: number
  task_timeout_seconds: number
  escalation_rules: EscalationRule[]
  created_at: string
  updated_at: string
}

export interface Webhook {
  id: string
  url: string
  secret?: string
  event_types: string[]
  status: WebhookStatus
  description: string
  created_at: string
  updated_at: string
}

export interface WebhookDelivery {
  id: string
  webhook_id: string
  event_type: string
  payload: unknown
  response_code: number
  response_body: string
  status: string
  attempts: number
  last_attempt_at: string
  next_retry_at: string
  created_at: string
}

export interface AgentAttribute {
  agent_id: string
  namespace: string
  value: unknown
  created_at: string
  updated_at: string
}

export interface CreateAgentRequest {
  name: string
  metadata?: Record<string, string>
  labels?: Record<string, string>
}

export interface UpdateAgentRequest {
  name?: string
  metadata?: Record<string, string>
  labels?: Record<string, string>
}

export interface CreateConfigRequest {
  model_provider: string
  model_id: string
  system_prompt?: string
  api_credential?: string
  model_params?: Partial<ModelParams>
  tool_permissions?: ToolPermissions
  max_iterations?: number
  token_budget?: number
  task_timeout_seconds?: number
}

export interface UpdateConfigRequest extends CreateConfigRequest {}

export interface CreateWebhookRequest {
  url: string
  description?: string
  event_types: string[]
}

export interface UpdateWebhookRequest {
  url?: string
  description?: string
  event_types?: string[]
  status?: WebhookStatus
}

// Sources
export interface Source {
  id: string
  name: string
  secret?: string
  created_at: string
  updated_at: string
}

// Triggers
export type TaskStatus = 'pending' | 'assigned' | 'running' | 'completed' | 'failed'

export interface FilterCondition {
  path: string
  op: 'eq' | 'neq' | 'contains' | 'exists'
  value?: string
}

export interface TriggerFilter {
  conditions: FilterCondition[]
}

export interface Trigger {
  id: string
  source_id: string
  agent_id: string
  name: string
  filter: TriggerFilter
  task_template: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreateTriggerRequest {
  source_id: string
  agent_id: string
  name: string
  filter?: TriggerFilter
  task_template: string
}

export interface UpdateTriggerRequest {
  name?: string
  filter?: TriggerFilter
  task_template?: string
  enabled?: boolean
}

// Tasks
export interface Task {
  id: string
  agent_id: string
  trigger_id?: string
  status: TaskStatus
  prompt: string
  context?: unknown
  result?: string
  created_at: string
  updated_at: string
  completed_at?: string
}

export interface CreateTaskRequest {
  agent_id: string
  prompt: string
  context?: unknown
}
