# TDB Milvus 服务集成设计文档

> 本文档记录 TDB 项目与 Milvus 服务框架的集成过程
> 创建日期: 2026-03-15

***

## 概述

TDB (Temporal Database) 作为 Milvus 的一个独立组件，需要集成到 Milvus 的服务框架中。本文档详细记录了集成的各个步骤和实现细节。

***

## 集成架构

```
┌─────────────────────────────────────────────────────────────────┐
│                     Milvus Cluster                               │
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │   Proxy     │  │  MixCoord   │  │  QueryNode  │  │  TDB    │ │
│  │             │  │             │  │             │  │         │ │
│  │  gRPC/API   │  │  Coordinator│  │  Search     │  │ Agent   │ │
│  │  Gateway    │  │  Management │  │  Execution  │  │ Memory  │ │
│  │             │  │             │  │             │  │ Event   │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └────┬────┘ │
│         │                │                │              │      │
│         └────────────────┴────────────────┴──────────────┘      │
│                              │                                   │
│                         etcd Cluster                            │
│                    (Service Discovery & KV)                      │
└─────────────────────────────────────────────────────────────────┘
```

***

## 集成步骤

### 1. 角色定义

**文件**: `pkg/util/typeutil/type.go`

添加 TDB 角色常量：

```go
const (
    // ... 其他角色
    
    // TDBRole is a constant represent TDB (Temporal Database)
    TDBRole = "tdb"
)
```

将 TDBRole 添加到服务类型集合：

```go
var (
    serverTypeSet = NewSet(
        // ... 其他角色
        CDCRole,
        TDBRole,  // 新增
    )
)
```

**作用**: 让 Milvus 识别 TDB 作为有效的服务类型。

### 2. 配置管理

**文件**: `pkg/util/paramtable/component_param.go`

添加 TDB 配置结构体：

```go
// ComponentParam 中添加
type ComponentParam struct {
    // ... 其他配置
    TDBCfg         tdbConfig
    TDBGrpcServerCfg GrpcServerConfig
}
```

定义 tdbConfig：

```go
type tdbConfig struct {
    Enabled             ParamItem `refreshable:"false"`
    GracefulStopTimeout ParamItem `refreshable:"true"`
}

func (p *tdbConfig) init(base *BaseTable) {
    p.Enabled = ParamItem{
        Key:          "tdb.enabled",
        Version:      "2.6.0",
        DefaultValue: "false",
        Doc:          "Enable TDB (Temporal Database) component",
        Export:       true,
    }
    p.Enabled.Init(base.mgr)

    p.GracefulStopTimeout = ParamItem{
        Key:          "tdb.gracefulStopTimeout",
        Version:      "2.6.0",
        DefaultValue: "5",
        Doc:          "TDB graceful stop timeout in seconds",
        Export:       true,
    }
    p.GracefulStopTimeout.Init(base.mgr)
}
```

初始化配置：

```go
func (p *ComponentParam) init(bt *BaseTable) {
    // ... 其他初始化
    p.TDBCfg.init(bt)
    p.TDBGrpcServerCfg.Init("tdb", bt)
}
```

**作用**: 提供 TDB 的配置管理能力。

### 3. 组件实现

**文件**: `cmd/components/tdb.go`

实现 TDB 组件：

```go
package components

// TDB implements TDB grpc server
type TDB struct {
    ctx context.Context
    svr *tdb.Server
}

// NewTDB creates a new TDB component
func NewTDB(ctx context.Context, factory dependency.Factory) (*TDB, error) {
    svr, err := tdb.NewServer(ctx, factory)
    if err != nil {
        return nil, err
    }
    return &TDB{
        ctx: ctx,
        svr: svr,
    }, nil
}

// Prepare prepares the TDB component
func (t *TDB) Prepare() error {
    return t.svr.Prepare()
}

// Run starts the TDB service
func (t *TDB) Run() error {
    if err := t.svr.Run(); err != nil {
        log.Ctx(t.ctx).Error("TDB starts error", zap.Error(err))
        return err
    }
    log.Ctx(t.ctx).Info("TDB successfully started")
    return nil
}

// Stop terminates the TDB service
func (t *TDB) Stop() error {
    timeout := paramtable.Get().TDBCfg.GracefulStopTimeout.GetAsDuration(time.Second)
    return exitWhenStopTimeout(t.svr.Stop, timeout)
}

// Health returns TDB's health status
func (t *TDB) Health(ctx context.Context) commonpb.StateCode {
    return commonpb.StateCode_Healthy
}

// GetName returns the component name
func (t *TDB) GetName() string {
    return typeutil.TDBRole
}
```

**作用**: 实现 TDB 组件的生命周期管理。

### 4. 启动流程集成

**文件**: `cmd/roles/roles.go`

在 MilvusRoles 结构体中添加 TDB 支持：

```go
type MilvusRoles struct {
    // ... 其他字段
    EnableTDB           bool `env:"ENABLE_TDB"`
}
```

添加 runTDB 方法：

```go
func (mr *MilvusRoles) runTDB(ctx context.Context, localMsg bool) *conc.Future[component] {
    return runComponent(ctx, localMsg, components.NewTDB, metrics.RegisterTDB)
}
```

在 Run 方法中启动 TDB：

```go
func (mr *MilvusRoles) Run() {
    // ... 其他初始化

    if mr.EnableTDB {
        paramtable.SetLocalComponentEnabled(typeutil.TDBRole)
        tdb := mr.runTDB(ctx, local)
        componentFutureMap[typeutil.TDBRole] = tdb
    }

    // ... 等待所有组件就绪
}
```

**作用**: 在 Milvus 启动流程中集成 TDB 组件。

### 5. 命令行支持

**文件**: `cmd/milvus/util.go`

在 GetMilvusRoles 函数中添加 TDBRole 处理：

```go
func GetMilvusRoles(args []string, flags *flag.FlagSet) *MilvusRoles {
    // ...
    switch serverType {
    // ... 其他角色
    case typeutil.TDBRole:
        role.EnableTDB = true
    default:
        fmt.Fprintf(os.Stderr, "Unknown server type = %s\n%s", serverType, getHelp())
        os.Exit(-1)
    }
    return role
}
```

**作用**: 支持通过命令行启动 TDB 组件。

### 6. 监控指标

**文件**: `pkg/metrics/tdb_metrics.go`

定义 TDB 监控指标：

```go
var (
    TDBAgentTotal = prometheus.NewGaugeVec(...)
    TDBSessionTotal = prometheus.NewGaugeVec(...)
    TDBMemoryTotal = prometheus.NewGaugeVec(...)
    TDBEventTotal = prometheus.NewCounterVec(...)
    TDBRequestLatency = prometheus.NewHistogramVec(...)
    TDBRequestTotal = prometheus.NewCounterVec(...)
    TDBActiveConnections = prometheus.NewGaugeVec(...)
)

func RegisterTDB(registry *prometheus.Registry) {
    registry.MustRegister(TDBAgentTotal)
    registry.MustRegister(TDBSessionTotal)
    registry.MustRegister(TDBMemoryTotal)
    registry.MustRegister(TDBEventTotal)
    registry.MustRegister(TDBRequestLatency)
    registry.MustRegister(TDBRequestTotal)
    registry.MustRegister(TDBActiveConnections)
}
```

**作用**: 提供 TDB 组件的监控指标。

***

## 启动方式

### 单独启动 TDB

```bash
./milvus run tdb
```

### 在 Standalone 模式下启用 TDB

1. 修改配置文件 `milvus.yaml`:

```yaml
tdb:
  enabled: true
  gracefulStopTimeout: 5
```

2. 启动 Standalone:

```bash
./milvus run standalone
```

### 通过环境变量启用

```bash
ENABLE_TDB=true ./milvus run standalone
```

***

## 配置项

| 配置项 | 配置键 | 默认值 | 说明 |
|--------|--------|--------|------|
| 启用 TDB | tdb.enabled | false | 是否启用 TDB 组件 |
| 优雅停止超时 | tdb.gracefulStopTimeout | 5 | 优雅停止超时时间(秒) |
| gRPC IP | tdb.ip | 0.0.0.0 | TDB gRPC 服务 IP |
| gRPC Port | tdb.port | 自动分配 | TDB gRPC 服务端口 |
| 最大接收消息大小 | tdb.serverMaxRecvSize | 100MB | gRPC 最大接收消息大小 |
| 最大发送消息大小 | tdb.serverMaxSendSize | 100MB | gRPC 最大发送消息大小 |

***

## 服务发现

TDB 组件通过 etcd 进行服务注册和发现：

1. **服务注册**: TDB 启动时向 etcd 注册服务信息
2. **服务发现**: 其他组件通过 etcd 查询 TDB 服务地址
3. **健康检查**: TDB 定期向 etcd 发送心跳

服务注册路径：
```
/milvus/session/tdb-<server_id>
```

***

## 健康检查

TDB 组件实现了健康检查接口：

```go
func (t *TDB) Health(ctx context.Context) commonpb.StateCode {
    return commonpb.StateCode_Healthy
}
```

健康状态：
- **Healthy**: 服务正常运行
- **Abnormal**: 服务异常

***

## 代码统计

| 文件 | 行数 | 说明 |
|------|------|------|
| pkg/util/typeutil/type.go | +3 | 添加 TDBRole |
| pkg/util/paramtable/component_param.go | +35 | TDB 配置 |
| cmd/components/tdb.go | +98 | TDB 组件 |
| cmd/roles/roles.go | +12 | 启动集成 |
| cmd/milvus/util.go | +2 | 命令行支持 |
| pkg/metrics/tdb_metrics.go | +128 | 监控指标 |
| **总计** | **~278** | **服务集成** |

***

## 注意事项

1. **默认禁用**: TDB 默认禁用，需要显式启用
2. **端口分配**: TDB gRPC 端口默认自动分配，也可手动配置
3. **依赖 etcd**: TDB 依赖 etcd 进行服务注册和发现
4. **优雅停止**: TDB 支持优雅停止，超时时间可配置
5. **监控指标**: TDB 暴露 Prometheus 格式的监控指标

***

## 后续工作

1. **配置热更新**: 支持部分配置的热更新
2. **多实例支持**: 支持 TDB 多实例部署
3. **负载均衡**: 实现 TDB 客户端负载均衡
4. **自动扩缩容**: 根据负载自动扩缩容

***

## 参考

- [Milvus 架构文档](https://milvus.io/docs/architecture_overview.md)
- [Milvus 配置文档](https://milvus.io/docs/configure-docker.md)
- [etcd 服务发现](https://etcd.io/docs/v3.5/dev-guide/grpc_naming/)

***

## 变更记录

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|----------|------|
| 1.0 | 2026-03-15 | 初始版本，记录服务集成过程 | - |
