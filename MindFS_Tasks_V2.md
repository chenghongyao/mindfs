# Tasks: MindFS V2

**Input**: MindFS_Planning_V2.md
**基于**: V1 已完成的基础设施 (Phase 1-4 大部分已完成)

**核心变化**: 从"AI 增强的文件管理器"转向"Agent 远程访问网关 + 结果可视化平台"

## Format: `[ID] [Status] [Priority] Description`

Status: `[ ]` 待做, `[~]` 进行中, `[X]` 已完成, `[-]` 废弃

---

## Phase 1: Session 系统 (P0 核心)

**Goal**: 实现用户与 Agent 的对话能力，Session 作为交互载体

### 1.1 Session 数据层

- [X] T101 [P0] 定义 Session 数据结构 (server/internal/session/types.go)
  ```go
  type Session struct {
    Key             string
    Type            string           // chat/view/skill
    Agent           string           // claude/codex/gemini
    AgentSessionID  *string          // Agent 原生 session-id (用于恢复)
    Name            string           // AI 生成的摘要
    Status          string           // active/idle/closed
    Summary         *SessionSummary  // 关闭时生成
    Exchanges       []Exchange       // 对话记录 (降级恢复用)
    RelatedFiles    []RelatedFile
    GeneratedView   string
    CreatedAt       time.Time
    UpdatedAt       time.Time
    ClosedAt        *time.Time
  }

  type SessionSummary struct {
    Title       string
    Description string
    KeyActions  []string
    Outputs     []string
    GeneratedAt time.Time
  }
  ```

- [X] T102 [P0] 实现 Session 存储与读写 (server/internal/session/store.go)
  - 存储位置: `.mindfs/sessions/session-{key}.json`
  - 不使用 index.json，直接扫描目录
  - 支持创建、读取、更新、列表
  - 记录 agent 和 agent_session_id

- [X] T103 [P0] 实现 Session 生命周期管理 (server/internal/session/manager.go)
  - 创建新 Session (指定类型: chat/view/skill，指定 Agent)
  - 空闲超时检测 (10 分钟无操作 → idle)
  - 关闭 Session 时生成 Summary
  - 恢复 Session (优先原生 resume，降级 exchanges)

- [X] T104 [P0] 实现 Session 恢复逻辑 (server/internal/session/resume.go)
  - 优先使用 Agent 原生 `--resume` 机制
  - 失败时用 exchanges 构建上下文作为降级方案
  - 恢复后更新 Session 状态为 active

### 1.2 Session API

- [X] T112 [P0] 实现 Session REST API (server/internal/api/session.go)
  - `GET /api/sessions` - 列表
  - `GET /api/sessions/:key` - 详情
  - `POST /api/sessions` - 创建
  - `POST /api/sessions/:key/message` - 发送消息

- [X] T113 [P0] 实现 Session WebSocket 消息处理 (server/internal/api/ws.go)
  - `session.create` - 创建 Session
  - `session.message` - 发送消息
  - `session.stream` - 流式响应
  - `session.done` - 响应完成
  - `session.close` - 关闭 Session
  - `session.error` - 错误通知

- [X] T114 [P0] 定义 WebSocket 消息类型 (server/internal/api/ws_types.go)
  - 请求/响应消息格式
  - Session 相关消息类型
  - 视图相关消息类型 (view.update, view.switch)
  - 文件相关消息类型 (file.created, file.changed)

### 1.3 Agent 进程管理

**协议**: 参考 Happy 项目，支持三种协议
- **Claude**: stream-json (--output-format stream-json --input-format stream-json)
- **Gemini**: ACP (--experimental-acp)
- **Codex**: MCP (mcp-server)

- [X] T106 [P0] 实现 Agent 进程池 (server/internal/agent/pool.go)
  - 每个 Session 对应一个 Agent 进程
  - 通过 stdin/stdout 多轮对话，保持上下文
  - 进程崩溃检测与重启
  - 空闲超时关闭

- [-] T107 [P0] ~~实现 Agent 进程通信 (server/internal/agent/process.go)~~ → 重写为多协议支持

- [X] T107-MULTI [P0] 实现多协议 Agent 进程通信
  - **统一接口**: server/internal/agent/process.go (Process interface)
  - **Claude**: server/internal/agent/claude/process.go (stream-json)
  - **Gemini**: server/internal/agent/gemini/process.go (ACP)
  - **Codex**: server/internal/agent/codex/process.go (MCP)
  - **公共类型**: server/internal/agent/acp/ 包 (SessionUpdate 等)

- [X] T107-TRANSPORT [P0] 实现 Transport 配置 (server/internal/agent/config.go)
  - Protocol 字段: stream-json / acp / mcp
  - TransportConfig: InitTimeout, IdleTimeout, FilterPatterns
  - 内置默认配置: Claude(30s), Gemini(120s), Codex(30s)

- [X] T108 [P0] 实现 Agent 配置加载 (server/internal/agent/config.go)
  - 配置位置: `~/.config/mindfs/agents.json`
  - 内置默认配置 (无需配置文件即可使用)
  - 支持 Protocol, Args, Env, Transport 配置

- [-] T109 [P0] ~~实现流式输出推送~~ → 已合并到 T107-MULTI

- [X] T109-ACP [P0] 重写流式输出为 ACP SessionUpdate (server/internal/agent/process.go)
  - 解析 ndJSON 行为 SessionUpdate
  - SessionUpdate 类型:
    - `agent_message_chunk`: 文本增量
    - `agent_thought_chunk`: 思考过程
    - `tool_call`: 工具调用开始
    - `tool_call_update`: 工具状态更新
    - `agent_message_complete`: 消息完成
  - 转换为前端 StreamChunk 格式
  - WebSocket 推送到客户端
  - **实现**: process.go 的 toSessionUpdate() 方法

- [X] T109-2 [P0] 定义流式输出消息格式 (server/internal/agent/stream_types.go)
  - 文本块: `{ type: "text", content: "..." }`
  - 思考过程: `{ type: "thinking", content: "..." }`
  - 进度: `{ type: "progress", task: "...", percent: 60 }`
  - 文件操作: `{ type: "file_start/progress/done", path: "...", ... }`
  - 工具调用: `{ type: "tool_call/result", tool: "...", ... }`
  - **新增**: 权限请求: `{ type: "permission_request", id: "...", description: "...", options: [...] }`
  - 错误: `{ type: "error", code: "...", message: "..." }`

- [X] T109-3 [P1] 实现 Agent 可用性探测 (server/internal/agent/probe.go)
  - 启动时探测所有配置的 Agent
  - 定期重新探测 (每 5 分钟)
  - 返回可用状态、版本、错误信息
  - 使用失败后立即重新探测

- [X] T109-API [P1] 实现 Agent 状态 API (server/internal/api/agents.go)
  - `GET /api/agents` - 获取所有 Agent 状态
  - `POST /api/agents/:name/probe` - 手动触发单个 Agent 探测
  - 返回: name, available, version, error, last_probe
  - **实现**: AgentHandler 结构体，集成到 HTTPHandler

- [X] T109-PERMISSION [P1] 实现权限请求处理 (server/internal/agent/acp/permission.go)
  - 解析 Agent 的 `requestPermission` RPC 请求
  - 通过 WebSocket 推送给前端
  - 等待用户响应 (proceed_once/proceed_always/cancel)
  - 返回 PermissionResponse 给 Agent
  - 超时默认拒绝 (30s)
  - **实现**: PermissionHandler 结构体，支持异步请求/响应

- [X] T109-RETRY [P1] 实现 Agent 错误处理与重试 (server/internal/agent/retry.go)
  - 初始化超时: 重试 3 次，指数退避 (1s, 2s, 4s)
  - 进程崩溃: 自动重启，保留 session 上下文
  - 响应超时: 发送取消请求，等待 2s 后强制终止
  - 权限请求超时: 默认拒绝，通知用户
  - **实现**: Retry 函数 + RecoverablePool

- [X] T109-SHUTDOWN [P1] 实现 Agent 进程优雅关闭 (server/internal/agent/shutdown.go)
  - GracefulShutdown 函数: 关闭 stdin → 等待 → SIGTERM → 等待 → SIGKILL
  - ProcessCloser 封装优雅关闭逻辑

### 1.4 文件创建追踪

- [X] T110 [P0] 实现文件系统监听 (server/internal/fs/watcher.go)
  - 使用 fsnotify 监听目录变化
  - 新文件自动关联到当前活跃 Session
  - 更新 file-meta.json

- [-] T111 [P0] ~~实现 Agent 输出解析 (server/internal/agent/parser.go)~~ → 废弃
  - 原因: ACP 协议中文件操作通过 tool_call 结构化传递，无需正则解析
  - 替代: T109-ACP 中的 tool_call 解析

### 1.5 Agent 交互上下文

**原则**: 路径优先（让 Agent 自己读取），只传 Agent 无法获取的信息（选中内容、userDescription、catalog 等）

- [X] T115 [P0] 实现客户端上下文收集 (web/src/services/context.ts)
  - 收集当前目录/文件路径
  - 收集用户选中内容 (selection) - UI 状态，Agent 无法获取
  - 收集当前视图信息 (视图模式)
  - 不传文件内容（Agent 可自己读取）

- [X] T116 [P0] 实现服务端上下文构建 (server/internal/context/builder.go)
  - 通用上下文: root_path、userDescription、关联 Session
  - 视图模式: catalog、registry schema、server API、当前视图、few-shot 示例
  - 技能模式: 目录自定义 skill 列表 (.mindfs/skills/)
  - 不传目录结构（Agent 可自己 ls）
  - 不传 Agent 内置能力（Agent 自己知道）

- [X] T117 [P0] 实现 Catalog 与 Schema 导出 (server/internal/context/catalog.go)
  - 导出 json-render 组件白名单
  - 导出组件 props schema (zod)
  - 导出可用 action 列表

- [X] T118 [P0] 实现 Server API 列表导出 (server/internal/context/api_list.go)
  - 导出可用的 REST API endpoint
  - 导出可用的 WebSocket action
  - 包含参数定义和响应类型

- [X] T119 [P0] 实现视图示例加载 (server/internal/context/examples.go)
  - 加载 few-shot 视图示例
  - 根据场景选择最相关的示例
  - 存储位置: ~/.config/mindfs/view-examples/

- [X] T120 [P0] 实现关联 Session 查找 (server/internal/context/related.go)
  - 优先级 1: 涉及当前文件的 Session
  - 优先级 2: 涉及同目录文件的 Session
  - 优先级 3: 同管理目录下的其他 Session
  - 每组内按时间排序，限制返回数量 (最近 3 个)

- [X] T121 [P0] 实现 Agent 提示词构建 (server/internal/context/prompt.go)
  - 根据模式构建 System Prompt (工作目录、userDescription、能力边界)
  - 构建 User Prompt (用户消息 + 文件路径 + 选中内容)
  - 新建 Session 时构建完整上下文
  - 恢复/继续 Session 时仅传新消息和选中内容

### 1.6 目录自定义 Skill 调用

**核心思路**: Agent 启动目录设为 .mindfs/，通过 --add-dir 添加用户目录，Agent 可自己发现 skill

- [X] T125 [P0] 更新 Agent 配置支持工作目录模板 (server/internal/agent/config.go)
  - cwdTemplate: Agent 工作目录模板 (如 "{root}/.mindfs")
  - addDirArgs: 添加用户目录的参数 (如 ["--add-dir", "{root}"])
  - 启动时替换 {root} 为实际路径

- [X] T126 [P0] 实现 Skill 执行 API (server/internal/api/skill.go)
  - `POST /api/skills/:id/execute` - 执行 skill
  - 权限校验
  - 执行 handler
  - 记录审计日志

- [X] T127 [P1] 实现 Skill 降级方案 (server/internal/agent/prompt.go)
  - Agent 不支持 --add-dir 时，在提示词中传递 skill 列表
  - 告知 Agent 通过 API 调用 skill

### 1.7 目录设置

- [X] T128 [P0] 实现目录设置 API (server/internal/api/config.go)
  - `GET /api/dirs/:id/config` - 获取目录配置
  - `PUT /api/dirs/:id/config` - 更新目录配置
  - 支持 userDescription、defaultAgent、viewCreateAgent

- [X] T129 [P0] 实现目录设置面板 (web/src/components/SettingsPanel.tsx)
  - userDescription 文本框
  - defaultAgent 下拉框
  - Agent 状态显示 + 刷新按钮
  - 保存/取消按钮
  - **实现**: 接入 /api/dirs/:id/config 和 /api/agents API

- [X] T130 [P0] 实现目录设置入口
  - 右侧边栏顶部设置图标 (⚙️)
  - 文件树目录右键菜单 "目录设置"
  - 首次添加目录时自动弹出
  - **实现**: RightSidebar 设置图标，FileTree 右键菜单

### 1.8 Session 前端

- [X] T131 [P0] 实现 Session 列表组件 (web/src/components/SessionList.tsx) [重构]
  - 显示 Session 列表
  - 类型图标 (💬/🎨/⚡)
  - Agent 标签 [Claude]/[Codex]
  - 状态标记 (活跃/空闲/已关闭)
  - 已关闭 Session 显示 [↻ 恢复] 按钮
  - 关联文件缩略显示
  - **点击活跃 Session (active/idle) → 展开浮框** (可交互)
  - **点击已关闭 Session (closed) → 主视图显示历史** (只读)
  - **实现**: SessionCard 组件，按状态分组显示

- [X] T132 [P0] 实现 Agent 交互浮框 (web/src/components/AgentFloatingPanel.tsx)
  - 浮框占主视图 80% 区域
  - 顶部显示 Session 信息 (类型、Agent、状态、收起按钮)
  - 显示对话历史 (用户消息 + Agent 响应)
  - 实时流式输出显示
  - 进度指示器 (文件下载、任务执行等)
  - 关联文件列表 (点击跳转到文件视图，浮框收起)
  - 底部输入框 + 发送按钮 (继续对话)
  - 点击浮框外部收起为气泡
  - **实现**: 完整浮框组件，集成 StreamMessage 和 PermissionDialog

- [X] T132-STREAM [P0] 实现流式消息渲染组件 (web/src/components/stream/)
  - StreamMessage.tsx: 流式消息容器，按类型分组渲染
  - TextChunk.tsx: 文本块渲染
  - ThinkingBlock.tsx: 思考过程（可折叠）
  - ToolCallCard.tsx: 工具调用卡片（可折叠，显示参数和结果）
  - **实现**: 完整组件集

- [X] T132-PERMISSION [P0] 实现权限请求对话框 (web/src/components/dialog/PermissionDialog.tsx)
  - 显示权限类型图标 (file_write/command_exec/network)
  - 显示操作描述
  - 显示选项按钮 (允许一次/始终允许/拒绝)
  - 响应后关闭对话框
  - 超时倒计时显示 (30秒自动拒绝)

- [X] T133 [P0] 实现 Agent 交互气泡 (web/src/components/AgentBubble.tsx)
  - 显示在主视图右下角
  - 显示当前活跃 Session 名称和状态
  - 点击展开浮框
  - 无活跃 Session 时不显示
  - **实现**: 完整气泡组件，带流式动画

- [X] T134 [P0] 实现 Session 历史视图 (web/src/components/SessionHistory.tsx)
  - 用于显示已关闭 Session 的历史记录 (主视图，只读)
  - 顶部显示 Summary
  - 显示对话历史
  - 关联文件列表
  - [↻ 恢复] 按钮 (恢复后切换到浮框)
  - **实现**: 完整历史视图组件

- [X] T135 [P0] 实现 Session 服务 (web/src/services/session.ts)
  - 创建 Session (指定类型和 Agent)
  - 发送消息 (附带客户端上下文)
  - 订阅 Session 更新
  - 管理浮框展开/收起状态
  - **新增**: 处理权限请求响应
  - **实现**: SessionService 类，WebSocket 连接管理

- [X] T135-HOOK [P0] 实现 WebSocket 消息处理 Hook (web/src/hooks/useSessionStream.ts)
  - chunks 状态管理
  - isStreaming 状态
  - permissionRequest 状态
  - respondToPermission 回调
  - **实现**: useSessionStream hook

- [X] T136 [P0] 实现 Agent 可用性状态展示 (web/src/services/agents.ts)
  - 获取可用 Agent 列表 (`GET /api/agents`)
  - 缓存机制 (30秒 TTL)
  - probeAgent 手动刷新
  - **实现**: fetchAgents, probeAgent, isAgentAvailable 函数

- [X] T137 [P0] 实现 Session 空闲检测定时器 (server/internal/session/idle_checker.go)
  - 定时检查所有 Session 的最后活动时间
  - active → idle (10 分钟无操作)
  - idle → closed (30 分钟无操作)
  - 超过 max_idle_sessions 时关闭最老的 idle Session
  - 状态变更时通知前端

**Checkpoint**: 用户可以通过 ActionBar 与 Agent 对话，Agent 交互以浮框形式展示，不打断主视图

---

## Phase 2: ActionBar 重构 (P0 核心)

**Goal**: ActionBar 成为 Agent 交互的主入口，支持三种模式

### 2.1 模式切换

- [X] T201 [P0] 重构 ActionBar 支持三种模式 + Agent 选择 (web/src/components/ActionBar.tsx)
  - 对话模式 (默认) → 创建 chat 类型 Session
  - 生成视图模式 → 创建 view 类型 Session
  - 执行技能模式 → 创建 skill 类型 Session
  - **实现**: 完整重构，集成 ModeAgentSelector

- [X] T202 [P0] 实现模式+Agent 下拉框组件 (web/src/components/ModeAgentSelector.tsx)
  - 左侧: 模式列表 (对话/生成视图/执行技能)
  - 右侧: Agent 列表 (Claude/Codex/Gemini + 可用状态)
  - 按钮显示: `对话 · Claude ▼`
  - 记住每种模式的 Agent 偏好
  - **实现**: 完整下拉组件

- [X] T203 [P0] 实现 Agent 偏好存储 (server/internal/config/preferences.go)
  - 存储位置: `~/.config/mindfs/preferences.json`
  - 记录每种模式的默认 Agent
  - **实现**: PreferencesStore 结构体

### 2.2 对话模式

- [X] T203 [P0] 实现对话输入与发送 (web/src/components/ActionBar.tsx)
  - 输入框绑定到当前 Session
  - 发送消息到 Agent
  - 无 Session 时自动创建
  - **实现**: 已集成到 ActionBar 重构

- [X] T204 [P0] 实现 Session 状态联动
  - 选中 Session 时显示 Session 名称
  - 选中文件时显示 "Connected"
  - **实现**: ActionBar 支持 currentSession prop，显示 Session 类型图标和状态

- [X] T209 [P1] 实现统一错误处理 (web/src/services/error.ts)
  - 错误码定义 (session.*/agent.*/view.*/file.*/skill.*)
  - Toast 通知组件 (临时错误 3 秒自动消失)
  - 可恢复错误显示重试按钮
  - 阻断性错误显示模态对话框
  - **实现**: ErrorService + ToastContainer

- [X] T210 [P1] 实现错误边界组件 (web/src/components/ErrorBoundary.tsx)
  - 主视图错误边界
  - 浮框错误边界
  - 错误日志记录到审计
  - **实现**: ErrorBoundary + MainViewErrorBoundary + FloatingPanelErrorBoundary

### 2.3 执行技能模式

- [X] T205 [P0] 实现 Skill 下拉框 (web/src/components/SkillSelector.tsx)
  - 合并展示 Agent 能力 + 目录 Skill
  - 分组显示 (Agent 能力 / 当前目录)
  - 选择后更新输入框 placeholder
  - **实现**: 完整下拉组件，支持搜索

- [X] T206 [P0] 实现 Skill 执行调用 (web/src/services/skills.ts)
  - 获取可用 Skill 列表
  - 执行 Skill 并传递参数
  - **实现**: fetchSkills, executeSkill, getAllSkills

- [X] T207 [P0] 实现 Agent 能力配置 (server/internal/agent/capabilities.go)
  - 定义 Agent 能力列表
  - 存储位置: `~/.config/mindfs/capabilities.json`
  - **实现**: CapabilitiesStore 结构体

### 2.4 生成视图模式

- [X] T208 [P1] 实现视图生成输入 (web/src/components/ActionBar.tsx)
  - 切换到视图模式时更新 placeholder
  - 发送触发 ViewCreateSkill
  - **实现**: 已集成到 ActionBar 重构，modePlaceholders

**Checkpoint**: ActionBar 支持三种模式切换，可以对话、执行技能、生成视图

---

## Phase 3: 文件-Session 关联 (P0 核心)

**Goal**: 记录文件来源，支持文件和 Session 双向跳转

### 3.1 文件元数据

- [X] T301 [P0] 实现文件元数据存储 (server/internal/fs/file_meta.go)
  - 存储位置: `.mindfs/file-meta.json`
  - 记录文件来源 Session、创建时间、创建者
  - **实现**: 扩展 FileMetaEntry，添加 SessionName, Agent, UpdatedAt

- [X] T302 [P0] Agent 创建文件时记录元数据 (server/internal/agent/file_writer.go)
  - 写入文件时自动记录到 file-meta.json
  - 关联到当前 Session
  - **实现**: FileWriteTracker 结构体，跟踪 tool_call 和 tool_result

- [X] T303 [P0] 实现文件元数据 API (server/internal/api/http.go)
  - `GET /api/file/meta?path=xxx` - 获取文件元数据
  - **实现**: handleFileMeta 处理函数

### 3.2 前端关联展示

- [X] T304 [P0] 文件视图显示来源 Session (web/src/components/FileViewer.tsx)
  - 文件头部显示 [来源: Session名称]
  - 点击跳转到 Session 视图
  - **实现**: 添加 meta prop，显示来源 Session 信息

- [X] T305 [P0] 文件树显示 Session 关联标记 (web/src/components/FileTree.tsx)
  - ◆ 当前选中 Session 生成的文件
  - ◇ 其他 Session 生成的文件
  - Hover 显示来源信息
  - **实现**: 添加 fileMetas 和 activeSessionKey props

- [X] T306 [P0] Session 视图显示关联文件 (web/src/components/SessionViewer.tsx)
  - 显示关联文件列表
  - 点击 [查看全部] 展开完整列表
  - 点击文件跳转到文件视图
  - **实现**: 添加 related_files 和 onFileClick

**Checkpoint**: 文件和 Session 可以双向跳转，关联关系清晰可见

---

## Phase 4: 视图路由系统 (P1)

**Goal**: 基于路径/类型的规则匹配，支持多视图切换

### 4.1 路由配置

- [X] T401 [P1] 重构 view.json 为路由配置 (server/internal/router/view_router.go)
  - 支持 routes 数组
  - 支持 match 规则 (path/ext/mime/name/any/all)
  - 支持 priority 优先级
  - **实现**: ViewRoute, MatchRule, ViewRouterConfig

- [X] T402 [P1] 实现路由匹配逻辑 (server/internal/router/matcher.go)
  - glob 路径匹配
  - 扩展名匹配
  - 组合规则 (any/all)
  - **实现**: matchesRule, matchGlob, matchExtension, matchMime

- [X] T403 [P1] 实现多视图解析 (server/internal/router/resolver.go)
  - 返回所有匹配的视图
  - 按 priority 排序
  - 考虑用户上次选择
  - **实现**: ViewResolver, ResolvedView

### 4.2 视图状态

- [X] T404 [P1] 实现 view.status.json 管理 (server/internal/router/view_status.go)
  - active_versions: 各规则的激活版本
  - last_selected: 用户上次选择
  - pending: 正在生成的版本
  - **实现**: ViewStatus, PendingView

- [X] T405 [P1] 实现用户选择记忆 (server/internal/router/preference.go)
  - 用户切换视图时保存
  - 下次打开同一路径时使用
  - **实现**: ViewPreferences

### 4.3 前端视图切换

- [X] T406 [P1] 实现视图+版本合并下拉框 (web/src/components/ViewVersionSelector.tsx)
  - 参考 ModeAgentSelector 设计
  - 左侧: 视图类型列表 (匹配当前路径的所有视图)
  - 右侧: 当前视图的版本列表
  - 按钮显示: `{视图名称} · {版本}` 如 `小说阅读器 · v2`
  - 切换视图类型时右侧版本列表同步更新
  - 无自定义视图时仅显示视图名称
  - **实现**: 完整下拉组件

- [X] T407 [P1] 实现视图切换 API 调用 (web/src/services/view.ts)
  - 切换视图类型
  - 切换版本
  - 保存用户选择
  - **实现**: fetchViewRoutes, switchViewRoute, switchViewVersion, saveViewPreference

**Checkpoint**: 用户可以在多个匹配视图间切换，选择自动记忆

---

## Phase 5: 视图版本管理 (P1)

**Goal**: 支持视图版本回退、重新生成

### 5.1 版本存储

- [X] T501 [P1] 重构视图存储结构 (server/internal/fs/view_store.go)
  ```
  .mindfs/views/
  ├── _default/
  ├── {rule-id}/
  │   ├── v1.json
  │   ├── v1.meta.json
  │   ├── v2.json
  │   └── v2.meta.json
  ```
  - **实现**: ViewVersionStore, ViewVersionMeta

- [X] T502 [P1] 实现版本元数据 (server/internal/fs/view_store.go)
  - 记录 prompt、agent、parent、created_at
  - 支持版本追溯
  - **实现**: ViewVersionMeta 结构体

### 5.2 版本操作

- [X] T503 [P1] 实现版本切换 API (server/internal/api/view.go)
  - `POST /api/view/switch` - 切换版本
  - `GET /api/view/versions/:ruleId` - 获取版本列表
  - **实现**: ViewHandler.handleSwitch, handleVersions

- [X] T504 [P1] 实现重新生成 API (server/internal/api/view.go)
  - `POST /api/view/generate` - 生成新版本
  - 支持基于现有版本或全新生成
  - **实现**: ViewHandler.handleGenerate

### 5.3 前端版本管理

- [X] T505 [P1] 实现重新生成对话框 (web/src/components/RegenerateDialog.tsx)
  - 显示上次 prompt
  - 输入新 prompt
  - 选择应用范围
  - 选择基于现有版本或全新生成
  - **实现**: 完整对话框组件

**Checkpoint**: 用户可以回退视图版本，可以重新生成视图（版本切换已合并到 T406 ViewVersionSelector）

---

## Phase 6: 快捷操作 (P1)

**Goal**: 快捷操作作为视图的一部分，显示在主视图底部

### 6.1 快捷操作定义

- [X] T601 [P1] 扩展 view.json 支持 shortcuts (server/internal/router/shortcuts.go)
  ```json
  {
    "shortcuts": [
      { "id": "prev", "label": "← 上一章", "action": "prev_chapter", "position": "left" }
    ]
  }
  ```
  - **实现**: Shortcut 结构体

- [X] T602 [P1] 实现快捷操作解析 (server/internal/router/shortcuts.go)
  - 从 view.json 读取 shortcuts
  - 返回给前端渲染
  - **实现**: LoadShortcuts, LoadShortcutsFromView, GroupShortcutsByPosition

### 6.2 前端快捷操作

- [X] T603 [P1] 实现快捷操作栏组件 (web/src/components/ShortcutBar.tsx)
  - 渲染在主视图底部
  - 支持 left/center/right 位置
  - 支持按钮和文本类型
  - **实现**: ShortcutBar, ShortcutButton

- [X] T604 [P1] 实现快捷操作触发 (web/src/components/ShortcutBar.tsx)
  - 点击按钮触发 action
  - 通过 ActionClient 发送
  - **实现**: onAction 回调

- [X] T605 [P1] 集成快捷操作到主视图 (web/src/renderer/merge.ts)
  - 视图有 shortcuts 时显示快捷操作栏
  - 无 shortcuts 时隐藏
  - **实现**: extractShortcuts, mergeViewWithShortcuts

**Checkpoint**: AI 生成的视图可以包含快捷操作，显示在主视图底部

---

## Phase 7: 体验优化 (P2)

**Goal**: 任务状态、通知、移动端适配

### 7.1 任务状态

- [X] T701 [P2] 实现任务状态追踪 (server/internal/agent/task.go)
  - 任务队列
  - 执行状态 (pending/running/completed/failed)
  - 执行进度
  - **实现**: TaskQueue, Task, TaskUpdate, TaskListener

- [X] T702 [P2] 实现任务状态 WebSocket 推送 (server/internal/api/ws.go)
  - 任务开始/进度/完成通知
  - 推送到所有连接的客户端
  - **实现**: InitTaskListener, broadcastTaskUpdate, handleTaskList, handleTaskGet

- [X] T703 [P2] 实现任务状态 UI (web/src/components/TaskStatus.tsx)
  - 显示当前任务状态
  - 进度条
  - 完成通知
  - **实现**: TaskStatus, TaskCard, useTaskStatus hook, TaskCompletionToast

### 7.2 认证

- [X] T704 [P2] 实现 Token 认证 (server/internal/api/auth.go)
  - 生成访问 Token
  - 验证 Token
  - Token 过期处理
  - **实现**: TokenStore, AuthService, Token 结构体

- [X] T705 [P2] 实现认证中间件 (server/internal/api/middleware.go)
  - HTTP 请求认证
  - WebSocket 连接认证
  - **实现**: AuthMiddleware, WSAuthMiddleware, RequireAuth, OptionalAuth

- [X] T706 [P2] 实现登录 UI (web/src/components/Login.tsx)
  - Token 输入
  - 连接状态显示
  - **实现**: Login, LoginStatus, useAuth hook

### 7.3 移动端适配

- [X] T707 [P2] 实现响应式布局 (web/src/layout/AppShell.tsx)
  - 移动端隐藏侧边栏
  - 底部导航切换
  - 手势操作支持
  - **实现**: useResponsive hook, MobileNavButton, 移动端布局

---

## Phase 8: 审计日志 (P1)

**Goal**: 记录必要的用户操作流水

- [X] T801 [P1] 实现审计日志写入 (server/internal/audit/writer.go)
  - 写入 `.mindfs/history.jsonl`
  - 支持多种操作类型
  - **实现**: Writer, WriterPool, Logger

- [X] T802 [P1] 定义审计事件类型 (server/internal/audit/types.go)
  ```go
  type AuditEntry struct {
    Ts      int64  `json:"ts"`
    Type    string `json:"type"`    // session/file/view/skill/dir
    Action  string `json:"action"`
    Actor   string `json:"actor"`   // user/agent/system
    Session string `json:"session,omitempty"`
    Path    string `json:"path,omitempty"`
    // ...其他上下文字段
  }
  ```
  - **实现**: Entry, EntryType, Action, Actor 常量

- [X] T803 [P1] 集成审计到各模块
  - Session: create, message, close, resume
  - File: open, create, delete, rename
  - View: generate, switch, revert
  - Skill: execute, cancel
  - Dir: add, remove
  - **实现**: HTTPHandler.getAuditLogger, WSHandler.getAuditLogger, 集成到 session handlers

- [X] T804 [P1] 定义统一错误码 (server/internal/api/errors.go)
  - session.*: not_found, already_closed, resume_failed
  - agent.*: not_available, process_crashed, timeout, invalid_response
  - view.*: not_found, invalid_schema, generation_failed
  - file.*: not_found, permission_denied, read_failed
  - skill.*: not_found, permission_denied, execution_failed
  - **实现**: ErrorCode 常量, APIError 结构体
  - 通用: internal_error, invalid_request, rate_limited

- [X] T805 [P1] 实现错误恢复策略 (server/internal/api/recovery.go)
  - agent.timeout: 自动重试 1 次
  - agent.process_crashed: 自动重启进程
  - session.resume_failed: 降级到 exchanges
  - 记录错误到审计日志
  - **实现**: RecoveryStrategy, WithAgentTimeoutRecovery, WithAgentCrashRecovery, WithSessionResumeRecovery

---

## Phase 10: 待优化 (Backlog)

- [ ] T1001 [P2] 优化多 Session 文件监控
  - **问题**: 多个 active session 时，每个 session 创建独立 FileWatcher 监控同一 rootPath，浪费资源
  - **问题**: 多个 session 同时写同一文件时，file-meta 归属不确定
  - **方案**: 改为每个 rootPath 一个 watcher，根据 Agent 进程实际输出判断文件归属
  - **涉及文件**: server/internal/fs/watcher.go, server/internal/api/ws.go

---

## Phase 9: 遗留任务 (从 V1 继承)

**Goal**: 完成 V1 未完成的 Polish 任务

**评估结果**: 大部分 V1 遗留任务已被 V2 新设计覆盖或废弃

- [-] T040 [P] ~~实现技能注册与版本匹配~~ → 废弃
  - 原因: V2 中 Agent 自己发现 skill，无需服务端注册
- [-] T041 [P] ~~实现技能权限校验与用户确认入口~~ → 替代为 T109-PERMISSION
  - 原因: 权限请求改为 ACP 协议中的 requestPermission RPC
- [-] T042 [P] ~~实现权限请求 UI 与确认反馈~~ → 替代为 T132-PERMISSION
  - 原因: 已在前端流式组件中实现
- [-] T047 [P] ~~实现知识库按需写入与去重策略~~ → 废弃
  - 原因: V2 不再有独立知识库，Agent 自己管理上下文
- [-] T049 ~~资源约束与超时控制~~ → 替代为 T107-TRANSPORT
  - 原因: 超时控制已在 Transport Handler 中实现

---

## Dependencies

```
Phase 1.1-1.4 (Session 后端) ───→ Phase 1.5-1.8 (Session 前端 + 上下文)
                                          │
Phase 2 (ActionBar) ──────────────────────┤
                                          ↓
                                   Phase 3 (文件-Session 关联)

Phase 4 (视图路由) ────→ Phase 5 (版本管理) ────→ Phase 6 (快捷操作)

Phase 7 (体验优化) - 独立，可并行
Phase 8 (审计日志) - 独立，可并行
Phase 9 (遗留任务) - 独立，可并行
```

---

## 实施分组与验收点

为便于逐步验收，将 Phase 1 拆分为更小的分组：

### Sprint 1: Session 数据层 + API (后端基础)

**任务**: T101, T102, T103, T104, T112, T113, T114

**验收标准**:
- Session CRUD API 可用 (REST + WebSocket)
- Session 数据持久化到 `.mindfs/sessions/`
- WebSocket 消息类型定义完整

---

### Sprint 2: Agent 进程管理 - ACP 协议重写 (后端核心)

**任务**: T106, T107-ACP, T107-TRANSPORT, T108, T109-ACP, T109-2, T109-3, T109-API, T109-PERMISSION, T109-RETRY, T109-SHUTDOWN

**变更说明**:
- T107, T109 废弃，重写为 ACP 协议版本
- 新增 Transport Handler 抽象层
- 新增权限请求处理
- 新增 Agent 状态 API
- 新增错误处理与重试
- 新增优雅关闭

**验收标准**:
- Agent 进程使用 ACP + ndJSON 协议通信
- 流式输出基于 SessionUpdate 解析
- 响应结束使用 idle 超时检测 (非 End Marker)
- Transport Handler 支持不同 Agent 配置
- 权限请求可推送到前端并等待响应
- `GET /api/agents` 返回所有 Agent 状态
- 进程崩溃自动重启，初始化失败自动重试
- 进程优雅关闭 (cancel → SIGTERM → SIGKILL)

---

### Sprint 3: 文件追踪 + 上下文构建 (后端完善)

**任务**: T110, T115, T116, T117, T118, T119, T120, T121

**变更说明**:
- T111 废弃 (文件操作通过 ACP tool_call 结构化传递)

**验收标准**:
- 文件创建自动关联到 Session
- 上下文构建逻辑完整 (客户端 + 服务端)
- Catalog/Schema/API 列表可导出

---

### Sprint 4: 目录 Skill + 设置 (后端扩展)

**任务**: T125, T126, T127, T128, T137

**验收标准**:
- Agent 工作目录模板生效
- Skill 执行 API 可用
- 目录配置 API 可用
- Session 空闲检测工作正常

---

### Sprint 5: Session 前端 (前端核心)

**任务**: T129, T130, T131, T132, T132-STREAM, T132-PERMISSION, T133, T134, T135, T135-HOOK, T136

**变更说明**:
- 新增 T132-STREAM: 流式消息渲染组件
- 新增 T132-PERMISSION: 权限请求对话框
- 新增 T135-HOOK: WebSocket 消息处理 Hook

**验收标准**:
- Session 列表显示正常
- 浮框/气泡交互正常
- 活跃 Session 点击展开浮框
- 已关闭 Session 点击显示历史
- Agent 可用性状态显示
- **新增**: 流式消息渲染正常 (文本/思考/工具调用)
- **新增**: 权限请求对话框正常工作

---

### Sprint 6: ActionBar 重构 (前端交互)

**任务**: T201, T202, T203, T204, T205, T206, T207, T208, T209, T210

**验收标准**:
- 三种模式切换正常
- Agent 选择下拉框正常
- 对话/技能/视图模式均可触发
- 错误处理和边界组件工作正常

---

### Sprint 7: 文件-Session 关联 (前后端联调)

**任务**: T301, T302, T303, T304, T305, T306

**验收标准**:
- 文件元数据记录正常
- 文件视图显示来源 Session
- 文件树显示关联标记
- Session 视图显示关联文件
- 双向跳转正常

**MVP 完成**: Sprint 1-7 完成后，用户可以与 Agent 对话，Agent 可以创建文件，文件和 Session 关联可见

---

### Sprint 8: 视图路由系统

**任务**: T401, T402, T403, T404, T405, T406, T407

**验收标准**:
- 路由配置解析正常
- 多视图匹配和切换正常
- 用户选择记忆正常

---

### Sprint 9: 视图版本管理

**任务**: T501, T502, T503, T504, T505

**验收标准**:
- 版本存储结构正确
- 版本切换 API 正常
- 重新生成对话框正常

---

### Sprint 10: 快捷操作

**任务**: T601, T602, T603, T604, T605

**验收标准**:
- 快捷操作定义解析正常
- 快捷操作栏渲染正常
- 快捷操作触发正常

---

### Sprint 11: 审计日志 + 错误处理

**任务**: T801, T802, T803, T804, T805

**验收标准**:
- 审计日志写入正常
- 各模块审计集成完成
- 错误码定义完整
- 错误恢复策略生效

---

### Sprint 12: 体验优化

**任务**: T701, T702, T703, T704, T705, T706, T707

**验收标准**:
- 任务状态追踪和推送正常
- Token 认证正常
- 移动端布局适配

---

### Sprint 13: 遗留任务

**任务**: 全部废弃或已被其他任务替代

**说明**:
- T040, T047 废弃 (V2 设计不再需要)
- T041, T042 替代为 T109-PERMISSION, T132-PERMISSION
- T049 替代为 T107-TRANSPORT

**验收标准**: 无 (Sprint 已废弃)

---

## Sprint 总览

| Sprint | 名称 | 任务数 | 优先级 | 依赖 | 状态 |
|--------|------|-------|-------|------|------|
| 1 | Session 数据层 + API | 7 | P0 | - | ✅ 完成 |
| 2 | Agent 进程管理 (ACP 重写) | 11 | P0 | Sprint 1 | 🔄 需重写 |
| 3 | 文件追踪 + 上下文 | 8 | P0 | Sprint 2 | ✅ 完成 |
| 4 | 目录 Skill + 设置 | 5 | P0 | Sprint 3 | ✅ 完成 |
| 5 | Session 前端 | 11 | P0 | Sprint 4 | ⏳ 待做 |
| 6 | ActionBar 重构 | 10 | P0 | Sprint 5 | ⏳ 待做 |
| 7 | 文件-Session 关联 | 6 | P0 | Sprint 6 | ⏳ 待做 |
| **MVP** | | **58** | | | |
| 8 | 视图路由系统 | 7 | P1 | MVP | ⏳ 待做 |
| 9 | 视图版本管理 | 5 | P1 | Sprint 8 | ⏳ 待做 |
| 10 | 快捷操作 | 5 | P1 | Sprint 9 | ⏳ 待做 |
| 11 | 审计日志 + 错误 | 5 | P1 | MVP | ⏳ 待做 |
| 12 | 体验优化 | 7 | P2 | MVP | ⏳ 待做 |
| 13 | 遗留任务 | 0 | - | - | ❌ 废弃 |

**变更说明**:
- Sprint 2: 任务数从 6 增加到 11 (新增 ACP 协议、API、重试、关闭相关任务)
- Sprint 3: 任务数从 9 减少到 8 (T111 废弃)
- Sprint 5: 任务数从 8 增加到 11 (新增流式组件和权限对话框)
- Sprint 13: 全部任务废弃或已被替代
- MVP 总任务数: 58
