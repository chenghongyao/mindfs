# MindFS (Mind File System) 产品规划与技术架构 (v7.0)

**核心理念**：Server 作为资源与 AI 能力的载体，Client 作为远程访问接口。通过“原子组件 + AI 编排”实现 UI 自我进化，通过“按需知识库”实现深度记忆。

---

## 一、 产品形态 & 界面框架 (Product Form)

### 1. 固定的 App Shell (The Frame)
无论中间的内容如何变化，应用的外框保持一致，保证用户习惯的延续性：
*   **左侧栏 (Sidebar)**: 可折叠的文件树列表 (File Tree)。用于基础导航。
*   **底部栏 (Action Bar)**:
    *   **AI 交互区**: 聊天输入框 (Chat Input)。
    *   **快捷操作区**: 显示当前上下文的常用 Action (如“上传”、“下载”、“上一章”)。
    *   **状态区**: 系统状态、传输进度。
*   **主视图 (Main Canvas)**: **核心动态区域**。完全由 AI 编排的原子组件构成。

### 2. 生成式 UI (Generative UI)
*   **模式**: JSON Schema 驱动 + json-render 渲染器。
*   **流程**: Agent 生成 `view.json`（UI Schema），前端固定渲染器读取并渲染。

---

## 二、 核心交互流程 (The Workflow)

1.  **Initialize (mindfs .)**:
    *   用户在终端运行 `mindfs .` 或在前端添加目录。
    *   Server 启动，挂载工作区，生成 `.mindfs` 隐藏目录。
    *   **Level 0**: Server 立即返回默认文件列表视图（保证秒开）。
2.  **Analyze & Generate**:
    *   Server 在目录加入管理后异步调用 **元技能 `ViewCreateSkill`**。
    *   外部 Agent 分析目录特征与用户 session，生成 **`view.json` + 运行时技能包**（skill bundle）。
    *   Server 将生成物写入 `.mindfs/` 并通知前端热更新。
3.  **Client Interaction**:
    *   用户触发操作（点击文件、输入文字）。
    *   Client 打包请求：`{ path: '/book/ch1.txt', action: 'open', context: {...} }` 发送给 Server。
4.  **Server Routing (Skill Interception)**:
    *   Server 检查 `.mindfs/skills/` 下注册的处理器。
    *   **Hit**: 如果有 Skill 拦截该 Action（例如 NovelSkill 拦截了 'open'），则执行 Skill 逻辑（如：记录阅读位置，流式读取文本）。
    *   **Miss**: 如果无 Skill 处理，执行系统默认逻辑。
    *   Server 返回结果（数据或新的 UI 状态）。

---

## 三、 关键机制 (Key Mechanisms)

### 1. Skill 系统：用户操作处理器
*   **.mindfs/skills/** 下存放 Agent 生成的技能包。
*   **核心职能**: 向 Server 注册 **Action Handler**。
    *   *示例*: 用户点击“下一章”。
    *   *Default*: 打开下一个文件。
    *   *NovelSkill*: 1. 记录当前章节已读；2. 将下一章内容加入知识库；3. 返回下一章数据。

### 2. Skill 最小运行协议 (Action -> Handler -> Response)
**目标**: 统一 Skill 与 Server 的交互边界，保证可插拔与可追踪。

**Action (Client -> Server)**:
```json
{
  "action": "open",
  "path": "/book/ch1.txt",
  "context": { "cursor": 0, "mode": "text" },
  "meta": { "traceId": "uuid", "ts": 1738200000 }
}
```

**Handler (Skill -> Server)**:
```ts
type ActionRequest = {
  action: string;
  path?: string;
  context?: Record<string, unknown>;
  meta?: { traceId?: string; ts?: number };
};

type ActionResponse = {
  status: "ok" | "error";
  handled: boolean;
  data?: unknown;
  view?: { type: "patch" | "full"; payload: unknown };
  effects?: Array<{ type: "log" | "index"; payload: unknown }>;
  error?: { code: string; message: string };
};
```

**Response 约束**:
*   `handled=false` 表示 Skill 放弃处理，回落到默认逻辑。
*   `view` 为可选 UI 更新（patch 或 full），用于驱动主视图刷新。
*   `effects` 为可选副作用（日志/索引/缓存等），由 Server 统一调度。
*   `error` 仅在 `status=error` 时出现，默认逻辑可根据 `code` 决定降级策略。

**快捷操作生命周期 (Session)**:
*   定义用户 session：连续操作中断超过 10 分钟即视为新 session。
*   由 AI 分析最近 n 个 session 中的用户操作序列，推断最合适的快捷操作。
*   快捷操作按使用次数从低到高排序，低频项优先被新的替换。
*   快捷操作存在数量上限（见 `shortcut_policy.max_items`），不会过期。
*   无 session 历史时，快捷操作为空。

### 3. 运行时职责确认 (Clarifications)
*   每个管理目录下的 `view.json` 与 `skills/` 均由 `ViewCreateSkill` 生成。
*   **元技能 `ViewCreateSkill` 仅负责生成**（view.json + skill bundle）；**运行时 handler 由 Server 统一加载/注册/执行**。
*   **Handler 逻辑不由 AI 直接生成可执行代码**，而是受控于 **action schema + 能力白名单**（类似 catalog/registry）。
*   后续用户操作的 Action 由目录内 Skill 响应，**不改变 UI 外框**，仅返回数据或追加快捷操作。
*   Skill 拦截异常时，回退到默认逻辑处理。
*   知识库写入由目录内 Skill 决定（而非统一全局规则）。
*   超出默认权限的授权由目录内 Skill 管理与提示用户确认。
*   `view.json` 与 Skill 版本一致；action 中的版本号来源于 `view.json`。
*   Skill 注册带版本号的 handler；路由时需版本号 + action 同时匹配才命中。
*   版本未命中时直接报错。
*   Skill 可注册新的或覆盖默认的 handler；action 按注册路由处理。
*   外部 Agent 异常时直接报错（不中断现有 UI/Skill）。
*   `view.json` 与 Skill 完全生成后以原子方式移动到管理目录（避免半更新）。
*   `ViewCreateSkill` 生成失败无需回滚或特殊处理。

### 4. Skill 模板 (Minimal Template)
**必须包含的模块**:
*   **注册与路由**: action -> handler 映射与可用 action 列表。
*   **权限声明**: read / write_generated / write_user / net 等能力清单。
*   **快捷操作管理**: 可追加/更新/撤回快捷操作（Action Bar）。
*   **知识库策略**: 触发条件、去重/摘要规则、索引范围。
*   **状态管理**: 位置/进度/游标等最小状态持久化。
*   **错误与降级**: 超时/异常处理与回退策略。
*   **审计与日志**: 写删/索引/权限升级的记录字段。

**模板文件**:
`skills/<skill-name>/config.json`
```json
{
  "name": "novel-reader",
  "version": "0.1.0",
  "actions": ["open", "next_chapter", "prev_chapter"],
  "permissions": {
    "read": true,
    "write_generated": true,
    "write_user": false,
    "net": false
  },
  "shortcuts": [
    { "id": "next", "label": "下一章", "action": "next_chapter" }
  ],
  "shortcut_policy": {
    "session_idle_minutes": 10,
    "max_items": 6,
    "expires": false
  },
  "knowledge": {
    "enabled": true,
    "triggers": ["open", "next_chapter"],
    "dedupe": "hash",
    "summary": "chapter"
  },
  "state": {
    "store": ".mindfs/state.json",
    "keys": ["cursor", "chapter"]
  },
  "errors": {
    "fallback": "default",
    "timeout_ms": 8000
  },
  "audit": {
    "enabled": true,
    "fields": ["actor", "action", "path", "ts", "summary"]
  }
}
```

`skills/<skill-name>/handlers.js`
```js
module.exports = function register(server, ctx) {
  const { fs, knowledge, audit } = ctx;

  server.on("open", async (req) => {
    const data = await fs.read(req.path);
    return { status: "ok", handled: true, data };
  });

  server.on("next_chapter", async (req) => {
    const nextPath = await ctx.nav.next(req.path);
    const data = await fs.read(nextPath);
    await knowledge.index(nextPath, data);
    audit.log({ action: "next_chapter", path: nextPath });
    return {
      status: "ok",
      handled: true,
      data,
      effects: [
        { type: "shortcut", payload: { id: "next", label: "下一章" } }
      ]
    };
  });
};
```

### 4.1 生成物边界 (Generated Artifacts Boundary)
**元技能 (ViewCreateSkill)**:
* 输入：目录特征 + 用户 session + catalog/registry 约束
* 输出：`view.json` + 运行时技能包（config + handlers 逻辑描述/调用计划）
* 不直接生成任意可执行 handler 代码；handler 由 Server 运行时根据白名单能力执行
* **Agent 驱动**：通过受控 CLI Agent（Claude/Codex/Gemini）生成 view.json；Server 负责探测可用 agent、选择优先级并传递目录偏好

**运行时技能包 (Skill Bundle)**:
* 位置：`.mindfs/skills/<skill-name>/`
* 组成：`config.json`（actions/permissions/shortcuts/knowledge/state/...）
* Handler 逻辑：以 **受控 action 调用计划** 或 **Server 认可的脚本** 表达
* Server 加载后注册 `action + version` 的 handler

### 5. 渐进式知识库 (Progressive Knowledge Base)
*   **拒绝全量扫描**: 不自动扫描所有文件（避免浪费算力和存储）。
*   **意图驱动**: 仅当用户“阅读”或“交互”时，才将相关内容存入向量库。
    *   *场景*: 看小说时，每读完一章，该章内容才会被索引。用户问“上一章讲了啥”时，AI 才有记忆。
*   **删除策略**: 文件删除时默认同步删除对应索引；可选保留“墓碑”记录用于审计/回滚。

### 6. 分层设计与文件读写规则
**分层职责**:
*   **L0: FS Core**: 统一文件访问入口，负责权限校验、审计、版本/锁。
*   **L1: Server Runtime**: Action 路由、Skill 生命周期、默认逻辑。
*   **L2: AI/Skills**: 仅通过 FS Core 调用读写能力，不直接触磁盘。
*   **L3: Client**: 纯 UI 与交互，不直接访问文件系统。

**默认读写规则**:
*   **用户文件 (workspace)**: 默认只读，禁止 Skill/AI 直接写删。
*   **Server/AI 创建文件**: 可读写删（例如 `.mindfs/*` 或显式创建的生成物）。
*   **写入升级**: 如需修改用户文件，必须显式请求用户确认或通过特殊 action（如 `write_user_file`）。
*   **审计**: 所有写删操作落 `history.jsonl`，含发起者、路径、摘要、时间。

### 7. 安全隔离建议 (Security Isolation)
*   **UI 执行隔离**: 动态渲染仅允许预置原子组件白名单，禁止任意 import/require。
*   **运行沙箱**: UI 运行环境禁网络/禁 FS，避免生成代码越权。
*   **Skill 能力隔离**: Skill 仅能通过 Server/FS Core API 执行读写与副作用。
*   **权限白名单**: Skill 声明所需能力（read, write_generated, write_user, net 等），Server 校验并提示确认。
*   **资源约束**: Skill 执行超时、并发与内存限制，防止阻塞主流程。

### 8. 生成式 UI 实现建议 (JSON Render)
**采用实现**: 直接使用 `vercel-labs/json-render` 作为生成式 UI 引擎。

**落地机制 (Quick Start 对齐)**:
1. **Catalog**: 由 `ViewCreateSkill` 维护组件白名单与 action schema（zod），作为 AI 允许输出的能力边界。
2. **Registry**: 为 catalog 中的每个组件注册 React 渲染实现（前端固定渲染器）。
3. **生成流程**: 目录加入管理后，`ViewCreateSkill` 分析目录与 session，调用 json-render 生成 `view.json`（UI tree），同时生成 skill bundle（handler 计划 + config）。
4. **渲染器**: 前端使用 DataProvider/ActionProvider/VisibilityProvider + Renderer 渲染 UI 树。
5. **文件类型渲染策略**: catalog 只提供少量“策略组件”（如 `FileViewer`, `MarkdownViewer`, `CodeViewer`），由组件内部根据 `ext/mime` 选择渲染方式，避免为每种文件类型扩展 catalog。

**Agent 生成流程 (ViewCreateSkill 标准化)**:
1. **探测可用 Agent**: 依次尝试 `claude`, `codex`, `gemini`（可配置白名单），通过一次轻量提示词/短超时运行判断可用性。
   - 若需要连续会话，需在全局 config 中为该 Agent 设置 `mode: "stdio"` 与 `sessionArgs`（按 CLI 具体参数配置）。
2. **读取目录偏好**: 每个受管目录读取 `.mindfs/config.json`（包含 viewCreateAgent/defaultAgent/偏好/禁用能力），作为用户提示的一部分。
3. **生成系统提示词**: 使用 json-render 的 `generateCatalogPrompt` + handler 白名单 + 目录结构摘要，形成系统提示词。
4. **生成用户提示词**: 汇总用户操作序列 + 目录结构 + 偏好，生成用户提示（例如“生成小说阅读书架”）。
5. **调用 Agent 生成 view.json**: 以系统提示词 + 用户提示词调用 Agent，输出 `view.json`，必要时返回 `skill bundle` 配置。
6. **原子写入与状态**: 写入 `.mindfs/view.json`，同时写入 `.mindfs/view.status.json` 标记 pending；由用户确认后接受。

**选择策略**:
* 若多个 Agent 可用，按目录配置的 `viewCreateAgent` 或偏好优先级选择；未配置时使用默认顺序 `claude > codex > gemini`。
* 任何 Agent 输出必须通过 catalog/registry 白名单校验，未通过则回退默认视图。
* 会话结束前生成 `session_summary.jsonl` 记录，便于后续偏好与快捷操作推断。

**要点**:
*   **JSON Schema** 驱动结构，渲染器只允许 catalog 内注册组件。
*   **Slot 等价**: catalog 组件即受控 slot，AI 只能选择与配置。
*   **可流式渲染**: 支持模型逐步生成 UI 树并实时渲染。
*   **Action 回传**: UI 触发 action 直接映射到 Skill 协议。

### 9. C/S 架构与远程访问
*   **Server**: 
    *   维护已管理目录列表。
    *   提供统一 FS API (Stream Reading, File Watcher)。
    *   调度外部 CLI Agent。
*   **Client**:
    *   纯 UI 呈现（Web/Electron）。
    *   支持远程连接（输入 Server IP + Token）。

---

## 四、 数据结构 (.mindfs)

```text
/TargetFolder/
  ├── .mindfs/
  │   ├── config.json         # [Meta] 目录偏好与 Agent 选择（含 sessionScope）
  │   ├── view.json           # [UI] Agent 生成的 JSON Schema
  │   ├── knowledge.db        # [RAG] 向量数据 (按需写入)
  │   ├── history.jsonl       # [Log] 操作与交互流水
  │   ├── session_summary.jsonl # [Log] 会话结束总结
  │   └── skills/             # [Logic] 技能配置
  │       └── novel-reader/
  │           ├── handlers.js # [后端逻辑] 注册 server.on('next_chapter', ...)
  │           └── config.json
```

**history.jsonl 格式 (JSON Lines)**:
```json
{"ts":1738200000,"actor":"user","origin":"client","dir":"/book","action":"open","path":"/book/ch1.txt","status":"ok","handled":true,"latency_ms":35,"summary":"open file","meta":{"traceId":"uuid"}}
{"ts":1738200123,"actor":"skill:novel-reader","origin":"skill","dir":"/book","action":"next_chapter","path":"/book/ch2.txt","status":"ok","handled":true,"latency_ms":82,"summary":"index chapter","effects":["index","shortcut"],"meta":{"traceId":"uuid"}}
{"ts":1738200200,"actor":"server","origin":"default","dir":"/book","action":"open","path":"/book/ch3.txt","status":"error","handled":false,"latency_ms":12,"error":{"code":"E_READ","message":"permission denied"},"meta":{"traceId":"uuid"}}
```

**字段说明**:
*   `ts`: Unix 时间戳（秒）
*   `actor`: user / server / skill:<name>
*   `origin`: client / skill / default / system
*   `dir`: 当前管理目录根路径
*   `action`: 触发的 action
*   `path`: 目标文件路径（可选）
*   `status`: ok / error
*   `handled`: 是否由 Skill 处理
*   `latency_ms`: 处理耗时（毫秒）
*   `summary`: 简要描述（可选）
*   `effects`: 副作用类型数组（可选）
*   `error`: 错误对象（可选）
*   `meta`: traceId / extra（可选）

**可选数据库 (SQLite)**:
*   `history.jsonl` 作为审计事实源，SQLite 用于查询与统计。
*   建议启用 WAL 模式，减少写入阻塞。

**表结构建议**:
```sql
CREATE TABLE IF NOT EXISTS history (
  ts INTEGER,
  actor TEXT,
  origin TEXT,
  dir TEXT,
  action TEXT,
  path TEXT,
  status TEXT,
  handled INTEGER,
  latency_ms INTEGER,
  summary TEXT,
  effects TEXT,
  error TEXT
);
CREATE INDEX IF NOT EXISTS idx_history_ts ON history(ts);
CREATE INDEX IF NOT EXISTS idx_history_dir ON history(dir);
CREATE INDEX IF NOT EXISTS idx_history_action ON history(action);
```

**config.json 模板建议**:
```json
{
  "viewCreateAgent": "claude",
  "defaultAgent": "codex",
  "userDescription": "这是一个小说目录，用于按章节阅读与追踪进度。",
  "sessionScope": "file"
}
```

---

## 五、 MVP 实施路线图 (Phase 1)

**目标**：搭建 Monorepo，实现 `mindfs .` 启动 Server，前端显示带有左侧树和底部栏的框架。

1.  **Monorepo Setup**: 初始化 `cli`, `server`, `web`。
2.  **Server Core**:
    *   实现 HTTP/WS 服务。
    *   实现“目录挂载”逻辑。
3.  **Web Layout**:
    *   实现固定布局：Sidebar (Tree) + Main + BottomBar。
    *   集成 `@swc/wasm-web` 准备动态渲染。
4.  **CLI**:
    *   实现启动命令。

---

## 六、 Server 技术选型 (Go)

**语言**: Go（优先考虑调试报错清晰度与 AI 生成代码可修复性）

**服务端框架**:
*   **HTTP 路由**: `go-chi/chi`
*   **WebSocket**: `gorilla/websocket`
*   **中间件**: `chi/middleware`（日志、recover、request-id）

**替代选项**:
*   极简模式：`net/http` + `gorilla/websocket`（牺牲路由组织）

---

## 七、 Server 架构建议（合并版）

**核心分层**:
*   **API Layer**: HTTP + WS 入口，鉴权、限流、版本协商。
*   **Action Router**: 按目录/Skill 路由 action，处理默认逻辑与回退。
*   **Skill Runtime**: 加载、生命周期、超时控制、权限校验、隔离执行。
*   **FS Core**: 统一文件访问入口（读/写/删/stream/watch）+ 审计。
*   **Knowledge Engine**: 按目录隔离的索引/检索/摘要管线。
*   **UI Orchestrator**: json-render schema 生成/缓存/热更新。
*   **Audit & History**: history.jsonl 与可选集中日志。

**服务通信**:
*   **Client ↔ Server**: WS 用于 action + UI patch，HTTP 用于静态资源/下载。
*   **Server ↔ Agent**: 任务队列 + 回调/事件，避免阻塞主流程。
*   **Skill ↔ Server**: 仅通过 ctx API（fs/knowledge/audit/shortcut）。

**关键机制**:
*   **目录隔离**: 每个目录独立 Skill/知识库/状态/权限。
*   **单目录 UI**: UI 同一时间仅操作一个管理目录。
*   **失败回退**: Skill 异常回退默认逻辑。
*   **缓存策略**: 目录树缓存 + schema 缓存 + LRU 内容缓存（大文件流式）。
*   **一致性**: view/schema 版本号 + action 兼容检查，保证热更新可用。
