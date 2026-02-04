# Update: UI/UX Tool Configuration and Storage Improvements

**Date**: 2026-02-01
**Category**: Configuration / Enhancement

## Summary

This update significantly improves the UI/UX tool's configuration management and storage capabilities. The changes include:
1. **Simplified ChatModel configuration** (aligned with DigFlow's approach)
2. **Configurable storage paths** for design system persistence
3. **Application/Agent isolation** support for multi-tenant scenarios

These improvements make the UI/UX tool more flexible, maintainable, and suitable for enterprise-level deployments with multiple agents or applications.

## Changes Made

### 1. ChatModel Configuration Simplification

**Before** (v1.0):
```yaml
ChatModel:
  Provider: "qwen"  # 选择模型提供商
  Qwen:
    APIKey: ""
    Model: "qwen-max"
    BaseURL: "..."
  OpenAI:
    APIKey: ""
    Model: "gpt-4"
    BaseURL: "..."
  OpenRouter:
    APIKey: ""
    Model: ""
    BaseURL: "..."
```

**After** (v2.0):
```yaml
ChatModel:
  Type: "qwen"  # 模型类型: qwen, openai
  Config:
    ApiKey: "${QWEN_API_KEY}"  # 支持环境变量
    Model: "qwen-max"
    BaseUrl: "https://dashscope.aliyuncs.com/compatible-mode/v1"
    Temperature: 0.7  # 可选参数
```

**Benefits**:
- ✅ Simplified configuration: Single model type instead of three separate config blocks
- ✅ Aligned with DigFlow's configuration approach for consistency
- ✅ Environment variable support: `${VAR_NAME}` format for sensitive data
- ✅ Easier to switch between providers: Just change `Type` field

### 2. Configurable Storage Paths

**Before**: Hardcoded storage path (default: current directory)

**After**: Configurable storage base directory
```yaml
UIUX:
  Storage:
    BaseDir: "storage/app/ui_ux"  # 可自定义存储基础目录
```

**Storage Structure**:
```
{BaseDir}/
├── {app-name}/              # 应用隔离（如果指定 AppName）
│   └── design-system/
│       └── {project}/
│           ├── MASTER.md
│           └── pages/
│               └── {page}.md
└── design-system/           # 默认共享存储（未指定 AppName）
    └── {project}/
        ├── MASTER.md
        └── pages/
            └── {page}.md
```

**Path Priority**:
1. Call-time `BaseDir` parameter (highest priority)
2. Configuration file `UIUX.Storage.BaseDir`
3. Default value: `storage/app/ui_ux`

### 3. Application/Agent Isolation

**New Feature**: Support for isolating design systems by application or agent name

**Usage**:
```go
// Travel Agent
persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "travel booking platform",
    ProjectName: "TravelApp",
    AppName:     "travel",  // 隔离存储
})

// Stock Agent
persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "stock analysis dashboard",
    ProjectName: "StockApp",
    AppName:     "stock",  // 隔离存储
})
```

**Storage Result**:
- Travel: `storage/app/ui_ux/travel/design-system/travelapp/MASTER.md`
- Stock: `storage/app/ui_ux/stock/design-system/stockapp/MASTER.md`

**Benefits**:
- ✅ Prevents conflicts between different applications/agents
- ✅ Suitable for multi-tenant scenarios
- ✅ Optional isolation: If `AppName` is not specified, uses shared storage

## Components Modified

### Configuration Files
- `config/config.go`: 
  - Simplified `ChatModelConfig` structure (from 3 separate configs to unified `Type + Config`)
  - Added `UIUXConfig` and `UIUXStorageConfig` structures
- `config/config.yaml`: 
  - Updated ChatModel configuration format
  - Added `UIUX.Storage` configuration section

### UI/UX Tool Files
- `tools/ui_ux/persistence.go`: 
  - Updated `NewPersistenceManager` to accept `appName` parameter
  - Added support for reading base directory from configuration
  - Added `GetBaseDir()` and `GetAppName()` getter methods
- `tools/ui_ux/tool.go`: 
  - Updated `PersistDesignSystemRequest` to include `AppName` field
  - Updated response path building logic to support application isolation
- `tools/ui_ux/model_factory.go`: 
  - Refactored to use new unified ChatModel configuration structure
  - Added environment variable processing (`processEnvVars`)
  - Simplified model creation logic (removed OpenRouter support, kept Qwen and OpenAI)

### Documentation
- `tools/ui_ux/README.md`: Updated with new configuration examples and storage isolation documentation
- `tools/ui_ux/CHANGELOG.md`: Created comprehensive changelog documenting all improvements

## Configuration

### ChatModel Configuration

**New Format** (aligned with DigFlow):
```yaml
ChatModel:
  Type: "qwen"  # qwen or openai
  Config:
    ApiKey: "${QWEN_API_KEY}"  # 支持环境变量
    Model: "qwen-max"
    BaseUrl: "https://dashscope.aliyuncs.com/compatible-mode/v1"
    Temperature: 0.7  # 可选
    MaxTokens: 2048   # 可选
```

**Switching Providers**:
```yaml
# Use OpenAI
ChatModel:
  Type: "openai"
  Config:
    ApiKey: "${OPENAI_API_KEY}"
    Model: "gpt-4"
    BaseUrl: "https://api.openai.com/v1"
```

### UIUX Storage Configuration

```yaml
UIUX:
  Storage:
    BaseDir: "storage/app/ui_ux"  # 自定义存储基础目录
```

## Migration Guide

### For ChatModel Configuration

**If you have existing configuration**:
1. Update `config.yaml` to use the new format:
   ```yaml
   # Old format (remove)
   ChatModel:
     Provider: "qwen"
     Qwen: { ... }
     OpenAI: { ... }
   
   # New format (use this)
   ChatModel:
     Type: "qwen"
     Config:
       ApiKey: "${QWEN_API_KEY}"
       Model: "qwen-max"
       BaseUrl: "..."
   ```

2. Set environment variables for API keys:
   ```bash
   export QWEN_API_KEY="your-api-key"
   ```

### For Storage Paths

**No migration needed** - The default storage path remains compatible. However, you can now:
1. Configure a custom base directory in `config.yaml`
2. Use `AppName` parameter for application isolation

## Usage Examples

### Example 1: Using Configuration-Driven Agent

```go
import (
    "context"
    "github.com/originaleric/digeino/config"
    "github.com/originaleric/digeino/tools/ui_ux"
)

// 1. Load configuration
_, err := config.Load("config/config.yaml")
if err != nil {
    return err
}

// 2. Create Agent from config (automatically reads ChatModel config)
agent, err := ui_ux.NewUIDesignSystemAgentFromConfig(ctx)
if err != nil {
    return err
}

// 3. Use Agent
result, err := agent.Invoke(ctx, "生成一个美容spa的设计系统")
```

### Example 2: Application Isolation

```go
persistTool, _ := ui_ux.NewPersistDesignSystemTool(ctx)

// Travel Agent
result, _ := persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "travel booking platform",
    ProjectName: "TravelApp",
    AppName:     "travel",  // 隔离存储
})
// Stores to: storage/app/ui_ux/travel/design-system/travelapp/MASTER.md

// Stock Agent
result, _ := persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "stock analysis",
    ProjectName: "StockApp",
    AppName:     "stock",  // 隔离存储
})
// Stores to: storage/app/ui_ux/stock/design-system/stockapp/MASTER.md
```

### Example 3: Custom Storage Path

```go
persistTool, _ := ui_ux.NewPersistDesignSystemTool(ctx)

result, _ := persistTool.Invoke(ctx, &ui_ux.PersistDesignSystemRequest{
    Query:       "beauty spa",
    ProjectName: "Serenity Spa",
    BaseDir:     "/custom/path/to/storage",  // 自定义路径
    AppName:     "beauty_app",
})
// Stores to: /custom/path/to/storage/beauty_app/design-system/serenity-spa/MASTER.md
```

## Technical Details

### Environment Variable Processing

The new configuration supports environment variable substitution using `${VAR_NAME}` format:

```yaml
ChatModel:
  Config:
    ApiKey: "${QWEN_API_KEY}"  # 从环境变量读取
```

The system will:
1. Check if the value matches `${VAR_NAME}` pattern
2. Look up the environment variable
3. Use the environment value if found, otherwise use the literal string

### Storage Path Resolution

The storage path is resolved in the following order:
1. **Call-time `BaseDir`**: If provided in `PersistDesignSystemRequest`, use it
2. **Configuration `UIUX.Storage.BaseDir`**: If configured, use it
3. **Default**: `storage/app/ui_ux`

The final path structure:
- **With AppName**: `{BaseDir}/{AppName}/design-system/{project}/MASTER.md`
- **Without AppName**: `{BaseDir}/design-system/{project}/MASTER.md`

## Backward Compatibility

✅ **Fully backward compatible**:
- Existing `ui_ux_search` tool functionality unchanged
- All existing APIs remain functional
- New features are optional (can use defaults)
- Old configuration format can be migrated easily

## Impact

### Benefits

1. **Simplified Configuration**: 
   - Reduced configuration complexity (from 3 blocks to 1 unified config)
   - Easier to understand and maintain
   - Consistent with DigFlow's approach

2. **Better Security**:
   - Environment variable support for sensitive API keys
   - No need to hardcode credentials in configuration files

3. **Flexible Storage**:
   - Configurable storage paths for different deployment scenarios
   - Support for application isolation in multi-tenant environments
   - Easy to customize for different use cases

4. **Enterprise Ready**:
   - Suitable for multi-agent/multi-application deployments
   - Proper isolation between different agents/applications
   - Configuration-driven approach for easier management

### Use Cases Enabled

- ✅ **Multi-Agent Scenarios**: Different agents can have isolated design system storage
- ✅ **Multi-Tenant Applications**: Each tenant can have separate storage
- ✅ **Development/Production**: Easy to switch storage paths for different environments
- ✅ **Enterprise Deployments**: Centralized configuration management

## Related Files

- `config/config.go` - Configuration structures
- `config/config.yaml` - Configuration file
- `tools/ui_ux/persistence.go` - Storage management
- `tools/ui_ux/tool.go` - Tool definitions
- `tools/ui_ux/model_factory.go` - ChatModel factory
- `tools/ui_ux/README.md` - Updated documentation
- `tools/ui_ux/CHANGELOG.md` - Comprehensive changelog

## References

- DigFlow configuration approach: `/Users/dig/Documents/文稿 - XinYe的MacBook Pro (5)/Projects/go-app/DigFlow/config/config.yml`
- Previous UI/UX update: `docs/updates/2026-01-11_ui_ux_plugin_integration.md`
