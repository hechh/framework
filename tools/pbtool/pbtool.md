# pbtool — .pb.go 辅助方法生成器

## 概述

pbtool 扫描 `.pb.go` 文件，自动为 protobuf message 生成四类辅助方法：

- **`*Rsp` 类型** → `SetRspHead` / `GetRspHead`（响应头快捷操作）
- **其他所有类型**（排除 `*Rsp`/`*Req`/`*Config`/`*ConfigAry`/`*ConfigS`） → `ToDB` / `FromDB`（proto 序列化/反序列化）
- **`*Config` 类型**（配置表） → `XxxConfigS` 只读包装体 + Getter 方法
- **声明 `@pbtool:dbpool` 注解的类型**（分库分表） → `TableName() string`（MySQL 分表名）

输出文件：`{src目录}/common.gen.pb.go`

---

## 生成的方法

### SetRspHead / GetRspHead

为所有以 `Rsp` 结尾且包含 `.RspHead` 类型字段的 struct 生成：

```go
func (d *XxxRsp) SetRspHead(code int32, msg string) {
    d.Head = &packet.RspHead{Code: code, Msg: msg}
}
func (d *XxxRsp) GetRspHead() (int32, string) {
    return d.Head.Code, d.Head.Msg
}
```

工具通过检测字段类型名是否以 `.RspHead` 结尾来自动匹配字段（不限字段名）。

### ToDB / FromDB

为所有**非** `*Rsp`/`*Req`/`*Config`/`*ConfigAry`/`*ConfigS` 的 struct 生成：

```go
func (d *Xxx) ToDB() ([]byte, error) {
    if d == nil { return nil, nil }
    return proto.Marshal(d)
}
func (d *Xxx) FromDB(val []byte) error {
    if len(val) <= 0 { return nil }
    return proto.Unmarshal(val, d)
}
```

### ConfigS 只读包装

为所有以 `Config` 结尾（但不以 `ConfigAry` 或 `ConfigS` 结尾）的配置 struct 生成只读包装体：

```go
type XxxConfigS struct {
    inner *XxxConfig
}
```
并生成所有导出字段的 Getter 方法：

- **普通字段**（基本类型、枚举等）：直接返回 `s.inner.FieldName`
- **`*Reward` 字段**：返回 `s.inner.FieldName.CloneVT()` 深拷贝
- **`[]*Reward` 字段**：返回新的切片，每个元素通过 `CloneVT()` 深拷贝

```go
// 示例：普通字段
func (s *PveRoomConfigS) GetRoomId() uint32 {
    return s.inner.RoomId
}

// 示例：*Reward 字段 — 深拷贝
func (s *PveRoomConfigS) GetEntryFee() *Reward {
    return s.inner.EntryFee.CloneVT()
}

// 示例：[]*Reward 字段 — 深拷贝
func (s *PveRankRewardConfigS) GetRankPrizes() []*Reward {
    if s.inner.RankPrizes == nil { return nil }
    rets := make([]*Reward, len(s.inner.RankPrizes))
    for i, v := range s.inner.RankPrizes {
        rets[i] = v.CloneVT()
    }
    return rets
}
```

**设计意图**：配置表数据在内存中是全局共享的只读缓存。通过 `ConfigS` 包装，Reward 类型字段自动进行 `CloneVT` 深拷贝，确保业务层无论如何修改返回值，都不会污染全局配置缓存。

---

### TableName() 分库分表

为声明了 `@pbtool:dbpool` 注解的数据库 struct 生成 xorm 分表名方法。**注解写在 `.proto` 的 message 上方**，经 protoc 生成 `.pb.go` 后由 pbtool 解析。核心理念与 redistool 的 `@dbtool` 一致：**proto 中声明 → 工具生成 → 业务直接调用**。

#### @pbtool 注解语法

```
// @pbtool:dbpool|shard:N|table:field@type
```

| 段 | 说明 | 示例 |
|----|------|------|
| `dbpool` | 存储类型标识（MySQL 分库分表） | `dbpool` |
| `shard:N` | 分片数量 | `shard:64` |
| `table:field@type` | 基础表名 + 分片字段（proto 字段名）+ 字段类型 | `bingo_contest_result_data:uid@uint64` |

```proto
// Bingo 比赛结果数据
// @pbtool:dbpool|shard:64|bingo_contest_result_data:uid@uint64
message BingoContestResultData {
  uint64 Uid = 3;  // @inject_tag: xorm:"bigint index notnull comment('玩家UID')"
  // ...
}
```

pbtool 将注解字段名 `uid` 自动转为首字母大写的 Go 字段 `Uid`，并校验该字段确实存在于 struct 中（缺失时打印告警并跳过）。生成：

```go
func (d *BingoContestResultData) TableName() string {
    return fmt.Sprintf("bingo_contest_result_data_%d", d.Uid%64)
}
```

> 说明：xorm 在 Insert/Find/Update 时若 bean 实现了 `TableName() string` 接口，会用它作为实际表名。每个分片记录通过自身分片字段（如 `Uid % 64`）定位到具体分表。

---

## 使用方法

```bash
# 直接运行（注意：必须用双横线 --src，单横线 -src 会被 pflag 误解析为 -s 简写 + 值 "rc=..."）
go run server/framework/tools/pbtool/ --src=server/common/pb

# 通过 Makefile（make pb 自动包含此步骤）
make pb
```

| 参数 | 说明 | 示例 |
|------|------|------|
| `--src` | `.pb.go` 文件所在目录 | `server/common/pb` |

> `make pb` 的执行顺序：**先 `protoc` 生成 `.pb.go` → 再 `pbtool` 生成 `common.gen.pb.go`**

---

## 排除规则

以下后缀的 struct **不会**生成 `ToDB`/`FromDB`：

| 后缀 | 原因 |
|------|------|
| `Rsp` | 响应体，不直接存 DB |
| `Req` | 请求体，不直接存 DB |
| `Config` | 配置表（由 xlsxtool 管理，生成 ConfigS 包装） |
| `ConfigAry` | 配置表数组（由 xlsxtool 管理） |
| `ConfigS` | 只读包装体（pbtool 自己生成，避免重复） |

`SetRspHead`/`GetRspHead` **仅**为 `*Rsp` 类型生成。

`ConfigS` 只读包装 **仅**为 `*Config` 类型生成（排除 `ConfigAry`）。

---

## 常见问题

| 问题 | 解决 |
|------|------|
| 新增 Rsp 后没有 SetRspHead | 确保 Rsp struct 中包含 `packet.RspHead` 类型字段；运行 `make pb` |
| 新增 Config 后没有 ConfigS | 运行 `make pb`，pbtool 会自动为所有 Config 生成只读包装 |
| ToDB/FromDB 未生成 | 检查 struct 名称是否以 `Rsp`/`Req`/`Config`/`ConfigAry`/`ConfigS` 结尾（这些会被跳过） |
| 新增分表消息后没有 TableName | 在 mysql.proto 的 message 上方声明 `@pbtool:dbpool|shard:N|table:field@type`，确认分片字段存在于 message 中，运行 `make pb` |
| common.gen.pb.go 过期 | 重新运行 `make pb`，pbtool 每次都会覆盖生成 |
| ConfigS 的 Get 方法没有 CloneVT | 检查字段类型名是否为 `Reward`（裸名）或 `pkg.Reward`（选择器名）；pbtool 会自动识别两种形式 |
