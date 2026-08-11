package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// ---- 测试基建:临时 SQLite + mock DeepSeek ----

type mockDeepSeek struct {
	responses []string
	calls     int
}

func newTestServer(t *testing.T, responses []string) (*httptest.Server, *Store) {
	t.Helper()
	store, err := OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	md := &mockDeepSeek{responses: responses}
	llmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		t.Logf("[mock DeepSeek] 请求 %d: %s", md.calls, truncate(string(body), 140))
		resp := `{"intent":"unknown","error":"mock 无响应","explanation":"mock 无响应"}`
		if md.calls < len(md.responses) {
			resp = md.responses[md.calls]
		}
		md.calls++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{
				"message": map[string]any{"role": "assistant", "content": resp},
			}},
		})
	}))
	ds := &DeepSeekClient{APIKey: "test", BaseURL: llmSrv.URL, Model: "mock", HTTP: llmSrv.Client()}
	srv := &Server{store: store, llm: ds}
	api := httptest.NewServer(srv.routes())
	t.Cleanup(func() {
		api.Close()
		llmSrv.Close()
		store.Close()
	})
	return api, store
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

func postJSON(t *testing.T, url string, body any) map[string]any {
	t.Helper()
	return requestJSON(t, http.MethodPost, url, body)
}

func requestJSON(t *testing.T, method, url string, body any) map[string]any {
	t.Helper()
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s %s 返回 %d: %s", method, url, resp.StatusCode, string(data))
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func createSession(t *testing.T, api string) string {
	t.Helper()
	out := postJSON(t, api+"/api/sessions", map[string]string{"title": "测试会话"})
	id, _ := out["id"].(string)
	if id == "" {
		t.Fatal("创建会话失败,没有 id")
	}
	return id
}

func sendChat(t *testing.T, api, sessionID, message string) map[string]any {
	t.Helper()
	return postJSON(t, api+"/api/chat", map[string]string{"sessionId": sessionID, "message": message})
}

func getSessionJSON(t *testing.T, api, id string) map[string]any {
	t.Helper()
	resp, err := http.Get(api + "/api/sessions/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func equationCount(gs map[string]any) int {
	eqs, _ := gs["equations"].([]any)
	return len(eqs)
}

// ---- 验收用例(设计稿 §22)----

const mockPlot = `{"intent":"plot","equations":[{"type":"function","expression":"y = x","normalizedExpression":"x","label":"y = x","color":"#2563eb","visible":true,"lineWidth":2}],"explanation":"已绘制 y = x"}`
const mockPlotQuad = `{"intent":"plot","equations":[{"type":"function","expression":"y = x^2","normalizedExpression":"x^2","label":"y = x²","color":"#2563eb","visible":true,"lineWidth":2}],"explanation":"这是一个开口向上的抛物线"}`
const mockAddSin = `{"intent":"add_equation","equations":[{"type":"function","expression":"y = sin(x)","normalizedExpression":"sin(x)","label":"y = sin(x)","color":"#22c55e","visible":true,"lineWidth":2}],"explanation":"已添加 y = sin(x)"}`
const mockUpdateRed = `{"intent":"update_equation","targetEquationId":"eq_001","updates":{"color":"#ef4444"},"explanation":"已改为红色"}`
const mockViewport = `{"intent":"update_viewport","viewport":{"xMin":-5,"xMax":5,"yMin":-5,"yMax":5},"explanation":"坐标范围已调整"}`
const mockUnknown = `{"intent":"unknown","error":"无法识别明确的数学方程或绘图需求","explanation":"我还无法确定你想绘制的具体方程"}`

// 测试1:绘制一次函数
func TestPlotOnceFunction(t *testing.T) {
	api, _ := newTestServer(t, []string{mockPlot})
	id := createSession(t, api.URL)
	out := sendChat(t, api.URL, id, "画 y = x")
	gs, _ := out["graphState"].(map[string]any)
	if equationCount(gs) != 1 {
		t.Fatalf("应有一条方程,实际 %v", gs)
	}
	msg, _ := out["message"].(map[string]any)
	if msg["role"] != "assistant" || !strings.Contains(msg["content"].(string), "y = x") {
		t.Fatalf("assistant 消息异常: %v", msg)
	}
}

// 测试2:绘制二次函数
func TestPlotQuadratic(t *testing.T) {
	api, _ := newTestServer(t, []string{mockPlotQuad})
	id := createSession(t, api.URL)
	out := sendChat(t, api.URL, id, "画 y = x^2")
	gs, _ := out["graphState"].(map[string]any)
	eqs, _ := gs["equations"].([]any)
	first := eqs[0].(map[string]any)
	if first["normalizedExpression"] != "x^2" || first["label"] != "y = x²" {
		t.Fatalf("方程字段不符: %v", first)
	}
}

// 测试3:追加函数,旧曲线不被覆盖
func TestAddEquationKeepsOld(t *testing.T) {
	api, _ := newTestServer(t, []string{mockPlotQuad, mockAddSin})
	id := createSession(t, api.URL)
	sendChat(t, api.URL, id, "画 y = x^2")
	out := sendChat(t, api.URL, id, "再加一条 y = sin(x)")
	gs, _ := out["graphState"].(map[string]any)
	if equationCount(gs) != 2 {
		t.Fatalf("应有两条方程,实际 %d", equationCount(gs))
	}
}

// 测试4:修改第一条曲线颜色
func TestUpdateColor(t *testing.T) {
	api, _ := newTestServer(t, []string{mockPlot, mockUpdateRed})
	id := createSession(t, api.URL)
	sendChat(t, api.URL, id, "画 y = x")
	out := sendChat(t, api.URL, id, "把第一条曲线改成红色")
	gs, _ := out["graphState"].(map[string]any)
	eqs, _ := gs["equations"].([]any)
	first := eqs[0].(map[string]any)
	if first["color"] != "#ef4444" {
		t.Fatalf("颜色应改为红色,实际 %v", first["color"])
	}
}

// 测试5:修改坐标范围
func TestUpdateViewport(t *testing.T) {
	api, _ := newTestServer(t, []string{mockViewport})
	id := createSession(t, api.URL)
	out := sendChat(t, api.URL, id, "把坐标范围改成 -5 到 5")
	gs, _ := out["graphState"].(map[string]any)
	vp, _ := gs["viewport"].(map[string]any)
	if vp["xMin"] != -5.0 || vp["xMax"] != 5.0 || vp["yMin"] != -5.0 || vp["yMax"] != 5.0 {
		t.Fatalf("viewport 不符: %v", vp)
	}
}

// 测试6:非法公式 -> unknown,原图像不丢失
func TestUnknownIntent(t *testing.T) {
	api, _ := newTestServer(t, []string{mockPlot, mockUnknown})
	id := createSession(t, api.URL)
	sendChat(t, api.URL, id, "画 y = x")
	out := sendChat(t, api.URL, id, "画一个很厉害的东西")
	gs, _ := out["graphState"].(map[string]any)
	if equationCount(gs) != 1 {
		t.Fatalf("非法请求不应清空图像,实际 %d 条", equationCount(gs))
	}
	msg, _ := out["message"].(map[string]any)
	if !strings.Contains(msg["content"].(string), "无法完成") {
		t.Fatalf("应返回错误提示: %v", msg["content"])
	}
}

// 测试7:会话隔离——A 画图不影响 B
func TestSessionIsolation(t *testing.T) {
	api, _ := newTestServer(t, []string{mockPlot})
	idA := createSession(t, api.URL)
	idB := createSession(t, api.URL)
	sendChat(t, api.URL, idA, "画 y = x")
	gsB := getSessionJSON(t, api.URL, idB)
	if equationCount(gsB["graphState"].(map[string]any)) != 0 {
		t.Fatal("会话 B 不应受到会话 A 影响")
	}
	msgs, _ := gsB["messages"].([]any)
	if len(msgs) != 0 {
		t.Fatalf("会话 B 不应有消息,实际 %d", len(msgs))
	}
}

// 测试8:方程列表删除(通过 PATCH 更新 graphState)后持久化
func TestDeleteEquationViaPatch(t *testing.T) {
	api, _ := newTestServer(t, nil)
	id := createSession(t, api.URL)
	// 先塞一条方程
	gs := map[string]any{
		"equations": []any{map[string]any{
			"id": "eq_001", "type": "function", "expression": "y = x",
			"normalizedExpression": "x", "label": "y = x", "color": "#2563eb",
			"visible": true, "lineWidth": 2,
		}},
		"viewport": map[string]any{"xMin": -10.0, "xMax": 10.0, "yMin": -10.0, "yMax": 10.0},
		"settings": map[string]any{"showGrid": true, "showAxis": true, "showLegend": true, "sampleCount": 1000},
	}
	requestJSON(t, http.MethodPatch, api.URL+"/api/sessions/"+id, map[string]any{"graphState": gs})
	// 删除方程(清空)
	gs["equations"] = []any{}
	requestJSON(t, http.MethodPatch, api.URL+"/api/sessions/"+id, map[string]any{"graphState": gs})
	got := getSessionJSON(t, api.URL, id)
	if equationCount(got["graphState"].(map[string]any)) != 0 {
		t.Fatal("PATCH 删除方程后应持久化为空")
	}
}

// 测试9:导出 PNG 是前端(Plotly.toImage)功能,见 web/README 与浏览器手测。

// ---- 单元测试:解析与状态更新规则 ----

func TestParseStructuredResultStripsFence(t *testing.T) {
	raw := "```json\n" + mockPlot + "\n```"
	res, err := ParseStructuredResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if res.Intent != "plot" || len(res.Equations) != 1 {
		t.Fatalf("解析失败: %+v", res)
	}
	if res.Equations[0].Color != "#2563eb" || res.Equations[0].LineWidth != 2 {
		t.Fatalf("默认值补齐失败: %+v", res.Equations[0])
	}
}

func TestApplyViewportPartial(t *testing.T) {
	gs := defaultGraphState()
	neg := -5.0
	res := &LLMStructuredResult{Intent: "update_viewport", Viewport: &ViewportUpdate{XMin: &neg, XMax: &neg}}
	ApplyResult(&gs, res)
	if gs.Viewport.XMin != -5 || gs.Viewport.XMax != -5 || gs.Viewport.YMin != -10 || gs.Viewport.YMax != 10 {
		t.Fatalf("部分更新不应影响未指定维度: %+v", gs.Viewport)
	}
}

func TestApplyUpdateDefaultLast(t *testing.T) {
	gs := defaultGraphState()
	gs.Equations = []EquationItem{
		{ID: "eq_001", Expression: "y = x", Visible: true, Color: "#2563eb", LineWidth: 2},
		{ID: "eq_002", Expression: "y = x^2", Visible: true, Color: "#ef4444", LineWidth: 2},
	}
	red := "#22c55e"
	res := &LLMStructuredResult{Intent: "update_equation", Updates: &EquationUpdate{Color: &red}}
	ApplyResult(&gs, res)
	if gs.Equations[1].Color != "#22c55e" {
		t.Fatalf("无目标时应默认修改最后一条: %+v", gs.Equations[1])
	}
}

func TestApplyRemove(t *testing.T) {
	gs := defaultGraphState()
	gs.Equations = []EquationItem{
		{ID: "eq_001", Expression: "y = x", Visible: true},
		{ID: "eq_002", Expression: "y = sin(x)", Visible: true},
	}
	res := &LLMStructuredResult{Intent: "remove_equation", TargetEquationID: "eq_001"}
	ApplyResult(&gs, res)
	if len(gs.Equations) != 1 || string(gs.Equations[0].ID) != "eq_002" {
		t.Fatalf("删除失败: %+v", gs.Equations)
	}
}

// 用户说"再加"但模型误判为 plot 时,应被兜底改成 add_equation。
func TestCoerceAddIntent(t *testing.T) {
	gs := defaultGraphState()
	gs.Equations = []EquationItem{{ID: "eq_001", Expression: "y = x", Visible: true}}
	res := &LLMStructuredResult{Intent: "plot", Equations: []EquationItem{
		{Expression: "y = sin(x)", NormalizedExpression: "sin(x)", Visible: true},
	}}
	CoerceIntent(res, "再加一条 y = sin(x)", &gs)
	if res.Intent != "add_equation" {
		t.Fatalf("应被兜底为 add_equation,实际 %s", res.Intent)
	}
	ApplyResult(&gs, res)
	if len(gs.Equations) != 2 {
		t.Fatalf("应追加而非替换,实际 %d 条", len(gs.Equations))
	}
}

// 模型给 analysis.zeros 返回字符串(而非数组)时,解析不应失败。
func TestExplainLenientAnalysis(t *testing.T) {
	raw := `{"intent":"explain","analysis":{"type":"sine_wave","zeros":"x = kπ (k 为整数)","monotonicity":"周期函数"},"explanation":"正弦函数"}`
	res, err := ParseStructuredResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if res.Intent != "explain" || res.Analysis == nil {
		t.Fatalf("解析失败: %+v", res)
	}
}

// 模型把 id 输出成数字时,解析不应失败。
func TestFlexIDAcceptsNumber(t *testing.T) {
	raw := `{"intent":"plot","equations":[{"id":1,"expression":"y = x","normalizedExpression":"x"}]}`
	res, err := ParseStructuredResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if string(res.Equations[0].ID) != "1" {
		t.Fatalf("数字 id 应转成字符串: %q", res.Equations[0].ID)
	}
}
