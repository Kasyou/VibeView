---
name: vibeview
description: Whiteboard for visual output. Start when user asks, push every response, stop when done.
allowed-tools: [Bash(vibeview *), Bash(taskkill /F /IM vibeview.exe)]
---

## Rules

1. User says preview → `vibeview &`
2. **Push before replying.** Whiteboard = full content, chat = one-line summary.
3. New topic → `preview_clear`. Done → `preview_stop`.

## Preview Show

```js
preview_show({ title: "Topic", content: "## Markdown\n- points\n| table |" })
```

Batch: `cards: [{title,content}, ...]` pushes multiple at once.

## Modes

`vibeview` (Claude whiteboard :51820) | `vibeview design` (UI preview :51821)
