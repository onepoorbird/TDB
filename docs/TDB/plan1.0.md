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

**状态**: ⚠️ 部分完成

### 任务清单

- [ ] 生成 agent.pb.go
- [ ] 生成 agent\_grpc.pb.go
- [ ] 生成 memory.pb.go
- [ ] 生成 memory\_grpc.pb.go
- [ ] 生成 event.pb.go
- [ ] 生成 event\_grpc.pb.go

### 已生成文件

| 文件          | 路径                              |
| ----------- | ------------------------------- |
| agent.pb.go | `pkg/proto/agentpb/agent.pb.go` |

### 待生成文件

| 文件                 | 目标路径                                   |
| ------------------ | -------------------------------------- |
| agent\_grpc.pb.go  | `pkg/proto/agentpb/agent_grpc.pb.go`   |
| memory.pb.go       | `pkg/proto/memorypb/memory.pb.go`      |
| memory\_grpc.pb.go | `pkg/proto/memorypb/memory_grpc.pb.go` |
| event.pb.go        | `pkg/proto/eventpb/event.pb.go`        |
| event\_grpc.pb.go  | `pkg/proto/eventpb/event_grpc.pb.go`   |

### 生成命令

```bash
# Agent
protoc --go_out=pkg/proto/agentpb --go_opt=paths=source_relative \
       --go-grpc_out=pkg/proto/agentpb --go-grpc_opt=paths=source_relative \
       pkg/proto/agent.proto

# Memory  
protoc --go_out=pkg/proto/memorypb --go_opt=paths=source_relative \
       --go-grpc_out=pkg/proto/memorypb --go-grpc_opt=paths=source_relative \
       pkg/proto/memory.proto

# Event
protoc --go_out=pkg/proto/eventpb --go_opt=paths=source_relative \
       --go-grpc_out=pkg/proto/eventpb --go-grpc_opt=paths=source_relative \
       pkg/proto/event.proto
```

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

**状态**: ✗ 待实现

### 任务清单

- [ ] 实现 AgentService Server
- [ ] 实现 MemoryService Server
- [ ] 实现 EventService Server
- [ ] 实现 Server 注册与启动

### 计划文件位置

| 组件            | 文件路径                                        | 说明                       |
| ------------- | ------------------------------------------- | ------------------------ |
| Agent Server  | `internal/distributed/tdb/agent_server.go`  | 实现 AgentService gRPC 接口  |
| Memory Server | `internal/distributed/tdb/memory_server.go` | 实现 MemoryService gRPC 接口 |
| Event Server  | `internal/distributed/tdb/event_server.go`  | 实现 EventService gRPC 接口  |
| Server        | `internal/distributed/tdb/server.go`        | Server 注册与启动逻辑           |

### 职责

- 接收 gRPC 请求
- 参数校验和转换
- 调用 Coordinator 层方法
- 处理响应和错误
- 实现流式接口 (SubscribeEvents)

***

## 阶段 7: 服务注册与集成

**状态**: ✗ 待实现

### 任务清单

- [ ] Milvus 服务框架集成
- [ ] 配置管理
- [ ] 启动流程集成

### 工作内容

1. **服务注册**: 将 TDB 服务注册到 Milvus 的组件体系中
2. **配置管理**: 添加 TDB 相关配置项到 Milvus 配置系统
3. **启动流程**: 在 Milvus 启动时初始化 TDB 组件
4. **依赖注入**: 将 TDB Server 注入到 Milvus 的 gRPC 服务中

***

## 阶段 8: 测试与优化

**状态**: ✗ 待实现

### 任务清单

- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能测试
- [ ] 文档完善

### 工作内容

1. **单元测试**: Coordinator 层各方法的单元测试
2. **集成测试**: 端到端测试，验证完整流程
3. **性能测试**: 压力测试和性能优化
4. **文档完善**: API 文档、使用指南、架构文档

***

## 进度统计

| 阶段             | 状态    | 完成度  | 代码行数   |
| -------------- | ----- | ---- | ------ |
| 1. Proto 定义    | ✓ 完成  | 100% | \~500  |
| 2. etcd 常量     | ✓ 完成  | 100% | \~200  |
| 3. Go 代码生成     | ⚠️ 部分 | 33%  | \~1000 |
| 4. KV Catalog  | ✓ 完成  | 100% | \~1500 |
| 5. Coordinator | ✓ 完成  | 100% | \~2600 |
| 6. gRPC Server | ✗ 未开始 | 0%   | -      |
| 7. 服务集成        | ✗ 未开始 | 0%   | -      |
| 8. 测试优化        | ✗ 未开始 | 0%   | -      |

**整体完成度**: 约 60%

***

## 文档记录

| 文档             | 路径                      | 说明     |
| -------------- | ----------------------- | ------ |
| 项目规划           | `docs/TDB/plan1.0.md`   | 本文档    |
| Schema 设计      | `docs/TDB/fixed1.01.md` | 数据模型设计 |
| KV Catalog 设计  | `docs/TDB/fixed1.02.md` | 存储层设计  |
| Coordinator 设计 | `docs/TDB/fixed1.03.md` | 服务层设计  |

***

## 后续行动计划

### 下一步 (优先级高)

1. 补全 protobuf Go 代码生成 (阶段 3)
2. 实现 gRPC Server 层 (阶段 6)

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

***

## 备注

- 本文档为动态文档，后续直接在本文档上更新进度
- 每完成一个阶段，更新对应状态的复选框
- 如需调整规划，在此文档中直接修改并记录变更原因

