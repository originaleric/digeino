# Update: UI/UX Design Intelligence Plugin Integration

**Date**: 2026-01-11
**Category**: Feature / Tools

## Summary
Integrated a comprehensive UI/UX Design Intelligence engine as a reusable tool within the Digeino framework. This enables any AI agent powered by Digeino to access professional design guidelines, color palettes, typography pairings, and interaction principles.

## New Features
- **Native search engine**: Efficient BM25-based retrieval system for design knowledge.
- **Embedded Knowledge Base**: 400+ professional design rules and patterns embedded directly into the binary using `//go:embed`.
- **Domain Detection**: Automatic detection of query intent (e.g., distinguishing between a request for "colors" vs. "layout").
- **Stack-Specific Guidelines**: Support for technology-specific advice (React, Vue, Tailwind, SwiftUI, Flutter, etc.).
- **Eino Tool Wrapper**: Out-of-the-box compatibility with the Eino framework.

## Components Added
- `tools/ui_ux/bm25.go`: Implementation of the BM25 ranking algorithm.
- `tools/ui_ux/detector.go`: Natural language intent detector for design domains.
- `tools/ui_ux/service.go`: Core orchestration service managing embedded CSV data.
- `tools/ui_ux/tool.go`: Eino-compliant tool factory.
- `tools/ui_ux/data/`: Full dataset including styles, colors, UX guidelines, and implementation prompts.

## How to Use
Import the `ui_ux` package and register the tool:

```go
import "github.com/originaleric/digeino/tools/ui_ux"

// ... 

uiTool, err := ui_ux.NewUIUXSearchTool(ctx)
if err == nil {
    tools = append(tools, uiTool)
}
```

## Impact
This update significantly lowers the barrier for AI agents to provide visually stunning and UX-sound UI recommendations without requiring external API calls or complex local file management.
