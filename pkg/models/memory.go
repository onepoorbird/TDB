package models

// MemoryType 记忆类型
type MemoryType int32

const (
	MemoryUnknown MemoryType = 0
	Episodic      MemoryType = 1 // 情景记忆：发生过什么
	Semantic      MemoryType = 2 // 语义记忆：抽象事实与知识
	Procedural    MemoryType = 3 // 程序记忆：规则、流程、策略
	Social        MemoryType = 4 // 社交/共享记忆：共享约定、团队状态
	Reflective    MemoryType = 5 // 反思记忆：反思、修正、经验结论
)

// MemoryLevel 记忆蒸馏层级
type MemoryLevel int32

const (
	LevelUnknown MemoryLevel = 0
	LevelRaw     MemoryLevel = 1 // 原始记录
	LevelSummary MemoryLevel = 2 // 摘要
	LevelPattern MemoryLevel = 3 // 归纳规律
)

// MemoryState 记忆状态
type MemoryState int32

const (
	MemoryStateUnknown    MemoryState = 0
	MemoryActive          MemoryState = 1
	MemoryFading          MemoryState = 2
	MemoryArchived        MemoryState = 3
	MemoryQuarantined     MemoryState = 4
	MemoryDeleted         MemoryState = 5
)

// Memory 记忆对象 - Base Layer
type Memory struct {
	MemoryID       string            `json:"memory_id"`
	MemoryType     MemoryType        `json:"memory_type"`
	AgentID        string            `json:"agent_id"`
	SessionID      string            `json:"session_id"`
	Scope          string            `json:"scope"`
	Level          MemoryLevel       `json:"level"`
	Content        string            `json:"content"`
	Summary        string            `json:"summary"`
	SourceEventIDs []string          `json:"source_event_ids"`
	Confidence     float32           `json:"confidence"`
	Importance     float32           `json:"importance"`
	FreshnessScore float32           `json:"freshness_score"`
	TTL            int64             `json:"ttl"` // 存活时间（秒）
	ValidFrom      uint64            `json:"valid_from"`
	ValidTo        uint64            `json:"valid_to"`
	ProvenanceRef  string            `json:"provenance_ref"`
	Version        int64             `json:"version"`
	IsActive       bool              `json:"is_active"`
	State          MemoryState       `json:"state"`
	CreatedTs      uint64            `json:"created_ts"`
	UpdatedTs      uint64            `json:"updated_ts"`
	Metadata       map[string]string `json:"metadata"`
	EmbeddingRef   string            `json:"embedding_ref"`
	EmbeddingVector []float32        `json:"embedding_vector"`
}

// MemoryPolicy 记忆策略 - Policy Layer
type MemoryPolicy struct {
	MemoryID           string            `json:"memory_id"`
	SalienceWeight     float32           `json:"salience_weight"`
	TTL                int64             `json:"ttl"`
	DecayFn            string            `json:"decay_fn"` // "linear", "exponential", "step"
	Confidence         float32           `json:"confidence"`
	Verified           bool              `json:"verified"`
	VerifiedBy         string            `json:"verified_by"`
	VerifiedAt         uint64            `json:"verified_at"`
	Quarantined        bool              `json:"quarantined"`
	QuarantineReason   string            `json:"quarantine_reason"`
	VisibilityPolicy   string            `json:"visibility_policy"`
	ReadACL            []string          `json:"read_acl"`
	WriteACL           []string          `json:"write_acl"`
	DeriveACL          []string          `json:"derive_acl"`
	PolicyReason       string            `json:"policy_reason"`
	PolicySource       string            `json:"policy_source"`
	PolicyEventID      string            `json:"policy_event_id"`
	CreatedTs          uint64            `json:"created_ts"`
	UpdatedTs          uint64            `json:"updated_ts"`
}

// MemoryAdaptation 记忆适配 - Adaptation Layer
type MemoryAdaptation struct {
	MemoryID             string             `json:"memory_id"`
	RetrievalProfile     map[string]float32 `json:"retrieval_profile"`
	RankingParams        map[string]float32 `json:"ranking_params"`
	FilteringThresholds  map[string]float32 `json:"filtering_thresholds"`
	ProjectionWeights    map[string]float32 `json:"projection_weights"`
	EmbeddingFamily      string             `json:"embedding_family"`
	ModelID              string             `json:"model_id"`
	AdaptationReason     string             `json:"adaptation_reason"`
	AdaptationSource     string             `json:"adaptation_source"`
	CreatedTs            uint64             `json:"created_ts"`
	UpdatedTs            uint64             `json:"updated_ts"`
}

// State 状态对象
type State struct {
	StateID            string            `json:"state_id"`
	AgentID            string            `json:"agent_id"`
	SessionID          string            `json:"session_id"`
	StateType          string            `json:"state_type"`
	StateKey           string            `json:"state_key"`
	StateValue         []byte            `json:"state_value"`
	Version            int64             `json:"version"`
	DerivedFromEventID string            `json:"derived_from_event_id"`
	CheckpointTs       uint64            `json:"checkpoint_ts"`
	CreatedTs          uint64            `json:"created_ts"`
	UpdatedTs          uint64            `json:"updated_ts"`
	Metadata           map[string]string `json:"metadata"`
}

// Artifact 外部工件
type Artifact struct {
	ArtifactID        string            `json:"artifact_id"`
	SessionID         string            `json:"session_id"`
	OwnerAgentID      string            `json:"owner_agent_id"`
	ArtifactType      string            `json:"artifact_type"`
	URI               string            `json:"uri"`
	ContentRef        string            `json:"content_ref"`
	MimeType          string            `json:"mime_type"`
	Metadata          map[string]string `json:"metadata"`
	Hash              string            `json:"hash"`
	ProducedByEventID string            `json:"produced_by_event_id"`
	Version           int64             `json:"version"`
	CreatedTs         uint64            `json:"created_ts"`
}

// Relation 关系/边
type Relation struct {
	EdgeID          string            `json:"edge_id"`
	SrcObjectID     string            `json:"src_object_id"`
	SrcType         string            `json:"src_type"` // "memory", "event", "state", "artifact"
	DstObjectID     string            `json:"dst_object_id"`
	DstType         string            `json:"dst_type"`
	RelationType    string            `json:"relation_type"` // "caused_by", "derived_from", "supports", "contradicts", "summarizes", "updates", "uses_tool", "belongs_to_task", "shared_with"
	Weight          float32           `json:"weight"`
	Properties      map[string]string `json:"properties"`
	CreatedTs       uint64            `json:"created_ts"`
	CreatedByEventID string           `json:"created_by_event_id"`
}

// MemoryFilter 记忆过滤条件
type MemoryFilter struct {
	AgentIDs       []string          `json:"agent_ids"`
	SessionIDs     []string          `json:"session_ids"`
	MemoryTypes    []MemoryType      `json:"memory_types"`
	Levels         []MemoryLevel     `json:"levels"`
	States         []MemoryState     `json:"states"`
	Scope          string            `json:"scope"`
	StartTime      uint64            `json:"start_time"`
	EndTime        uint64            `json:"end_time"`
	MinConfidence  float32           `json:"min_confidence"`
	MinImportance  float32           `json:"min_importance"`
	MetadataFilter map[string]string `json:"metadata_filter"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Memory      *Memory `json:"memory"`
	Score       float32 `json:"score"`
	Explanation string  `json:"explanation"`
}
