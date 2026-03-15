# 2026-03-13 Knowledge 对接 goDig 性能优化升级

## 概述

为充分利用最新版本 goDig 提供的 HTTP Server、JSON 编码和数据库连接池等性能优化能力，对 Knowledge 后端进行了适配与配置更新。本次更新在保持向后兼容的前提下，重点提升高并发场景下的吞吐和稳定性。

## 主要变更

### 1. HTTP Server 启动方式升级
- **文件**：`cmd/api/main.go`
- **调整内容**：
  - 原先使用 `router.Run(port)` 直接启动 Gin 服务。
  - 现改为使用标准库 `http.Server`，并对关键参数进行显式配置：
    - `ReadTimeout: 5 * time.Second`
    - `WriteTimeout: 5 * time.Second`
    - `MaxHeaderBytes: 1 << 20`
  - 通过 `ListenAndServe` 启动服务，并对异常错误（非 `http.ErrServerClosed`）进行 `log.Fatal` 级别输出。
- **收益**：
  - 避免慢客户端长时间占用连接拖垮服务端。
  - 限制请求头大小，防止异常流量占用过多内存。
  - 提升整体稳定性与可控性。

### 2. GORM 数据库连接池参数优化
- **文件**：`config/gorm.yml`
- **调整范围**：
  - `Mysql.Write` / `Mysql.Read`
  - `SqlServer.Write` / `SqlServer.Read`
  - `PostgreSql.Write` / `PostgreSql.Read`
- **调整内容**：
  - `SetMaxIdleConns`：从 `10` 提升到 `100`
  - `SetMaxOpenConns`：从 `128` 提升到 `200`
  - `SetConnMaxLifetime`：从 `60` 秒提升到 `300` 秒
- **收益**：
  - 减少频繁创建/销毁连接带来的开销。
  - 支持更高并发下的稳定访问。
  - 降低长时间不回收连接造成的资源浪费风险。

### 3. 新增性能相关配置块
- **文件**：`config/config.yml`
- **新增配置**：
  ```yaml
  Performance:
    JsonEncoder: "sonic"
    EnablePprof: false
    Cache:
      DefaultTTL: 300
      Enabled: true
  ```
- **设计目的**：
  - 与 goDig 的性能能力保持统一配置入口，便于后续扩展。
  - 为未来在不同环境间切换 JSON 编码器、开启/关闭 pprof、接入通用缓存工具等提供集中化控制。
  - 当前 Knowledge 已通过 goDig 的 `response` 包默认使用 `sonic`，该配置主要作为显式声明与后续能力扩展预留。

## 构建与验证
- **依赖整理**：在 `Knowledge` 根目录执行：`go mod tidy`。
- **编译检查**：执行 `go build ./cmd/api`，构建通过，无编译错误。
- **静态检查**：针对本次改动的关键文件（`cmd/api/main.go`、`config/gorm.yml`、`config/config.yml`）运行 linter，无新增问题。

## 预期效果
- 在保持现有业务逻辑不变的前提下：
  - **HTTP 层**：慢客户端与异常请求对整体服务的影响收敛，可用性提升。
  - **数据库层**：在高并发访问下，连接等待时间减少，QPS 上限抬升。
  - **配置层**：为后续接入 goDig 的缓存、pprof 等能力提供统一的配置基础。

