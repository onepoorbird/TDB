# TDB Schema与对象模型层 - 版本 1.01

> 创建时间: 2026-03-14
> 本版本内容: Proto定义、etcd元数据Schema、Go模型代码

***

## 1. Proto定义文件

路径详见pkg/proto

### 1.1 agent.proto

```protobuf
syntax = "proto3";
package agent.proto;
option go_package="github.com/milvus-io/milvus/pkg/v2/proto/agentpb";

import "common.proto";

// AgentState 定义Agent的状态
enum AgentState {
    AGENT_UNKNOWN = 0;
    AGENT_CREATING = 1;
    AGENT_ACTIVE = 2;
    AGENT_PAUSED = 3;
    AGENT_TERMINATED = 4;
}

// Agent 表示执行主体
message AgentInfo {
    string agent_id = 1;
    string tenant_id = 2;
    string workspace_id = 3;
    string agent_type = 4;
    string role_profile = 5;
    string policy_ref = 6;
    repeated string capability_set = 7;
    string default_memory_policy = 8;
    uint64 created_ts = 9;
    uint64 updated_ts = 10;
    AgentState state = 11;
    map<string, string> metadata = 12;
}

// SessionState 定义Session的状态
enum SessionState {
    SESSION_UNKNOWN = 0;
    SESSION_CREATING = 1;
    SESSION_ACTIVE = 2;
    SESSION_PAUSED = 3;
    SESSION_COMPLETED = 4;
    SESSION_FAILED = 5;
}

// Session 表示具体任务、会话或推理线程
message SessionInfo {
    string session_id = 1;
    string agent_id = 2;
    string parent_session_id = 3;
    string task_type = 4;
    string goal = 5;
    string context_ref = 6;
    uint64 start_ts = 7;
    uint64 end_ts = 8;
    SessionState state = 9;
    int64 budget_token = 10;
    int64 budget_time_ms = 11;
    map<string, string> metadata = 12;
}

// CreateAgentRequest 创建Agent请求
message CreateAgentRequest {
    string tenant_id = 1;
    string workspace_id = 2;
    string agent_type = 3;
    string role_profile = 4;
    repeated string capability_set = 5;
    string default_memory_policy = 6;
    map<string, string> metadata = 7;
}

// CreateAgentResponse 创建Agent响应
message CreateAgentResponse {
    common.Status status = 1;
    string agent_id = 2;
}

// GetAgentRequest 获取Agent请求
message GetAgentRequest {
    string agent_id = 1;
}

// GetAgentResponse 获取Agent响应
message GetAgentResponse {
    common.Status status = 1;
    AgentInfo agent = 2;
}

// ListAgentsRequest 列出Agent请求
message ListAgentsRequest {
    string tenant_id = 1;
    string workspace_id = 2;
    string agent_type = 3;
}

// ListAgentsResponse 列出Agent响应
message ListAgentsResponse {
    common.Status status = 1;
    repeated AgentInfo agents = 2;
}

// UpdateAgentRequest 更新Agent请求
message UpdateAgentRequest {
    string agent_id = 1;
    string role_profile = 2;
    repeated string capability_set = 3;
    string default_memory_policy = 4;
    map<string, string> metadata = 5;
}

// UpdateAgentResponse 更新Agent响应
message UpdateAgentResponse {
    common.Status status = 1;
}

// DeleteAgentRequest 删除Agent请求
message DeleteAgentRequest {
    string agent_id = 1;
}

// DeleteAgentResponse 删除Agent响应
message DeleteAgentResponse {
    common.Status status = 1;
}

// CreateSessionRequest 创建Session请求
message CreateSessionRequest {
    string agent_id = 1;
    string parent_session_id = 2;
    string task_type = 3;
    string goal = 4;
    string context_ref = 5;
    int64 budget_token = 6;
    int64 budget_time_ms = 7;
    map<string, string> metadata = 8;
}

// CreateSessionResponse 创建Session响应
message CreateSessionResponse {
    common.Status status = 1;
    string session_id = 2;
}

// GetSessionRequest 获取Session请求
message GetSessionRequest {
    string session_id = 1;
}

// GetSessionResponse 获取Session响应
message GetSessionResponse {
    common.Status status = 1;
    SessionInfo session = 2;
}

// ListSessionsRequest 列出Session请求
message ListSessionsRequest {
    string agent_id = 1;
    string parent_session_id = 2;
    SessionState state = 3;
}

// ListSessionsResponse 列出Session响应
message ListSessionsResponse {
    common.Status status = 1;
    repeated SessionInfo sessions = 2;
}

// UpdateSessionRequest 更新Session请求
message UpdateSessionRequest {
    string session_id = 1;
    SessionState state = 2;
    string goal = 3;
    map<string, string> metadata = 4;
}

// UpdateSessionResponse 更新Session响应
message UpdateSessionResponse {
    common.Status status = 1;
}

// AgentService Agent服务定义
service AgentService {
    rpc CreateAgent(CreateAgentRequest) returns (CreateAgentResponse);
    rpc GetAgent(GetAgentRequest) returns (GetAgentResponse);
    rpc ListAgents(ListAgentsRequest) returns (ListAgentsResponse);
    rpc UpdateAgent(UpdateAgentRequest) returns (UpdateAgentResponse);
    rpc DeleteAgent(DeleteAgentRequest) returns (DeleteAgentResponse);
    
    rpc CreateSession(CreateSessionRequest) returns (CreateSessionResponse);
    rpc GetSession(GetSessionRequest) returns (GetSessionResponse);
    rpc ListSessions(ListSessionsRequest) returns (ListSessionsResponse);
    rpc UpdateSession(UpdateSessionRequest) returns (UpdateSessionResponse);
}
```

***

### 1.2 event.proto

```protobuf
syntax = "proto3";
package event.proto;
option go_package="github.com/milvus-io/milvus/pkg/v2/proto/eventpb";

import "common.proto";

// EventType 定义事件类型
enum EventType {
    EVENT_UNKNOWN = 0;
    
    // 消息类事件
    USER_MESSAGE = 1;
    ASSISTANT_MESSAGE = 2;
    
    // 工具类事件
    TOOL_CALL_ISSUED = 10;
    TOOL_RESULT_RETURNED = 11;
    
    // 检索类事件
    RETRIEVAL_EXECUTED = 20;
    
    // 记忆类事件
    MEMORY_WRITE_REQUESTED = 30;
    MEMORY_CONSOLIDATED = 31;
    MEMORY_UPDATED = 32;
    MEMORY_DELETED = 33;
    
    // 计划类事件
    PLAN_UPDATED = 40;
    PLAN_EXECUTED = 41;
    
    // 反思类事件
    CRITIQUE_GENERATED = 50;
    REFLECTION_CREATED = 51;
    
    // 任务类事件
    TASK_STARTED = 60;
    TASK_FINISHED = 61;
    TASK_FAILED = 62;
    
    // 协作类事件
    HANDOFF_OCCURRED = 70;
    SHARED_MEMORY_ACCESSED = 71;
    
    // 系统类事件
    SESSION_CREATED = 80;
    SESSION_ENDED = 81;
    AGENT_REGISTERED = 82;
    AGENT_DEREGISTERED = 83;
}

// Event 系统事实来源
message Event {
    string event_id = 1;
    string tenant_id = 2;
    string workspace_id = 3;
    string agent_id = 4;
    string session_id = 5;
    EventType event_type = 6;
    
    // 时间戳
    uint64 event_time = 7;      // 事件实际发生时间
    uint64 ingest_time = 8;     // 进入WAL时间
    uint64 visible_time = 9;    // 对查询可见时间
    uint64 logical_ts = 10;     // 逻辑时间戳
    
    // 因果关系
    string parent_event_id = 11;
    repeated string causal_refs = 12;
    
    // 内容
    bytes payload = 13;
    string source = 14;
    
    // 重要性
    float importance = 15;
    
    // 可见性
    string visibility = 16;
    
    // 版本
    int64 version = 17;
    
    // 元数据
    map<string, string> metadata = 18;
}

// EventBatch 事件批次
message EventBatch {
    repeated Event events = 1;
    uint64 batch_timestamp = 2;
    string batch_id = 3;
}

// EventFilter 事件过滤条件
message EventFilter {
    repeated string agent_ids = 1;
    repeated string session_ids = 2;
    repeated EventType event_types = 3;
    uint64 start_time = 4;
    uint64 end_time = 5;
    float min_importance = 6;
    float max_importance = 7;
    map<string, string> metadata_filter = 8;
}

// AppendEventRequest 追加事件请求
message AppendEventRequest {
    string tenant_id = 1;
    string workspace_id = 2;
    string agent_id = 3;
    string session_id = 4;
    EventType event_type = 5;
    bytes payload = 6;
    float importance = 7;
    string visibility = 8;
    string parent_event_id = 9;
    repeated string causal_refs = 10;
    map<string, string> metadata = 11;
}

// AppendEventResponse 追加事件响应
message AppendEventResponse {
    common.Status status = 1;
    string event_id = 2;
    uint64 logical_ts = 3;
}

// GetEventRequest 获取事件请求
message GetEventRequest {
    string event_id = 1;
}

// GetEventResponse 获取事件响应
message GetEventResponse {
    common.Status status = 1;
    Event event = 2;
}

// QueryEventsRequest 查询事件请求
message QueryEventsRequest {
    EventFilter filter = 1;
    int64 limit = 2;
    int64 offset = 3;
    string order_by = 4;  // "time", "importance", "logical_ts"
    bool ascending = 5;
}

// QueryEventsResponse 查询事件响应
message QueryEventsResponse {
    common.Status status = 1;
    repeated Event events = 2;
    int64 total_count = 3;
}

// SubscribeEventsRequest 订阅事件请求
message SubscribeEventsRequest {
    string subscriber_id = 1;
    EventFilter filter = 2;
    uint64 start_timestamp = 3;
}

// SubscribeEventsResponse 订阅事件响应（流式）
message SubscribeEventsResponse {
    oneof response {
        Event event = 1;
        common.Status status = 2;
    }
}

// EventLogPosition 事件日志位置
message EventLogPosition {
    string channel_name = 1;
    uint64 log_id = 2;
    uint64 timestamp = 3;
}

// EventService 事件服务定义
service EventService {
    // 单次事件操作
    rpc AppendEvent(AppendEventRequest) returns (AppendEventResponse);
    rpc GetEvent(GetEventRequest) returns (GetEventResponse);
    rpc QueryEvents(QueryEventsRequest) returns (QueryEventsResponse);
    
    // 流式订阅
    rpc SubscribeEvents(SubscribeEventsRequest) returns (stream SubscribeEventsResponse);
}
```

***

### 1.3 memory.proto

```protobuf
syntax = "proto3";
package memory.proto;
option go_package="github.com/milvus-io/milvus/pkg/v2/proto/memorypb";

import "common.proto";
import "schema.proto";

// MemoryType 记忆类型
enum MemoryType {
    MEMORY_UNKNOWN = 0;
    EPISODIC = 1;       // 情景记忆：发生过什么
    SEMANTIC = 2;       // 语义记忆：抽象事实与知识
    PROCEDURAL = 3;     // 程序记忆：规则、流程、策略
    SOCIAL = 4;         // 社交/共享记忆：共享约定、团队状态
    REFLECTIVE = 5;     // 反思记忆：反思、修正、经验结论
}

// MemoryLevel 记忆蒸馏层级
enum MemoryLevel {
    LEVEL_UNKNOWN = 0;
    LEVEL_RAW = 1;      // 原始记录
    LEVEL_SUMMARY = 2;  // 摘要
    LEVEL_PATTERN = 3;  // 归纳规律
}

// MemoryState 记忆状态
enum MemoryState {
    MEMORY_STATE_UNKNOWN = 0;
    MEMORY_ACTIVE = 1;
    MEMORY_FADING = 2;
    MEMORY_ARCHIVED = 3;
    MEMORY_QUARANTINED = 4;
    MEMORY_DELETED = 5;
}

// Memory 记忆对象
message Memory {
    string memory_id = 1;
    MemoryType memory_type = 2;
    string agent_id = 3;
    string session_id = 4;
    string scope = 5;
    MemoryLevel level = 6;
    
    // 内容
    string content = 7;
    string summary = 8;
    repeated string source_event_ids = 9;
    
    // 质量指标
    float confidence = 10;
    float importance = 11;
    float freshness_score = 12;
    
    // 生命周期
    int64 ttl = 13;  // 存活时间（秒）
    uint64 valid_from = 14;
    uint64 valid_to = 15;
    
    // 溯源
    string provenance_ref = 16;
    
    // 版本
    int64 version = 17;
    bool is_active = 18;
    MemoryState state = 19;
    
    // 时间戳
    uint64 created_ts = 20;
    uint64 updated_ts = 21;
    
    // 元数据
    map<string, string> metadata = 22;
    
    // 向量引用（实际向量存储在Milvus中）
    string embedding_ref = 23;
    repeated float embedding_vector = 24;
}

// MemoryPolicy 记忆策略（Policy Layer）
message MemoryPolicy {
    string memory_id = 1;
    
    // 显著性权重
    float salience_weight = 2;
    
    // TTL策略
    int64 ttl = 3;
    string decay_fn = 4;  // "linear", "exponential", "step"
    
    // 置信度
    float confidence = 5;
    
    // 验证状态
    bool verified = 6;
    string verified_by = 7;
    uint64 verified_at = 8;
    
    // 隔离状态
    bool quarantined = 9;
    string quarantine_reason = 10;
    
    // 可见性策略
    string visibility_policy = 11;
    
    // ACL
    repeated string read_acl = 12;
    repeated string write_acl = 13;
    repeated string derive_acl = 14;
    
    // 策略来源
    string policy_reason = 15;
    string policy_source = 16;
    string policy_event_id = 17;
    
    // 时间戳
    uint64 created_ts = 18;
    uint64 updated_ts = 19;
}

// MemoryAdaptation 记忆适配（Adaptation Layer）
message MemoryAdaptation {
    string memory_id = 1;
    
    // 检索配置
    map<string, float> retrieval_profile = 2;
    
    // 排序参数
    map<string, float> ranking_params = 3;
    
    // 过滤阈值
    map<string, float> filtering_thresholds = 4;
    
    // 投影权重
    map<string, float> projection_weights = 5;
    
    // 嵌入模型选择
    string embedding_family = 6;
    string model_id = 7;
    
    // 适配来源
    string adaptation_reason = 8;
    string adaptation_source = 9;
    
    // 时间戳
    uint64 created_ts = 10;
    uint64 updated_ts = 11;
}

// State 状态对象
message State {
    string state_id = 1;
    string agent_id = 2;
    string session_id = 3;
    string state_type = 4;
    string state_key = 5;
    bytes state_value = 6;
    int64 version = 7;
    string derived_from_event_id = 8;
    uint64 checkpoint_ts = 9;
    uint64 created_ts = 10;
    uint64 updated_ts = 11;
    map<string, string> metadata = 12;
}

// Artifact 外部工件
message Artifact {
    string artifact_id = 1;
    string session_id = 2;
    string owner_agent_id = 3;
    string artifact_type = 4;
    string uri = 5;
    string content_ref = 6;
    string mime_type = 7;
    map<string, string> metadata = 8;
    string hash = 9;
    string produced_by_event_id = 10;
    int64 version = 11;
    uint64 created_ts = 12;
}

// Relation 关系/边
message Relation {
    string edge_id = 1;
    string src_object_id = 2;
    string src_type = 3;  // "memory", "event", "state", "artifact"
    string dst_object_id = 4;
    string dst_type = 5;
    string relation_type = 6;  // "caused_by", "derived_from", "supports", "contradicts", "summarizes", "updates", "uses_tool", "belongs_to_task", "shared_with"
    float weight = 7;
    map<string, string> properties = 8;
    uint64 created_ts = 9;
    string created_by_event_id = 10;
}

// CreateMemoryRequest 创建记忆请求
message CreateMemoryRequest {
    string tenant_id = 1;
    string workspace_id = 2;
    string agent_id = 3;
    string session_id = 4;
    MemoryType memory_type = 5;
    string scope = 6;
    string content = 7;
    string summary = 8;
    repeated string source_event_ids = 9;
    float confidence = 10;
    float importance = 11;
    int64 ttl = 12;
    map<string, string> metadata = 13;
    repeated float embedding_vector = 14;
}

// CreateMemoryResponse 创建记忆响应
message CreateMemoryResponse {
    common.Status status = 1;
    string memory_id = 2;
    uint64 version = 3;
}

// GetMemoryRequest 获取记忆请求
message GetMemoryRequest {
    string memory_id = 1;
    uint64 version = 2;  // 0表示最新版本
    uint64 timestamp = 3;  // 时间旅行查询
}

// GetMemoryResponse 获取记忆响应
message GetMemoryResponse {
    common.Status status = 1;
    Memory memory = 2;
    MemoryPolicy policy = 3;
    MemoryAdaptation adaptation = 4;
}

// UpdateMemoryRequest 更新记忆请求
message UpdateMemoryRequest {
    string memory_id = 1;
    string content = 2;
    string summary = 3;
    float confidence = 4;
    float importance = 5;
    int64 ttl = 6;
    map<string, string> metadata = 7;
    repeated float embedding_vector = 8;
}

// UpdateMemoryResponse 更新记忆响应
message UpdateMemoryResponse {
    common.Status status = 1;
    uint64 new_version = 2;
}

// DeleteMemoryRequest 删除记忆请求
message DeleteMemoryRequest {
    string memory_id = 1;
    bool hard_delete = 2;  // false表示软删除
}

// DeleteMemoryResponse 删除记忆响应
message DeleteMemoryResponse {
    common.Status status = 1;
}

// MemoryFilter 记忆过滤条件
message MemoryFilter {
    repeated string agent_ids = 1;
    repeated string session_ids = 2;
    repeated MemoryType memory_types = 3;
    repeated MemoryLevel levels = 4;
    repeated MemoryState states = 5;
    string scope = 6;
    uint64 start_time = 7;
    uint64 end_time = 8;
    float min_confidence = 9;
    float min_importance = 10;
    map<string, string> metadata_filter = 11;
}

// QueryMemoriesRequest 查询记忆请求
message QueryMemoriesRequest {
    MemoryFilter filter = 1;
    int64 limit = 2;
    int64 offset = 3;
    string order_by = 4;
    bool ascending = 5;
}

// QueryMemoriesResponse 查询记忆响应
message QueryMemoriesResponse {
    common.Status status = 1;
    repeated Memory memories = 2;
    int64 total_count = 3;
}

// SearchMemoriesRequest 向量搜索记忆请求
message SearchMemoriesRequest {
    string tenant_id = 1;
    string workspace_id = 2;
    repeated float query_vector = 3;
    int64 top_k = 4;
    MemoryFilter filter = 5;
    float min_score = 6;
}

// SearchResult 搜索结果
message SearchResult {
    Memory memory = 1;
    float score = 2;
    string explanation = 3;
}

// SearchMemoriesResponse 向量搜索记忆响应
message SearchMemoriesResponse {
    common.Status status = 1;
    repeated SearchResult results = 2;
}

// GetRelationsRequest 获取关系请求
message GetRelationsRequest {
    string object_id = 1;
    string object_type = 2;
    string relation_type = 3;
    int64 hop = 4;
}

// GetRelationsResponse 获取关系响应
message GetRelationsResponse {
    common.Status status = 1;
    repeated Relation relations = 2;
}

// CreateRelationRequest 创建关系请求
message CreateRelationRequest {
    string src_object_id = 1;
    string src_type = 2;
    string dst_object_id = 3;
    string dst_type = 4;
    string relation_type = 5;
    float weight = 6;
    map<string, string> properties = 7;
    string created_by_event_id = 8;
}

// CreateRelationResponse 创建关系响应
message CreateRelationResponse {
    common.Status status = 1;
    string edge_id = 2;
}

// MemoryService 记忆服务定义
service MemoryService {
    // 记忆CRUD
    rpc CreateMemory(CreateMemoryRequest) returns (CreateMemoryResponse);
    rpc GetMemory(GetMemoryRequest) returns (GetMemoryResponse);
    rpc UpdateMemory(UpdateMemoryRequest) returns (UpdateMemoryResponse);
    rpc DeleteMemory(DeleteMemoryRequest) returns (DeleteMemoryResponse);
    rpc QueryMemories(QueryMemoriesRequest) returns (QueryMemoriesResponse);
    
    // 向量搜索
    rpc SearchMemories(SearchMemoriesRequest) returns (SearchMemoriesResponse);
    
    // 关系操作
    rpc GetRelations(GetRelationsRequest) returns (GetRelationsResponse);
    rpc CreateRelation(CreateRelationRequest) returns (CreateRelationResponse);
}
```

***

### 1.4 governance.proto

```protobuf
syntax = "proto3";
package governance.proto;
option go_package="github.com/milvus-io/milvus/pkg/v2/proto/governancepb";

import "common.proto";

// ConsistencyLevel 一致性级别
enum ConsistencyLevel {
    CONSISTENCY_UNKNOWN = 0;
    STRONG = 1;
    BOUNDED_STALENESS = 2;
    SESSION = 3;
    EVENTUAL = 4;
}

// MergePolicy 合并策略
enum MergePolicy {
    MERGE_UNKNOWN = 0;
    LAST_WRITER_WINS = 1;
    CAUSAL_MERGE = 2;
    WEIGHTED_MERGE = 3;
    CRDT_MERGE = 4;
}

// ShareContract 共享契约
message ShareContract {
    string contract_id = 1;
    string scope = 2;
    string owner_agent_id = 3;
    
    // ACL
    repeated string read_acl = 4;
    repeated string write_acl = 5;
    repeated string derive_acl = 6;
    
    // TTL策略
    int64 ttl_policy = 7;
    
    // 一致性级别
    ConsistencyLevel consistency_level = 8;
    
    // 合并策略
    MergePolicy merge_policy = 9;
    
    // 隔离策略
    bool quarantine_enabled = 10;
    string quarantine_policy = 11;
    
    // 审计策略
    bool audit_enabled = 12;
    string audit_policy = 13;
    
    // 元数据
    map<string, string> metadata = 14;
    
    // 时间戳
    uint64 created_ts = 15;
    uint64 updated_ts = 16;
    string created_by_event_id = 17;
}

// Conflict 冲突信息
message Conflict {
    string conflict_id = 1;
    string conflict_type = 2;  // "fact", "plan", "state"
    string object_id = 3;
    string object_type = 4;
    repeated string conflicting_versions = 5;
    repeated string conflicting_agents = 6;
    string description = 7;
    uint64 detected_at = 8;
    ConflictState state = 9;
}

// ConflictState 冲突状态
enum ConflictState {
    CONFLICT_UNKNOWN = 0;
    CONFLICT_DETECTED = 1;
    CONFLICT_RESOLVING = 2;
    CONFLICT_RESOLVED = 3;
    CONFLICT_ESCALATED = 4;
}

// ConflictResolution 冲突解决结果
message ConflictResolution {
    string conflict_id = 1;
    string resolution_strategy = 2;
    string resolved_version = 3;
    map<string, string> merge_details = 4;
    string resolved_by = 5;
    uint64 resolved_at = 6;
}

// AuditLog 审计日志
message AuditLog {
    string log_id = 1;
    string tenant_id = 2;
    string workspace_id = 3;
    string agent_id = 4;
    string session_id = 5;
    string operation = 6;
    string object_type = 7;
    string object_id = 8;
    string action = 9;  // "create", "read", "update", "delete", "share"
    string details = 10;
    uint64 timestamp = 11;
    string source_ip = 12;
    string user_identity = 13;
    bool success = 14;
    string error_message = 15;
}

// CreateShareContractRequest 创建共享契约请求
message CreateShareContractRequest {
    string tenant_id = 1;
    string workspace_id = 2;
    string scope = 3;
    string owner_agent_id = 4;
    repeated string read_acl = 5;
    repeated string write_acl = 6;
    repeated string derive_acl = 7;
    int64 ttl_policy = 8;
    ConsistencyLevel consistency_level = 9;
    MergePolicy merge_policy = 10;
    bool quarantine_enabled = 11;
    string quarantine_policy = 12;
    bool audit_enabled = 13;
    string audit_policy = 14;
    map<string, string> metadata = 15;
}

// CreateShareContractResponse 创建共享契约响应
message CreateShareContractResponse {
    common.Status status = 1;
    string contract_id = 2;
}

// GetShareContractRequest 获取共享契约请求
message GetShareContractRequest {
    string contract_id = 1;
}

// GetShareContractResponse 获取共享契约响应
message GetShareContractResponse {
    common.Status status = 1;
    ShareContract contract = 2;
}

// UpdateShareContractRequest 更新共享契约请求
message UpdateShareContractRequest {
    string contract_id = 1;
    repeated string read_acl = 2;
    repeated string write_acl = 3;
    repeated string derive_acl = 4;
    int64 ttl_policy = 5;
    ConsistencyLevel consistency_level = 6;
    MergePolicy merge_policy = 7;
}

// UpdateShareContractResponse 更新共享契约响应
message UpdateShareContractResponse {
    common.Status status = 1;
}

// DetectConflictRequest 检测冲突请求
message DetectConflictRequest {
    string object_id = 1;
    string object_type = 2;
}

// DetectConflictResponse 检测冲突响应
message DetectConflictResponse {
    common.Status status = 1;
    repeated Conflict conflicts = 2;
}

// ResolveConflictRequest 解决冲突请求
message ResolveConflictRequest {
    string conflict_id = 1;
    string resolution_strategy = 2;
    string preferred_version = 3;
    map<string, float> version_weights = 4;
}

// ResolveConflictResponse 解决冲突响应
message ResolveConflictResponse {
    common.Status status = 1;
    ConflictResolution resolution = 2;
}

// QueryAuditLogsRequest 查询审计日志请求
message QueryAuditLogsRequest {
    string tenant_id = 1;
    string workspace_id = 2;
    string agent_id = 3;
    string object_type = 4;
    string object_id = 5;
    string action = 6;
    uint64 start_time = 7;
    uint64 end_time = 8;
    int64 limit = 9;
    int64 offset = 10;
}

// QueryAuditLogsResponse 查询审计日志响应
message QueryAuditLogsResponse {
    common.Status status = 1;
    repeated AuditLog logs = 2;
    int64 total_count = 3;
}

// CheckAccessRequest 检查访问权限请求
message CheckAccessRequest {
    string agent_id = 1;
    string object_id = 2;
    string object_type = 3;
    string action = 4;  // "read", "write", "derive"
}

// CheckAccessResponse 检查访问权限响应
message CheckAccessResponse {
    common.Status status = 1;
    bool allowed = 2;
    string reason = 3;
}

// GovernanceService 治理服务定义
service GovernanceService {
    // 共享契约管理
    rpc CreateShareContract(CreateShareContractRequest) returns (CreateShareContractResponse);
    rpc GetShareContract(GetShareContractRequest) returns (GetShareContractResponse);
    rpc UpdateShareContract(UpdateShareContractRequest) returns (UpdateShareContractResponse);
    
    // 冲突管理
    rpc DetectConflict(DetectConflictRequest) returns (DetectConflictResponse);
    rpc ResolveConflict(ResolveConflictRequest) returns (ResolveConflictResponse);
    
    // 审计日志
    rpc QueryAuditLogs(QueryAuditLogsRequest) returns (QueryAuditLogsResponse);
    
    // 权限检查
    rpc CheckAccess(CheckAccessRequest) returns (CheckAccessResponse);
}
```

***

### 1.5 agent\_meta.proto

```protobuf
syntax = "proto3";
package milvus.proto.etcd;
option go_package="github.com/milvus-io/milvus/pkg/v2/proto/etcdpb";

import "common.proto";

// AgentState 定义Agent的状态
enum AgentState {
    AGENT_UNKNOWN = 0;
    AGENT_CREATING = 1;
    AGENT_ACTIVE = 2;
    AGENT_PAUSED = 3;
    AGENT_TERMINATED = 4;
}

// AgentInfo 存储在etcd中的Agent信息
message AgentInfo {
    string agent_id = 1;
    string tenant_id = 2;
    string workspace_id = 3;
    string agent_type = 4;
    string role_profile = 5;
    string policy_ref = 6;
    repeated string capability_set = 7;
    string default_memory_policy = 8;
    uint64 created_ts = 9;
    uint64 updated_ts = 10;
    AgentState state = 11;
    map<string, string> metadata = 12;
}

// SessionState 定义Session的状态
enum SessionState {
    SESSION_UNKNOWN = 0;
    SESSION_CREATING = 1;
    SESSION_ACTIVE = 2;
    SESSION_PAUSED = 3;
    SESSION_COMPLETED = 4;
    SESSION_FAILED = 5;
}

// SessionInfo 存储在etcd中的Session信息
message SessionInfo {
    string session_id = 1;
    string agent_id = 2;
    string parent_session_id = 3;
    string task_type = 4;
    string goal = 5;
    string context_ref = 6;
    uint64 start_ts = 7;
    uint64 end_ts = 8;
    SessionState state = 9;
    int64 budget_token = 10;
    int64 budget_time_ms = 11;
    map<string, string> metadata = 12;
}

// MemoryType 记忆类型
enum MemoryType {
    MEMORY_UNKNOWN = 0;
    EPISODIC = 1;
    SEMANTIC = 2;
    PROCEDURAL = 3;
    SOCIAL = 4;
    REFLECTIVE = 5;
}

// MemoryLevel 记忆蒸馏层级
enum MemoryLevel {
    LEVEL_UNKNOWN = 0;
    LEVEL_RAW = 1;
    LEVEL_SUMMARY = 2;
    LEVEL_PATTERN = 3;
}

// MemoryState 记忆状态
enum MemoryState {
    MEMORY_STATE_UNKNOWN = 0;
    MEMORY_ACTIVE = 1;
    MEMORY_FADING = 2;
    MEMORY_ARCHIVED = 3;
    MEMORY_QUARANTINED = 4;
    MEMORY_DELETED = 5;
}

// MemoryInfo 存储在etcd中的Memory元数据（不包含实际内容）
message MemoryInfo {
    string memory_id = 1;
    MemoryType memory_type = 2;
    string agent_id = 3;
    string session_id = 4;
    string scope = 5;
    MemoryLevel level = 6;
    string summary = 7;
    repeated string source_event_ids = 8;
    float confidence = 9;
    float importance = 10;
    float freshness_score = 11;
    int64 ttl = 12;
    uint64 valid_from = 13;
    uint64 valid_to = 14;
    string provenance_ref = 15;
    int64 version = 16;
    bool is_active = 17;
    MemoryState state = 18;
    uint64 created_ts = 19;
    uint64 updated_ts = 20;
    map<string, string> metadata = 21;
    string embedding_ref = 22;
    int64 collection_id = 23;  // 存储实际向量的Milvus collection
    int64 partition_id = 24;   // 存储实际向量的Milvus partition
}

// MemoryPolicyInfo 存储在etcd中的Memory策略
message MemoryPolicyInfo {
    string memory_id = 1;
    float salience_weight = 2;
    int64 ttl = 3;
    string decay_fn = 4;
    float confidence = 5;
    bool verified = 6;
    string verified_by = 7;
    uint64 verified_at = 8;
    bool quarantined = 9;
    string quarantine_reason = 10;
    string visibility_policy = 11;
    repeated string read_acl = 12;
    repeated string write_acl = 13;
    repeated string derive_acl = 14;
    string policy_reason = 15;
    string policy_source = 16;
    string policy_event_id = 17;
    uint64 created_ts = 18;
    uint64 updated_ts = 19;
}

// StateInfo 存储在etcd中的State信息
message StateInfo {
    string state_id = 1;
    string agent_id = 2;
    string session_id = 3;
    string state_type = 4;
    string state_key = 5;
    int64 version = 6;
    string derived_from_event_id = 7;
    uint64 checkpoint_ts = 8;
    uint64 created_ts = 9;
    uint64 updated_ts = 10;
    map<string, string> metadata = 11;
    string storage_ref = 12;  // 实际状态值的存储位置
}

// ArtifactInfo 存储在etcd中的Artifact元数据
message ArtifactInfo {
    string artifact_id = 1;
    string session_id = 2;
    string owner_agent_id = 3;
    string artifact_type = 4;
    string uri = 5;
    string content_ref = 6;
    string mime_type = 7;
    map<string, string> metadata = 8;
    string hash = 9;
    string produced_by_event_id = 10;
    int64 version = 11;
    uint64 created_ts = 12;
}

// RelationInfo 存储在etcd中的Relation信息
message RelationInfo {
    string edge_id = 1;
    string src_object_id = 2;
    string src_type = 3;
    string dst_object_id = 4;
    string dst_type = 5;
    string relation_type = 6;
    float weight = 7;
    map<string, string> properties = 8;
    uint64 created_ts = 9;
    string created_by_event_id = 10;
}

// ShareContractInfo 存储在etcd中的ShareContract
message ShareContractInfo {
    string contract_id = 1;
    string scope = 2;
    string owner_agent_id = 3;
    repeated string read_acl = 4;
    repeated string write_acl = 5;
    repeated string derive_acl = 6;
    int64 ttl_policy = 7;
    int32 consistency_level = 8;  // ConsistencyLevel
    int32 merge_policy = 9;       // MergePolicy
    bool quarantine_enabled = 10;
    string quarantine_policy = 11;
    bool audit_enabled = 12;
    string audit_policy = 13;
    map<string, string> metadata = 14;
    uint64 created_ts = 15;
    uint64 updated_ts = 16;
    string created_by_event_id = 17;
}

// EventLogInfo 存储在etcd中的Event Log元数据
message EventLogInfo {
    string channel_name = 1;
    uint64 start_log_id = 2;
    uint64 end_log_id = 3;
    uint64 start_timestamp = 4;
    uint64 end_timestamp = 5;
    string storage_path = 6;
    int64 segment_count = 7;
}

// MemoryIndexInfo 存储在etcd中的Memory索引信息
message MemoryIndexInfo {
    int64 collection_id = 1;
    int64 partition_id = 2;
    int64 field_id = 3;
    int64 index_id = 4;
    int64 build_id = 5;
    bool enable_index = 6;
    uint64 create_time = 7;
    string index_type = 8;
    map<string, string> index_params = 9;
}

// AgentWorkspace 工作空间信息
message AgentWorkspace {
    string workspace_id = 1;
    string tenant_id = 2;
    string name = 3;
    string description = 4;
    repeated string agent_ids = 5;
    map<string, string> metadata = 6;
    uint64 created_ts = 7;
    uint64 updated_ts = 8;
}
```

***

### 1.6 common.proto

```protobuf
syntax = "proto3";
package common.proto;
option go_package="github.com/milvus-io/milvus/pkg/v2/proto/commonpb";

// ErrorCode 错误码
enum ErrorCode {
    Success = 0;
    UnexpectedError = 1;
    ConnectFailed = 2;
    PermissionDenied = 3;
    CollectionNotFound = 4;
    IllegalArgument = 5;
    IllegalDimension = 7;
    IllegalIndexType = 8;
    IllegalCollectionName = 9;
    IllegalTOPK = 10;
    IllegalRowRecord = 11;
    IllegalVectorID = 12;
    IllegalSearchResult = 13;
    FileNotFound = 14;
    MetaFailed = 15;
    CacheFailed = 16;
    CannotCreateFolder = 17;
    CannotCreateFile = 18;
    CannotDeleteFolder = 19;
    CannotDeleteFile = 20;
    BuildIndexError = 21;
    IllegalNLIST = 22;
    IllegalMetricType = 23;
    OutOfMemory = 24;
    IndexNotExist = 25;
    EmptyCollection = 26;
    UpdateImportTaskFailure = 27;
    CollectionNameNotFound = 28;
    CreateCredentialFailure = 29;
    UpdateCredentialFailure = 30;
    DeleteCredentialFailure = 31;
    GetCredentialFailure = 32;
    ListCredUsersFailure = 33;
    GetUserFailure = 34;
    CreateRoleFailure = 35;
    DropRoleFailure = 36;
    OperateUserRoleFailure = 37;
    SelectRoleFailure = 38;
    SelectUserFailure = 39;
    SelectResourceFailure = 40;
    OperatePrivilegeFailure = 41;
    SelectGrantFailure = 42;
    RefreshPolicyInfoCacheFailure = 43;
    ListPolicyFailure = 44;
    NotShardLeader = 45;
    NoReplicaAvailable = 46;
    SegmentNotFound = 47;
    ForceDeny = 48;
    RateLimit = 49;
    NodeIDNotMatch = 50;
    UpsertAutoIDTrue = 51;
    InsufficientMemoryToLoad = 52;
    MemoryQuotaExhausted = 53;
    DiskQuotaExhausted = 54;
    TimeTickLongDelay = 55;
    NotFoundTSafer = 56;
    DenyToRewrite = 57;
    InvalidPassword = 58;
    DDRequestRace = 59;

    // internal error code for DC
    DCNotReady = 100;

    // internal error code for DN
    DNNotReady = 101;

    // internal error code for CGO
    CGOError = 102;
}

// Status 状态信息
message Status {
    ErrorCode error_code = 1;
    string reason = 2;
    int32 code = 3;
}

// KeyValuePair 键值对
message KeyValuePair {
    string key = 1;
    string value = 2;
}

// KeyDataPair 键数据对
message KeyDataPair {
    string key = 1;
    bytes data = 2;
}

// Blob 二进制数据
message Blob {
    bytes value = 1;
}

// PlaceholderValue 占位符值
message PlaceholderValue {
    string tag = 1;
    bytes value = 2;
}

// PlaceholderGroup 占位符组
message PlaceholderGroup {
    repeated PlaceholderValue placeholders = 1;
}

// Address 地址信息
message Address {
    string ip = 1;
    int64 port = 2;
}

// MsgBase 消息基础信息
message MsgBase {
    int64 msg_type = 1;
    int64 msgID = 2;
    uint64 timestamp = 3;
    int64 sourceID = 4;
}

// MsgHeader 消息头
message MsgHeader {
    MsgBase base = 1;
}

// DMLMsgHeader DML消息头
message DMLMsgHeader {
    MsgBase base = 1;
    string shardName = 2;
}

// ConsistencyLevel 一致性级别
enum ConsistencyLevel {
    Strong = 0;
    Session = 1;
    Bounded = 2;
    Eventually = 3;
    Customized = 4;
}

// CompactionState 压缩状态
enum CompactionState {
    UndefiedState = 0;
    Executing = 1;
    Completed = 2;
}

// SegmentState Segment状态
enum SegmentState {
    SegmentStateNone = 0;
    NotExist = 1;
    Growing = 2;
    Sealed = 3;
    Flushed = 4;
    Flushing = 5;
    Dropped = 6;
    Importing = 7;
}

// ImportState 导入状态
enum ImportState {
    ImportPending = 0;
    ImportFailed = 1;
    ImportStarted = 2;
    ImportPersisted = 5;
    ImportCompleted = 6;
    ImportFailedAndCleaned = 7;
}

// ObjectType 对象类型
enum ObjectType {
    Collection = 0;
    Global = 1;
    User = 2;
    ObjectTypeCount = 3;
}

// ObjectPrivilege 对象权限
enum ObjectPrivilege {
    PrivilegeAll = 0;
    PrivilegeCreate = 1;
    PrivilegeDrop = 2;
    PrivilegeAlter = 3;
    PrivilegeRead = 4;
    PrivilegeLoad = 5;
    PrivilegeRelease = 6;
    PrivilegeCompact = 7;
    PrivilegeInsert = 8;
    PrivilegeDelete = 9;
    PrivilegeGetStatistics = 10;
    PrivilegeCreateIndex = 11;
    PrivilegeIndexDetail = 12;
    PrivilegeDropIndex = 13;
    PrivilegeSearch = 14;
    PrivilegeFlush = 15;
    PrivilegeQuery = 16;
    PrivilegeLoadBalance = 17;
    PrivilegeImport = 18;
    PrivilegeBackup = 19;
    PrivilegeRestore = 20;
    PrivilegeListIndexes = 21;
    PrivilegeCreateDatabase = 22;
    PrivilegeDropDatabase = 23;
    PrivilegeListDatabases = 24;
    PrivilegeDescribeDatabase = 25;
    PrivilegeAlterDatabase = 26;
    PrivilegeDescribeCollection = 27;
    PrivilegeCount = 28;
}

// StateCode 状态码
enum StateCode {
    Initializing = 0;
    Healthy = 1;
    Abnormal = 2;
    StandBy = 3;
    Stopping = 4;
}

// LoadState 加载状态
enum LoadState {
    LoadStateNotExist = 0;
    LoadStateNotLoad = 1;
    LoadStateLoading = 2;
    LoadStateLoaded = 3;
    LoadStateUnloading = 4;
}
```

***

## 2. etcd元数据常量定义

### 2.1 agentcoord/constant.go

见internal\metastore\kv\agentcoord\constant.go

```go
package agentcoord

import "fmt"

const (
	// ComponentPrefix prefix for agentcoord component
	ComponentPrefix = "agent-coord"

	// AgentMetaPrefix prefix for agent meta
	AgentMetaPrefix = ComponentPrefix + "/agent"

	// SessionMetaPrefix prefix for session meta
	SessionMetaPrefix = ComponentPrefix + "/session"

	// WorkspaceMetaPrefix prefix for workspace meta
	WorkspaceMetaPrefix = ComponentPrefix + "/workspace"

	// SnapshotPrefix prefix for snapshots
	SnapshotPrefix = ComponentPrefix + "/snapshots"
)

// BuildAgentKey builds agent key
func BuildAgentKey(agentID string) string {
	return fmt.Sprintf("%s/%s", AgentMetaPrefix, agentID)
}

// BuildAgentPrefix builds agent prefix for listing
func BuildAgentPrefix(tenantID, workspaceID string) string {
	if tenantID != "" && workspaceID != "" {
		return fmt.Sprintf("%s/%s/%s", AgentMetaPrefix, tenantID, workspaceID)
	}
	if tenantID != "" {
		return fmt.Sprintf("%s/%s", AgentMetaPrefix, tenantID)
	}
	return AgentMetaPrefix
}

// BuildSessionKey builds session key
func BuildSessionKey(sessionID string) string {
	return fmt.Sprintf("%s/%s", SessionMetaPrefix, sessionID)
}

// BuildSessionPrefix builds session prefix for listing
func BuildSessionPrefix(agentID string) string {
	if agentID != "" {
		return fmt.Sprintf("%s/%s", SessionMetaPrefix, agentID)
	}
	return SessionMetaPrefix
}

// BuildWorkspaceKey builds workspace key
func BuildWorkspaceKey(workspaceID string) string {
	return fmt.Sprintf("%s/%s", WorkspaceMetaPrefix, workspaceID)
}

// BuildWorkspacePrefix builds workspace prefix for listing
func BuildWorkspacePrefix(tenantID string) string {
	if tenantID != "" {
		return fmt.Sprintf("%s/%s", WorkspaceMetaPrefix, tenantID)
	}
	return WorkspaceMetaPrefix
}

// BuildSnapshotKey builds snapshot key
func BuildSnapshotKey(objectType, objectID string, version int64) string {
	return fmt.Sprintf("%s/%s/%s/%d", SnapshotPrefix, objectType, objectID, version)
}
```

***

### 2.2 memorycoord/constant.go

见internal\metastore\kv\memorycoord

```go
package memorycoord

import "fmt"

const (
	// ComponentPrefix prefix for memorycoord component
	ComponentPrefix = "memory-coord"

	// MemoryMetaPrefix prefix for memory meta
	MemoryMetaPrefix = ComponentPrefix + "/memory"

	// MemoryPolicyPrefix prefix for memory policy
	MemoryPolicyPrefix = ComponentPrefix + "/memory-policy"

	// MemoryAdaptationPrefix prefix for memory adaptation
	MemoryAdaptationPrefix = ComponentPrefix + "/memory-adaptation"

	// StateMetaPrefix prefix for state meta
	StateMetaPrefix = ComponentPrefix + "/state"

	// ArtifactMetaPrefix prefix for artifact meta
	ArtifactMetaPrefix = ComponentPrefix + "/artifact"

	// RelationMetaPrefix prefix for relation meta
	RelationMetaPrefix = ComponentPrefix + "/relation"

	// ShareContractPrefix prefix for share contract
	ShareContractPrefix = ComponentPrefix + "/share-contract"

	// MemoryIndexPrefix prefix for memory index
	MemoryIndexPrefix = ComponentPrefix + "/index"

	// SnapshotPrefix prefix for snapshots
	SnapshotPrefix = ComponentPrefix + "/snapshots"
)

// BuildMemoryKey builds memory key
func BuildMemoryKey(memoryID string) string {
	return fmt.Sprintf("%s/%s", MemoryMetaPrefix, memoryID)
}

// BuildMemoryPrefix builds memory prefix for listing
func BuildMemoryPrefix(agentID, sessionID string) string {
	if agentID != "" && sessionID != "" {
		return fmt.Sprintf("%s/%s/%s", MemoryMetaPrefix, agentID, sessionID)
	}
	if agentID != "" {
		return fmt.Sprintf("%s/%s", MemoryMetaPrefix, agentID)
	}
	return MemoryMetaPrefix
}

// BuildMemoryPolicyKey builds memory policy key
func BuildMemoryPolicyKey(memoryID string) string {
	return fmt.Sprintf("%s/%s", MemoryPolicyPrefix, memoryID)
}

// BuildMemoryAdaptationKey builds memory adaptation key
func BuildMemoryAdaptationKey(memoryID string) string {
	return fmt.Sprintf("%s/%s", MemoryAdaptationPrefix, memoryID)
}

// BuildStateKey builds state key
func BuildStateKey(stateID string) string {
	return fmt.Sprintf("%s/%s", StateMetaPrefix, stateID)
}

// BuildStatePrefix builds state prefix for listing
func BuildStatePrefix(agentID, sessionID string) string {
	if agentID != "" && sessionID != "" {
		return fmt.Sprintf("%s/%s/%s", StateMetaPrefix, agentID, sessionID)
	}
	if agentID != "" {
		return fmt.Sprintf("%s/%s", StateMetaPrefix, agentID)
	}
	return StateMetaPrefix
}

// BuildArtifactKey builds artifact key
func BuildArtifactKey(artifactID string) string {
	return fmt.Sprintf("%s/%s", ArtifactMetaPrefix, artifactID)
}

// BuildArtifactPrefix builds artifact prefix for listing
func BuildArtifactPrefix(sessionID string) string {
	if sessionID != "" {
		return fmt.Sprintf("%s/%s", ArtifactMetaPrefix, sessionID)
	}
	return ArtifactMetaPrefix
}

// BuildRelationKey builds relation key
func BuildRelationKey(edgeID string) string {
	return fmt.Sprintf("%s/%s", RelationMetaPrefix, edgeID)
}

// BuildRelationPrefix builds relation prefix for listing
func BuildRelationPrefix(objectID string) string {
	if objectID != "" {
		return fmt.Sprintf("%s/%s", RelationMetaPrefix, objectID)
	}
	return RelationMetaPrefix
}

// BuildShareContractKey builds share contract key
func BuildShareContractKey(contractID string) string {
	return fmt.Sprintf("%s/%s", ShareContractPrefix, contractID)
}

// BuildShareContractPrefix builds share contract prefix for listing
func BuildShareContractPrefix(scope string) string {
	if scope != "" {
		return fmt.Sprintf("%s/%s", ShareContractPrefix, scope)
	}
	return ShareContractPrefix
}

// BuildMemoryIndexKey builds memory index key
func BuildMemoryIndexKey(collectionID, indexID int64) string {
	return fmt.Sprintf("%s/%d/%d", MemoryIndexPrefix, collectionID, indexID)
}

// BuildSnapshotKey builds snapshot key
func BuildSnapshotKey(objectType, objectID string, version int64) string {
	return fmt.Sprintf("%s/%s/%s/%d", SnapshotPrefix, objectType, objectID, version)
}
```

***

### 2.3 event/constant.go

见internal\metastore\kv\event

```go
package event

import "fmt"

const (
	// ComponentPrefix prefix for event component
	ComponentPrefix = "event"

	// EventLogPrefix prefix for event log
	EventLogPrefix = ComponentPrefix + "/log"

	// EventMetaPrefix prefix for event meta
	EventMetaPrefix = ComponentPrefix + "/meta"

	// EventChannelPrefix prefix for event channel
	EventChannelPrefix = ComponentPrefix + "/channel"

	// EventSubscriberPrefix prefix for event subscriber
	EventSubscriberPrefix = ComponentPrefix + "/subscriber"

	// EventPositionPrefix prefix for event position
	EventPositionPrefix = ComponentPrefix + "/position"
)

// BuildEventLogKey builds event log key
func BuildEventLogKey(channelName string, logID uint64) string {
	return fmt.Sprintf("%s/%s/%d", EventLogPrefix, channelName, logID)
}

// BuildEventLogPrefix builds event log prefix
func BuildEventLogPrefix(channelName string) string {
	if channelName != "" {
		return fmt.Sprintf("%s/%s", EventLogPrefix, channelName)
	}
	return EventLogPrefix
}

// BuildEventMetaKey builds event meta key
func BuildEventMetaKey(eventID string) string {
	return fmt.Sprintf("%s/%s", EventMetaPrefix, eventID)
}

// BuildEventChannelKey builds event channel key
func BuildEventChannelKey(channelName string) string {
	return fmt.Sprintf("%s/%s", EventChannelPrefix, channelName)
}

// BuildEventSubscriberKey builds event subscriber key
func BuildEventSubscriberKey(subscriberID string) string {
	return fmt.Sprintf("%s/%s", EventSubscriberPrefix, subscriberID)
}

// BuildEventPositionKey builds event position key
func BuildEventPositionKey(subscriberID, channelName string) string {
	return fmt.Sprintf("%s/%s/%s", EventPositionPrefix, subscriberID, channelName)
}
```

***

## 3. Go模型代码

见pkg\models

### 3.1 models/agent.go

```go
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
```

***

### 3.2 models/event.go

```go
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
	MemoryDeleted        EventType = 33

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
```

***

### 3.3 models/memory.go

```go
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
```

***

### 3.4 models/governance.go

```go
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
```

***

## 4. 目录结构

```
TDB/
├── pkg/
│   ├── proto/
│   │   ├── agent.proto
│   │   ├── event.proto
│   │   ├── memory.proto
│   │   ├── governance.proto
│   │   ├── agent_meta.proto
│   │   └── common.proto
│   └── models/
│       ├── agent.go
│       ├── event.go
│       ├── memory.go
│       └── governance.go
└── internal/
    └── metastore/kv/
        ├── agentcoord/constant.go
        ├── memorycoord/constant.go
        └── event/constant.go
```

***

## 5. 后续计划

### 5.1 下一步建议

1. **实现KV Catalog层** - 基于etcd元数据常量，实现元数据的CRUD操作
2. **实现Agent Coordinator** - 第一个核心组件，验证架构可行性
3. **生成Proto代码** - 在有网络环境时，运行`make generated-proto`生成正式的Go代码

### 5.2 文件位置汇总

| 文件类型    | 路径                                    |
| ------- | ------------------------------------- |
| Proto定义 | `pkg/proto/*.proto`                   |
| Go模型    | `pkg/models/*.go`                     |
| etcd常量  | `internal/metastore/kv/*/constant.go` |
| 本文档     | `docs/TDB/fixed1.01.md`               |

***

*文档版本: 1.01*\
*最后更新: 2026-03-14*
