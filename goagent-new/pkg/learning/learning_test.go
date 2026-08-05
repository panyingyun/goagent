package learning

import (
	"path/filepath"
	"testing"
)

func TestExperienceRequiresRegressionGateBeforeRetrieval(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "experiences.json"))
	if err != nil {
		t.Fatal(err)
	}
	experience, err := store.Propose(Experience{TaskType: "go_test", Situation: "测试因端口占用失败", Action: "检查端口", Outcome: "恢复", Lesson: "启动服务前检查端口", Tags: []string{"go", "端口"}, Confidence: 0.9, Evidence: []string{"trace-1", "trace-2"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := store.Search("go_test", "端口", "global", 3); len(got) != 0 {
		t.Fatal("candidate experience leaked into retrieval")
	}
	verified, err := store.Verify(experience.ID, Gate{MinEvidence: 2, MinConfidence: 0.8, MinPassRate: 0.9}, Review{Passed: 10, Total: 10})
	if err != nil {
		t.Fatal(err)
	}
	if verified.Status != Verified {
		t.Fatalf("unexpected verified status: %s", verified.Status)
	}
	promoted, err := store.Promote(experience.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if promoted.Status != Promoted {
		t.Fatalf("unexpected status: %s", promoted.Status)
	}
	if got := store.Search("go_test", "端口失败", "global", 3); len(got) != 1 || got[0].ID != experience.ID {
		t.Fatalf("promoted experience was not retrieved: %+v", got)
	}
	reopened, err := Open(filepath.Join(filepath.Dir(store.path), "experiences.json"))
	if err != nil || len(reopened.List()) != 1 {
		t.Fatalf("experience was not persisted: %v", err)
	}
}

func TestHighRiskExperienceRequiresHumanApproval(t *testing.T) {
	store, _ := Open("")
	experience, err := store.Propose(Experience{TaskType: "deployment", Situation: "发布失败", Lesson: "自动回滚生产", Confidence: 1, Evidence: []string{"e1", "e2"}, Risk: HighRisk})
	if err != nil {
		t.Fatal(err)
	}
	gate := Gate{MinEvidence: 2, MinConfidence: 0.8, MinPassRate: 1}
	if _, err := store.Verify(experience.ID, gate, Review{Passed: 5, Total: 5}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Promote(experience.ID, false); err == nil {
		t.Fatal("high-risk experience was promoted without human approval")
	}
	if _, err := store.Promote(experience.ID, true); err != nil {
		t.Fatal(err)
	}
}

func TestRegressionBlocksPromotion(t *testing.T) {
	store, _ := Open("")
	experience, _ := store.Propose(Experience{TaskType: "review", Situation: "重复错误", Lesson: "增加检查", Confidence: 0.9, Evidence: []string{"e1", "e2"}})
	_, err := store.Verify(experience.ID, Gate{MinEvidence: 2, MinConfidence: 0.8, MinPassRate: 0.8}, Review{Passed: 9, Total: 10, Regressions: 1})
	if err == nil {
		t.Fatal("regression unexpectedly passed the gate")
	}
}
