# -*- coding: utf-8 -*-
"""
为 ch01~ch26 生成架构图:
  chXX/docs/architecture-code-map.svg|png    代码架构图(文件结构与调用流)
  chXX/docs/architecture-diagram.svg|png     业务架构图(概念/数据流)

同一套布局数据同时输出 SVG 和 PNG(PNG 用 2x 超采样 + LANCZOS 保证文字清晰)。
代码更新后,重新运行本脚本即可同步:
  python scripts/generate_arch_diagrams.py
"""

import os
import xml.sax.saxutils as sax

from PIL import Image, ImageDraw, ImageFont

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
FONT_PATH = "C:/Windows/Fonts/msyh.ttc"

# 配色(与 ch26 主题色设计稿接近)
PRIMARY = "#2563eb"
SECONDARY = "#ef4444"
GREEN = "#22c55e"
BG = "#faf8ff"
BOX_FILL = "#ffffff"
BORDER = "#c3c6d7"
TEXT = "#191b23"
MUTED = "#434655"

W, H = 1200, 760


def font(size):
    return ImageFont.truetype(FONT_PATH, size)


def wrap_text(text, fnt, max_w):
    lines = []
    for raw in text.split("\n"):
        line = ""
        for ch in raw:
            if fnt.getbbox(line + ch)[2] <= max_w:
                line += ch
            else:
                lines.append(line)
                line = ch
        lines.append(line)
    return lines


def box_size(label, sub, fnt_label, fnt_sub, max_w):
    w = 0
    for l in wrap_text(label, fnt_label, max_w - 40):
        w = max(w, fnt_label.getbbox(l)[2])
    if sub:
        for l in wrap_text(sub, fnt_sub, max_w - 40):
            w = max(w, fnt_sub.getbbox(l)[2])
    return min(max_w, w + 40)


def layout_flow(flow, x, w, y0, step, box_h, fnt_label, fnt_sub):
    """返回节点矩形列表。"""
    boxes = []
    for label, sub in flow:
        bw = box_size(label, sub, fnt_label, fnt_sub, w)
        boxes.append((x + (w - bw) // 2, y0, bw, box_h, label, sub))
        y0 += step
    return boxes


# ---------------- SVG 输出 ----------------

def svg_escape(s):
    return sax.escape(s, {'"': "&quot;"})


def svg_node(box, fill=BOX_FILL, border=BORDER, label_color=TEXT, sub_color=MUTED):
    x, y, w, h, label, sub = box
    out = []
    out.append(
        f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="10" fill="{fill}" '
        f'stroke="{border}" stroke-width="1.5"/>'
    )
    ly = y + h / 2
    if sub:
        ly = y + h / 2 - 12
    out.append(
        f'<text x="{x + w / 2}" y="{ly}" fill="{label_color}" font-family="Microsoft YaHei, sans-serif" '
        f'font-size="17" font-weight="600" text-anchor="middle">{svg_escape(label)}</text>'
    )
    if sub:
        yy = y + h / 2 + 16
        for line in wrap_text(sub, font(13), w - 20):
            out.append(
                f'<text x="{x + w / 2}" y="{yy}" fill="{sub_color}" font-family="Microsoft YaHei, sans-serif" '
                f'font-size="13" text-anchor="middle">{svg_escape(line)}</text>'
            )
            yy += 17
    return "".join(out)


def svg_arrow(x1, y1, x2, y2, color=PRIMARY):
    return (
        f'<line x1="{x1}" y1="{y1}" x2="{x2}" y2="{y2}" stroke="{color}" '
        f'stroke-width="2" marker-end="url(#arrow)"/>'
    )


def svg_header(title, kind):
    return (
        '<?xml version="1.0" encoding="UTF-8"?>\n'
        '<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="760" viewBox="0 0 1200 760">\n'
        '<defs><marker id="arrow" markerWidth="10" markerHeight="10" refX="8" refY="3" '
        'orient="auto"><path d="M0,0 L8,3 L0,6 z" fill="#2563eb"/></marker></defs>\n'
        f'<rect width="1200" height="760" fill="{BG}"/>\n'
        f'<text x="40" y="48" fill="{TEXT}" font-family="Microsoft YaHei, sans-serif" font-size="26" '
        f'font-weight="700">{svg_escape(title)}</text>\n'
        f'<text x="40" y="76" fill="{MUTED}" font-family="Microsoft YaHei, sans-serif" font-size="14">'
        f'{svg_escape(kind)}</text>\n'
    )


def svg_footer():
    return (
        '<text x="1190" y="744" fill="#737686" font-family="Microsoft YaHei, sans-serif" '
        'font-size="11" text-anchor="end">Go Agent 课程 ch01~ch26</text>\n</svg>\n'
    )


def render_svg(path, title, kind, flow_boxes, arrows, extras=None):
    parts = [svg_header(title, kind)]
    for box in flow_boxes:
        parts.append(svg_node(box))
    for a in arrows:
        parts.append(svg_arrow(*a))
    if extras:
        parts.append(
            '<text x="700" y="110" fill="'
            + MUTED
            + '" font-family="Microsoft YaHei, sans-serif" font-size="15" font-weight="600">其他文件</text>'
        )
        y = 138
        for name, note in extras:
            parts.append(
                f'<rect x="700" y="{y}" width="450" height="52" rx="8" fill="#f3f3fe" '
                f'stroke="{BORDER}" stroke-width="1"/>'
            )
            parts.append(
                f'<text x="716" y="{y + 22}" fill="{TEXT}" font-family="Microsoft YaHei, sans-serif" '
                f'font-size="14" font-weight="600">{svg_escape(name)}</text>'
            )
            if note:
                parts.append(
                    f'<text x="716" y="{y + 40}" fill="{MUTED}" font-family="Microsoft YaHei, sans-serif" '
                    f'font-size="12">{svg_escape(note)}</text>'
                )
            y += 62
    parts.append(svg_footer())
    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(parts))


# ---------------- PNG 输出 ----------------

SCALE = 2


def png_font(size):
    return font(size * SCALE)


def png_node(draw, box, fnt_label, fnt_sub):
    x, y, w, h, label, sub = box
    s = SCALE
    draw.rounded_rectangle(
        [x * s, y * s, (x + w) * s, (y + h) * s], radius=10 * s, fill=BOX_FILL, outline=BORDER, width=2
    )
    ly = (y + h / 2 - 6) * s
    if sub:
        ly = (y + h / 2 - 14) * s
    draw.text(((x + w / 2) * s, ly), label, font=fnt_label, fill=TEXT, anchor="ma")
    if sub:
        yy = (y + h / 2 + 8) * s
        for line in wrap_text(sub, png_font(13), (w - 20) * s):
            draw.text(((x + w / 2) * s, yy), line, font=fnt_sub, fill=MUTED, anchor="ma")
            yy += 17 * s


def png_arrow(draw, x1, y1, x2, y2, color=PRIMARY):
    s = SCALE
    draw.line([x1 * s, y1 * s, x2 * s, y2 * s], fill=color, width=3)
    # 箭头
    dx, dy = x2 - x1, y2 - y1
    length = max(1, (dx * dx + dy * dy) ** 0.5)
    ux, uy = dx / length, dy / length
    px, py = -uy, ux
    tip = (x2 * s, y2 * s)
    base = (x2 * s - ux * 12 * s, y2 * s - uy * 12 * s)
    left = (base[0] + px * 6 * s, base[1] + py * 6 * s)
    right = (base[0] - px * 6 * s, base[1] - py * 6 * s)
    draw.polygon([tip, left, right], fill=color)


def render_png(path, title, kind, flow_boxes, arrows, extras=None):
    img = Image.new("RGB", (W * SCALE, H * SCALE), BG)
    draw = ImageDraw.Draw(img)
    fnt_title = png_font(26)
    fnt_kind = png_font(14)
    fnt_label = png_font(17)
    fnt_sub = png_font(13)
    fnt_small = png_font(14)
    fnt_note = png_font(12)
    draw.text((40 * SCALE, 26 * SCALE), title, font=fnt_title, fill=TEXT)
    draw.text((40 * SCALE, 58 * SCALE), kind, font=fnt_kind, fill=MUTED)
    for box in flow_boxes:
        png_node(draw, box, fnt_label, fnt_sub)
    for a in arrows:
        png_arrow(draw, *a)
    if extras:
        draw.text((700 * SCALE, 86 * SCALE), "其他文件", font=fnt_small, fill=MUTED)
        y = 112
        for name, note in extras:
            draw.rounded_rectangle(
                [700 * SCALE, y * SCALE, 1150 * SCALE, (y + 50) * SCALE],
                radius=8 * SCALE, fill="#f3f3fe", outline=BORDER, width=2,
            )
            draw.text((716 * SCALE, (y + 12) * SCALE), name, font=fnt_small, fill=TEXT)
            if note:
                draw.text((716 * SCALE, (y + 33) * SCALE), note, font=fnt_note, fill=MUTED)
            y += 60
    img = img.resize((W, H), Image.LANCZOS)
    img.save(path)


# ---------------- 章节数据 ----------------

CHAPTERS = {
    "ch01": {
        "title": "ch01 Go 调用 LLM(DeepSeek)",
        "code": {
            "flow": [
                ("main.go", "读取问题(参数/标准输入)"),
                ("chat()", "组装 system + user 消息"),
                ("POST /chat/completions", "Authorization: Bearer Key"),
                ("解析响应", "choices[0].message.content"),
                ("main.go", "打印模型回复"),
            ],
            "extra": [("README.md", "运行方式与概念讲解")],
        },
        "biz": [
            ("用户问题", "如:用一句话介绍你自己"),
            ("DeepSeek Chat Completions", "OpenAI 兼容接口"),
            ("模型回复", "deepseek-v4-flash"),
            ("控制台输出", "模型说的话"),
        ],
    },
    "ch02": {
        "title": "ch02 Go 实现 MCP 服务器与客户端",
        "code": {
            "flow": [
                ("client/main.go", "启动服务器子进程"),
                ("MCPClient", "stdio + JSON-RPC 2.0"),
                ("initialize / tools/list / tools/call", "协议消息"),
                ("server/main.go", "handle() 按方法分发"),
                ("callTool()", "echo / add 工具执行"),
            ],
            "extra": [("README.md", "协议讲解与运行方式")],
        },
        "biz": [
            ("客户端", "JSON-RPC 请求"),
            ("stdio 管道", "每行一个 JSON 消息"),
            ("MCP 服务器", "initialize 握手"),
            ("tools/list", "列出工具"),
            ("tools/call", "调用 echo / add"),
            ("返回结果", "result.content"),
        ],
    },
    "ch03": {
        "title": "ch03 Go 实现 SKILL 调用",
        "code": {
            "flow": [
                ("main.go", "读取用户问题"),
                ("loadSkills()", "读取 skills/*.md"),
                ("拼接 system 提示词", "技能注入提示词"),
                ("chat()", "调用 DeepSeek"),
                ("输出回复", "模型遵守技能"),
            ],
            "extra": [
                ("skills/math-tutor.md", "数学三步回答规范"),
                ("skills/translator.md", "翻译格式规范"),
                ("README.md", "SKILL 概念讲解"),
            ],
        },
        "biz": [
            ("skills/*.md", "行为说明书(提示词资产)"),
            ("loadSkills()", "读取并解析文件名"),
            ("注入 system", "## 你掌握的技能(SKILL)"),
            ("DeepSeek", "遵守技能规范"),
            ("回答", "按技能格式输出"),
        ],
    },
    "ch04": {
        "title": "ch04 Go 实现 ReAct 循环",
        "code": {
            "flow": [
                ("main.go", "入口:读取问题"),
                ("agent.go runReAct()", "思考-行动-观察循环"),
                ("llm.go Chat()", "DeepSeek 调用"),
                ("tools.go Registry", "calculator / get_time"),
                ("parseReAct()", "解析 Action / Final Answer"),
            ],
            "extra": [
                ("agent_test.go", "mock LLM 离线测试"),
                ("README.md", "ReAct 范式讲解"),
            ],
        },
        "biz": [
            ("用户问题", "12 * 8 + 4 等于多少?"),
            ("LLM 思考", "Thought"),
            ("Action + Action Input", "calculator {\"expression\": ...}"),
            ("执行工具", "Observation: 100"),
            ("循环", "直到 Final Answer"),
            ("最终答案", "返回给用户"),
        ],
    },
    "ch05": {
        "title": "ch05 综合 Agent(ReAct + MCP + SKILL + LLM)",
        "code": {
            "flow": [
                ("main.go", "模式选择 react / tool"),
                ("agent.go", "runReAct / runToolCall"),
                ("llm.go", "DeepSeek 客户端"),
                ("tools.go + skills.go", "工具注册表 + 技能加载"),
                ("mcpclient.go → mcp/server", "MCP 工具接入"),
            ],
            "extra": [
                ("react_test.go", "ReAct + MCP 集成测试"),
                ("skills/math-skill.md", "技能文件"),
                ("README.md", "综合架构讲解"),
            ],
        },
        "biz": [
            ("用户", "自然语言指令"),
            ("LLM DeepSeek", "负责思考"),
            ("工具注册表", "calculator / get_time"),
            ("SKILL 技能", "行为规范注入"),
            ("MCP 客户端", "stdio JSON-RPC"),
            ("MCP 服务器", "echo / get_time"),
            ("最终回答", "Final Answer"),
        ],
    },
    "ch06": {
        "title": "ch06 多轮对话与记忆",
        "code": {
            "flow": [
                ("main.go", "交互式 CLI 循环"),
                ("memory.go History", "对话历史 + 摘要"),
                ("CompressIfNeeded()", "token 超预算→LLM 压缩"),
                ("memory.go Preferences", "偏好持久化 memory.json"),
                ("llm.go", "DeepSeek 调用"),
            ],
            "extra": [
                ("memory_test.go", "压缩/持久化测试"),
                ("README.md", "记忆机制讲解"),
            ],
        },
        "biz": [
            ("用户输入", "记住我喜欢喝咖啡"),
            ("追加历史", "History"),
            ("token 估算", "超预算?"),
            ("LLM 压缩", "旧对话→摘要"),
            ("摘要注入 system", "早期对话摘要"),
            ("LLM 回复", "记得偏好"),
            ("持久化", "memory.json"),
        ],
    },
    "ch07": {
        "title": "ch07 流式输出(SSE)",
        "code": {
            "flow": [
                ("main.go", "读取问题"),
                ("llm.go StreamChat()", "stream: true 请求"),
                ("SSE 解析", "逐行 data: {...}"),
                ("onDelta 回调", "逐字打印"),
                ("data: [DONE]", "结束"),
            ],
            "extra": [
                ("llm_test.go", "mock SSE 服务器测试"),
                ("README.md", "SSE 格式讲解"),
            ],
        },
        "biz": [
            ("用户问题", "写一首短诗"),
            ("请求 stream: true", "DeepSeek"),
            ("SSE 分片", "data: {\"delta\": ...}"),
            ("逐字输出", "打字机效果"),
            ("[DONE]", "结束标志"),
        ],
    },
    "ch08": {
        "title": "ch08 RAG 入门(路线B:OpenAI embedding)",
        "code": {
            "flow": [
                ("main.go", "索引 + 问答入口"),
                ("rag.go", "切块 / Index 构建 / 检索"),
                ("embedder.go", "OpenAI text-embedding-3-small"),
                ("vector.go", "余弦相似度"),
                ("llm.go", "DeepSeek 回答"),
            ],
            "extra": [
                ("docs/sample.txt", "示例公司制度"),
                ("rag_test.go", "切块/检索/mock 测试"),
                ("README.md", "RAG 流程讲解"),
            ],
        },
        "biz": [
            ("文档 docs/*.txt", "私有知识"),
            ("切块 chunk", "每块 ≤300 字"),
            ("向量化", "OpenAI embedding"),
            ("索引 index.json", "片段 + 向量"),
            ("问题向量化", "余弦检索 top-3"),
            ("片段注入提示词", "system 资料"),
            ("DeepSeek 回答", "基于资料"),
        ],
    },
    "ch09": {
        "title": "ch09 真实工具(HTTP/文件/命令/存储)",
        "code": {
            "flow": [
                ("main.go", "入口 + 示例文件准备"),
                ("agent.go runReAct()", "ReAct 循环"),
                ("tools.go Registry", "6 个真实工具"),
                ("safePath() / allowedCommand()", "安全校验"),
                ("KVStore", "data/kv.json 持久化"),
            ],
            "extra": [
                ("tools_test.go", "路径穿越/白名单测试"),
                ("README.md", "工具安全三原则"),
            ],
        },
        "biz": [
            ("用户指令", "执行 go version"),
            ("ReAct 解析", "Action: run_command"),
            ("命令白名单", "拒绝危险命令"),
            ("执行(5s 超时)", "cmd / sh"),
            ("Observation", "go1.25.5 windows/amd64"),
            ("最终回答", "环境信息"),
        ],
    },
    "ch10": {
        "title": "ch10 评估、安全与工程化",
        "code": {
            "flow": [
                ("main.go", "-eval / HTTP 服务双模式"),
                ("agent.go runReAct()", "ChatLLM 接口注入"),
                ("eval.go", "评估用例 + 断言"),
                ("retry.go / security.go", "重试 + 注入防护"),
                ("logging.go", "结构化 JSON 日志"),
            ],
            "extra": [
                ("eval/cases.json", "3 个回归用例"),
                ("ch10_test.go", "重试/防护/HTTP 测试"),
                ("README.md", "工程化讲解"),
            ],
        },
        "biz": [
            ("HTTP 请求 /api/chat", "POST JSON"),
            ("Guard 检测", "提示注入防御"),
            ("Retry 重试", "指数退避"),
            ("ReAct 执行", "工具调用"),
            ("结构化日志", "trace 记录"),
            ("评估模式", "-eval 回归测试"),
        ],
    },
    "ch11": {
        "title": "ch11 提示词工程",
        "code": {
            "flow": [
                ("main.go", "读取待分析文本"),
                ("技巧1 系统提示词", "角色 + 规则"),
                ("技巧2 few-shot", "例子引导"),
                ("技巧3 结构化输出", "JSON + 校验"),
                ("extractJSON() / validateResult()", "解析与兜底"),
            ],
            "extra": [
                ("prompt_test.go", "few-shot/JSON 测试"),
                ("README.md", "三大技巧讲解"),
            ],
        },
        "biz": [
            ("文本输入", "情感分析"),
            ("系统提示词", "角色 + 规则"),
            ("few-shot 示例", "正面/负面/中性"),
            ("JSON 输出要求", "sentiment/confidence"),
            ("解析校验", "枚举值检查"),
            ("结论", "结构化结果"),
        ],
    },
    "ch12": {
        "title": "ch12 向量数据库与高级 RAG",
        "code": {
            "flow": [
                ("main.go", "索引 + 问答"),
                ("rag.go chunkText()", "文档切块"),
                ("embedder.go HashEmbedder", "本地哈希向量"),
                ("hybrid.go HybridRetriever", "关键词 + 向量融合"),
                ("rerank.go", "去重重排"),
            ],
            "extra": [
                ("docs/sample.txt", "示例文档"),
                ("hybrid_test.go", "混合检索测试"),
                ("README.md", "检索质量优化"),
            ],
        },
        "biz": [
            ("文档切块", "每块 ≤300 字"),
            ("哈希向量化", "256 维,离线可用"),
            ("关键词得分", "精确匹配"),
            ("向量得分", "语义相似"),
            ("融合排序", "top-5"),
            ("重排去重", "多样性 top-3"),
            ("DeepSeek 回答", "基于片段"),
        ],
    },
    "ch13": {
        "title": "ch13 工具进阶与并行调用",
        "code": {
            "flow": [
                ("main.go", "入口 + 演示工具"),
                ("agent.go runToolCallParallel()", "function calling 循环"),
                ("tools.go Executor", "校验/缓存/并行"),
                ("ExecuteParallel()", "worker 池并发执行"),
                ("Cache", "TTL 结果缓存"),
            ],
            "extra": [
                ("tools_test.go", "并行加速/缓存测试"),
                ("README.md", "三层升级讲解"),
            ],
        },
        "biz": [
            ("模型返回多个 tool_calls", "add / mul / sub"),
            ("参数校验", "必填字段检查"),
            ("缓存命中?", "60s 内直接返回"),
            ("并行执行", "worker 池 ≤4"),
            ("按序回填", "tool_call_id 对应"),
            ("LLM 汇总", "最终回答"),
        ],
    },
    "ch14": {
        "title": "ch14 Agent 工作流与任务规划",
        "code": {
            "flow": [
                ("main.go", "-plan / -run / -resume"),
                ("workflow.go Plan()", "LLM 拆解步骤"),
                ("RunStep()", "逐步骤执行 + 上下文"),
                ("NextStep() / Resume()", "状态推进"),
                ("Save() / LoadWorkflow()", "checkpoint.json"),
            ],
            "extra": [
                ("workflow_test.go", "规划/恢复测试"),
                ("README.md", "状态机讲解"),
            ],
        },
        "biz": [
            ("目标", "写周报提纲"),
            ("LLM 规划", "拆成 3~5 步 JSON"),
            ("步骤状态机", "pending/done/failed"),
            ("逐步骤执行", "携带前序产出"),
            ("检查点保存", "checkpoint.json"),
            ("断点续跑", "-resume"),
        ],
    },
    "ch15": {
        "title": "ch15 多 Agent 协作",
        "code": {
            "flow": [
                ("main.go", "创建三个角色 Agent"),
                ("agent.go Agent", "角色 + 独立记忆"),
                ("agent.go Orchestra.Run()", "编排循环"),
                ("Planner → Executor → Reviewer", "职责链"),
                ("llm.go", "共享 DeepSeek"),
            ],
            "extra": [
                ("agent_test.go", "返工/通过测试"),
                ("README.md", "多 Agent 思想"),
            ],
        },
        "biz": [
            ("任务", "写产品介绍"),
            ("规划者", "拆解执行计划"),
            ("执行者", "产出内容"),
            ("审查者", "质量检查"),
            ("不通过?", "返工 ≤N 轮"),
            ("通过", "最终交付"),
        ],
    },
    "ch16": {
        "title": "ch16 MCP 进阶(HTTP 传输 + 路由)",
        "code": {
            "flow": [
                ("main.go", "Router 挂载服务器"),
                ("mcp.go Router", "server.tool 命名空间"),
                ("httpmcp.go HTTPMCPHandler", "POST /mcp"),
                ("httpmcp.go HTTPMCPClient", "HTTP 调用"),
                ("handleRPC()", "JSON-RPC 分发"),
            ],
            "extra": [
                ("mcp_test.go", "HTTP/路由测试"),
                ("README.md", "传输与路由对比"),
            ],
        },
        "biz": [
            ("客户端", "echo.echo / mem.add"),
            ("Router 路由", "按服务器名前缀"),
            ("HTTP 传输", "POST /mcp"),
            ("MCP 服务器", "initialize/tools/list/call"),
            ("工具执行", "echo / add"),
            ("结果返回", "原路送回"),
        ],
    },
    "ch17": {
        "title": "ch17 Agent Web 界面",
        "code": {
            "flow": [
                ("main.go", "HTTP 服务器"),
                ("handleChat()", "POST /chat JSON"),
                ("handleStream()", "GET /chat/stream SSE"),
                ("llm.go StreamChat()", "流式调用"),
                ("web/index.html", "浏览器聊天界面"),
            ],
            "extra": [
                ("server_test.go", "JSON/SSE 测试"),
                ("README.md", "Web 界面讲解"),
            ],
        },
        "biz": [
            ("浏览器", "输入问题"),
            ("POST /chat", "普通问答"),
            ("GET /chat/stream", "SSE 流式"),
            ("LLM 回复", "逐字推送"),
            ("页面渲染", "打字机效果"),
        ],
    },
    "ch18": {
        "title": "ch18 模型路由与成本优化",
        "code": {
            "flow": [
                ("main.go", "提问入口"),
                ("router.go Choose()", "复杂度启发式"),
                ("Record()", "token 成本记账"),
                ("RespCache", "10 分钟缓存"),
                ("chatWithModel()", "按模型调用"),
            ],
            "extra": [
                ("router_test.go", "路由/预算/缓存测试"),
                ("README.md", "成本优化策略"),
            ],
        },
        "biz": [
            ("问题", "简单 or 复杂?"),
            ("选模型", "flash / pro"),
            ("预算检查", "超半预算降级"),
            ("缓存?", "相同问题直接返回"),
            ("回答", "省钱省时"),
        ],
    },
    "ch19": {
        "title": "ch19 评估进阶:LLM-as-Judge",
        "code": {
            "flow": [
                ("main.go", "答题 + 评审两阶段"),
                ("llm.go", "答题模型"),
                ("judge.go Judge.Score()", "评审模型打分"),
                ("extractJSONObject()", "解析评分 JSON"),
                ("分数夹取", "0~10 兜底"),
            ],
            "extra": [
                ("judge_test.go", "解析/夹取测试"),
                ("README.md", "主观评估方案"),
            ],
        },
        "biz": [
            ("问题", "goroutine 是什么?"),
            ("答题模型", "生成回答"),
            ("评审标准", "准确性/完整性/清晰度"),
            ("评审模型", "输出 score + reason"),
            ("分数校验", "0~10 夹取"),
            ("评估结论", "得分 + 理由"),
        ],
    },
    "ch20": {
        "title": "ch20 可观测性与日志",
        "code": {
            "flow": [
                ("main.go", "模拟 Agent 任务"),
                ("trace.go Tracer", "span 树(父子)"),
                ("Start() / End()", "耗时记录"),
                ("metrics.go Metrics", "计数 + 直方图"),
                ("Export()", "JSON 导出"),
            ],
            "extra": [
                ("trace_test.go", "span 树/指标测试"),
                ("README.md", "可观测性讲解"),
            ],
        },
        "biz": [
            ("Agent 任务", "llm.think"),
            ("span 入栈", "tool.calculator"),
            ("span 出栈", "llm.answer"),
            ("指标", "计数/平均耗时"),
            ("导出", "span 树 JSON"),
        ],
    },
    "ch21": {
        "title": "ch21 部署与运维",
        "code": {
            "flow": [
                ("main.go", "HTTP 服务"),
                ("handleReady()", "就绪检查"),
                ("rateLimiter", "每 IP 限流"),
                ("signal.NotifyContext", "优雅关闭"),
                ("Dockerfile", "容器化部署"),
            ],
            "extra": [
                ("main_test.go", "限流/健康测试"),
                ("README.md", "上线要素讲解"),
            ],
        },
        "biz": [
            ("请求", "限流检查"),
            ("/health /ready", "健康检查"),
            ("业务处理", "Agent 调用"),
            ("退出信号", "优雅关闭"),
            ("Docker", "容器部署"),
        ],
    },
    "ch22": {
        "title": "ch22 迷你 LangChain",
        "code": {
            "flow": [
                ("main.go", "链 + Agent 演示"),
                ("chain.go Chain 接口", "输入→输出抽象"),
                ("LLMChain / SequentialChain", "翻译链→总结链"),
                ("Memory / Tool", "记忆 + 工具"),
                ("AgentExecutor", "ReAct 循环"),
            ],
            "extra": [
                ("chain_test.go", "链/执行器测试"),
                ("README.md", "与 LangChain 对照表"),
            ],
        },
        "biz": [
            ("输入", "你好,世界"),
            ("LLMChain 翻译", "英文译文"),
            ("SequentialChain", "串联下一链"),
            ("LLMChain 总结", "一句话"),
            ("AgentExecutor", "计算器工具"),
            ("输出", "最终结果"),
        ],
    },
    "ch23": {
        "title": "ch23 迷你 LangGraph",
        "code": {
            "flow": [
                ("main.go", "构建工作流图"),
                ("graph.go Graph", "节点 + 边 + 状态"),
                ("AddNode / AddEdge", "图结构"),
                ("Invoke()", "执行 + __next__ 路由"),
                ("agent_graph.go", "plan/execute/review"),
            ],
            "extra": [
                ("graph_test.go", "线性/循环/检查点测试"),
                ("README.md", "与 LangGraph 对照"),
            ],
        },
        "biz": [
            ("任务", "写产品介绍"),
            ("plan 节点", "规划"),
            ("execute 节点", "执行"),
            ("review 节点", "审查"),
            ("不通过?", "__next__ 回 execute"),
            ("检查点", "checkpoint.json"),
        ],
    },
    "ch24": {
        "title": "ch24 LangSmith 风格追踪",
        "code": {
            "flow": [
                ("main.go", "Agent 执行包裹"),
                ("tracer.go Run", "run 树"),
                ("StartRun / EndRun", "父子嵌套"),
                ("FailRun()", "错误记录"),
                ("Export()", "嵌套 JSON"),
            ],
            "extra": [
                ("tracer_test.go", "树结构测试"),
                ("README.md", "langsmith-go 接入"),
            ],
        },
        "biz": [
            ("agent run", "根节点"),
            ("llm.think", "子 run"),
            ("tool.calculator", "子 run"),
            ("llm.answer", "子 run"),
            ("属性", "tokens/耗时"),
            ("导出", "run 树 JSON"),
        ],
    },
    "ch25": {
        "title": "ch25 LangServe 风格服务",
        "code": {
            "flow": [
                ("main.go", "REST 服务"),
                ("/invoke", "输入→输出"),
                ("/chat/stream", "SSE 流式"),
                ("/playground", "调试页面"),
                ("/docs", "接口自描述"),
            ],
            "extra": [
                ("web/playground.html", "浏览器调试台"),
                ("main_test.go", "接口测试"),
                ("README.md", "与 LangServe 对照"),
            ],
        },
        "biz": [
            ("客户端", "POST /invoke"),
            ("Go 服务", "调用 DeepSeek"),
            ("流式", "/chat/stream SSE"),
            ("Playground", "可视化调试"),
            ("Docs", "接口文档"),
        ],
    },
    "ch26": {
        "title": "ch26 综合案例:AI 数学方程绘图智能体",
        "code": {
            "flow": [
                ("backend/main.go", "API + 前端托管"),
                ("backend/api.go handleChat()", "对话主流程"),
                ("backend/llm.go + prompt.txt", "DeepSeek + 提示词"),
                ("backend/apply.go", "graphState 更新规则"),
                ("backend/store.go", "SQLite 持久化"),
            ],
            "extra": [
                ("frontend/src/App.tsx", "三栏工作台"),
                ("frontend/src/plot.ts", "mathjs 安全采样"),
                ("frontend/src/components/GraphViewer.tsx", "Plotly 绘图"),
                ("backend/api_test.go", "9 条验收用例"),
                ("docker-compose.yml", "一键启动"),
            ],
        },
        "biz": [
            ("自然语言", "画 y = x^2"),
            ("DeepSeek 解析", "结构化 JSON(intent/equations)"),
            ("校验与兜底", "FlexID/意图修正/重试"),
            ("graphState 规则", "apply.go 更新状态"),
            ("SQLite", "会话/消息持久化"),
            ("React 前端", "Plotly 绘制"),
            ("多轮修改", "改色/追加/坐标/解释"),
        ],
    },
}


def build_diagrams(ch, data):
    doc_dir = os.path.join(ROOT, ch, "docs")
    os.makedirs(doc_dir, exist_ok=True)
    title = data["title"]

    # ---- 代码架构图 ----
    code = data["code"]
    flow = [(label, sub or "") for label, sub in code["flow"]]
    extras = code.get("extra", [])
    step = 118
    boxes = layout_flow(flow, 60, 520, 112, step, 96, font(17), font(13))
    arrows = []
    for i in range(len(boxes) - 1):
        x1 = boxes[i][0] + boxes[i][2] / 2
        y1 = boxes[i][1] + boxes[i][3]
        x2 = boxes[i + 1][0] + boxes[i + 1][2] / 2
        y2 = boxes[i + 1][1]
        arrows.append((x1, y1, x2, y2))
    kind = "代码架构图 · 文件结构与调用流(基于最新代码生成)"
    render_svg(os.path.join(doc_dir, "architecture-code-map.svg"), title, kind, boxes, arrows, extras)
    render_png(os.path.join(doc_dir, "architecture-code-map.png"), title, kind, boxes, arrows, extras)

    # ---- 业务架构图 ----
    biz = [(label, sub or "") for label, sub in data["biz"]]
    step_b = 100
    boxes_b = layout_flow(biz, 210, 780, 112, step_b, 86, font(17), font(13))
    arrows_b = []
    for i in range(len(boxes_b) - 1):
        x1 = boxes_b[i][0] + boxes_b[i][2] / 2
        y1 = boxes_b[i][1] + boxes_b[i][3]
        x2 = boxes_b[i + 1][0] + boxes_b[i + 1][2] / 2
        y2 = boxes_b[i + 1][1]
        arrows_b.append((x1, y1, x2, y2))
    kind_b = "业务架构图 · 概念与数据流(基于最新代码生成)"
    render_svg(os.path.join(doc_dir, "architecture-diagram.svg"), title, kind_b, boxes_b, arrows_b)
    render_png(os.path.join(doc_dir, "architecture-diagram.png"), title, kind_b, boxes_b, arrows_b)
    print(f"{ch}: 已生成 4 个文件 -> {doc_dir}")


def main():
    for ch, data in CHAPTERS.items():
        build_diagrams(ch, data)
    print("全部完成:ch01~ch26 × (architecture-code-map + architecture-diagram) × (svg + png)")


if __name__ == "__main__":
    main()
