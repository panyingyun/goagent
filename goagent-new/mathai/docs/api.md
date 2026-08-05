# HTTP API

默认地址为 `http://localhost:8080`，请求和响应均使用 UTF-8 JSON。若设置 `MATHAI_API_TOKEN`，所有 `/api` 请求都需要：

```http
Authorization: Bearer local-dev-token
```

## 健康检查

- `GET /livez`：进程存活即返回 200；
- `GET /readyz`：`LLM_API_KEY` 与 `LLM_MODEL` 已配置才返回 200。

## 会话

### `GET /api/sessions`

返回按更新时间倒序排列的摘要。

### `POST /api/sessions`

```json
{"title":"三角函数比较"}
```

`title` 可为空，未知字段会被拒绝。成功返回 201 和完整 Session。

### `GET /api/sessions/{id}`

返回消息与 GraphState。不存在返回 404。

### `DELETE /api/sessions/{id}`

删除会话并返回 204。该操作当前不可恢复。

## 聊天

### `POST /api/sessions/{id}/chat`

```json
{"message":"画出 y=x 和 y=x^2，并比较它们"}
```

成功响应：

```json
{
  "answer": "已经画出直线和抛物线……",
  "traceId": "trace_...",
  "session": {
    "id": "session_...",
    "version": 2,
    "messages": [],
    "graphState": {
      "equations": [],
      "viewport": {"xMin": -10, "xMax": 10, "yMin": -10, "yMax": 10},
      "settings": {"showGrid": true, "showAxes": true, "showLegend": true, "sampleCount": 700}
    }
  }
}
```

错误语义：

| 状态码 | 含义 |
|---|---|
| 400 | 非法 JSON、字段或消息长度 |
| 401 | Token 缺失或错误 |
| 404 | 会话或 trace 不存在 |
| 409 | 同一会话在模型执行期间已被其他请求更新 |
| 429 | 客户端超过进程内限流 |
| 502 | LLM 或 Agent 执行失败 |

## Trace

### `GET /api/traces/{traceId}`

返回该请求的 spans 与 Agent events。Trace 仅保留最近 200 条，进程重启后清空，因此 404 也可能表示已过期。
