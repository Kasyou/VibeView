# 开源工具开发 — npm 包 / CLI 工具 / 框架

从 ViteToolProject/MCPX 项目提炼。做一个 TypeScript 开源项目，目标是 1000+ GitHub Stars。

---

## 核心思路

**做一个框架/工具型开源项目 = 极致的 DX + 明确的产品定位 + 多平台营销。**

"10 倍更简单"是最有效的差异点。

---

## 一、产品设计原则

### DX（开发者体验）优先

- **Zero config**: 开箱即用，不需要看文档就能跑起来
- **Zero boilerplate**: 用户写最少代码达到目的
- **一行安装**: `npm install xxx` 然后就能用
- **API 直觉化**: 方法名和参数符合开发者的心理预期

### 定位策略

用开发者熟悉的类比来定位："Next.js for MCP"、"Express.js for WebSocket"、"Lodash for arrays"。

好定位 = 一个众所周知的类比 + 你的差异化价值。

### 什么不要做（MVP 阶段）

- 不要做插件系统（v1 不需要）
- 不要做 Web UI / Dashboard
- 不要做太多 transports（先做一个，稳了再加）
- 不要做注册中心 / 市场 / 平台

---

## 二、项目结构

### 推荐结构（npm 包 + CLI）

```
项目名/
├── CLAUDE.md
├── src/
│   ├── index.ts          ← 公共导出入口
│   ├── core/
│   │   ├── index.ts      ← 核心模块导出
│   │   ├── types.ts      ← 类型定义
│   │   └── server.ts     ← 核心逻辑
│   └── cli/
│       ├── index.ts      ← CLI 入口 (#!/usr/bin/env node)
│       ├── init.ts       ← init 命令
│       ├── dev.ts        ← dev 命令
│       └── build.ts      ← build 命令
├── templates/            ← 项目模板（供 init 命令使用）
│   └── basic/
│       ├── tsconfig.json
│       └── src/index.ts
├── docs/superpowers/specs/  ← 设计文档
├── tsconfig.json
├── package.json
└── README.md             ← 核心营销资产
```

### CLI + Core 分离

Core 是纯库，可以被其他项目 `import` 使用。CLI 是命令行工具，调用 Core。两者在同一个包里但职责分离。

---

## 三、API 设计模式

### "流畅的链式 API"模式

```typescript
const server = mcpx({ name: 'my-mcp', version: '1.0.0' });

server
  .tool({ name: 'greet', handler: () => 'hello' })
  .tool({ name: 'calc', handler: () => 42 })
  .resource({ name: 'readme', handler: () => '...' })
  .start();
```

返回自身 (`return this`) 实现链式调用。

### 返回值的灵活性

Handler 可以返回多种类型，框架自动处理：

- `string` → text content
- `object` → JSON stringified
- `CallToolResult` → 直接透传

减少用户的认知负担。

### 类型安全

用 TypeScript 泛型从参数定义自动推导 handler 参数类型：

```typescript
// 定义参数时就确定了 handler 的入参类型
server.tool({
  name: 'greet',
  parameters: { name: { type: 'string' } },
  handler: ({ name }) => `Hello, ${name}!` // name 自动推断为 string
});
```

---

## 四、增长策略

### Launch Checklist

- [ ] 15 秒终端 GIF（展示从零到运行的全过程）
- [ ] 代码对比表（用你的工具 vs 不用）
- [ ] 一行安装命令
- [ ] 3-4 个示例项目（覆盖典型使用场景）
- [ ] README 清晰、短小、有视觉冲击力

### 分发渠道

1. **Show HN** — 最关键的首次曝光
2. **Reddit** — r/programming, r/LocalLLaMA 等
3. **Twitter/X** — AI dev 社区
4. **中文平台** — 掘金、知乎（国内开发者社区）
5. **周报/Newsletter** — 联系相关领域的周刊作者

### 增长节奏预估

| 时间 | Stars | 增长动力 |
|------|-------|----------|
| 第 1 周 | 100 | 发布日流量 |
| 第 2 周 | 300 | 社区传播 |
| 第 2 月 | 1000+ | 口碑 + 搜索引擎 |

### 关键：多平台同时发布

第一天就要在所有平台（GitHub, HN, Reddit, 掘金, 知乎）同时发力，制造"到处都是"的感觉。

---

## 五、CLI 设计

### 命令最少化

- `init` — 创建新项目
- `dev` — 开发模式（hot reload）
- `build` — 生产构建

三个命令覆盖开发全流程。不要加太多命令。

### init 模板系统

- 用文件系统模板（不是字符串拼接）
- `cp` 整个模板目录 → 替换模板变量
- 模板里的 `TEMPLATE_NAME` 替换为实际项目名

---

## 六、技术选型备忘

| 用途 | 推荐 | 原因 |
|------|------|------|
| MCP 协议 | `@modelcontextprotocol/sdk` | 官方 SDK |
| CLI | `commander` | 最流行，社区大 |
| Hot Reload | `chokidar` | 文件监听 |
| 参数验证 | `zod` | 类型安全 + 运行时验证 |
| TypeScript | strict mode | 类型安全是 DX 的一部分 |
| 模块 | Node16 + ESM | 2026 年的标准 |
