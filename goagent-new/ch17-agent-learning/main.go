package main

import (
	"fmt"

	"goagent-new/pkg/learning"
)

func runLearningDemo() (learning.Experience, []learning.Experience, error) {
	store, err := learning.Open("")
	if err != nil {
		return learning.Experience{}, nil, err
	}
	candidate, err := store.Propose(learning.Experience{
		TaskType: "go_http", Situation: "测试服务启动失败，trace 显示端口被占用", Action: "启动前检测监听端口",
		Outcome: "后续两次任务均提前发现冲突", Lesson: "启动本地服务前先执行无副作用的端口检查",
		Tags: []string{"go", "http", "port"}, Confidence: 0.92, Evidence: []string{"trace-101", "trace-117"},
	})
	if err != nil {
		return learning.Experience{}, nil, err
	}
	_, err = store.Verify(candidate.ID, learning.Gate{MinEvidence: 2, MinConfidence: 0.8, MinPassRate: 0.9}, learning.Review{Passed: 10, Total: 10})
	if err != nil {
		return learning.Experience{}, nil, err
	}
	promoted, err := store.Promote(candidate.ID, false)
	if err != nil {
		return learning.Experience{}, nil, err
	}
	return promoted, store.Search("go_http", "启动 HTTP 端口", "global", 3), nil
}

func main() {
	experience, matches, err := runLearningDemo()
	if err != nil {
		panic(err)
	}
	fmt.Printf("experience=%s status=%s retrieved=%d lesson=%q\n", experience.ID, experience.Status, len(matches), experience.Lesson)
}
