# VibeView

[English](README.md)

**让 Claude Code 的输出可视化——实时图表、结构化卡片、代码高亮。不止是聊天。**

Claude Code 可以写代码、做分析、给方案，但这些产出都被困在聊天窗口里。VibeView 给 Claude 一个可以"展示"的地方：架构图渲染成 SVG、方案对比用表格呈现、分析结论打印为带时间戳的卡片。这些都是聊天框做不到的事情。

<p align="center">
  <img src="screenshots/vv-capability-demo.png" alt="VibeView 白板：思维导图与对比表格" width="720">
</p>

## 解决什么问题

当你和 Claude Code 协作时，Claude 的分析、架构决策、代码审查结果全部以文字形式输出。文字多了就找不到重点。

VibeView 把这些文字**转换成可视化内容**：

- **思维导图和流程图** — Claude 写 Mermaid 语法，VibeView 渲染为 SVG 图表。聊天框做不到这一点
- **结构化对比表格** — 方案对比不再是纯文本，而是带深色主题的格式化表格，一眼看懂
- **语法高亮代码块** — GitHub 暗色主题，多语言支持，可滚动
- **带时间戳的分析卡片** — 每张卡片有编号和时间，形成可回溯的历史记录

## 为什么不能只看聊天框

聊天框是线性的、临时的。往上翻几页就找不到了。

VibeView 的不同之处：

1. **持久化** — 每条分析以卡片形式保留，带编号和时间戳。用 `preview_history` 查询任意历史卡片
2. **图表能力** — Mermaid 思维导图和流程图在聊天框中根本无法渲染，这是 VibeView 独有的
3. **可扫读** — 结构化表格和标题层级让你一眼定位信息，不需要重读大段文字
4. **关注点分离** — 代码在编辑器，预览在右边，Claude 的分析在白板。各司其职

## 两种模式

### 白板模式 (`vibeview`)

Claude 把分析、决策、总结推送到浏览器白板。端口 51820。

```
用户: "消息队列用 Redis 还是 Kafka？"
Claude: [分析需求、对比方案]
        → preview_show 推送对比表格 + 架构图
        → 白板实时显示：格式化表格 + Mermaid 流程图
```

### 设计预览模式 (`vibeview design`)

代码编辑器的实时预览伴侣。文件变更检测 + 即时刷新。iPhone/Pixel/iPad 设备框。自动识别 React、Vue、Svelte、HTML 项目。端口 51821。

<p align="center">
  <img src="screenshots/design-demo-v0.3.1.png" alt="代码编辑器与实时预览并排" width="720">
</p>

## 快速开始

```bash
# 安装
go install github.com/Kasyou/VibeView@latest

# 启动白板
vibeview
# 浏览器打开 http://localhost:51820

# 启动设计预览
vibeview design
# 浏览器打开 http://localhost:51821
```

## Claude Code 集成

VibeView 注册了 9 个 MCP 工具，Claude 可以在对话中直接调用：

| 工具 | Claude 能做什么 |
|------|----------------|
| `preview_show` | 推送分析到白板，渲染为格式化卡片 |
| `preview_clear` | 清空白板，开始新话题 |
| `preview_history` | 分页查询历史卡片 |
| `preview_screenshot` | 截取当前白板画面 |
| `preview_inspect` | 用 CSS 选择器查询 DOM 元素 |
| `preview_console` | 读取浏览器控制台错误 |
| `preview_diff` | 对比两张截图的变化 |
| `preview_reload` | 强制刷新预览 |
| `preview_stop` | 关闭服务器释放资源 |

### 智能浏览器

VibeView 会检测推送内容是否包含图表。如果有图表，返回 `hasDiagrams: true`，Claude 会提示你打开外置浏览器（Cursor 内置浏览器不完全支持 SVG 渲染）。纯文字和表格内容可以直接在 Cursor 内置浏览器中查看。

## 从源码编译

```bash
git clone https://github.com/Kasyou/VibeView.git
cd VibeView
go build -o vibeview .

# 二进制包含内嵌的 Mermaid.js（约 12MB）
# 37 个测试通过，Go 1.23+
```

## License

MIT
