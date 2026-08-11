package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Skill 代表一个"技能包":一段给 LLM 的行为规范/知识说明。
// 这是 SKILL 概念的最小形态:把可复用的说明打包成文件,启动时注入提示词。
type Skill struct {
	Name         string
	Instructions string
}

// loadSkills 读取 skills/ 目录下的所有 *.md 文件。
// 约定:文件第一行是 "# 技能名",其余是技能内容。
func loadSkills(dir string) ([]Skill, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var skills []Skill
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			continue
		}
		lines := strings.SplitN(text, "\n", 2)
		name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[0]), "#"))
		instructions := ""
		if len(lines) == 2 {
			instructions = strings.TrimSpace(lines[1])
		}
		skills = append(skills, Skill{Name: name, Instructions: instructions})
	}
	return skills, nil
}
