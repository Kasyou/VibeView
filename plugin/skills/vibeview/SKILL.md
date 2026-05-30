---
name: vibeview
description: Whiteboard for visual output. Start when user asks, push every response, stop when done.
allowed-tools: [Bash(vibeview *), Bash(taskkill /F /IM vibeview.exe)]
---

## Rules

1. User says preview → `vibeview &`. Done → `preview_stop`.
2. **Push before replying.** Whiteboard has the full answer; chat is one sentence.
3. Start new topic → `preview_clear` first.

## Diagrams

Use Mermaid in code blocks. VV renders them as SVG diagrams — impossible in plain chat.

````
```mermaid
graph TD
  A[Problem] → B[Options]
  B → C[Redis]
  B → D[Kafka]
  C → E[Low latency]
  D → F[High throughput]
```
````

Also: `mindmap`, `flowchart`, `sequenceDiagram`, `gantt`, `pie`, `classDiagram`.

## Example

```
User: "Should we use Redis?"
preview_show({title:"Cache Decision", content:"## Redis\n- 5M QPS\n\n**Use Redis**"})
Chat: "→ Whiteboard"
```

## Modes

`vibeview` (whiteboard :51820) | `vibeview design` (preview :51821)
