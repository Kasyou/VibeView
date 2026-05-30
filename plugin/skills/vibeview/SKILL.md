---
name: vibeview
description: When VibeView Claude whiteboard is active at localhost:51820, you MUST push every significant response to the whiteboard using preview_show. The whiteboard IS what the user sees — this chat is secondary. Use Design mode for project UI preview.
---

## CRITICAL: The Whiteboard Is Your Output

When VibeView Claude mode is running on port 51820, the browser whiteboard is **the user's primary view**. They are NOT reading this chat carefully — they are watching the whiteboard.

**Your job**: Push every key conclusion, analysis, or result to the whiteboard with `preview_show`. This is not optional.

```
User sees:  Whiteboard (browser) ←←← YOUR REAL OUTPUT
            This chat           ←←← secondary, reference only
```

## Push Rules

1. **Every major conclusion** → one card with `preview_show`
2. **Comparisons** → table in markdown
3. **Steps / processes** → numbered list
4. **Architecture** → heading hierarchy + bullet points
5. **Code explanations** → code block + bullet list of key points
6. **Summaries** → bold key terms + concise bullet

## Example

```
User asks: "Should we use Redis or Kafka?"

Your response here: "Kafka because of throughput and persistence needs."

preview_show({
  title: "消息队列选型结论",
  content: "## 推荐: Kafka\n\n| 维度 | Redis | Kafka |\n|------|-------|-------|\n| 吞吐量 | 中 | **高** |\n| 持久化 | 弱 | **强** |\n\n### 决策因素\n- 需要高吞吐量的事件流\n- 消息不能丢失\n- 已有 Kafka 运维经验"
})
```

## MCP Tools

| Tool | When |
|------|------|
| `preview_show` | Push EVERY response to whiteboard |
| `preview_clear` | Starting a new topic |
| `preview_screenshot` | Verify whiteboard appearance |

## Design Mode

For project UI preview:
```bash
vibeview --mode design --port 51821
```
