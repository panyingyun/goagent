package contextkit

import (
	"strings"
	"testing"
)

func TestAssemblerKeepsRequiredAndRelevantSources(t *testing.T) {
	sources := []Source{
		{ID: "rules", Kind: "policy", Content: "所有修改必须运行测试", Trust: Trusted, Required: true},
		{ID: "go", Kind: "document", Content: "Go HTTP 服务使用 context 控制取消", Trust: Untrusted, Priority: 2},
		{ID: "python", Kind: "document", Content: "Python notebook", Trust: Untrusted, Priority: 10},
	}
	bundle, err := (Assembler{MaxRunes: 250, MaxItems: 2}).Build("修改 Go HTTP 服务", sources)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Items) != 2 || bundle.Items[0].Source.ID != "rules" || bundle.Items[1].Source.ID != "go" {
		t.Fatalf("unexpected selection: %+v", bundle.Items)
	}
	if !strings.Contains(bundle.Text, `trust="untrusted"`) || !strings.Contains(bundle.Text, "CONTENT_JSON") {
		t.Fatalf("untrusted source was not framed: %s", bundle.Text)
	}
}

func TestAssemblerFailsClosedForRequiredOverflow(t *testing.T) {
	_, err := (Assembler{MaxRunes: 20, MaxItems: 1}).Build("", []Source{{ID: "rules", Kind: "policy", Content: "required content", Trust: Trusted, Required: true}})
	if err == nil {
		t.Fatal("required source overflow unexpectedly succeeded")
	}
}

func TestAssemblerRejectsDuplicateSources(t *testing.T) {
	_, err := (Assembler{MaxRunes: 1000, MaxItems: 2}).Build("", []Source{{ID: "same", Kind: "a"}, {ID: "same", Kind: "b"}})
	if err == nil {
		t.Fatal("duplicate source unexpectedly accepted")
	}
}
