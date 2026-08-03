package main

import (
	"fmt"
	"strings"
)

// ApplyResult 按设计稿 §14 的规则,把 structuredResult 应用到 graphState。
// 返回要展示给用户的 assistant 文本。
func ApplyResult(gs *GraphState, res *LLMStructuredResult) string {
	switch res.Intent {
	case "plot", "add_equation":
		// plot:设计稿 MVP 建议替换旧方程;add_equation:追加。
		if res.Intent == "plot" {
			gs.Equations = []EquationItem{}
		}
		for _, e := range res.Equations {
			e.ID = nextEquationID(gs)
			gs.Equations = append(gs.Equations, e)
		}

	case "update_equation":
		idx := findEquation(gs, string(res.TargetEquationID))
		if idx < 0 {
			idx = len(gs.Equations) - 1 // 无明确目标默认最后一条(设计稿 §14.3)
		}
		if idx >= 0 && res.Updates != nil {
			e := &gs.Equations[idx]
			u := res.Updates
			if u.Color != nil {
				e.Color = *u.Color
			}
			if u.LineWidth != nil {
				e.LineWidth = *u.LineWidth
			}
			if u.Visible != nil {
				e.Visible = *u.Visible
			}
			if u.Label != nil {
				e.Label = *u.Label
			}
			if u.Expression != nil {
				e.Expression = *u.Expression
			}
			if u.NormalizedExpression != nil {
				e.NormalizedExpression = *u.NormalizedExpression
			}
		}

	case "remove_equation":
		idx := findEquation(gs, string(res.TargetEquationID))
		if idx >= 0 {
			gs.Equations = append(gs.Equations[:idx], gs.Equations[idx+1:]...)
		}

	case "update_viewport":
		if res.Viewport != nil {
			v := res.Viewport
			if v.XMin != nil {
				gs.Viewport.XMin = *v.XMin
			}
			if v.XMax != nil {
				gs.Viewport.XMax = *v.XMax
			}
			if v.YMin != nil {
				gs.Viewport.YMin = *v.YMin
			}
			if v.YMax != nil {
				gs.Viewport.YMax = *v.YMax
			}
		}

	case "analyze", "explain":
		gs.Analysis = res.Analysis
	}

	if res.Error != "" {
		return "抱歉,我无法完成这个请求:" + res.Error
	}
	if res.Explanation != "" {
		return res.Explanation
	}
	return "已完成操作: " + res.Intent
}

func nextEquationID(gs *GraphState) FlexID {
	return FlexID(fmt.Sprintf("eq_%03d", len(gs.Equations)+1))
}

func findEquation(gs *GraphState, id string) int {
	if id == "" {
		return -1
	}
	for i, e := range gs.Equations {
		if string(e.ID) == id {
			return i
		}
	}
	return -1
}

// addKeywords 是"追加曲线"的用户措辞;命中时即使模型返回 plot 也强制改为
// add_equation,避免覆盖已有图像(设计稿 §13:不要直接信任模型结果)。
var addKeywords = []string{"再加", "再画", "增加", "追加", "多画", "加上", "再来一条", "补充"}

func CoerceIntent(res *LLMStructuredResult, userMsg string, gs *GraphState) {
	if res.Intent != "plot" || len(gs.Equations) == 0 {
		return
	}
	for _, kw := range addKeywords {
		if strings.Contains(userMsg, kw) {
			res.Intent = "add_equation"
			return
		}
	}
}
