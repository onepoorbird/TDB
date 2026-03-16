# TDB 测试与优化文档

> 本文档记录 TDB 项目的测试策略和性能优化建议
> 创建日期: 2026-03-15

***

## 测试覆盖

### 单元测试

#### 1. Coordinator 层测试

**文件**: `internal/agentcoord/agent_coord_test.go`
- 测试 AgentCoord 的所有方法
- 使用 MockCatalog 隔离存储层依赖
- 覆盖正常路径和错误路径
- 包含基准测试

**测试方法**:
- `TestCreateAgent`: 创建 Agent
- `TestGetAgent`: 获取 Agent
- `TestGetAgent_NotFound`: 获取不存在的 Agent
- `TestListAgents`: 列出 Agents
- `TestUpdateAgent`: 更新 Agent
- `TestDeleteAgent`: 删除 Agent
- `TestCreateSession`: 创建 Session
- `TestGetSession`: 获取 Session
- `TestListSessions`: 列出 Sessions
- `TestUpdateSession`: 更新 Session

**文件**: `internal/memorycoord/memory_coord_test.go`
- 测试 MemoryCoord 的所有方法
- 覆盖 Memory 和 Relation 操作

**测试方法**:
- `TestCreateMemory`: 创建记忆
- `TestGetMemory`: 获取记忆
- `TestUpdateMemory`: 更新记忆
- `TestDeleteMemory`: 删除记忆
- `TestQueryMemories`: 查询记忆
- `TestSearchMemories`: 搜索记忆
- `TestGetRelations`: 获取关系
- `TestCreateRelation`: 创建关系

**文件**: `internal/eventcoord/event_coord_test.go`
- 测试 EventCoord 的所有方法
- 覆盖事件操作和订阅

**测试方法**:
- `TestAppendEvent`: 追加事件
- `TestGetEvent`: 获取事件
- `TestQueryEvents`: 查询事件

#### 2. gRPC Server 层测试

**文件**: `internal/distributed/tdb/agent_server_test.go`
- 测试 AgentServer 的 gRPC 接口
- 使用 MockAgentCoord 隔离 Coordinator 层

**测试方法**:
- `TestCreateAgent`: 创建 Agent gRPC 调用
- `TestGetAgent`: 获取 Agent gRPC 调用

### 集成测试

**文件**: `tests/integration/tdb_integration_test.go`
- 端到端测试，需要运行中的 TDB 服务器
- 测试完整的业务流程

**测试方法**:
- `TestAgentLifecycle`: Agent 完整生命周期
- `TestSessionLifecycle`: Session 完整生命周期
- `TestMemoryOperations`: Memory CRUD 操作
- `TestEventOperations`: Event 操作
- `TestEventSubscription`: Event 订阅流式接口

### 基准测试

```go
// AgentCoord 基准测试
func BenchmarkCreateAgent(b *testing.B)
func BenchmarkGenerateSessionID(b *testing.B)

// MemoryCoord 基准测试
func BenchmarkCreateMemory(b *testing.B)
func BenchmarkGenerateRelationID(b *testing.B)

// EventCoord 基准测试
func BenchmarkAppendEvent(b *testing.B)
```

***

## 性能优化建议

### 1. 存储层优化

#### etcd 优化
- **批量操作**: 使用 Txn 批量写入减少网络往返
- **压缩策略**: 定期压缩历史版本减少存储
- **Watch 优化**: 使用前缀 Watch 减少连接数

```go
// 批量写入示例
func (c *Catalog) BatchCreateAgents(agents []*models.Agent) error {
    // 使用 Txn 批量写入
}
```

#### 缓存策略
- **本地缓存**: 热点数据本地缓存
- **缓存失效**: 基于版本号的缓存失效策略
- **LRU 缓存**: 使用 LRU 算法管理缓存

```go
// 缓存示例
type CachedCatalog struct {
    catalog Catalog
    cache   *lru.Cache
    mu      sync.RWMutex
}
```

### 2. 查询优化

#### 索引优化
- **二级索引**: 为常用查询字段建立二级索引
- **复合索引**: 多字段查询使用复合索引
- **前缀索引**: 长字符串字段使用前缀索引

#### 分页优化
- **游标分页**: 大数据量使用游标分页
- **预加载**: 预加载下一页数据
- **限制返回字段**: 只返回需要的字段

### 3. 向量搜索优化

#### Milvus 集成优化
- **Collection 设计**: 合理设计 Collection 和 Partition
- **索引类型**: 选择合适的索引类型 (IVF_FLAT, HNSW 等)
- **批量搜索**: 批量向量搜索减少 RPC 调用

```go
// 批量搜索示例
func (c *MemoryCoord) BatchSearchMemories(vectors [][]float32, topK int) ([]*MemorySearchResult, error) {
    // 批量搜索减少 RPC 调用
}
```

### 4. gRPC 优化

#### 连接管理
- **连接池**: 使用连接池管理 gRPC 连接
- **Keepalive**: 配置合理的 Keepalive 参数
- **负载均衡**: 客户端负载均衡

#### 序列化优化
- **Protobuf**: 使用 Protobuf 高效序列化
- **压缩**: 大消息启用压缩
- **流式传输**: 大数据量使用流式传输

```go
// gRPC 配置优化
grpcOpts := []grpc.ServerOption{
    grpc.MaxRecvMsgSize(100 * 1024 * 1024), // 100MB
    grpc.MaxSendMsgSize(100 * 1024 * 1024),
    grpc.KeepaliveParams(keepalive.ServerParameters{
        Time:    60 * time.Second,
        Timeout: 10 * time.Second,
    }),
}
```

### 5. 并发优化

#### 协程池
- **工作池**: 使用工作池管理并发任务
- **限流**: 限制并发数防止资源耗尽
- **背压**: 实现背压机制

```go
// 工作池示例
type WorkerPool struct {
    workers int
    jobs    chan func()
    wg      sync.WaitGroup
}
```

#### 锁优化
- **读写锁**: 读多写少场景使用 RWMutex
- **分段锁**: 高并发场景使用分段锁
- **无锁结构**: 使用原子操作和无锁队列

### 6. 内存优化

#### 对象池
- **sync.Pool**: 使用对象池复用对象
- **内存预分配**: 预分配内存减少 GC

```go
// 对象池示例
var agentPool = sync.Pool{
    New: func() interface{} {
        return &models.Agent{}
    },
}
```

#### GC 优化
- **减少对象分配**: 减少临时对象分配
- **大对象处理**: 大对象单独处理避免堆分配

### 7. 监控与调优

#### 性能指标
- **延迟**: P50, P95, P99 延迟
- **吞吐量**: QPS, TPS
- **资源使用**: CPU, 内存, 网络, 磁盘

#### 性能分析
- **pprof**: 使用 pprof 进行性能分析
- **trace**: 使用 trace 分析协程调度
- **火焰图**: 生成火焰图定位热点

```go
// 启用 pprof
import _ "net/http/pprof"

func init() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}
```

***

## 测试运行

### 运行单元测试

```bash
# 运行所有单元测试
go test ./internal/agentcoord/... -v
go test ./internal/memorycoord/... -v
go test ./internal/eventcoord/... -v
go test ./internal/distributed/tdb/... -v

# 运行基准测试
go test ./internal/agentcoord/... -bench=.
go test ./internal/memorycoord/... -bench=.
```

### 运行集成测试

```bash
# 需要先启动 TDB 服务
go test ./tests/integration/... -v
```

### 代码覆盖率

```bash
# 生成覆盖率报告
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

***

## 优化检查清单

### 存储层
- [ ] etcd 批量操作优化
- [ ] 本地缓存实现
- [ ] 索引优化
- [ ] 压缩策略

### 查询层
- [ ] 分页优化
- [ ] 预加载策略
- [ ] 字段选择

### 向量搜索
- [ ] Milvus 索引优化
- [ ] 批量搜索
- [ ] 结果缓存

### gRPC
- [ ] 连接池
- [ ] 消息压缩
- [ ] 流式传输

### 并发
- [ ] 协程池
- [ ] 限流机制
- [ ] 锁优化

### 内存
- [ ] 对象池
- [ ] 预分配
- [ ] GC 优化

### 监控
- [ ] 性能指标收集
- [ ] 告警配置
- [ ] 日志优化

***

## 性能目标

| 指标 | 目标值 | 说明 |
|------|--------|------|
| P99 延迟 | < 100ms | 99% 请求延迟小于 100ms |
| 吞吐量 | > 10000 QPS | 单实例 QPS |
| 内存使用 | < 2GB | 单实例内存使用 |
| CPU 使用 | < 80% | 正常负载 CPU 使用 |
| 错误率 | < 0.1% | 请求错误率 |

***

## 参考

- [Go 性能优化指南](https://go.dev/doc/diagnostics)
- [etcd 性能优化](https://etcd.io/docs/v3.5/tuning/)
- [Milvus 性能优化](https://milvus.io/docs/performance_faq.md)
- [gRPC 最佳实践](https://grpc.io/docs/guides/performance/)

***

## 变更记录

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|----------|------|
| 1.0 | 2026-03-15 | 初始版本，记录测试和优化 | - |
