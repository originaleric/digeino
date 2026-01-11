# DigEino Performance Optimization Update - 2026-01-11

## Overview
This update focuses on improving the responsiveness of Eino-based applications by eliminating blocking operations in the status collection and webhook delivery layer.

## Key Changes

### 1. Asynchronous Status Collection
The `StatusCollector` has been refactored to prevent slowing down the main execution graph during state persistence and webhook delivery.
- **Async Execution**: `StatusStore.AddStatus` and `WebhookClient.SendStatus` now run in their own goroutines.
- **Benefit**: Zero-latency impact on the Eino Graph from slow database writes or network-constrained webhook endpoints.
- **Affected File**: `webhook/status_collector.go`

### 2. HTTP Connection Pooling for Webhooks
`WebhookClient` now utilizes a shared, high-performance HTTP Transport.
- **Connection Reuse**: Configured `MaxIdleConns` and `MaxIdleConnsPerHost` to `100` to support high-frequency event delivery.
- **Benefit**: Significantly reduces per-request overhead and prevents port exhaustion under load.
- **Affected File**: `webhook/webhook_client.go`

### 3. Improved Token Usage Tracking
- **Aggregation**: Enhanced the `CollectTokenUsage` logic to accurately sum tokens across complex parallel and nested nodes.
- **Benefit**: Provides a reliable metric for cost and performance monitoring.

## Impact
These changes collectively allow `DigFlow` and other dependent agents to achieve significantly higher throughput and lower end-to-end latency, especially in multi-node graph architectures.
