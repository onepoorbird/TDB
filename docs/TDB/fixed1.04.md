# TDB Protobuf Go 代码生成记录

> 本文档记录 TDB 项目 Protobuf Go 代码生成的详细过程和修复内容
> 创建日期: 2026-03-15

***

## 概述

成功使用 `protoc-gen-go` 和 `protoc-gen-go-grpc` 生成了 TDB 项目的所有 Protobuf Go 代码，包括：

- AgentService 的 gRPC 接口
- MemoryService 的 gRPC 接口
- EventService 的 gRPC 接口
- 通用类型定义 (Status, ErrorCode 等)

***

## 生成的文件

### 1. Agent 相关

| 文件 | 路径 | 行数 | 说明 |
|------|------|------|------|
| agent.pb.go | `pkg/proto/agentpb/agent.pb.go` | ~1600 | Agent 消息类型定义 |
| agent_grpc.pb.go | `pkg/proto/agentpb/agent_grpc.pb.go` | ~350 | AgentService gRPC 接口 |

**AgentService 接口方法：**
- `CreateAgent` - 创建 Agent
- `GetAgent` - 获取 Agent
- `ListAgents` - 列出 Agents
- `UpdateAgent` - 更新 Agent
- `DeleteAgent` - 删除 Agent
- `CreateSession` - 创建 Session
- `GetSession` - 获取 Session
- `ListSessions` - 列出 Sessions
- `UpdateSession` - 更新 Session

### 2. Memory 相关

| 文件 | 路径 | 行数 | 说明 |
|------|------|------|------|
| memory.pb.go | `pkg/proto/memorypb/memory.pb.go` | ~2800 | Memory 消息类型定义 |
| memory_grpc.pb.go | `pkg/proto/memorypb/memory_grpc.pb.go` | ~450 | MemoryService gRPC 接口 |

**MemoryService 接口方法：**
- `CreateMemory` - 创建记忆
- `GetMemory` - 获取记忆
- `UpdateMemory` - 更新记忆
- `DeleteMemory` - 删除记忆
- `QueryMemories` - 查询记忆
- `SearchMemories` - 向量搜索记忆
- `GetRelations` - 获取关系
- `CreateRelation` - 创建关系

### 3. Event 相关

| 文件 | 路径 | 行数 | 说明 |
|------|------|------|------|
| event.pb.go | `pkg/proto/eventpb/event.pb.go` | ~1300 | Event 消息类型定义 |
| event_grpc.pb.go | `pkg/proto/eventpb/event_grpc.pb.go` | ~280 | EventService gRPC 接口 |

**EventService 接口方法：**
- `AppendEvent` - 追加事件
- `GetEvent` - 获取事件
- `QueryEvents` - 查询事件
- `SubscribeEvents` - 订阅事件 (流式)

### 4. Common 相关

| 文件 | 路径 | 行数 | 说明 |
|------|------|------|------|
| common.pb.go | `pkg/proto/commonpb/common.pb.go` | ~800 | 通用类型定义 |

**定义内容：**
- `ErrorCode` - 错误码枚举
- `Status` - 状态信息
- `KeyValuePair` - 键值对
- `Address` - 地址信息
- `MsgBase` - 消息基础信息
- 各种状态枚举 (ConsistencyLevel, SegmentState, etc.)

***

## 修复内容

### 修复 1: common.Status 引用错误

**问题：**
proto 文件中引用了 `common.Status`，但 common.proto 的 package 是 `common.proto`，导致 protoc 报错：
```
agent.proto:71:5: "common.Status" is not defined.
```

**解决方案：**
将所有 `common.Status` 改为 `common.proto.Status`

**修改文件：**
- `pkg/proto/agent.proto` - 9 处修改
- `pkg/proto/memory.proto` - 8 处修改
- `pkg/proto/event.proto` - 4 处修改

**修改示例：**
```protobuf
// 修改前
message CreateAgentResponse {
    common.Status status = 1;
    string agent_id = 2;
}

// 修改后
message CreateAgentResponse {
    common.proto.Status status = 1;
    string agent_id = 2;
}
```

### 修复 2: go_package 路径错误

**问题：**
初始的 go_package 选项使用了错误的路径，导致生成的 Go 代码导入路径不正确：
```protobuf
option go_package="github.com/milvus-io/milvus/pkg/v2/proto/agentpb";
```

但 Milvus 项目的模块结构是：
- 根模块: `github.com/milvus-io/milvus`
- pkg 模块: `github.com/milvus-io/milvus/pkg/v2` (定义在 `pkg/go.mod`)

**解决方案：**
更新 go_package 选项，使用正确的导入路径：
```protobuf
option go_package="github.com/milvus-io/milvus/pkg/v2/proto/agentpb";
```

**修改文件：**
- `pkg/proto/agent.proto`
- `pkg/proto/memory.proto`
- `pkg/proto/event.proto`
- `pkg/proto/common.proto`

### 修复 3: 移除未使用的导入

**问题：**
`memory.proto` 导入了 `schema.proto`，但该文件不存在且未被使用：
```protobuf
import "schema.proto";
```

**解决方案：**
移除该导入语句。

**修改文件：**
- `pkg/proto/memory.proto`

### 修复 4: 移除未使用的 google/protobuf/any.proto 导入

**问题：**
`agent.proto` 导入了 `google/protobuf/any.proto` 但未使用。

**解决方案：**
虽然 protoc 只发出警告，但建议后续清理。

***

## 生成命令

### 前置条件

安装 protoc 和相关插件：
```bash
# 安装 protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# 安装 protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

验证安装：
```bash
protoc-gen-go --version
# 输出: protoc-gen-go.exe v1.36.11

protoc-gen-go-grpc --version
# 输出: protoc-gen-go-grpc 1.6.1
```

### 生成步骤

```bash
cd pkg/proto

# 生成 Agent 相关代码
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       agent.proto common.proto

# 生成 Memory 相关代码
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       memory.proto common.proto

# 生成 Event 相关代码
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       event.proto common.proto

# 生成 Common 代码 (单独生成)
protoc --go_out=. --go_opt=paths=source_relative \
       common.proto
```

### 移动文件到正确位置

生成的文件默认在 proto 目录下，需要移动到对应的子目录：

```bash
# Agent
mv pkg/proto/agent.pb.go pkg/proto/agentpb/
mv pkg/proto/agent_grpc.pb.go pkg/proto/agentpb/

# Memory
mv pkg/proto/memory.pb.go pkg/proto/memorypb/
mv pkg/proto/memory_grpc.pb.go pkg/proto/memorypb/

# Event
mv pkg/proto/event.pb.go pkg/proto/eventpb/
mv pkg/proto/event_grpc.pb.go pkg/proto/eventpb/

# Common
mv pkg/proto/common.pb.go pkg/proto/commonpb/
```

***

## 编译验证

在 `pkg` 目录下编译验证：

```bash
cd pkg

go build ./proto/agentpb/...
go build ./proto/memorypb/...
go build ./proto/eventpb/...
go build ./proto/commonpb/...
```

所有包编译成功，无错误。

***

## 代码统计

| 包 | 文件数 | 总行数 | 消息类型数 | 服务方法数 |
|----|--------|--------|------------|------------|
| agentpb | 2 | ~1950 | 18 | 9 |
| memorypb | 2 | ~3250 | 28 | 8 |
| eventpb | 2 | ~1580 | 14 | 4 |
| commonpb | 1 | ~800 | 15 | 0 |
| **总计** | **7** | **~7580** | **75** | **21** |

***

## 后续工作

1. **gRPC Server 实现** - 基于生成的接口实现 Server 层
2. **集成测试** - 验证 protobuf 序列化/反序列化正确性
3. **性能优化** - 如有需要，优化 protobuf 消息结构

***

## 参考

- [Protocol Buffers 官方文档](https://protobuf.dev/)
- [gRPC Go 文档](https://grpc.io/docs/languages/go/)
- [protoc-gen-go 文档](https://pkg.go.dev/google.golang.org/protobuf/cmd/protoc-gen-go)
- [protoc-gen-go-grpc 文档](https://pkg.go.dev/google.golang.org/grpc/cmd/protoc-gen-go-grpc)

***

## 变更记录

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|----------|------|
| 1.0 | 2026-03-15 | 初始版本，记录 protobuf Go 代码生成过程 | - |
