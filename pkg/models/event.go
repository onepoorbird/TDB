package models

// EventType 定义事件类型
type EventType int32

const (
	EventUnknown EventType = 0

	// 消息类事件
	UserMessage       EventType = 1
	AssistantMessage  EventType = 2

	// 工具类事件
	ToolCallIssued     EventType = 10
	ToolResultReturned EventType = 11

	// 检索类事件
	RetrievalExecuted EventType = 20

	// 记忆类事件
	MemoryWriteRequested EventType = 30
	MemoryConsolidated   EventType = 31
	MemoryUpdated        EventType = 32
	EventMemoryDeleted   EventType = 33

	// 计划类事件
	PlanUpdated  EventType = 40
	PlanExecuted EventType = 41

	// 反思类事件
	CritiqueGenerated  EventType = 50
	ReflectionCreated  EventType = 51

	// 任务类事件
	TaskStarted  EventType = 60
	TaskFinished EventType = 61
	TaskFailed   EventType = 62

	// 协作类事件
	HandoffOccurred       EventType = 70
	SharedMemoryAccessed  EventType = 71

	// 系统类事件
	SessionCreated    EventType = 80
	SessionEnded      EventType = 81
	AgentRegistered   EventType = 82
	AgentDeregistered EventType = 83
)

// Event 系统事实来源
type Event struct {
	EventID       string            `json:"event_id"`
	TenantID      string            `json:"tenant_id"`
	WorkspaceID   string            `json:"workspace_id"`
	AgentID       string            `json:"agent_id"`
	SessionID     string            `json:"session_id"`
	EventType     EventType         `json:"event_type"`
	EventTime     uint64            `json:"event_time"`      // 事件实际发生时间
	IngestTime    uint64            `json:"ingest_time"`     // 进入WAL时间
	VisibleTime   uint64            `json:"visible_time"`    // 对查询可见时间
	LogicalTs     uint64            `json:"logical_ts"`      // 逻辑时间戳
	ParentEventID string            `json:"parent_event_id"`
	CausalRefs    []string          `json:"causal_refs"`
	Payload       []byte            `json:"payload"`
	Source        string            `json:"source"`
	Importance    float32           `json:"importance"`
	Visibility    string            `json:"visibility"`
	Version       int64             `json:"version"`
	Metadata      map[string]string `json:"metadata"`
}

// EventBatch 事件批次
type EventBatch struct {
	Events         []*Event `json:"events"`
	BatchTimestamp uint64   `json:"batch_timestamp"`
	BatchID        string   `json:"batch_id"`
}

// EventFilter 事件过滤条件
type EventFilter struct {
	AgentIDs       []string          `json:"agent_ids"`
	SessionIDs     []string          `json:"session_ids"`
	EventTypes     []EventType       `json:"event_types"`
	StartTime      uint64            `json:"start_time"`
	EndTime        uint64            `json:"end_time"`
	MinImportance  float32           `json:"min_importance"`
	MaxImportance  float32           `json:"max_importance"`
	MetadataFilter map[string]string `json:"metadata_filter"`
}

// EventLogPosition 事件日志位置
type EventLogPosition struct {
	ChannelName string `json:"channel_name"`
	LogID       uint64 `json:"log_id"`
	Timestamp   uint64 `json:"timestamp"`
}
