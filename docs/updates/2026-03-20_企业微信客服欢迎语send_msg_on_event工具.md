# 2026-03-20 企业微信客服欢迎语 send_msg_on_event 工具

## 概述

基于 [docs/ideas/2026-03-20_企业微信客服欢迎语send_msg_on_event工具.md](../ideas/2026-03-20_企业微信客服欢迎语send_msg_on_event工具.md)，已完成 DigEino 企业微信客服「发送欢迎语」能力的实现。当用户首次进入客服会话（或 48 小时内未收过欢迎语且未发过消息）时，可通过 sync_msg 事件中的 `welcome_code` 调用该工具发送欢迎语。

## 实施状态

| 修改项 | 状态 | 修改文件 |
|--------|------|----------|
| 新增类型 SendWeComMsgOnEventRequest/Response | ✅ 已完成 | `tools/wx/wx_types.go` |
| sendWeComMsgOnEventAPI 底层调用 | ✅ 已完成 | `tools/wx/wecom_message.go` |
| SendWeComMsgOnEvent 发送函数 | ✅ 已完成 | `tools/wx/wecom_message.go` |
| NewWeComMsgOnEventTool 工具 | ✅ 已完成 | `tools/wx/wecom_tool.go` |
| BaseTools 注册 | ✅ 已完成 | `tools/tools.go` |

## 新增内容

### 1. 工具信息

- **工具名**：`send_wecom_msg_on_event`
- **描述**：通过企业微信客服发送欢迎语（事件响应消息）。用于用户进入会话时，使用 sync_msg 事件中的 welcome_code 发送欢迎语。code 仅可使用一次，收到事件后 20 秒内有效。

### 2. 请求参数

| 参数 | 必填 | 说明 |
|------|------|------|
| code | 是 | welcome_code，来自 sync_msg 事件，仅可使用一次 |
| content | 是 | 文本欢迎语内容，最长 2048 字节 |
| msgid | 否 | 消息 ID，不传则系统自动生成 |
| access_token | 否 | 第三方传入的 access_token，不传则使用 ManageAllKFSession 应用自动获取 |

### 3. 与 send_wecom_customer_message 的区别

| 维度 | send_wecom_customer_message | send_wecom_msg_on_event |
|------|-----------------------------|--------------------------|
| API | cgi-bin/kf/send_msg | cgi-bin/kf/send_msg_on_event |
| 参数 | open_kf_id + customer_id + content | code + content |
| 适用场景 | 用户发消息后 48 小时内回复 | 用户进入会话时发送欢迎语 |
| 限制 | 48 小时 + 5 条 | 20 秒内、仅一次 |

## 使用场景

1. **DigUserService**：收到 kf_msg_or_event → 调用 sync_msg → 解析到 welcome_code → 调用 `send_wecom_msg_on_event` 工具发送欢迎语
2. **Knowledge**：若 DigUserService 将 welcome_code 事件转发给 Knowledge，Knowledge 可通过该工具发送欢迎语

## 注意事项

1. **时效**：welcome_code 收到后 20 秒内有效，调用方需尽快完成调用
2. **一次性**：code 仅可使用一次，调用失败后不可重试
3. **Token**：必须使用 ManageAllKFSession 应用的 access_token（与 send_wecom_customer_message 相同）
4. **长度限制**：欢迎语内容最长 2048 字节，code 仅一次有效，不支持分条发送

## 参考

- 方案文档：`docs/ideas/2026-03-20_企业微信客服欢迎语send_msg_on_event工具.md`
- 企业微信 API：https://kf.weixin.qq.com/api/doc/path/95123
