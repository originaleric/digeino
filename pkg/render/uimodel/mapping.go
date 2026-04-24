package uimodel

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// BlockProfile names the card type for one render.BlockKind (首版仅三类块)。
type BlockProfile struct {
	Card string `yaml:"card"`
}

// Mapping is YAML-driven block→card mapping plus governance metadata.
type Mapping struct {
	SchemaVersion    int                     `yaml:"schema_version"`
	MappingVersion   string                  `yaml:"mapping_version"`
	MappingSource    string                  `yaml:"mapping_source"`
	MappingChangedAt string                  `yaml:"mapping_changed_at"` // RFC3339 optional
	BlockProfiles    map[string]BlockProfile `yaml:"block_profiles"`
}

// BuiltinMapping returns frozen 首版默认（与计划一致：MarkdownCard / ReasoningCard / CodeCard）。
func BuiltinMapping() Mapping {
	return Mapping{
		SchemaVersion:  1,
		MappingVersion: "2026-04-24",
		MappingSource:  "github.com/originaleric/digeino/pkg/render/uimodel",
		BlockProfiles: map[string]BlockProfile{
			"markdown": {Card: "MarkdownCard"},
			"thinking": {Card: "ReasoningCard"},
			"code":     {Card: "CodeCard"},
		},
	}
}

// LoadMappingFromBytes parses YAML into Mapping and merges missing profiles from BuiltinMapping().
func LoadMappingFromBytes(data []byte) (Mapping, error) {
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return BuiltinMapping(), nil
	}
	var m Mapping
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Mapping{}, err
	}
	return normalizeMapping(m)
}

// LoadMappingFromFile reads YAML from path.
func LoadMappingFromFile(path string) (Mapping, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Mapping{}, err
	}
	m, err := LoadMappingFromBytes(b)
	if err != nil {
		return Mapping{}, err
	}
	if strings.TrimSpace(m.MappingSource) == "" {
		m.MappingSource = path
	}
	return m, nil
}

// LoadMappingFromReader reads YAML until EOF.
func LoadMappingFromReader(r io.Reader) (Mapping, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return Mapping{}, err
	}
	return LoadMappingFromBytes(b)
}

func normalizeMapping(m Mapping) (Mapping, error) {
	def := BuiltinMapping()
	if m.SchemaVersion == 0 {
		m.SchemaVersion = def.SchemaVersion
	}
	if m.MappingVersion == "" {
		m.MappingVersion = def.MappingVersion
	}
	if m.MappingSource == "" {
		m.MappingSource = def.MappingSource
	}
	if m.BlockProfiles == nil {
		m.BlockProfiles = make(map[string]BlockProfile)
	}
	for k, v := range def.BlockProfiles {
		if _, ok := m.BlockProfiles[k]; !ok {
			m.BlockProfiles[k] = v
		}
	}
	for k, p := range m.BlockProfiles {
		if strings.TrimSpace(p.Card) == "" {
			return Mapping{}, fmt.Errorf("block_profiles.%s: card must be non-empty", k)
		}
	}
	return m, nil
}
