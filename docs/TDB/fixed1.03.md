# TDB Fixed 1.03 - Coordinator 服务层实现

## 概述

本文档记录了 TDB (Agent-Native Database) 的三个核心 Coordinator 服务层的实现：
- AgentCoordinator - Agent 管理服务
- MemoryCoordinator - 记忆管理服务
- EventCoordinator - 事件流管理服务

这三个 Coordinator 构成了 TDB 的核心服务层，位于 gRPC 接口层和 KV Catalog 存储层之间。

---

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    gRPC API Gateway                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Coordinator Layer                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ AgentCoord   │  │ MemoryCoord  │  │ EventCoord   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    KV Catalog Layer                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ AgentCatalog │  │MemoryCatalog │  │ EventCatalog │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    etcd Storage Layer                       │
└─────────────────────────────────────────────────────────────┘
```

### Coordinator 设计模式

所有 Coordinator 遵循统一的设计模式：

1. **生命周期管理**: Init() → Start() → Stop()
2. **状态机**: Initializing → Healthy → Stopping
3. **后台任务**: 使用 time.Ticker 定期执行维护任务
4. **并发安全**: 使用 sync.Once 确保初始化/启动/停止只执行一次
5. **错误处理**: 使用 cockroachdb/errors 进行错误包装
6. **日志记录**: 使用 zap 进行结构化日志记录

---

## 1. AgentCoordinator

### 文件位置
`internal/agentcoord/agent_coord.go` (790 行)

### 核心功能

AgentCoordinator 负责管理 Agent 的生命周期和元数据。

#### 管理实体
- **Agent**: Agent 基本信息
- **AgentProfile**: Agent 配置档案
- **AgentCapability**: Agent 能力定义
- **AgentACL**: Agent 访问控制
- **AgentLifecycle**: Agent 生命周期状态

#### 主要方法

```go
// Agent 管理
CreateAgent(ctx, agentID, agentType, tenantID, name, description, config string) (*models.Agent, error)
GetAgent(ctx, agentID string) (*models.Agent, error)
ListAgents(ctx, tenantID string) ([]*models.Agent, error)
UpdateAgent(ctx, agentID string, updates map[string]interface{}) (*models.Agent, error)
DeleteAgent(ctx, agentID string) error

// Profile 管理
CreateAgentProfile(ctx, agentID, profileType string, config map[string]interface{}) (*models.AgentProfile, error)
GetAgentProfile(ctx, profileID string) (*models.AgentProfile, error)

// Capability 管理
RegisterCapability(ctx, agentID, capabilityType, capabilityName, version string, config map[string]interface{}) (*models.AgentCapability, error)
ListCapabilities(ctx, agentID string) ([]*models.AgentCapability, error)

// ACL 管理
CreateACL(ctx, agentID string, permissions []models.AgentPermission, resourcePatterns []string) (*models.AgentACL, error)

// Lifecycle 管理
CreateLifecycle(ctx, agentID string, state models.AgentState, reason string) (*models.AgentLifecycle, error)
UpdateLifecycleState(ctx, agentID string, state models.AgentState, reason string) (*models.AgentLifecycle, error)
```

#### 状态码
```go
const (
    StateCode_Initializing StateCode = 0
    StateCode_Healthy      StateCode = 1
    StateCode_Abnormal     StateCode = 2
    StateCode_Stopping     StateCode = 3
)
```

---

## 2. MemoryCoordinator

### 文件位置
`internal/memorycoord/memory_coord.go` (1013 行)

### 核心功能

MemoryCoordinator 负责管理 Agent 的记忆系统，包括记忆、状态、工件和关系。

#### 管理实体
- **Memory**: 记忆内容
- **MemoryPolicy**: 记忆策略
- **MemoryAdaptation**: 记忆适配参数
- **State**: Agent 状态快照
- **Artifact**: 会话工件
- **Relation**: 对象间关系
- **ShareContract**: 记忆共享合约

#### 主要方法

```go
// Memory 管理
CreateMemory(ctx, agentID, sessionID string, memoryType models.MemoryType, scope, content, summary string, sourceEventIDs []string, confidence, importance float32, ttl int64, metadata map[string]string, embeddingVector []float32) (*models.Memory, error)
GetMemory(ctx, memoryID string) (*models.Memory, *models.MemoryPolicy, *models.MemoryAdaptation, error)
ListMemories(ctx, agentID, sessionID string) ([]*models.Memory, error)
ListMemoriesByType(ctx, agentID string, memoryType models.MemoryType) ([]*models.Memory, error)
UpdateMemory(ctx, memoryID string, updates map[string]interface{}) (*models.Memory, error)
DeleteMemory(ctx, memoryID string, hardDelete bool) error
ArchiveMemory(ctx, memoryID string) error
QuarantineMemory(ctx, memoryID, reason string) error

// MemoryPolicy 管理
UpdateMemoryPolicy(ctx, memoryID string, updates map[string]interface{}) (*models.MemoryPolicy, error)

// MemoryAdaptation 管理
UpdateMemoryAdaptation(ctx, memoryID string, updates map[string]interface{}) (*models.MemoryAdaptation, error)

// State 管理
CreateState(ctx, agentID, sessionID, stateType, stateKey string, stateValue []byte, derivedFromEventID string, metadata map[string]string) (*models.State, error)
GetState(ctx, stateID string) (*models.State, error)
ListStates(ctx, agentID, sessionID string) ([]*models.State, error)
UpdateState(ctx, stateID string, stateValue []byte) (*models.State, error)
DeleteState(ctx, stateID string) error

// Artifact 管理
CreateArtifact(ctx, sessionID, ownerAgentID, artifactType, uri, contentRef, mimeType string, metadata map[string]string, hash string) (*models.Artifact, error)
GetArtifact(ctx, artifactID string) (*models.Artifact, error)
ListArtifacts(ctx, sessionID string) ([]*models.Artifact, error)
DeleteArtifact(ctx, artifactID string) error

// Relation 管理
CreateRelation(ctx, srcObjectID, srcType, dstObjectID, dstType, relationType string, weight float32, properties map[string]string, createdByEventID string) (*models.Relation, error)
GetRelation(ctx, edgeID string) (*models.Relation, error)
ListRelations(ctx, objectID string) ([]*models.Relation, error)
DeleteRelation(ctx, edgeID string) error

// ShareContract 管理
CreateShareContract(ctx, scope, ownerAgentID string, readACL, writeACL, deriveACL []string, ttlPolicy int64, consistencyLevel models.ConsistencyLevel, mergePolicy models.MergePolicy, metadata map[string]string) (*models.ShareContract, error)
GetShareContract(ctx, contractID string) (*models.ShareContract, error)
ListShareContracts(ctx, scope string) ([]*models.ShareContract, error)
UpdateShareContract(ctx, contractID string, updates map[string]interface{}) (*models.ShareContract, error)
DeleteShareContract(ctx, contractID string) error
```

#### 状态码
```go
const (
    StateCode_Initializing StateCode = 0
    StateCode_Healthy      StateCode = 1
    StateCode_Abnormal     StateCode = 2
    StateCode_Stopping     StateCode = 3
)
```

---

## 3. EventCoordinator

### 文件位置
`internal/eventcoord/event_coord.go` (773 行)

### 核心功能

EventCoordinator 负责管理事件流，提供事件日志、订阅和流式传输功能。

#### 管理实体
- **Event**: 事件内容
- **EventFilter**: 事件过滤条件
- **EventLogPosition**: 事件日志位置
- **Channel**: 事件通道
- **Subscriber**: 事件订阅者

#### 主要方法

```go
// Event Log 操作
AppendEvent(ctx, channelName string, agentID, sessionID string, eventType models.EventType, payload []byte, parentEventID string, causalRefs []string, source string, importance float32, visibility string, metadata map[string]string) (*models.Event, uint64, error)
GetEvent(ctx, channelName string, logID uint64) (*models.Event, error)
GetEvents(ctx, channelName string, startLogID, endLogID uint64) ([]*models.Event, error)
ListEventsByTimeRange(ctx, channelName string, startTime, endTime uint64) ([]*models.Event, error)
QueryEvents(ctx, channelName string, filter *models.EventFilter, limit int64) ([]*models.Event, error)
AppendEventBatch(ctx, channelName string, events []*models.Event) ([]uint64, error)

// Event Meta 操作
SaveEventMeta(ctx, eventID string, meta map[string]string) error
GetEventMeta(ctx, eventID string) (map[string]string, error)

// Channel 操作
CreateChannel(ctx, channelName string) error
GetChannel(ctx, channelName string) (map[string]interface{}, error)
ListChannels(ctx) ([]string, error)
DeleteChannel(ctx, channelName string) error

// Subscriber 操作
RegisterSubscriber(ctx, subscriberID string, filter *models.EventFilter) error
GetSubscriber(ctx, subscriberID string) (map[string]interface{}, error)
UnregisterSubscriber(ctx, subscriberID string) error
ListSubscribers(ctx) ([]string, error)

// Position 操作
SavePosition(ctx, subscriberID, channelName string, position *models.EventLogPosition) error
GetPosition(ctx, subscriberID, channelName string) (*models.EventLogPosition, error)
GetCurrentLogID(ctx, channelName string) (uint64, error)

// Event Stream 操作
Subscribe(ctx, subscriberID, channelName string, filter *models.EventFilter, startPosition *models.EventLogPosition) (chan *models.Event, error)
Unsubscribe(ctx, subscriberID, channelName string) error
```

#### 状态码
```go
const (
    StateCode_Initializing StateCode = 0
    StateCode_Healthy      StateCode = 1
    StateCode_Abnormal     StateCode = 2
    StateCode_Stopping     StateCode = 3
)
```

---

## 代码统计

| Coordinator | 文件 | 行数 | 主要功能 |
|------------|------|------|----------|
| AgentCoordinator | internal/agentcoord/agent_coord.go | 790 | Agent 生命周期、Profile、Capability、ACL |
| MemoryCoordinator | internal/memorycoord/memory_coord.go | 1013 | Memory、State、Artifact、Relation、ShareContract |
| EventCoordinator | internal/eventcoord/event_coord.go | 773 | Event Log、Channel、Subscriber、Stream |

**总计**: 2576 行代码

---

## 依赖关系

### 外部依赖
```go
import (
    "github.com/cockroachdb/errors"
    "go.uber.org/atomic"
    "go.uber.org/zap"
)
```

### Milvus 内部依赖
```go
import (
    "github.com/milvus-io/milvus/internal/metastore/kv/agentcoord"
    "github.com/milvus-io/milvus/internal/metastore/kv/memorycoord"
    "github.com/milvus-io/milvus/internal/metastore/kv/event"
    "github.com/milvus-io/milvus/internal/util/sessionutil"
    "github.com/milvus-io/milvus/pkg/v2/kv"
    "github.com/milvus-io/milvus/pkg/v2/log"
    "github.com/milvus-io/milvus/pkg/v2/models"
    "github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)
```

---

## 后续工作

### 待实现功能

1. **AgentCoordinator**
   - 状态恢复机制
   - Agent 健康检查
   - 批量操作优化

2. **MemoryCoordinator**
   - 记忆衰减和清理
   - TTL 过期检查
   - 索引维护
   - 隔离区检查

3. **EventCoordinator**
   - 通道健康检查
   - 订阅者超时检查
   - 位置清理
   - 事件保留策略执行

### 下一步

1. 实现 gRPC 服务层，将 Coordinator 方法暴露为 API
2. 添加单元测试和集成测试
3. 实现配置管理和动态更新
4. 添加监控和指标收集

---

## 变更记录

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| 1.03 | 2026-03-14 | 实现 AgentCoordinator、MemoryCoordinator、EventCoordinator |
