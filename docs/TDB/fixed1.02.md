# TDB Schema与对象模型层 - 版本 1.02

> 创建时间: 2026-03-14
> 本版本内容: Proto定义、etcd元数据Schema、Go模型代码、KV Catalog层实现

---

## 更新说明

本版本(1.02)在1.01的基础上增加了KV Catalog层实现，包括：
- AgentCoord Catalog - Agent/Session/Workspace的CRUD操作
- MemoryCoord Catalog - Memory/State/Artifact/Relation/ShareContract的CRUD操作
- Event Catalog - Event日志和订阅管理

---

## 1. Proto定义文件

*(与1.01版本相同，此处省略，详见fixed1.01.md)*

---

## 2. etcd元数据常量定义

*(与1.01版本相同，此处省略，详见fixed1.01.md)*

---

## 3. Go模型代码

*(与1.01版本相同，此处省略，详见fixed1.01.md)*

---

## 4. KV Catalog层实现

### 4.1 AgentCoord Catalog

**文件位置**: `internal/metastore/kv/agentcoord/kv_catalog.go`

```go
package agentcoord

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/milvus-io/milvus/pkg/v2/kv"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// Catalog provides access to agent metadata stored in etcd.
type Catalog struct {
	Txn      kv.TxnKV
	Snapshot kv.SnapShotKV
}

// NewCatalog creates a new Catalog instance.
func NewCatalog(metaKV kv.TxnKV, ss kv.SnapShotKV) *Catalog {
	return &Catalog{Txn: metaKV, Snapshot: ss}
}

// ==================== Agent Operations ====================

// CreateAgent creates a new agent in the catalog.
func (c *Catalog) CreateAgent(ctx context.Context, agent *models.Agent, ts typeutil.Timestamp) error {
	key := BuildAgentKey(agent.AgentID)
	value, err := json.Marshal(agent)
	if err != nil {
		return errors.Wrap(err, "failed to marshal agent")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetAgent retrieves an agent by ID.
func (c *Catalog) GetAgent(ctx context.Context, agentID string, ts typeutil.Timestamp) (*models.Agent, error) {
	key := BuildAgentKey(agentID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("agent not found: %s", agentID)
		}
		return nil, errors.Wrap(err, "failed to load agent")
	}

	var agent models.Agent
	if err := json.Unmarshal([]byte(value), &agent); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal agent")
	}
	return &agent, nil
}

// ListAgents lists all agents in a workspace.
func (c *Catalog) ListAgents(ctx context.Context, tenantID, workspaceID string, ts typeutil.Timestamp) ([]*models.Agent, error) {
	prefix := BuildAgentPrefix(tenantID, workspaceID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list agents")
	}

	agents := make([]*models.Agent, 0, len(values))
	for _, value := range values {
		var agent models.Agent
		if err := json.Unmarshal([]byte(value), &agent); err != nil {
			log.Warn("failed to unmarshal agent", log.Error(err))
			continue
		}
		agents = append(agents, &agent)
	}
	return agents, nil
}

// UpdateAgent updates an existing agent.
func (c *Catalog) UpdateAgent(ctx context.Context, agent *models.Agent, ts typeutil.Timestamp) error {
	key := BuildAgentKey(agent.AgentID)
	value, err := json.Marshal(agent)
	if err != nil {
		return errors.Wrap(err, "failed to marshal agent")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteAgent deletes an agent.
func (c *Catalog) DeleteAgent(ctx context.Context, agentID string, ts typeutil.Timestamp) error {
	key := BuildAgentKey(agentID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// AgentExists checks if an agent exists.
func (c *Catalog) AgentExists(ctx context.Context, agentID string, ts typeutil.Timestamp) bool {
	key := BuildAgentKey(agentID)
	_, err := c.Snapshot.Load(ctx, key, ts)
	return err == nil
}

// ==================== Session Operations ====================

// CreateSession creates a new session in the catalog.
func (c *Catalog) CreateSession(ctx context.Context, session *models.Session, ts typeutil.Timestamp) error {
	key := BuildSessionKey(session.SessionID)
	value, err := json.Marshal(session)
	if err != nil {
		return errors.Wrap(err, "failed to marshal session")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetSession retrieves a session by ID.
func (c *Catalog) GetSession(ctx context.Context, sessionID string, ts typeutil.Timestamp) (*models.Session, error) {
	key := BuildSessionKey(sessionID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("session not found: %s", sessionID)
		}
		return nil, errors.Wrap(err, "failed to load session")
	}

	var session models.Session
	if err := json.Unmarshal([]byte(value), &session); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal session")
	}
	return &session, nil
}

// ListSessions lists all sessions for an agent.
func (c *Catalog) ListSessions(ctx context.Context, agentID string, ts typeutil.Timestamp) ([]*models.Session, error) {
	prefix := BuildSessionPrefix(agentID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list sessions")
	}

	sessions := make([]*models.Session, 0, len(values))
	for _, value := range values {
		var session models.Session
		if err := json.Unmarshal([]byte(value), &session); err != nil {
			log.Warn("failed to unmarshal session", log.Error(err))
			continue
		}
		sessions = append(sessions, &session)
	}
	return sessions, nil
}

// ListSessionsByState lists sessions filtered by state.
func (c *Catalog) ListSessionsByState(ctx context.Context, agentID string, state models.SessionState, ts typeutil.Timestamp) ([]*models.Session, error) {
	sessions, err := c.ListSessions(ctx, agentID, ts)
	if err != nil {
		return nil, err
	}

	filtered := make([]*models.Session, 0)
	for _, session := range sessions {
		if session.State == state {
			filtered = append(filtered, session)
		}
	}
	return filtered, nil
}

// UpdateSession updates an existing session.
func (c *Catalog) UpdateSession(ctx context.Context, session *models.Session, ts typeutil.Timestamp) error {
	key := BuildSessionKey(session.SessionID)
	value, err := json.Marshal(session)
	if err != nil {
		return errors.Wrap(err, "failed to marshal session")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteSession deletes a session.
func (c *Catalog) DeleteSession(ctx context.Context, sessionID string, ts typeutil.Timestamp) error {
	key := BuildSessionKey(sessionID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== Workspace Operations ====================

// CreateWorkspace creates a new workspace in the catalog.
func (c *Catalog) CreateWorkspace(ctx context.Context, workspace *models.Workspace, ts typeutil.Timestamp) error {
	key := BuildWorkspaceKey(workspace.WorkspaceID)
	value, err := json.Marshal(workspace)
	if err != nil {
		return errors.Wrap(err, "failed to marshal workspace")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetWorkspace retrieves a workspace by ID.
func (c *Catalog) GetWorkspace(ctx context.Context, workspaceID string, ts typeutil.Timestamp) (*models.Workspace, error) {
	key := BuildWorkspaceKey(workspaceID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("workspace not found: %s", workspaceID)
		}
		return nil, errors.Wrap(err, "failed to load workspace")
	}

	var workspace models.Workspace
	if err := json.Unmarshal([]byte(value), &workspace); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal workspace")
	}
	return &workspace, nil
}

// ListWorkspaces lists all workspaces for a tenant.
func (c *Catalog) ListWorkspaces(ctx context.Context, tenantID string, ts typeutil.Timestamp) ([]*models.Workspace, error) {
	prefix := BuildWorkspacePrefix(tenantID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list workspaces")
	}

	workspaces := make([]*models.Workspace, 0, len(values))
	for _, value := range values {
		var workspace models.Workspace
		if err := json.Unmarshal([]byte(value), &workspace); err != nil {
			log.Warn("failed to unmarshal workspace", log.Error(err))
			continue
		}
		workspaces = append(workspaces, &workspace)
	}
	return workspaces, nil
}

// UpdateWorkspace updates an existing workspace.
func (c *Catalog) UpdateWorkspace(ctx context.Context, workspace *models.Workspace, ts typeutil.Timestamp) error {
	key := BuildWorkspaceKey(workspace.WorkspaceID)
	value, err := json.Marshal(workspace)
	if err != nil {
		return errors.Wrap(err, "failed to marshal workspace")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteWorkspace deletes a workspace.
func (c *Catalog) DeleteWorkspace(ctx context.Context, workspaceID string, ts typeutil.Timestamp) error {
	key := BuildWorkspaceKey(workspaceID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== Snapshot Operations ====================

// CreateSnapshot creates a snapshot of an object.
func (c *Catalog) CreateSnapshot(ctx context.Context, objectType, objectID string, data []byte, version int64, ts typeutil.Timestamp) error {
	key := BuildSnapshotKey(objectType, objectID, version)
	return c.Snapshot.Save(ctx, key, string(data), ts)
}

// GetSnapshot retrieves a snapshot.
func (c *Catalog) GetSnapshot(ctx context.Context, objectType, objectID string, version int64, ts typeutil.Timestamp) ([]byte, error) {
	key := BuildSnapshotKey(objectType, objectID, version)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load snapshot")
	}
	return []byte(value), nil
}

// ListSnapshots lists all snapshots for an object.
func (c *Catalog) ListSnapshots(ctx context.Context, objectType, objectID string, ts typeutil.Timestamp) ([]int64, error) {
	prefix := fmt.Sprintf("%s/%s/%s", SnapshotPrefix, objectType, objectID)
	keys, _, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list snapshots")
	}

	versions := make([]int64, 0, len(keys))
	for _, key := range keys {
		var version int64
		if _, err := fmt.Sscanf(key, prefix+"/%d", &version); err == nil {
			versions = append(versions, version)
		}
	}
	return versions, nil
}

// ==================== Transaction Operations ====================

// MultiSave saves multiple key-value pairs in a transaction.
func (c *Catalog) MultiSave(ctx context.Context, kvs map[string]string, ts typeutil.Timestamp) error {
	return c.Snapshot.MultiSave(ctx, kvs, ts)
}

// MultiSaveAndRemove saves and removes multiple keys in a transaction.
func (c *Catalog) MultiSaveAndRemove(ctx context.Context, saves map[string]string, removals []string, ts typeutil.Timestamp) error {
	return c.Snapshot.MultiSaveAndRemove(ctx, saves, removals, ts)
}
```

---

### 4.2 MemoryCoord Catalog

**文件位置**: `internal/metastore/kv/memorycoord/kv_catalog.go`

主要功能：
- **Memory Operations**: Create/Get/List/Update/Delete Memory
- **MemoryPolicy Operations**: 记忆策略的CRUD
- **MemoryAdaptation Operations**: 记忆适配层的CRUD
- **State Operations**: 状态对象的CRUD
- **Artifact Operations**: 外部工件的CRUD
- **Relation Operations**: 关系/边的CRUD
- **ShareContract Operations**: 共享契约的CRUD

核心方法示例：
```go
// CreateMemory creates a new memory in the catalog.
func (c *Catalog) CreateMemory(ctx context.Context, memory *models.Memory, ts typeutil.Timestamp) error

// GetMemory retrieves a memory by ID.
func (c *Catalog) GetMemory(ctx context.Context, memoryID string, ts typeutil.Timestamp) (*models.Memory, error)

// ListMemories lists all memories for an agent or session.
func (c *Catalog) ListMemories(ctx context.Context, agentID, sessionID string, ts typeutil.Timestamp) ([]*models.Memory, error)

// DeleteMemory deletes a memory (soft delete by default).
func (c *Catalog) DeleteMemory(ctx context.Context, memoryID string, hardDelete bool, ts typeutil.Timestamp) error
```

---

### 4.3 Event Catalog

**文件位置**: `internal/metastore/kv/event/kv_catalog.go`

主要功能：
- **Event Log Operations**: 事件日志的追加和查询
- **Event Meta Operations**: 事件元数据管理
- **Channel Operations**: 事件通道管理
- **Subscriber Operations**: 订阅者注册/注销
- **Position Operations**: 消费位置管理
- **Log ID Management**: 日志ID自增管理
- **Batch Operations**: 批量事件写入
- **Query Operations**: 基于过滤器的事件查询

核心方法示例：
```go
// AppendEvent appends an event to the event log.
func (c *Catalog) AppendEvent(ctx context.Context, channelName string, event *models.Event, ts typeutil.Timestamp) (uint64, error)

// GetEvent retrieves an event by log ID.
func (c *Catalog) GetEvent(ctx context.Context, channelName string, logID uint64, ts typeutil.Timestamp) (*models.Event, error)

// QueryEvents queries events based on a filter.
func (c *Catalog) QueryEvents(ctx context.Context, channelName string, filter *models.EventFilter, limit int64, ts typeutil.Timestamp) ([]*models.Event, error)

// AppendEventBatch appends a batch of events.
func (c *Catalog) AppendEventBatch(ctx context.Context, channelName string, events []*models.Event, ts typeutil.Timestamp) ([]uint64, error)
```

---

## 5. 目录结构

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
        ├── agentcoord/
        │   ├── constant.go
        │   └── kv_catalog.go       # 新增
        ├── memorycoord/
        │   ├── constant.go
        │   └── kv_catalog.go       # 新增
        └── event/
            ├── constant.go
            └── kv_catalog.go       # 新增
```

---

## 6. 技术特点

### 6.1 存储架构
- 使用 **etcd** 作为元数据存储
- 使用 **SnapshotKV** 支持时间戳版本控制
- 使用 **TxnKV** 支持事务操作

### 6.2 序列化
- 使用 **JSON** 进行数据序列化（替代protobuf，避免生成依赖）
- 保持与模型定义的一致性

### 6.3 错误处理
- 使用 `github.com/cockroachdb/errors` 进行错误包装
- 支持错误类型判断（如 `kv.ErrKeyNotFound`）

### 6.4 日志记录
- 使用 Milvus 的 `log` 包进行日志记录
- 在反序列化失败等场景下记录警告日志

---

## 7. 后续计划

### 7.1 下一步建议

1. **实现 Coordinator 服务层**
   - AgentCoordinator: 管理Agent生命周期
   - MemoryCoordinator: 管理Memory存储和检索
   - EventCoordinator: 管理Event流

2. **实现 gRPC 服务接口**
   - 基于proto定义实现服务
   - 添加认证和权限控制

3. **集成测试**
   - 编写Catalog层的单元测试
   - 集成etcd进行端到端测试

### 7.2 文件位置汇总

| 文件类型 | 路径 |
|---------|------|
| Proto定义 | `pkg/proto/*.proto` |
| Go模型 | `pkg/models/*.go` |
| etcd常量 | `internal/metastore/kv/*/constant.go` |
| KV Catalog | `internal/metastore/kv/*/kv_catalog.go` |
| 本文档 | `docs/TDB/fixed1.02.md` |

---

*文档版本: 1.02*  
*最后更新: 2026-03-14*
