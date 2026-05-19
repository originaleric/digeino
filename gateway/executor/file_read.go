package executor

import (
	"context"

	"github.com/originaleric/digeino/gateway/protocol"
	"github.com/originaleric/digeino/gateway/registry"
	"github.com/originaleric/digeino/tools/research"
)

const ToolFileRead = "file.read"

// FileReadEntry returns registry entry for file.read.
func FileReadEntry(allowedPaths []string) registry.Entry {
	return registry.Entry{
		Descriptor: protocol.ToolDescriptor{
			Name:        ToolFileRead,
			Description: "Read a local file from an allowed directory prefix.",
			InputSchema: registry.MustSchema(map[string]any{
				"type":     "object",
				"required": []string{"path"},
				"properties": map[string]any{
					"path": map[string]string{"type": "string"},
				},
			}),
			Capabilities:         []string{"file.local"},
			Risk:                 "filesystem",
			RequiresUserApproval: true,
		},
		Handler: func(ctx context.Context, call *protocol.ToolCall) (map[string]any, []protocol.Artifact, error) {
			in, err := decodeInput[research.ReadFileRequest](call)
			if err != nil {
				return nil, nil, err
			}
			if err := validateReadPath(in.Path, allowedPaths); err != nil {
				return nil, nil, err
			}
			resp, err := research.ReadFile(ctx, &in)
			if err != nil {
				return nil, nil, err
			}
			return map[string]any{
				"path":    in.Path,
				"content": resp.Content,
			}, nil, nil
		},
	}
}
