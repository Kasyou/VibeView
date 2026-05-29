# 游戏开发 — Canvas / 回合制 / 卡牌游戏

从 GameDevelopmentProject/RougeLike-CardGame 项目提炼。多人肉鸽抽卡冒险桌游，纯前端实现。

---

## 一、游戏架构模式

### 状态机驱动

用 `mode` 字段驱动游戏流程，每个 mode 有独立的渲染和处理逻辑：

```
"map" → "combat" → "discard" → "reward" → "map" (循环)
  ↓                    ↓
"gameOver"          "victory"
```

```javascript
// 主循环：模式 → 渲染
function gameLoop() {
  switch (gameState.mode) {
    case "map":     drawMap(gameState); break;
    case "combat":  drawCombat(gameState); break;
    case "reward":  drawReward(gameState); break;
    case "gameOver": drawGameOver(gameState); break;
    case "victory": drawVictory(gameState); break;
  }
  requestAnimationFrame(gameLoop);
}
```

### 模块划分

```
main.js                    ← 入口：init + gameLoop + 事件绑定
src/
  core/                    ← 核心系统
    gameState.js           ← 全局状态初始化 + 管理
    turnManager.js         ← 回合流转逻辑
  data/                    ← 静态数据
    characters.js          ← 角色定义
    enemies.js             ← 敌人数据
    cards.js               ← 卡牌定义（baseCards + poolCards）
    mapConfig.js           ← 地图配置
  mechanics/               ← 游戏机制
    combat.js              ← 战斗系统
    map.js                 ← 地图生成 + 节点导航
    reward.js              ← 奖励/抽卡（三选一）
  ui/                      ← 渲染 & 输入
    draw.js                ← Canvas 渲染
    input.js               ← 鼠标/键盘处理
  utils/                   ← 通用工具
```

---

## 二、Canvas 游戏开发要点

### 为什么用 Canvas 而不是 DOM

- 游戏有大量自定义绘制（地图、血条、动画）
- 性能要求高（60fps 渲染循环）
- 不需要 DOM 的布局能力

### 渲染循环

```javascript
// requestAnimationFrame 驱动，每帧重绘
function gameLoop() {
  drawUI(gameState);  // 根据当前 mode 绘制对应画面
  requestAnimationFrame(gameLoop);
}
```

### 输入处理

- Canvas 点击：`canvas.addEventListener("click", (e) => handleCanvasClick(e, gameState, canvas))`
- 坐标转换：鼠标坐标 → Canvas 坐标系
- 键盘快捷键：`Space` 结束回合等便捷操作

---

## 三、回合制战斗系统

### 卡牌抽牌堆循环

```
deck → drawPile → hand → discardPile → drawPile (循环)
```

```javascript
// 抽牌逻辑
function drawCards(character, count) {
  while (drawn < count) {
    if (character.drawPile.length === 0) {
      // 弃牌堆洗入抽牌堆
      character.drawPile = [...character.discardPile];
      shuffle(character.drawPile);
      character.discardPile = [];
    }
    character.hand.push(character.drawPile.pop());
  }
}
```

### 回合流转

```
玩家 A 回合 → 玩家 B 回合 → 玩家 C 回合 → 敌人回合 → 玩家 A 回合...
```

- 每个角色每回合 +2 能量（上限 4-5）
- 手牌上限 5 张，超出的需要丢弃
- 弃牌阶段：选择要丢弃的牌

### 伤害计算

```javascript
function applyDamage(target, amount, isBlockable) {
  let dmg = amount;
  if (isBlockable && target.block > 0) {
    const absorbed = Math.min(target.block, dmg);
    target.block -= absorbed;
    dmg -= absorbed;
  }
  target.hp = Math.max(0, target.hp - dmg);
}
```

格挡值每回合清零，优先吸收伤害。

### 卡牌效果系统

用 effects 数组定义特殊效果：

```javascript
effects: ["taunt"]          // 嘲讽 — 敌人优先攻击
effects: ["chainLightning"] // 闪电链 — 额外伤害随机敌人
effects: ["freeze"]         // 冻结 — 跳过敌人下次攻击
effects: ["exhaust"]        // 消耗 — 使用后移出牌组
effects: ["dot"]            // 持续伤害 — 每回合扣血
effects: ["strengthUp"]     // 力量提升 — 全体 Buff
```

---

## 四、地图与进度系统

### 层级地图

```
Layer 1: [战斗]           ← 起点
Layer 2: [战斗] [战斗]    ← 同层多节点
Layer 3: [Boss]           ← 关底
```

- 同层节点需逐个清完
- 清除一层后解锁下一层
- Boss 在最后一层

### Boss 机制

- Boss 有 AOE 技能，每 2 回合对所有玩家造成固定伤害
- Boss 血量高，攻击力高
- 需要团队成员配合（战士嘲讽、牧师治疗、法师输出）

---

## 五、奖励系统

### 战斗胜利后抽卡

```
每个存活的角色依次抽卡 → 三选一（2张牌 + 1个治疗选项）
```

- 从角色对应职业的卡池中随机抽 3 张
- 玩家选择 1 张加入牌组
- 或者选择"治疗"（全体回复 6 点生命）
- 所有存活角色抽完后回到地图

### 牌组构建策略

- baseCards：初始卡组（每个角色 5 张基础牌）
- poolCards：奖励卡池（每个职业约 6 张可抽取的牌）
- 随着游戏推进，牌组越来越强

---

## 六、多人热座模式设计

### 为什么先做热座而非联网

1. **开发成本低** — 不需要后端和网络层
2. **立即可玩** — 共用同一浏览器窗口
3. **后期可扩展** — 架构预留了 Socket.IO 接入点

### 热座交互

- 每个玩家在自己的回合操作
- 共享屏幕，按顺序轮流
- Canvas 显示"当前是 X 的回合"

---

## 七、Canvas 游戏快速启动模板

```html
<!DOCTYPE html>
<html>
<head><title>Game</title><link rel="stylesheet" href="style.css"></head>
<body>
  <canvas id="gameCanvas" width="1000" height="700"></canvas>
  <script type="module" src="main.js"></script>
</body>
</html>
```

```javascript
// main.js
import { initGame, gameState } from "./src/core/gameState.js";
import { drawUI } from "./src/ui/draw.js";
import { handleCanvasClick } from "./src/ui/input.js";

const canvas = document.getElementById("gameCanvas");
initGame(initialData, config);

function gameLoop() {
  drawUI(gameState);  // 分发到各模式的绘制函数
  requestAnimationFrame(gameLoop);
}

canvas.addEventListener("click", (e) => handleCanvasClick(e, gameState, canvas));
gameLoop();
```

---

## 八、游戏开发的 Claude 协作要点

1. **先设计数据结构再写逻辑** — 角色、卡牌、敌人、地图的数据结构决定了后面所有的代码
2. **状态机模式非常适合 Claude** — Claude 能很好地理解"在不同状态下做什么"
3. **数据分离** — 数据文件（cards.js, enemies.js）和逻辑文件分开，方便 Claude 单独修改
4. **渐进开发** — 先做单机热座原型，再考虑联网。MVP 能玩是最高优先级
5. **Canvas 渲染代码天然较长** — 接受 UI 代码会比较长的事实，重点是逻辑清晰
