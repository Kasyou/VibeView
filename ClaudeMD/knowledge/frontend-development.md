# 前端项目开发 — React 全栈应用 & 求职作品集

从 FrontedProjcet/CoBoard 项目提炼。一个 React + TypeScript + Vite 的跨部门协作看板，作为求职作品集。

---

## 一、技术选型决策框架

做技术选型时，不只是看"哪个更好"，而是回答三个问题：
1. 这个项目最需要什么能力？
2. 候选方案在这个能力上谁更强？
3. 如果以后要迁移，成本多大？

### Zustand vs Redux Toolkit

| 维度 | Zustand | Redux Toolkit |
|------|---------|---------------|
| API 复杂度 | 极简，`create(set => ...)` | 需理解 slice/reducer/action |
| 持久化 | `persist` 中间件一行代码 | 需 `redux-persist` 额外库 |
| 包体积 | < 1KB | ~11KB |
| DevTools | 基础 | 时间旅行调试 |
| 适用场景 | 中小型应用、快速开发 | 大型团队、复杂状态 |

**选择建议**：数据规模小、追求开发速度选 Zustand。团队大、状态逻辑复杂选 Redux Toolkit。核心概念相通，迁移成本可控。

### @dnd-kit vs react-beautiful-dnd

**默认选 @dnd-kit**。react-beautiful-dnd 已停止维护（Atlassian 2023年停止）。

@dnd-kit 优势：
- `closestCorners` 碰撞检测适合 Kanban 多列场景
- `PointerSensor` 原生支持移动端触控
- 支持键盘操作和无障碍
- 动画流畅度更好

### Tailwind CSS vs CSS Modules vs styled-components

| 场景 | 推荐 |
|------|------|
| 设计 token 多（品牌色、主题色） | Tailwind — 直接映射到 className |
| 组件样式隔离、复杂选择器 | CSS Modules |
| 动态样式、运行时主题 | styled-components |

**选择建议**：新项目优先 Tailwind。设计 token 在 `tailwind.config.ts` 定义后直接使用，避免样式文件跳转。

### Vite vs Create React App

**默认选 Vite**。CRA 已停更。Vite 优势：
- HMR 秒级响应
- 零配置 TS/JSX
- 构建速度显著快于 Webpack

---

## 二、权限系统设计模式

### "集中式规则引擎 + 分布式调用"

所有权限判断集中在 `utils/permissions.ts` 的纯函数中：

```typescript
// 每个权限函数 = 纯函数，输入用户+资源，输出 boolean
function canEditTask(user: UserSafe, task: Task): boolean {
  if (user.role === 'pm') return true;
  if (user.role === 'leader' && user.department === task.department) return true;
  return false;
}
```

组件只需调用，不用理解权限细节：
```tsx
{canEditTask(currentUser, task) && <EditButton />}
```

### 权限矩阵设计要点

1. **先列出所有操作**（创建、编辑、删除、变更状态等）
2. **再列出所有角色**（pm、leader、member）
3. **交叉填充规则**
4. **处理边界**：跨部门卡片、自己创建的内容等特殊规则

### 用户类型安全

```
User (完整信息，含 password) → 仅用于认证查找
UserSafe (Omit<User, 'password'>) → 用于所有组件和持久化
```

用类型系统防止密码泄漏到前端——TypeScript 编译器会在编译时就阻止错误。

---

## 三、状态管理架构

### 推荐模式：单一 Store + Derived Queries

```
Zustand Store ──persist──▶ localStorage
     │
     ├── 原始数据 (tasks[])
     ├── CRUD actions (add/update/delete/move)
     └── derived queries (getTasksByDepartment, getBlockedTasks...)
```

**关键原则**：
- 数据归一化：所有同类数据存在一个数组，不按视图拆分
- Derived queries 在 store 内实现，组件直接调用
- 乐观更新：直接改 store，持久化自动同步

### 为后端接入预留接口

即使是纯前端项目，设计 store 的 action 也要像调用 API：

```typescript
// 当前实现：直接改 localStorage
async function login(username: string, password: string): Promise<{success: boolean; error?: string}>
// 未来替换：指向后端 API，接口签名不变
```

---

## 四、拖拽实现要点

### Kanban 看板拖拽

1. `DndContext` 包裹整个看板，使用 `closestCorners` 碰撞检测
2. 每列是 `SortableContext`，子项用 `useSortable`
3. `onDragEnd` 获取目标列 status → 调用 store.moveTask(taskId, newStatus)
4. 权限控制：拖拽前判断 `canEditStatus`

### 阻塞标记模式

- `isBlocked` + `blockReason` 字段
- 被阻塞卡片：红色左边框 + Badge
- 全局仪表盘组件通过 `getBlockedTasks()` 查询所有阻塞卡片

---

## 五、首页动效模式

### Framer Motion 常用组合

| 效果 | API 组合 |
|------|----------|
| 鼠标跟随光斑 | `onMouseMove` 坐标 + `useSpring` 缓动 + `radial-gradient` |
| 滚动视差 | `useScroll` + `useTransform` 不同速率 |
| 交错入场 | `whileInView` + `delay: i * 0.18s` + spring physics |
| 数字递增 | `@react-spring/web` 的 `useSpring` |

### Spring Physics 参数

```typescript
{ stiffness: 55, damping: 16 }  // 自然弹性
{ stiffness: 100, damping: 20 } // 快速弹性
{ stiffness: 200, damping: 30 } // 干脆
```

---

## 六、求职作品集策略

### 作品集 vs 生产项目

| 维度 | 作品集 | 生产项目 |
|------|--------|----------|
| 优先级 | 视觉 > 测试 > 边界 | 测试 > 安全 > 视觉 |
| 动效投入 | 值得多花时间 | 适可而止 |
| 完美程度 | 核心流程完整即可 | 所有边界覆盖 |
| 文档 | 技术决策复盘 > API 文档 | API 文档 > 决策记录 |

### 作品集加分项

1. **技术决策有理由** — 不是"我用了 X"，而是"因为 Y 原因选了 X，对比了 Z"
2. **权限系统** — 展示架构设计能力
3. **动效实现** — 展示前端基本功
4. **类型安全** — 展示工程素养
5. **已知局限和改进方向** — 展示工程师思维（知道自己的代码哪里不够好）

---

## 七、渐进增强策略

即使是纯前端 demo，也要预留后端的扩展点：

- 数据层函数用 `async` + `Promise` 返回，即使现在是同步实现
- 类型定义独立于实现（`User` vs `UserSafe`）
- Store 的 action 接口保持稳定，实现可以替换

"为未来设计接口，为现在写实现" — 减少未来重构成本，同时不增加当前复杂度。
