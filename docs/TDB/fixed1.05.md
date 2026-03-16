# TDB gRPC Server 层设计文档

> 本文档记录 TDB 项目 gRPC Server 层的详细设计
> 创建日期: 2026-03-15

***

## 概述

TDB gRPC Server 层是 TDB 系统的对外接口层，负责接收和处理客户端的 gRPC 请求。它实现了三个核心服务的 gRPC 接口：

- **AgentService**: Agent 和 Session 管理
- **MemoryService**: 记忆和关系管理
- **EventService**: 事件日志和订阅

***

## 架构设计

### 目录结构

```
internal/distributed/tdb/
├── agent_server.go   # AgentService 实现
├── memory_server.go  # MemoryService 实现
├── event_server.go   # EventService 实现
└── server.go         # Server 主结构和启动逻辑
```

### 组件关系

```
┌─────────────────────────────────────────────────────────────┐
│                        gRPC Client                          │
└──────────────────────────┬──────────────────────────────────┘
                           │ gRPC Protocol
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                      TDB Server                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  gRPC Server                          │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐     │  │
│  │  │AgentServer  │ │MemoryServer │ │EventServer  │     │  │
│  │  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘     │  │
│  └─────────┼───────────────┼───────────────┼────────────┘  │
└────────────┼───────────────┼───────────────┼───────────────┘
             │               │               │
             ▼               ▼               ▼
        ┌─────────┐    ┌─────────┐    ┌─────────┐
        │AgentCoord│    │MemoryCoord│    │EventCoord│
        └────┬────┘    └────┬────┘    └────┬────┘
             │               │               │
             └───────────────┼───────────────┘
                             ▼
                    ┌─────────────────┐
                    │   KV Catalog    │
                    └─────────────────┘
```

***

## 核心组件

### 1. Server (server.go)

**职责**: 整合所有 gRPC 服务，管理生命周期

**主要结构**:
```go
type Server struct {
    agentServer  *AgentServer
    memoryServer *MemoryServer
    eventServer  *EventServer
    
    grpcServer *grpc.Server
    listener   *netutil.NetListener
    
    agentCoord  *agentcoord.AgentCoord
    memoryCoord *memorycoord.MemoryCoord
    eventCoord  *eventcoord.EventCoord
    
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}
```

**生命周期方法**:
- `NewServer()`: 创建 Server 实例，初始化 Coordinators
- `Prepare()`: 准备网络监听器
- `Run()`: 运行 Server (Init + Start)
- `Stop()`: 优雅停止 Server

**gRPC 配置**:
- Keepalive: MinTime=5s, Time=60s, Timeout=10s
- MaxRecvMsgSize: 可配置
- MaxSendMsgSize: 可配置
- 拦截器: TraceLogger, ClusterValidation, ServerIDValidation

### 2. AgentServer (agent_server.go)

**职责**: 实现 AgentService gRPC 接口

**实现方法**:
| 方法 | 功能 | 对应 Coordinator 方法 |
|------|------|---------------------|
| CreateAgent | 创建 Agent | CreateAgent |
| GetAgent | 获取 Agent | GetAgent |
| ListAgents | 列出 Agents | ListAgents |
| UpdateAgent | 更新 Agent | UpdateAgent |
| DeleteAgent | 删除 Agent | DeleteAgent |
| CreateSession | 创建 Session | CreateSession |
| GetSession | 获取 Session | GetSession |
| ListSessions | 列出 Sessions | ListSessions |
| UpdateSession | 更新 Session | UpdateSession |

**数据转换**:
- `convertAgentToProto()`: 将 models.Agent 转换为 agentpb.AgentInfo
- `convertSessionToProto()`: 将 models.Session 转换为 agentpb.SessionInfo

### 3. MemoryServer (memory_server.go)

**职责**: 实现 MemoryService gRPC 接口

**实现方法**:
| 方法 | 功能 | 对应 Coordinator 方法 |
|------|------|---------------------|
| CreateMemory | 创建记忆 | CreateMemory |
| GetMemory | 获取记忆 | GetMemory |
| UpdateMemory | 更新记忆 | UpdateMemory |
| DeleteMemory | 删除记忆 | DeleteMemory |
| QueryMemories | 查询记忆 | QueryMemories |
| SearchMemories | 向量搜索 | SearchMemories |
| GetRelations | 获取关系 | GetRelations |
| CreateRelation | 创建关系 | CreateRelation |

**数据转换**:
- `convertMemoryToProto()`: 将 models.Memory 转换为 memorypb.Memory
- `convertMemoryPolicyToProto()`: 转换 MemoryPolicy
- `convertMemoryAdaptationToProto()`: 转换 MemoryAdaptation
- `convertRelationToProto()`: 将 models.Relation 转换为 memorypb.Relation

### 4. EventServer (event_server.go)

**职责**: 实现 EventService gRPC 接口

**实现方法**:
| 方法 | 功能 | 对应 Coordinator 方法 |
|------|------|---------------------|
| AppendEvent | 追加事件 | AppendEvent |
| GetEvent | 获取事件 | GetEvent |
| QueryEvents | 查询事件 | QueryEvents |
| SubscribeEvents | 订阅事件流 | SubscribeEvents |

**流式接口 SubscribeEvents**:
- 创建事件订阅
- 通过 gRPC 流推送事件
- 支持客户端取消订阅
- 自动清理订阅资源

**数据转换**:
- `convertEventToProto()`: 将 models.Event 转换为 eventpb.Event

***

## 错误处理

### 错误码映射

| 场景 | ErrorCode | 说明 |
|------|-----------|------|
| 成功 | ErrorCode_Success | 操作成功 |
| 通用错误 | ErrorCode_UnexpectedError | 未预期的错误 |
| Key 不存在 | 自定义错误 | 使用 ErrKeyNotFound |

### 错误处理模式

```go
result, err := coord.Method(ctx, params)
if err != nil {
    log.Error("failed to execute method", zap.Error(err))
    return &pb.Response{
        Status: &commonpb.Status{
            ErrorCode: commonpb.ErrorCode_UnexpectedError,
            Reason:    err.Error(),
        },
    }, nil
}
```

***

## 数据类型转换

### Agent 状态映射

| models.AgentState | agentpb.AgentState |
|-------------------|-------------------|
| AgentStateCreating | AGENT_CREATING |
| AgentStateActive | AGENT_ACTIVE |
| AgentStatePaused | AGENT_PAUSED |
| AgentStateTerminated | AGENT_TERMINATED |

### Session 状态映射

| models.SessionState | agentpb.SessionState |
|---------------------|---------------------|
| SessionStateCreating | SESSION_CREATING |
| SessionStateActive | SESSION_ACTIVE |
| SessionStatePaused | SESSION_PAUSED |
| SessionStateCompleted | SESSION_COMPLETED |
| SessionStateFailed | SESSION_FAILED |

### Memory 类型映射

| models.MemoryType | memorypb.MemoryType |
|-------------------|---------------------|
| MemoryTypeEpisodic | EPISODIC |
| MemoryTypeSemantic | SEMANTIC |
| MemoryTypeProcedural | PROCEDURAL |
| MemoryTypeSocial | SOCIAL |
| MemoryTypeReflective | REFLECTIVE |

### Event 类型映射

| models.EventType | eventpb.EventType |
|------------------|-------------------|
| UserMessage | USER_MESSAGE |
| AssistantMessage | ASSISTANT_MESSAGE |
| ToolCallIssued | TOOL_CALL_ISSUED |
| MemoryWriteRequested | MEMORY_WRITE_REQUESTED |
| ... | ... |

***

## 代码统计

| 文件 | 行数 | 方法数 | 说明 |
|------|------|--------|------|
| agent_server.go | ~350 | 9 | Agent 服务实现 |
| memory_server.go | ~450 | 8 | Memory 服务实现 |
| event_server.go | ~280 | 4 | Event 服务实现 |
| server.go | ~310 | 10 | Server 主结构 |
| **总计** | **~1390** | **31** | **gRPC Server 层** |

***

## 使用示例

### 启动 TDB Server

```go
ctx := context.Background()
factory := dependency.NewDefaultFactory(true)

server, err := tdb.NewServer(ctx, factory)
if err != nil {
    log.Fatal("failed to create server", zap.Error(err))
}

if err := server.Prepare(); err != nil {
    log.Fatal("failed to prepare server", zap.Error(err))
}

if err := server.Run(); err != nil {
    log.Fatal("failed to run server", zap.Error(err))
}

// 优雅停止
defer server.Stop()
```

### 调用 AgentService

```go
client := agentpb.NewAgentServiceClient(conn)

resp, err := client.CreateAgent(ctx, &agentpb.CreateAgentRequest{
    TenantId:            "tenant-001",
    WorkspaceId:         "workspace-001",
    AgentType:           "assistant",
    RoleProfile:         "helpful assistant",
    CapabilitySet:       []string{"chat", "search"},
    DefaultMemoryPolicy: "standard",
})
```

### 订阅事件流

```go
stream, err := client.SubscribeEvents(ctx, &eventpb.SubscribeEventsRequest{
    SubscriberId: "subscriber-001",
    Filter: &eventpb.EventFilter{
        AgentIds: []string{"agent-001"},
    },
})

for {
    resp, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Error("stream error", zap.Error(err))
        break
    }
    // 处理事件
    event := resp.GetEvent()
    log.Info("received event", zap.String("event_id", event.EventId))
}
```

***

## 注意事项

1. **流式接口**: SubscribeEvents 是双向流式接口，需要正确处理连接生命周期
2. **错误处理**: 所有错误都转换为 commonpb.Status 返回，不直接返回 Go error
3. **数据转换**: 注意枚举类型的映射关系，确保 proto 和 model 一致
4. **资源清理**: Server.Stop() 会优雅停止所有组件，确保资源正确释放
5. **并发安全**: Server 内部使用 sync.WaitGroup 管理 goroutine 生命周期

***

## 后续工作

1. **服务注册**: 将 TDB Server 注册到 Milvus 的服务发现系统
2. **配置管理**: 添加更多可配置参数（如超时、重试策略）
3. **监控指标**: 添加 gRPC 调用指标收集
4. **限流保护**: 实现请求限流和熔断机制
5. **认证授权**: 添加 gRPC 认证和权限校验

***

## 参考

- [gRPC Go 文档](https://grpc.io/docs/languages/go/)
- [Protocol Buffers 文档](https://protobuf.dev/)
- [Milvus 分布式服务设计](https://github.com/milvus-io/milvus/tree/master/internal/distributed)

***

## 变更记录

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|----------|------|
| 1.0 | 2026-03-15 | 初始版本，记录 gRPC Server 层实现 | - |
