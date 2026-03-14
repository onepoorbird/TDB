package models

// ConsistencyLevel 一致性级别
type ConsistencyLevel int32

const (
	ConsistencyUnknown     ConsistencyLevel = 0
	ConsistencyStrong      ConsistencyLevel = 1
	ConsistencyBounded     ConsistencyLevel = 2
	ConsistencySession     ConsistencyLevel = 3
	ConsistencyEventual    ConsistencyLevel = 4
)

// MergePolicy 合并策略
type MergePolicy int32

const (
	MergeUnknown       MergePolicy = 0
	LastWriterWins     MergePolicy = 1
	CausalMerge        MergePolicy = 2
	WeightedMerge      MergePolicy = 3
	CRDTMerge          MergePolicy = 4
)

// ConflictState 冲突状态
type ConflictState int32

const (
	ConflictUnknown    ConflictState = 0
	ConflictDetected   ConflictState = 1
	ConflictResolving  ConflictState = 2
	ConflictResolved   ConflictState = 3
	ConflictEscalated  ConflictState = 4
)

// ShareContract 共享契约
type ShareContract struct {
	ContractID         string            `json:"contract_id"`
	Scope              string            `json:"scope"`
	OwnerAgentID       string            `json:"owner_agent_id"`
	ReadACL            []string          `json:"read_acl"`
	WriteACL           []string          `json:"write_acl"`
	DeriveACL          []string          `json:"derive_acl"`
	TTLPolicy          int64             `json:"ttl_policy"`
	ConsistencyLevel   ConsistencyLevel  `json:"consistency_level"`
	MergePolicy        MergePolicy       `json:"merge_policy"`
	QuarantineEnabled  bool              `json:"quarantine_enabled"`
	QuarantinePolicy   string            `json:"quarantine_policy"`
	AuditEnabled       bool              `json:"audit_enabled"`
	AuditPolicy        string            `json:"audit_policy"`
	Metadata           map[string]string `json:"metadata"`
	CreatedTs          uint64            `json:"created_ts"`
	UpdatedTs          uint64            `json:"updated_ts"`
	CreatedByEventID   string            `json:"created_by_event_id"`
}

// Conflict 冲突信息
type Conflict struct {
	ConflictID         string            `json:"conflict_id"`
	ConflictType       string            `json:"conflict_type"` // "fact", "plan", "state"
	ObjectID           string            `json:"object_id"`
	ObjectType         string            `json:"object_type"`
	ConflictingVersions []string         `json:"conflicting_versions"`
	ConflictingAgents  []string          `json:"conflicting_agents"`
	Description        string            `json:"description"`
	DetectedAt         uint64            `json:"detected_at"`
	State              ConflictState     `json:"state"`
}

// ConflictResolution 冲突解决结果
type ConflictResolution struct {
	ConflictID         string            `json:"conflict_id"`
	ResolutionStrategy string            `json:"resolution_strategy"`
	ResolvedVersion    string            `json:"resolved_version"`
	MergeDetails       map[string]string `json:"merge_details"`
	ResolvedBy         string            `json:"resolved_by"`
	ResolvedAt         uint64            `json:"resolved_at"`
}

// AuditLog 审计日志
type AuditLog struct {
	LogID          string `json:"log_id"`
	TenantID       string `json:"tenant_id"`
	WorkspaceID    string `json:"workspace_id"`
	AgentID        string `json:"agent_id"`
	SessionID      string `json:"session_id"`
	Operation      string `json:"operation"`
	ObjectType     string `json:"object_type"`
	ObjectID       string `json:"object_id"`
	Action         string `json:"action"` // "create", "read", "update", "delete", "share"
	Details        string `json:"details"`
	Timestamp      uint64 `json:"timestamp"`
	SourceIP       string `json:"source_ip"`
	UserIdentity   string `json:"user_identity"`
	Success        bool   `json:"success"`
	ErrorMessage   string `json:"error_message"`
}

// AccessCheckRequest 访问检查请求
type AccessCheckRequest struct {
	AgentID    string `json:"agent_id"`
	ObjectID   string `json:"object_id"`
	ObjectType string `json:"object_type"`
	Action     string `json:"action"` // "read", "write", "derive"
}

// AccessCheckResponse 访问检查响应
type AccessCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}
