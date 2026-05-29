# Electron 桌面应用 & 练习项目模式

从 PracticeProject 和多个项目的开发模式中提炼。

---

## 一、Electron + React + Vite 快速启动

### 项目结构

```
project/
├── electron/
│   ├── main.cjs        ← 主进程入口
│   └── preload.cjs     ← 预加载脚本
├── src/
│   ├── App.tsx         ← React 入口
│   ├── components/     ← 组件目录
│   ├── store/          ← 状态管理
│   ├── hooks/          ← 自定义 hooks
│   ├── services/       ← 业务逻辑
│   ├── utils/          ← 工具函数
│   ├── types/          ← TypeScript 类型
│   └── styles/         ← 全局样式
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

### Electron 主进程关键配置

```javascript
const win = new BrowserWindow({
  width: 1200,
  height: 760,
  minWidth: 900,
  minHeight: 560,
  webPreferences: {
    preload: path.join(__dirname, "preload.cjs"),
    contextIsolation: true,   // 安全：隔离渲染进程
    nodeIntegration: false,   // 安全：禁止渲染进程使用 Node
  },
  backgroundColor: "#0f0f12", // 避免白屏闪烁
  show: false,                // 等 ready-to-show 再显示
});

win.once("ready-to-show", () => win.show());

// 开发模式加载 Vite dev server
if (isDev) {
  win.loadURL("http://127.0.0.1:5173");
} else {
  win.loadFile(path.join(__dirname, "../dist/index.html"));
}
```

### 安全原则

- `contextIsolation: true` — 渲染进程无法直接访问 Node API
- `nodeIntegration: false` — 需要 Node 能力通过 preload 脚本暴露
- preload 脚本用 `contextBridge.exposeInMainWorld` 暴露安全的 API

---

## 二、目录骨架预设模式

项目初期用 `.gitkeep` 占位目录结构：

```
src/
  components/
    common/.gitkeep
    layout/.gitkeep
    player/.gitkeep
    playlist/.gitkeep
    library/.gitkeep
  store/.gitkeep
  hooks/.gitkeep
  services/.gitkeep
  utils/.gitkeep
  assets/.gitkeep
  styles/
    globals.css
dist/.gitkeep
public/.gitkeep
```

**优点**：
- 目录结构一目了然
- git 可以追踪空目录
- Claude 看到骨架就知道代码该放哪
- 新增文件时不用 `mkdir`

---

## 三、练习项目的正确姿势

### 练习项目的目的

不同于生产项目和作品集，练习项目的重点是**快速验证一个技术点**：

| 维度 | 练习项目 | 作品集项目 | 生产项目 |
|------|----------|-----------|----------|
| 目标 | 学一个东西 | 展示能力 | 解决问题 |
| 规模 | 单文件~几十文件 | 中等 | 不限 |
| 完美度 | 能跑就行 | 核心流程完整 | 边界全覆盖 |

### Practice-001：单文件 HTML 快速原型

适合验证：纯前端逻辑、UI 交互、CSS 动效。

```html
<!DOCTYPE html>
<html>
<head><title>Demo</title></head>
<body>
  <div id="app"></div>
  <script>
    // 所有逻辑在一个文件中
    // 适合快速验证想法
  </script>
</body>
</html>
```

### Practice-002：完整工程脚手架

适合验证：框架特性、工程化配置、Electron 集成。

- Vite + React + TypeScript 脚手架
- 完整的目录结构
- Electron 集成

---

## 四、通用项目初始化清单

无论什么类型的项目，初始化时应该：

1. **创建 CLAUDE.md** — AI 上下文恢复
2. **创建目录骨架** — .gitkeep 占位
3. **初始化 git** — 从第一个文件就开始版本控制
4. **安装依赖并验证能跑** — 不要等到写了很多代码才发现环境问题
5. **先 commit 脚手架** — `init: project scaffold`

### 脚手架 commit 内容

```
project/
├── CLAUDE.md            ← 项目概述
├── .gitignore
├── package.json         ← 依赖声明
├── tsconfig.json        ← TS 配置
├── vite.config.ts       ← 构建配置
├── index.html           ← 入口 HTML
└── src/
    └── App.tsx          ← 最小可运行组件
```

确保 `npm install && npm run dev` 能跑起来再开始写业务代码。
