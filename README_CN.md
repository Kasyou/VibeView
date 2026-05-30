# VibeView

**Claude Code 的可视化输出白板。** Claude 的分析、决策、总结以卡片形式呈现在浏览器中——不再埋没在聊天框里。

<p align="center">
  <img src="screenshots/whiteboard-demo-v0.3.1.png" alt="白板演示" width="720">
</p>

## 双模式

| 模式 | 命令 | 端口 | 用途 |
|------|------|------|------|
| Claude 白板 | `vibeview` | 51820 | AI 推理可视化 |
| Design 预览 | `vibeview design` | 51821 | 项目 UI 实时预览 |

### Claude 白板

Claude 将分析结论、架构决策、代码审查以卡片形式推送到浏览器。用户看白板而非刷聊天框。

### Design 预览

<p align="center">
  <img src="screenshots/design-demo-v0.3.1.png" alt="Design 预览" width="480">
</p>

Android Studio 式即时 UI 预览。文件监控 + WebSocket 热更新。自动识别 React/Vue/Svelte/HTML 项目。支持 iPhone/Pixel/iPad 设备框。

## 快速开始

```bash
go install github.com/Kasyou/VibeView@latest

# Claude 白板
vibeview
# → 浏览器打开 http://localhost:51820

# Design 预览
vibeview design --dir ./my-project
# → 浏览器打开 http://localhost:51821
```

## 9 个 MCP 工具

| 工具 | 用途 |
|------|------|
| `preview_show` | 推送 Markdown 卡片 |
| `preview_clear` | 清空白板 |
| `preview_history` | 分页历史查询 |
| `preview_screenshot` | 截取预览图 |
| `preview_inspect` | CSS 选择器查元素 |
| `preview_console` | 读浏览器错误 |
| `preview_diff` | 前后截图对比 |
| `preview_reload` | 强制刷新 |
| `preview_stop` | 关闭服务器 |

## 特性

- `#序号 · 时间戳` 卡片注释
- 离线消息队列（浏览器断开不丢卡）
- 30 张 DOM 限制 + 服务器无限历史
- 中文 UTF-8 完美渲染
- 37 个测试 | Go 1.23+

## 编译

```bash
git clone https://github.com/Kasyou/VibeView.git
cd VibeView
go build -o vibeview .
```

## License

MIT
