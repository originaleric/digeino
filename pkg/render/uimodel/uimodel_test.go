package uimodel

import (
	"strings"
	"testing"

	"github.com/originaleric/digeino/pkg/render"
)

func TestBuildUIModelBuiltin(t *testing.T) {
	blocks := []render.Block{
		{Kind: render.BlockKindMarkdown, Content: "# Hi"},
		{Kind: render.BlockKindThinking, Content: "plan"},
		{Kind: render.BlockKindCode, Language: "go", Content: "x"},
	}
	ui, err := BuildUIModel(blocks, BuiltinMapping())
	if err != nil {
		t.Fatal(err)
	}
	if len(ui.Cards) != 3 || ui.Cards[1].Type != "ReasoningCard" {
		t.Fatalf("%+v", ui)
	}
	if ui.Cards[0].Props["markdown"] != "# Hi" {
		t.Fatalf("%+v", ui.Cards[0])
	}
}

func TestBuildHybridPayloadModes(t *testing.T) {
	blocks := []render.Block{{Kind: render.BlockKindMarkdown, Content: "a"}}
	m := BuiltinMapping()
	pb, err := BuildHybridPayload(OutputBlocks, blocks, m)
	if err != nil || len(pb.Blocks) != 1 || pb.UIModel != nil {
		t.Fatalf("%+v", pb)
	}
	pu, err := BuildHybridPayload(OutputUIModel, blocks, m)
	if err != nil || len(pu.Blocks) != 0 || pu.UIModel == nil || len(pu.UIModel.Cards) != 1 {
		t.Fatalf("%+v", pu)
	}
	pboth, err := BuildHybridPayload(OutputBoth, blocks, m)
	if err != nil || len(pboth.Blocks) != 1 || pboth.UIModel == nil {
		t.Fatalf("%+v", pboth)
	}
	if pb.MappingHash == "" || pb.MappingVersion == "" {
		t.Fatal("expected mapping metadata")
	}
}

func TestParseOutputMode(t *testing.T) {
	if ParseOutputMode("BLOCKS") != OutputBlocks {
		t.Fatal()
	}
	if ParseOutputMode("ui_model") != OutputUIModel {
		t.Fatal()
	}
	if ParseOutputMode("") != OutputBoth {
		t.Fatal()
	}
}

func TestLoadMappingFromBytesOverride(t *testing.T) {
	m, err := LoadMappingFromBytes([]byte(`
mapping_version: "custom-1"
block_profiles:
  code:
    card: SnippetCard
`))
	if err != nil {
		t.Fatal(err)
	}
	if m.BlockProfiles["code"].Card != "SnippetCard" {
		t.Fatalf("%+v", m.BlockProfiles["code"])
	}
	if m.BlockProfiles["markdown"].Card != "MarkdownCard" {
		t.Fatalf("expected merge from builtin")
	}
}

func TestBuildUIModelUsesOverrideCard(t *testing.T) {
	m, err := LoadMappingFromBytes([]byte(`
block_profiles:
  markdown:
    card: ProseCard
`))
	if err != nil {
		t.Fatal(err)
	}
	ui, err := BuildUIModel([]render.Block{{Kind: render.BlockKindMarkdown, Content: "x"}}, m)
	if err != nil || ui.Cards[0].Type != "ProseCard" {
		t.Fatalf("%+v", err)
	}
}

func TestMappingFingerprintStable(t *testing.T) {
	a := BuiltinMapping()
	b := BuiltinMapping()
	if MappingFingerprint(a) != MappingFingerprint(b) {
		t.Fatal("fingerprint should be stable")
	}
	b.MappingSource = "other"
	if MappingFingerprint(a) == MappingFingerprint(b) {
		t.Fatal("expected different hash when source changes")
	}
}

func TestLoadMappingFromBytesInvalidCard(t *testing.T) {
	_, err := LoadMappingFromBytes([]byte(`
block_profiles:
  markdown:
    card: ""
`))
	if err == nil || !strings.Contains(err.Error(), "card must be non-empty") {
		t.Fatalf("err=%v", err)
	}
}
