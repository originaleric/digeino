package uimodel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// MappingFingerprint returns a stable SHA-256 hex over governance + card names (用于 mapping_hash)。
func MappingFingerprint(m Mapping) string {
	type row struct {
		Kind string `json:"kind"`
		Card string `json:"card"`
	}
	var rows []row
	for kind, p := range m.BlockProfiles {
		rows = append(rows, row{Kind: kind, Card: p.Card})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Kind != rows[j].Kind {
			return rows[i].Kind < rows[j].Kind
		}
		return rows[i].Card < rows[j].Card
	})
	payload := struct {
		SchemaVersion  int    `json:"schema_version"`
		MappingVersion string `json:"mapping_version"`
		MappingSource  string `json:"mapping_source"`
		Profiles       []row  `json:"profiles"`
	}{
		SchemaVersion:  m.SchemaVersion,
		MappingVersion: m.MappingVersion,
		MappingSource:  m.MappingSource,
		Profiles:       rows,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
