package models

// AgentState 定义Agent的状态
type AgentState int32

const (
	AgentUnknown    AgentState = 0
	AgentCreating   AgentState = 1
	AgentActive     AgentState = 2
	AgentPaused     AgentState = 3
	AgentTerminated AgentState = 4
)

// Agent 表示执行主体
type Agent struct {
	AgentID             string            `json:"agent_id"`
	TenantID            string            `json:"tenant_id"`
	WorkspaceID         string            `json:"workspace_id"`
	AgentType           string            `json:"agent_type"`
	RoleProfile         string            `json:"role_profile"`
	PolicyRef           string            `json:"policy_ref"`
	CapabilitySet       []string          `json:"capability_set"`
	DefaultMemoryPolicy string            `json:"default_memory_policy"`
	CreatedTs           uint64            `json:"created_ts"`
	UpdatedTs           uint64            `json:"updated_ts"`
	State               AgentState        `json:"state"`
	Metadata            map[string]string `json:"metadata"`
}

// SessionState 定义Session的状态
type SessionState int32

const (
	SessionUnknown   SessionState = 0
	SessionCreating  SessionState = 1
	SessionActive    SessionState = 2
	SessionPaused    SessionState = 3
	SessionCompleted SessionState = 4
	SessionFailed    SessionState = 5
)

// Session 表示具体任务、会话或推理线程
type Session struct {
	SessionID       string            `json:"session_id"`
	AgentID         string            `json:"agent_id"`
	ParentSessionID string            `json:"parent_session_id"`
	TaskType        string            `json:"task_type"`
	Goal            string            `json:"goal"`
	ContextRef      string            `json:"context_ref"`
	StartTs         uint64            `json:"start_ts"`
	EndTs           uint64            `json:"end_ts"`
	State           SessionState      `json:"state"`
	BudgetToken     int64             `json:"budget_token"`
	BudgetTimeMs    int64             `json:"budget_time_ms"`
	Metadata        map[string]string `json:"metadata"`
}

// Workspace 工作空间
type Workspace struct {
	WorkspaceID string            `json:"workspace_id"`
	TenantID    string            `json:"tenant_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	AgentIDs    []string          `json:"agent_ids"`
	Metadata    map[string]string `json:"metadata"`
	CreatedTs   uint64            `json:"created_ts"`
	UpdatedTs   uint64            `json:"updated_ts"`
}
