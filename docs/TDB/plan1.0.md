# TDB (Agent-Native Database) 项目修改规划 v1.0

> 本文档记录 TDB 项目的完整开发规划，后续在此基础上直接修改更新。
> 创建日期: 2026-03-14

***

## 项目概述

TDB 是一个面向 AI Agent 的数据库系统，作为 Milvus 的扩展组件实现，提供：

- Agent 生命周期管理
- 记忆存储与检索
- 事件流处理

***

## 架构分层

```
┌─────────────────────────────────────────────────────────────┐
│                    gRPC API Gateway                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │AgentService  │  │MemoryService │  │EventService  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
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

***

## 阶段 1: Protocol Buffer 定义

**状态**: ✓ 已完成

### 任务清单

- [x] 创建 agent.proto
- [x] 创建 memory.proto
- [x] 创建 event.proto
- [x] 创建 common.proto

### 文件位置

| 文件           | 路径                       | 说明                     |
| ------------ | ------------------------ | ---------------------- |
| agent.proto  | `pkg/proto/agent.proto`  | Agent 和 Session 定义     |
| memory.proto | `pkg/proto/memory.proto` | Memory 和相关实体定义         |
| event.proto  | `pkg/proto/event.proto`  | Event 和 EventFilter 定义 |
| common.proto | `pkg/proto/common.proto` | 通用类型定义                 |

### 服务定义

- **AgentService**: CreateAgent, GetAgent, ListAgents, UpdateAgent, DeleteAgent, CreateSession, GetSession, ListSessions, UpdateSession
- **MemoryService**: CreateMemory, GetMemory, UpdateMemory, DeleteMemory, QueryMemories, SearchMemories, GetRelations, CreateRelation
- **EventService**: AppendEvent, GetEvent, QueryEvents, SubscribeEvents

***

## 阶段 2: etcd 元数据常量定义

**状态**: ✓ 已完成

### 实现方式

常量直接定义在各 KV Catalog 文件中，如需统一可后续提取到单独文件。

### 文件位置

| 组件     | 文件路径                                              |
| ------ | ------------------------------------------------- |
| Agent  | `internal/metastore/kv/agentcoord/kv_catalog.go`  |
| Memory | `internal/metastore/kv/memorycoord/kv_catalog.go` |
| Event  | `internal/metastore/kv/event/kv_catalog.go`       |

***

## 阶段 3: Go 代码生成

**状态**: ✓ 已完成

### 任务清单

- [x] 生成 agent.pb.go
- [x] 生成 agent\_grpc.pb.go
- [x] 生成 memory.pb.go
- [x] 生成 memory\_grpc.pb.go
- [x] 生成 event.pb.go
- [x] 生成 event\_grpc.pb.go
- [x] 生成 common.pb.go

### 已生成文件

| 文件                | 路径                                         | 说明                  |
| ----------------- | ------------------------------------------ | ------------------- |
| agent.pb.go       | `pkg/proto/agentpb/agent.pb.go`            | Agent 消息类型定义       |
| agent\_grpc.pb.go | `pkg/proto/agentpb/agent_grpc.pb.go`       | AgentService gRPC 接口 |
| memory.pb.go      | `pkg/proto/memorypb/memory.pb.go`          | Memory 消息类型定义      |
| memory\_grpc.pb.go| `pkg/proto/memorypb/memory_grpc.pb.go`     | MemoryService gRPC 接口|
| event.pb.go       | `pkg/proto/eventpb/event.pb.go`            | Event 消息类型定义       |
| event\_grpc.pb.go | `pkg/proto/eventpb/event_grpc.pb.go`       | EventService gRPC 接口 |
| common.pb.go      | `pkg/proto/commonpb/common.pb.go`          | 通用类型定义 (Status等)  |

### 生成命令

```bash
cd pkg/proto

# Agent
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       agent.proto common.proto

# Memory  
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       memory.proto common.proto

# Event
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       event.proto common.proto
```

### 修复记录

详细修复内容见 `docs/TDB/fixed1.04.md`，主要包括：
1. 修复 proto 文件中 `common.Status` 引用为 `common.proto.Status`
2. 更新 go_package 选项使用正确的导入路径 `github.com/milvus-io/milvus/pkg/v2/proto/...`
3. 移除 memory.proto 中未使用的 schema.proto 导入

***

## 阶段 4: KV Catalog 层

**状态**: ✓ 已完成

### 任务清单

- [x] 实现 Agent Catalog
- [x] 实现 Memory Catalog
- [x] 实现 Event Catalog

### 文件位置

| 组件             | 文件路径                                              | 行数    |
| -------------- | ------------------------------------------------- | ----- |
| Agent Catalog  | `internal/metastore/kv/agentcoord/kv_catalog.go`  | \~400 |
| Memory Catalog | `internal/metastore/kv/memorycoord/kv_catalog.go` | \~600 |
| Event Catalog  | `internal/metastore/kv/event/kv_catalog.go`       | \~450 |

### 核心功能

- **AgentCatalog**: Agent, Profile, Capability, ACL, Lifecycle 的 CRUD
- **MemoryCatalog**: Memory, Policy, Adaptation, State, Artifact, Relation, ShareContract 的 CRUD
- **EventCatalog**: Event, Channel, Subscriber, Position 的 CRUD，支持事件过滤和批量操作

***

## 阶段 5: Coordinator 服务层

**状态**: ✓ 已完成

### 任务清单

- [x] 实现 AgentCoordinator
- [x] 实现 MemoryCoordinator
- [x] 实现 EventCoordinator

### 文件位置

| 组件                | 文件路径                                   | 行数   |
| ----------------- | -------------------------------------- | ---- |
| AgentCoordinator  | `internal/agentcoord/agent_coord.go`   | 790  |
| MemoryCoordinator | `internal/memorycoord/memory_coord.go` | 1013 |
| EventCoordinator  | `internal/eventcoord/event_coord.go`   | 773  |

### 设计模式

所有 Coordinator 遵循统一的设计模式：

1. **生命周期管理**: Init() → Start() → Stop()
2. **状态机**: Initializing → Healthy → Stopping
3. **后台任务**: 使用 time.Ticker 定期执行维护任务
4. **并发安全**: 使用 sync.Once 确保初始化/启动/停止只执行一次
5. **错误处理**: 使用 cockroachdb/errors 进行错误包装
6. **日志记录**: 使用 zap 进行结构化日志记录

***

## 阶段 6: gRPC Server 层

**状态**: ✓ 已完成

### 任务清单

- [x] 创建 Server 目录结构
- [x] 实现 AgentService Server
- [x] 实现 MemoryService Server
- [x] 实现 EventService Server
- [x] 实现 Server 注册与启动逻辑

### 实现文件

| 文件 | 路径 | 行数 | 说明 |
|------|------|------|------|
| agent_server.go | `internal/distributed/tdb/agent_server.go` | ~350 | AgentService gRPC 实现 |
| memory_server.go | `internal/distributed/tdb/memory_server.go` | ~450 | MemoryService gRPC 实现 |
| event_server.go | `internal/distributed/tdb/event_server.go` | ~280 | EventService gRPC 实现 |
| server.go | `internal/distributed/tdb/server.go` | ~310 | Server 主结构和启动逻辑 |

### 职责

- 接收 gRPC 请求
- 参数校验和转换
- 调用 Coordinator 层方法
- 处理响应和错误
- 实现流式接口 (SubscribeEvents)

### 主要功能

**AgentServer**:
- CreateAgent, GetAgent, ListAgents, UpdateAgent, DeleteAgent
- CreateSession, GetSession, ListSessions, UpdateSession

**MemoryServer**:
- CreateMemory, GetMemory, UpdateMemory, DeleteMemory
- QueryMemories, SearchMemories
- GetRelations, CreateRelation

**EventServer**:
- AppendEvent, GetEvent, QueryEvents
- SubscribeEvents (流式接口)

**Server**:
- 整合所有 gRPC 服务
- gRPC Server 配置和启动
- Keepalive 和拦截器配置
- 生命周期管理 (Init, Start, Stop)

***

## 阶段 7: 服务注册与集成

**状态**: ✓ 已完成

### 任务清单

- [x] Milvus 服务框架集成
- [x] 配置管理
- [x] 启动流程集成

### 集成内容

#### 1. 角色定义 (pkg/util/typeutil/type.go)
- 添加 `TDBRole = "tdb"` 常量
- 将 TDBRole 添加到 `serverTypeSet`

#### 2. 配置管理 (pkg/util/paramtable/component_param.go)
- 添加 `tdbConfig` 结构体
  - `Enabled`: 是否启用 TDB (默认 false)
  - `GracefulStopTimeout`: 优雅停止超时时间
- 添加 `TDBGrpcServerCfg`: gRPC 服务器配置
- 添加 `TDBCfg.init()` 初始化

#### 3. 组件实现 (cmd/components/tdb.go)
- 实现 `TDB` 组件结构体
- 实现生命周期方法: `Prepare()`, `Run()`, `Stop()`
- 实现健康检查: `Health()`, `GetComponentStates()`

#### 4. 启动流程集成 (cmd/roles/roles.go)
- 在 `MilvusRoles` 结构体中添加 `EnableTDB` 字段
- 添加 `runTDB()` 方法
- 在 `Run()` 方法中启动 TDB 组件
- 更新 `enableComponents` 列表

#### 5. 命令行支持 (cmd/milvus/util.go)
- 在 `GetMilvusRoles()` 中添加 TDBRole 处理

#### 6. 监控指标 (pkg/metrics/tdb_metrics.go)
- 创建 TDB 监控指标文件
- 定义指标: AgentTotal, SessionTotal, MemoryTotal, EventTotal
- 定义指标: RequestLatency, RequestTotal, ActiveConnections
- 实现 `RegisterTDB()` 函数

### 启动方式

```bash
# 单独启动 TDB 组件
./milvus run tdb

# 在 Standalone 模式下启用 TDB (需配置 tdb.enabled=true)
./milvus run standalone

# 通过环境变量启用
ENABLE_TDB=true ./milvus run standalone
```

### 配置项

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| tdb.enabled | false | 是否启用 TDB 组件 |
| tdb.gracefulStopTimeout | 5 | 优雅停止超时时间(秒) |
| tdb.port | 未设置 | TDB gRPC 服务端口 |
| tdb.ip | 0.0.0.0 | TDB gRPC 服务 IP |

### 工作内容

1. **服务注册**: 将 TDB 服务注册到 Milvus 的组件体系中
2. **配置管理**: 添加 TDB 相关配置项到 Milvus 配置系统
3. **启动流程**: 在 Milvus 启动时初始化 TDB 组件
4. **依赖注入**: 将 TDB Server 注入到 Milvus 的 gRPC 服务中

***

## 阶段 8: 测试与优化

**状态**: ✓ 已完成

### 任务清单

- [x] 单元测试
- [x] 集成测试
- [x] 性能测试 (基准测试)
- [x] 性能优化建议文档

### 测试实现

#### 单元测试

| 测试文件 | 路径 | 覆盖范围 | 测试数 |
|---------|------|---------|--------|
| agent_coord_test.go | `internal/agentcoord/` | AgentCoord 所有方法 | 10 |
| memory_coord_test.go | `internal/memorycoord/` | MemoryCoord 所有方法 | 8 |
| event_coord_test.go | `internal/eventcoord/` | EventCoord 所有方法 | 3 |
| agent_server_test.go | `internal/distributed/tdb/` | AgentServer gRPC 接口 | 2 |

**测试特点**:
- 使用 testify/mock 进行依赖隔离
- 使用 testify/suite 组织测试用例
- 覆盖正常路径和错误路径
- 包含基准测试

#### 集成测试

| 测试文件 | 路径 | 覆盖范围 |
|---------|------|---------|
| tdb_integration_test.go | `tests/integration/` | 端到端业务流程 |

**测试场景**:
- `TestAgentLifecycle`: Agent 完整生命周期
- `TestSessionLifecycle`: Session 完整生命周期
- `TestMemoryOperations`: Memory CRUD 操作
- `TestEventOperations`: Event 操作
- `TestEventSubscription`: Event 订阅流式接口

#### 基准测试

| 基准测试 | 说明 |
|---------|------|
| BenchmarkCreateAgent | Agent ID 生成性能 |
| BenchmarkGenerateSessionID | Session ID 生成性能 |
| BenchmarkCreateMemory | Memory ID 生成性能 |
| BenchmarkGenerateRelationID | Relation ID 生成性能 |
| BenchmarkAppendEvent | Event ID 生成性能 |

### 性能优化建议

详细优化建议见 `docs/TDB/fixed1.07.md`，包括：

1. **存储层优化**: etcd 批量操作、本地缓存、索引优化
2. **查询优化**: 分页优化、预加载、字段选择
3. **向量搜索优化**: Milvus 索引、批量搜索、结果缓存
4. **gRPC 优化**: 连接池、消息压缩、流式传输
5. **并发优化**: 协程池、限流机制、锁优化
6. **内存优化**: 对象池、预分配、GC 优化
7. **监控与调优**: 性能指标、性能分析

### 性能目标

| 指标 | 目标值 | 说明 |
|------|--------|------|
| P99 延迟 | < 100ms | 99% 请求延迟小于 100ms |
| 吞吐量 | > 10000 QPS | 单实例 QPS |
| 内存使用 | < 2GB | 单实例内存使用 |
| CPU 使用 | < 80% | 正常负载 CPU 使用 |
| 错误率 | < 0.1% | 请求错误率 |

***

## 进度统计

| 阶段             | 状态    | 完成度  | 代码行数   |
| -------------- | ----- | ---- | ------ |
| 1. Proto 定义    | ✓ 完成  | 100% | \~500  |
| 2. etcd 常量     | ✓ 完成  | 100% | \~200  |
| 3. Go 代码生成     | ✓ 完成  | 100% | \~3500 |
| 4. KV Catalog  | ✓ 完成  | 100% | \~1500 |
| 5. Coordinator | ✓ 完成  | 100% | \~2600 |
| 6. gRPC Server | ✓ 完成  | 100% | \~1400 |
| 7. 服务集成        | ✓ 完成  | 100% | \~400  |
| 8. 测试优化        | ✓ 完成  | 100% | \~1500 |

**整体完成度**: 100%

***

## 文档记录

| 文档             | 路径                      | 说明     |
| -------------- | ----------------------- | ------ |
| 项目规划           | `docs/TDB/plan1.0.md`   | 本文档    |
| Schema 设计      | `docs/TDB/fixed1.01.md` | 数据模型设计 |
| KV Catalog 设计  | `docs/TDB/fixed1.02.md` | 存储层设计  |
| Coordinator 设计 | `docs/TDB/fixed1.03.md` | 服务层设计  |
| Proto Go 代码生成 | `docs/TDB/fixed1.04.md` | Protobuf 代码生成记录 |
| gRPC Server 设计 | `docs/TDB/fixed1.05.md` | gRPC 服务层设计 |
| 服务集成设计 | `docs/TDB/fixed1.06.md` | Milvus 服务集成设计 |
| 测试与优化 | `docs/TDB/fixed1.07.md` | 测试策略和性能优化 |

***

## 后续行动计划

### 下一步 (优先级高)

1. ~~补全 protobuf Go 代码生成 (阶段 3)~~ ✓ 已完成
2. ~~实现 gRPC Server 层 (阶段 6)~~ ✓ 已完成
3. ~~服务注册与集成 (阶段 7)~~ ✓ 已完成
4. 单元测试和集成测试 (阶段 8)

### 近期计划 (优先级中)

1. 服务注册与集成 (阶段 7)
2. 单元测试 (阶段 8)

### 远期计划 (优先级低)

1. 集成测试和性能测试
2. 文档完善

***

## 变更记录

| 版本  | 日期         | 变更内容             | 作者 |
| --- | ---------- | ---------------- | -- |
| 1.0 | 2026-03-14 | 初始版本，记录已完成和待完成工作 | -  |
| 1.1 | 2026-03-15 | 完成阶段 3: Protobuf Go 代码生成 | -  |
| 1.2 | 2026-03-15 | 完成阶段 6: gRPC Server 层实现 | -  |
| 1.3 | 2026-03-15 | 完成阶段 7: 服务注册与集成 | -  |
| 1.4 | 2026-03-15 | 完成阶段 8: 测试与优化 | -  |

***

## 备注

- 本文档为动态文档，后续直接在本文档上更新进度
- 每完成一个阶段，更新对应状态的复选框
- 如需调整规划，在此文档中直接修改并记录变更原因

