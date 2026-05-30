---
name: vibeview
description: Automated visual whiteboard for Claude. Start once, push every response, stop when done. Resource-conscious lifecycle. Design mode for project UI preview.
allowed-tools: [Bash(vibeview *), Bash(taskkill /F /IM vibeview.exe)]
---

## Lifecycle (Fully Automated)

```
START → USE → STOP
```

### START — once per session

When user says "preview", "visualize", "show me", or "whiteboard":

```bash
vibeview &
```

Starts Claude mode on port 51820. Tell user: `Open http://localhost:51820`

If port 51820 is in use, try `vibeview --port 51822 &`.

**Design mode** for project UI preview:
```bash
vibeview design &
```

### USE — every response

**CRITICAL: Push EVERY significant conclusion to the whiteboard with `preview_show`.**

The whiteboard IS the user's view. This chat is secondary. Push:

- Architecture decisions → table comparison
- Analysis conclusions → headings + bullet points
- Code explanations → code block + key points
- Summaries → bold terms + concise bullets
- Process flows → numbered list

```json
preview_show({
  "title": "决策: 使用 Redis 做缓存",
  "content": "## 理由\n\n| 方案 | 延迟 | 成本 |\n|------|------|------|\n| Redis | <1ms | 低 |\n| Memcached | <1ms | 低 |\n\n### 结论\n- **Redis** 因为有持久化和丰富数据结构\n- 预估 QPS 10000+ 无压力"
})
```

Start new topic → `preview_clear` first.

### STOP — when session ends

When user says "done", "that's it", "stop preview", or conversation naturally concludes:

```bash
preview_stop
```

This frees the port and stops the file watcher.

## MCP Tools

| Tool | When |
|------|------|
| `preview_show` | Push every response |
| `preview_clear` | New topic |
| `preview_screenshot` | Verify appearance |
| `preview_stop` | Clean shutdown |

## Rules

1. Start once per session, stop when done. Don't leave servers running.
2. Push before explaining. The whiteboard gets the content first.
3. If whiteboard tools timeout, the user hasn't opened the browser. Remind them.
4. No auto-start on session open. Wait for user request.
