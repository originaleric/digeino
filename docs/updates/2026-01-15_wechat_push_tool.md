# Update: WeChat Push Tool Integration

**Date**: 2026-01-15
**Category**: Feature / Tools

## Summary
Integrated a WeChat push notification tool as a reusable component within the Digeino framework. This enables any AI agent powered by Digeino to send text messages to WeChat users through WeChat Official Account (服务号) API. The tool includes automatic access token management with caching and refresh capabilities.

## New Features
- **WeChat Message Sending**: Send text messages to WeChat users via Official Account API
- **Automatic Token Management**: Access token caching and automatic refresh with file-based persistence
- **Batch Sending Support**: Send messages to multiple users (configured OpenIDs) or a single user
- **Error Handling**: Comprehensive error handling with detailed failure reasons for each recipient
- **Configuration-Driven**: Fully configurable through global config with enable/disable toggle
- **Eino Tool Wrapper**: Out-of-the-box compatibility with the Eino framework

## Components Added
- `tools/wx/wx_types.go`: Type definitions for requests, responses, and API structures
- `tools/wx/wx_token.go`: Access token management with file-based caching and automatic refresh
- `tools/wx/wx_message.go`: Core message sending logic with error handling
- `tools/wx/tool.go`: Eino-compliant tool factory
- `config/config.go`: Added `WeChatConfig` structure to global configuration
- `config/config.yaml`: Added WeChat configuration section with example values

## Configuration

Add the following section to your `config/config.yaml`:

```yaml
WeChat:
  Enabled: true                    # Enable/disable WeChat push functionality
  AppID: "your_appid"              # WeChat Official Account AppID
  AppSecret: "your_app_secret"     # WeChat Official Account AppSecret
  OpenIDs:                         # Default recipient OpenID list
    - "openid1"
    - "openid2"
  TokenFilePath: "storage/app/wechat/access_token.json"  # Access token storage path
```

## How to Use

The tool is automatically registered in `BaseTools()` if enabled in configuration:

```go
import "github.com/originaleric/digeino/tools"

// Get all base tools (includes WeChat tool if enabled)
tools, err := tools.BaseTools(ctx)
if err != nil {
    // Handle error
}
```

The tool name is `send_wechat_message` and accepts the following parameters:
- `content` (required): Text message content (recommended max 2048 characters)
- `openid` (optional): Specific recipient OpenID. If not provided, sends to all configured OpenIDs

**Note**: Users must have interacted with the Official Account within the last 48 hours to receive customer service messages.

## Technical Details

### Access Token Management
- Tokens are cached in a JSON file to minimize API calls
- Automatic refresh when token expires (with 5-minute buffer)
- Thread-safe token access using mutex locks
- Graceful handling of file I/O errors

### Error Handling
The tool provides detailed error information:
- Common error codes (45015: user not interacted, 40001: invalid token, 40003: invalid openid, 45047: rate limit)
- Per-recipient success/failure tracking
- Partial success support (some recipients succeed, others fail)

## Impact
This update enables AI agents built on Digeino to send real-time notifications to WeChat users, improving user engagement and providing a native Chinese messaging channel. The tool is designed to be lightweight, reliable, and easy to configure, making it suitable for production use in various scenarios such as task completion notifications, error alerts, and status updates.
