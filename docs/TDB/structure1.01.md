# TDB 系统结构文档

> 本文档详细描述 TDB (Temporal Database) 系统的整体架构和组件关系
> 创建日期: 2026-03-15
> 版本: 1.01

***

## 一、整体架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           TDB 系统架构                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                        客户端层                                   │    │
│  │         gRPC / REST API / CLI / SDK                              │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                    │                                     │
│  ┌─────────────────────────────────▼─────────────────────────────────┐   │
│  │                      gRPC Server 层                               │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐            │   │
│  │  │ AgentServer  │  │ MemoryServer │  │ EventServer  │            │   │
│  │  │  (Agent服务)  │  │  (记忆服务)   │  │  (事件服务)   │            │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘            │   │
│  └───────────────────────────────────────────────────────────────────┘   │
│                                    │                                     │
│  ┌─────────────────────────────────▼─────────────────────────────────┐   │
│  │                     Coordinator 层                                │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐   │   │
│  │  │ AgentCoord  │  │ MemoryCoord │  │ EventCoord              │   │   │
│  │  │             │  │             │  │                         │   │   │
│  │  │ • Agent管理  │  │ • 记忆CRUD   │  │ • 事件日志              │   │   │
│  │  │ • Session   │  │ • 向量搜索   │  │ • 事件订阅              │   │   │
│  │  │ • Workspace │  │ • 关系管理   │  │ • 位置追踪              │   │   │
│  │  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘   │   │
│  │         │                │                     │                  │   │
│  │         └────────────────┼─────────────────────┘                  │   │
│  │                          │                                        │   │
│  │  ┌───────────────────────▼────────────────────────┐              │   │
│  │  │              Catalog (元数据目录)               │              │   │
│  │  │  ┌──────────┐ ┌──────────┐ ┌──────────┐       │              │   │
│  │  │  │Agent     │ │Memory    │ │Event     │       │              │   │
│  │  │  │Catalog   │ │Catalog   │ │Catalog   │       │              │   │
│  │  │  └──────────┘ └──────────┘ └──────────┘       │              │   │
│  │  └────────────────────────────────────────────────┘              │   │
│  └───────────────────────────────────────────────────────────────────┘   │
│                                    │                                     │
│  ┌─────────────────────────────────▼─────────────────────────────────┐   │
│  │                       存储层                                      │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐  │   │
│  │  │    etcd      │  │    TiKV      │  │       Milvus           │  │   │
│  │  │  (元数据/KV)  │  │  (元数据/KV)  │  │   (向量存储/检索)       │  │   │
│  │  └──────────────┘  └──────────────┘  └────────────────────────┘  │   │
│  └───────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

***

## 二、目录结构

```
f:\学习资料\我的论文\VDB\前置研究\code\TDB\
├── cmd/                                    # 入口点和组件启动
│   ├── components/                         # 组件定义
│   │   └── tdb.go                         # TDB组件实现
│   ├── milvus/                            # Milvus主程序
│   │   └── util.go                        # 工具函数
│   └── roles/                             # 角色管理
│       └── roles.go                       # MilvusRoles实现
│
├── internal/                               # 内部实现
│   ├── agentcoord/                        # Agent协调器
│   │   ├── agent_coord.go                 # AgentCoord实现
│   │   └── agent_coord_test.go            # 单元测试
│   ├── memorycoord/                       # 记忆协调器
│   │   ├── memory_coord.go                # MemoryCoord实现
│   │   └── memory_coord_test.go           # 单元测试
│   ├── eventcoord/                        # 事件协调器
│   │   ├── event_coord.go                 # EventCoord实现
│   │   └── event_coord_test.go            # 单元测试
│   ├── distributed/                       # 分布式服务
│   │   └── tdb/                          # TDB gRPC服务
│   │       ├── server.go                  # Server主结构
│   │       ├── agent_server.go            # AgentServer实现
│   │       ├── memory_server.go           # MemoryServer实现
│   │       ├── event_server.go            # EventServer实现
│   │       └── agent_server_test.go       # 单元测试
│   └── metastore/                         # 元数据存储
│       └── kv/                           # KV存储实现
│           ├── agentcoord/               # Agent元数据存储
│           │   └── kv_catalog.go         # AgentCatalog实现
│           ├── memorycoord/              # 记忆元数据存储
│           │   └── kv_catalog.go         # MemoryCatalog实现
│           └── event/                    # 事件元数据存储
│               └── kv_catalog.go         # EventCatalog实现
│
├── pkg/                                    # 公共包
│   ├── models/                            # 领域模型
│   │   ├── agent.go                       # Agent模型
│   │   ├── session.go                     # Session模型
│   │   ├── memory.go                      # Memory模型
│   │   ├── event.go                       # Event模型
│   │   ├── state.go                       # State模型
│   │   ├── artifact.go                    # Artifact模型
│   │   ├── relation.go                    # Relation模型
│   │   ├── policy.go                      # Policy模型
│   │   └── share.go                       # ShareContract模型
│   ├── proto/                             # Protocol Buffers
│   │   ├── agent.proto                    # Agent服务定义
│   │   ├── memory.proto                   # 记忆服务定义
│   │   ├── event.proto                    # 事件服务定义
│   │   ├── common.proto                   # 通用类型定义
│   │   ├── agentpb/                      # Agent protobuf生成代码
│   │   ├── memorypb/                     # 记忆protobuf生成代码
│   │   ├── eventpb/                      # 事件protobuf生成代码
│   │   └── commonpb/                     # 通用protobuf生成代码
│   ├── constants/                         # 常量定义
│   │   └── etcd.go                        # etcd key常量
│   └── util/                              # 工具函数
│       ├── paramtable/                    # 参数表
│       │   └── component_param.go         # 组件参数
│       └── typeutil/                      # 类型工具
│           └── type.go                    # 角色类型定义
│
├── tests/                                  # 测试
│   └── integration/                       # 集成测试
│       └── tdb_integration_test.go        # TDB集成测试
│
└── docs/                                   # 文档
    └── TDB/                               # TDB项目文档
        ├── plan1.0.md                     # 项目规划
        ├── fixed1.01.md                   # Schema设计
        ├── fixed1.02.md                   # KV Catalog设计
        ├── fixed1.03.md                   # Coordinator设计
        ├── fixed1.04.md                   # Protobuf代码生成
        ├── fixed1.05.md                   # gRPC Server设计
        ├── fixed1.06.md                   # 服务集成设计
        ├── fixed1.07.md                   # 测试与优化
        └── structure1.01.md               # 本文件
```

***

## 三、核心组件详解

### 3.1 gRPC Server 层

**位置**: `internal/distributed/tdb/`

| 组件               | 文件                 | 职责                        |
| ---------------- | ------------------ | ------------------------- |
| **Server**       | `server.go`        | 整合所有服务，管理生命周期，gRPC服务器配置   |
| **AgentServer**  | `agent_server.go`  | 处理Agent/Session相关gRPC请求   |
| **MemoryServer** | `memory_server.go` | 处理Memory/Relation相关gRPC请求 |
| **EventServer**  | `event_server.go`  | 处理Event日志和订阅的gRPC请求       |

**Server 生命周期**:

```
NewServer() → Prepare() → Run() → Stop()
                ↓
         Init() → Start()
```

### 3.2 Coordinator 层

#### 3.2.1 AgentCoord (`internal/agentcoord/`)

**核心功能**:

```go
Agent 管理:
├── CreateAgent(tenantID, workspaceID, agentType, roleProfile, capabilitySet, defaultMemoryPolicy, metadata)
├── GetAgent(agentID)
├── ListAgents(tenantID, workspaceID)
├── UpdateAgent(agentID, updates)
└── DeleteAgent(agentID)

Session 管理:
├── CreateSession(agentID, parentSessionID, taskType, goal, budgetToken, budgetTimeMs, metadata)
├── GetSession(sessionID)
├── ListSessions(agentID)
├── UpdateSession(sessionID, updates)
└── CompleteSession/FailSession(sessionID)

Workspace 管理:
├── CreateWorkspace(tenantID, workspaceName, description, metadata)
├── GetWorkspace(workspaceID)
└── ListWorkspaces(tenantID)
```

#### 3.2.2 MemoryCoord (`internal/memorycoord/`)

**记忆类型**:

| 类型         | 常量                     | 说明            |
| ---------- | ---------------------- | ------------- |
| Episodic   | `MemoryTypeEpisodic`   | 情景记忆：发生过什么    |
| Semantic   | `MemoryTypeSemantic`   | 语义记忆：抽象事实与知识  |
| Procedural | `MemoryTypeProcedural` | 程序记忆：规则、流程、策略 |
| Social     | `MemoryTypeSocial`     | 社交/共享记忆：共享约定  |
| Reflective | `MemoryTypeReflective` | 反思记忆：反思、修正、经验 |

**三层记忆架构**:

```
┌─────────────────────────────────────────┐
│      Adaptation Layer (适配层)           │
│  - RetrievalProfile (检索配置)           │
│  - RankingParams (排序参数)              │
│  - FilteringThresholds (过滤阈值)        │
│  - ProjectionWeights (投影权重)          │
│  - EmbeddingFamily (Embedding族)         │
├─────────────────────────────────────────┤
│       Policy Layer (策略层)             │
│  - SalienceWeight (显著性权重)           │
│  - TTL / DecayFn (衰减函数)              │
│  - Confidence / Verified (置信度/验证)   │
│  - Quarantined (隔离标志)                │
│  - VisibilityPolicy (可见性策略)         │
│  - Read/Write/Derive ACL (访问控制)      │
├─────────────────────────────────────────┤
│        Base Layer (基础层)              │
│  - Content / Summary (内容/摘要)         │
│  - Confidence / Importance (置信度/重要性)│
│  - EmbeddingVector (向量)                │
│  - Level (0=原始, 1=摘要, 2=归纳)        │
│  - Version / Provenance (版本/来源)      │
└─────────────────────────────────────────┘
```

**核心功能**:

```go
记忆管理:
├── CreateMemory(memory)
├── GetMemory(memoryID) → (memory, policy, adaptation)
├── UpdateMemory(memoryID, updates) → newVersion
├── DeleteMemory(memoryID, hardDelete)
├── QueryMemories(filter, limit, offset) → (memories, totalCount)
└── SearchMemories(queryVector, topK, minScore) → results

关系管理:
├── CreateRelation(relation) → edgeID
└── GetRelations(objectID, objectType, relationType, hop) → relations

状态管理:
├── SetState(agentID, sessionID, stateType, stateKey, stateValue)
└── GetState(agentID, sessionID, stateType, stateKey)
```

#### 3.2.3 EventCoord (`internal/eventcoord/`)

**事件类型**:

| 类别  | 事件类型                                                                   |
| --- | ---------------------------------------------------------------------- |
| 消息类 | UserMessage, AssistantMessage                                          |
| 工具类 | ToolCallIssued, ToolResultReturned                                     |
| 检索类 | RetrievalExecuted                                                      |
| 记忆类 | MemoryWriteRequested, MemoryConsolidated, MemoryUpdated, MemoryDeleted |
| 计划类 | PlanUpdated, PlanExecuted                                              |
| 反思类 | CritiqueGenerated, ReflectionCreated                                   |
| 任务类 | TaskStarted, TaskFinished, TaskFailed                                  |
| 协作类 | HandoffOccurred, SharedMemoryAccessed                                  |
| 系统类 | SessionCreated, SessionEnded, AgentRegistered, AgentDeregistered       |

**核心功能**:

```go
事件日志:
├── AppendEvent(event) → (eventID, logicalTs)
├── GetEvent(eventID)
├── QueryEvents(filter, limit, offset) → (events, totalCount)
└── SubscribeEvents(subscriberID, filter, startTimestamp) → EventSubscription

订阅管理:
├── UnsubscribeEvents(subscriberID)
└── EventSubscription (chan *Event)

位置追踪:
├── GetNextLogID(channelName) → logID
└── UpdateSubscriberPosition(subscriberID, channelName, position)
```

### 3.3 Catalog 层

**位置**: `internal/metastore/kv/`

| Catalog           | 路径                | 存储内容                                             | Key前缀                                                                        |
| ----------------- | ----------------- | ------------------------------------------------ | ---------------------------------------------------------------------------- |
| **AgentCatalog**  | `kv/agentcoord/`  | Agent、Session、Workspace                          | `/tdb/agents/`, `/tdb/sessions/`, `/tdb/workspaces/`                         |
| **MemoryCatalog** | `kv/memorycoord/` | Memory、Policy、Adaptation、State、Artifact、Relation | `/tdb/memories/`, `/tdb/memory_policies/`, `/tdb/states/`, `/tdb/relations/` |
| **EventCatalog**  | `kv/event/`       | Event Log、Channel、Subscriber、Position            | `/tdb/events/`, `/tdb/event_channels/`, `/tdb/event_subscribers/`            |

### 3.4 数据模型层

**位置**: `pkg/models/`

#### Agent 模型

```go
type Agent struct {
    AgentID             string            // 唯一标识
    TenantID            string            // 租户ID
    WorkspaceID         string            // 工作空间ID
    AgentType           string            // Agent类型
    RoleProfile         string            // 角色配置
    PolicyRef           string            // 策略引用
    CapabilitySet       []string          // 能力集合
    DefaultMemoryPolicy string            // 默认记忆策略
    CreatedAt           uint64            // 创建时间
    UpdatedAt           uint64            // 更新时间
    State               AgentState        // 状态
    Metadata            map[string]string // 元数据
}

type Session struct {
    SessionID       string       // 会话ID
    AgentID         string       // 所属Agent
    ParentSessionID string       // 父会话ID（支持层级）
    TaskType        string       // 任务类型
    Goal            string       // 目标
    ContextRef      string       // 上下文引用
    StartAt         uint64       // 开始时间
    EndAt           uint64       // 结束时间
    State           SessionState // 状态
    BudgetToken     int64        // Token预算
    BudgetTimeMs    int64        // 时间预算
    Metadata        map[string]string // 元数据
}
```

#### Memory 模型

```go
type Memory struct {
    MemoryID        string            // 记忆ID
    MemoryType      MemoryType        // 记忆类型
    AgentID         string            // 所属Agent
    SessionID       string            // 所属会话
    Scope           string            // 作用域
    Level           MemoryLevel       // 蒸馏层级
    Content         string            // 内容
    Summary         string            // 摘要
    SourceEventIDs  []string          // 来源事件ID
    Confidence      float32           // 置信度
    Importance      float32           // 重要性
    FreshnessScore  float32           // 新鲜度
    TTL             int64             // 存活时间
    ValidFrom       uint64            // 有效起始时间
    ValidTo         uint64            // 有效结束时间
    ProvenanceRef   string            // 来源引用
    Version         int64             // 版本
    IsActive        bool              // 是否激活
    State           MemoryState       // 状态
    CreatedAt       uint64            // 创建时间
    UpdatedAt       uint64            // 更新时间
    Metadata        map[string]string // 元数据
}

type MemoryPolicy struct {
    MemoryID         string   // 关联记忆ID
    SalienceWeight   float32  // 显著性权重
    TTL              int64    // TTL
    DecayFn          string   // 衰减函数
    Confidence       float32  // 置信度
    Verified         bool     // 是否验证
    VerifiedBy       string   // 验证者
    VerifiedAt       uint64   // 验证时间
    Quarantined      bool     // 是否隔离
    QuarantineReason string   // 隔离原因
    VisibilityPolicy string   // 可见性策略
    ReadACL          []string // 读ACL
    WriteACL         []string // 写ACL
    DeriveACL        []string // 派生ACL
    PolicyReason     string   // 策略原因
    PolicySource     string   // 策略来源
    PolicyEventID    string   // 策略事件ID
}

type MemoryAdaptation struct {
    MemoryID            string            // 关联记忆ID
    RetrievalProfile    string            // 检索配置
    RankingParams       map[string]float32 // 排序参数
    FilteringThresholds map[string]float32 // 过滤阈值
    ProjectionWeights   map[string]float32 // 投影权重
    EmbeddingFamily     string            // Embedding族
    ModelID             string            // 模型ID
    AdaptationReason    string            // 适配原因
    AdaptationSource    string            // 适配来源
}
```

#### Event 模型

```go
type Event struct {
    EventID       string            // 事件ID
    TenantID      string            // 租户ID
    WorkspaceID   string            // 工作空间ID
    AgentID       string            // Agent ID
    SessionID     string            // 会话ID
    EventType     EventType         // 事件类型
    EventTime     uint64            // 事件发生时间
    IngestTime    uint64            // 摄入时间
    VisibleTime   uint64            // 可见时间
    LogicalTs     uint64            // 逻辑时间戳
    ParentEventID string            // 父事件ID
    CausalRefs    []string          // 因果引用
    Payload       []byte            // 载荷
    Source        string            // 来源
    Importance    float32           // 重要性
    Visibility    string            // 可见性
    Version       int64             // 版本
    Metadata      map[string]string // 元数据
}
```

***

## 四、服务接口定义

### 4.1 gRPC 服务

**位置**: `pkg/proto/`

#### AgentService

```protobuf
service AgentService {
    // Agent 管理
    rpc CreateAgent(CreateAgentRequest) returns (CreateAgentResponse);
    rpc GetAgent(GetAgentRequest) returns (GetAgentResponse);
    rpc ListAgents(ListAgentsRequest) returns (ListAgentsResponse);
    rpc UpdateAgent(UpdateAgentRequest) returns (UpdateAgentResponse);
    rpc DeleteAgent(DeleteAgentRequest) returns (DeleteAgentResponse);
    
    // Session 管理
    rpc CreateSession(CreateSessionRequest) returns (CreateSessionResponse);
    rpc GetSession(GetSessionRequest) returns (GetSessionResponse);
    rpc ListSessions(ListSessionsRequest) returns (ListSessionsResponse);
    rpc UpdateSession(UpdateSessionRequest) returns (UpdateSessionResponse);
}
```

#### MemoryService

```protobuf
service MemoryService {
    // Memory 管理
    rpc CreateMemory(CreateMemoryRequest) returns (CreateMemoryResponse);
    rpc GetMemory(GetMemoryRequest) returns (GetMemoryResponse);
    rpc UpdateMemory(UpdateMemoryRequest) returns (UpdateMemoryResponse);
    rpc DeleteMemory(DeleteMemoryRequest) returns (DeleteMemoryResponse);
    rpc QueryMemories(QueryMemoriesRequest) returns (QueryMemoriesResponse);
    
    // 向量搜索
    rpc SearchMemories(SearchMemoriesRequest) returns (SearchMemoriesResponse);
    
    // 关系管理
    rpc GetRelations(GetRelationsRequest) returns (GetRelationsResponse);
    rpc CreateRelation(CreateRelationRequest) returns (CreateRelationResponse);
}
```

#### EventService

```protobuf
service EventService {
    // 事件日志
    rpc AppendEvent(AppendEventRequest) returns (AppendEventResponse);
    rpc GetEvent(GetEventRequest) returns (GetEventResponse);
    rpc QueryEvents(QueryEventsRequest) returns (QueryEventsResponse);
    
    // 流式订阅 (Server Streaming)
    rpc SubscribeEvents(SubscribeEventsRequest) returns (stream SubscribeEventsResponse);
}
```

***

## 五、存储层设计

### 5.1 etcd Key 结构

```
/tdb/
├── agents/
│   └── {agent_id}                          # Agent JSON
│
├── sessions/
│   └── {agent_id}/
│       └── {session_id}                    # Session JSON
│
├── workspaces/
│   └── {tenant_id}/
│       └── {workspace_id}                  # Workspace JSON
│
├── memories/
│   └── {agent_id}/
│       └── {memory_id}                     # Memory JSON
│
├── memory_policies/
│   └── {memory_id}                         # MemoryPolicy JSON
│
├── memory_adaptations/
│   └── {memory_id}                         # MemoryAdaptation JSON
│
├── states/
│   └── {agent_id}/
│       └── {state_id}                      # State JSON
│
├── artifacts/
│   └── {session_id}/
│       └── {artifact_id}                   # Artifact JSON
│
├── relations/
│   └── {object_id}/
│       └── {edge_id}                       # Relation JSON
│
├── events/
│   └── {channel_name}/
│       └── {log_id}                        # Event JSON
│
├── event_channels/
│   └── {channel_name}                      # Channel Metadata
│
├── event_subscribers/
│   └── {subscriber_id}                     # Subscriber Metadata
│
└── event_positions/
    └── {subscriber_id}/
        └── {channel_name}                  # Position JSON
```

### 5.2 存储后端

| 后端         | 用途            | 配置               |
| ---------- | ------------- | ---------------- |
| **etcd**   | 元数据存储、服务发现    | `etcd.endpoints` |
| **TiKV**   | 大规模元数据存储 (可选) | `tikv.endpoints` |
| **Milvus** | 向量存储、相似度检索    | `milvus.address` |

***

## 六、系统数据流

### 6.1 写入流程

```
Client Request
      ↓
┌─────────────┐
│ gRPC Server │
└──────┬──────┘
       ↓
┌─────────────┐
│ Coordinator │ → 业务逻辑处理
└──────┬──────┘
       ↓
┌─────────────┐
│   Catalog   │ → 元数据操作
└──────┬──────┘
       ↓
┌─────────────┐
│    etcd     │ → 持久化存储
└─────────────┘
       ↓
┌─────────────┐
│   Milvus    │ → 向量索引 (异步)
└─────────────┘
```

### 6.2 查询流程

```
Client Request
      ↓
┌─────────────┐
│ gRPC Server │
└──────┬──────┘
       ↓
┌─────────────┐
│ Coordinator │ → 权限检查、策略过滤
└──────┬──────┘
       ↓
┌─────────────┐
│   Catalog   │ → 查询元数据
└──────┬──────┘
       ↓
┌─────────────┐
│    etcd     │ → 读取数据
└──────┬──────┘
       ↓
┌─────────────┐
│   Milvus    │ → 向量检索 (如需要)
└──────┬──────┘
       ↓
   Response
```

### 6.3 事件订阅流程

```
Client Subscribe
      ↓
┌─────────────┐
│ EventServer │
└──────┬──────┘
       ↓
┌─────────────┐
│ EventCoord  │ → 创建订阅
└──────┬──────┘
       ↓
┌─────────────┐
│  EventCh    │ ← 事件通道
└──────┬──────┘
       ↓
   Stream Events → Client
```

***

## 七、Milvus 集成

### 7.1 组件集成关系

```
┌─────────────────────────────────────────┐
│           Milvus Cluster                │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐ │
│  │  Proxy  │  │ MixCoord│  │QueryNode│ │
│  └────┬────┘  └────┬────┘  └────┬────┘ │
│       │            │            │      │
│       └────────────┼────────────┘      │
│                    │                   │
│              ┌─────▼─────┐             │
│              │    TDB    │             │
│              │  ┌─────┐  │             │
│              │  │Agent│  │             │
│              │  │Mem  │  │             │
│              │  │Event│  │             │
│              │  └─────┘  │             │
│              └─────┬─────┘             │
│                    │                   │
│       ┌────────────┼────────────┐      │
│       ▼            ▼            ▼      │
│  ┌─────────┐  ┌─────────┐  ┌────────┐ │
│  │  etcd   │  │  TiKV   │  │ Milvus │ │
│  │(Metadata│  │(Metadata│  │(Vector)│ │
│  └─────────┘  └─────────┘  └────────┘ │
│                                         │
└─────────────────────────────────────────┘
```

### 7.2 配置集成

**位置**: `pkg/util/paramtable/component_param.go`

```go
type ComponentParam struct {
    // ... 其他配置
    
    // TDB 配置
    TDBCfg         tdbConfig
    TDBGrpcServerCfg GrpcServerConfig
}

type tdbConfig struct {
    Enabled             ParamItem // 是否启用
    GracefulStopTimeout ParamItem // 优雅停止超时
}
```

**配置文件**:

```yaml
# milvus.yaml
tdb:
  enabled: true                    # 启用TDB
  gracefulStopTimeout: 5           # 优雅停止超时(秒)
  
  # gRPC服务器配置
  ip: 0.0.0.0
  port: 19530
  serverMaxRecvSize: 104857600     # 100MB
  serverMaxSendSize: 104857600     # 100MB
```

### 7.3 启动流程

```
1. main.go
   └── milvus.RunMilvus()
       └── roles.MilvusRoles.Run()
           ├── paramtable.Init()          # 初始化配置
           ├── 启动HTTP服务               # 健康检查
           ├── 启动Prometheus指标         # 监控
           ├── 初始化各组件
           │   ├── MixCoord               # 混合协调器
           │   ├── QueryNode              # 查询节点
           │   ├── DataNode               # 数据节点
           │   ├── Proxy                  # 代理
           │   └── TDB                    # TDB服务
           │       └── components.NewTDB()
           │           └── tdb.NewServer()
           │               ├── NewAgentCoord()
           │               ├── NewMemoryCoord()
           │               └── NewEventCoord()
           └── 等待所有组件就绪
```

***

## 八、系统特性

### 8.1 核心能力

| 特性           | 说明                    |
| ------------ | --------------------- |
| **Agent 原生** | 专为AI Agent设计的对象模型     |
| **事件驱动**     | WAL + Event Log驱动状态演化 |
| **记忆管理**     | 5种记忆类型 + 3层架构         |
| **向量检索**     | 基于Milvus的高性能向量搜索      |
| **关系图谱**     | 支持多跳推理的关系网络           |
| **版本控制**     | 时间旅行查询和版本回滚           |
| **策略治理**     | 细粒度的ACL和策略控制          |
| **多租户**      | Tenant/Workspace隔离    |
| **云原生**      | Milvus分布式架构支撑         |

### 8.2 架构优势

- **模块化设计**: 各Coord独立演进，接口解耦
- **高可用**: 支持Active-Standby模式
- **可扩展**: 基于Milvus的分布式架构
- **强一致性**: 基于etcd/TiKV的强一致存储
- **高性能**: 向量检索 + 缓存优化
- **可观测**: 完整的监控和日志

***

## 九、文档索引

| 文档            | 路径                 | 说明        |
| ------------- | ------------------ | --------- |
| 项目规划          | `plan1.0.md`       | 整体规划和进度   |
| Schema设计      | `fixed1.01.md`     | 数据模型设计    |
| KV Catalog设计  | `fixed1.02.md`     | 存储层设计     |
| Coordinator设计 | `fixed1.03.md`     | 服务层设计     |
| Protobuf代码生成  | `fixed1.04.md`     | gRPC接口生成  |
| gRPC Server设计 | `fixed1.05.md`     | Server层设计 |
| 服务集成设计        | `fixed1.06.md`     | Milvus集成  |
| 测试与优化         | `fixed1.07.md`     | 测试策略      |
| **系统结构**      | `structure1.01.md` | **本文件**   |

***

## 十、变更记录

| 版本   | 日期         | 变更内容        | 作者 |
| ---- | ---------- | ----------- | -- |
| 1.01 | 2026-03-15 | 初始版本，完整系统结构 | -  |

***

