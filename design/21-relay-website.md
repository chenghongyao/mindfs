# Relay 网站设计

## 1. 文档边界

本文只定义 Relay 网站的产品形态和对外接口边界，不把尚未落地的 MindFS CLI 语法写死到文档里。

和当前仓库保持一致的约束：

- 现有 `mindfs` CLI 没有 `start` 子命令，因此本文不再出现 `mindfs start ...`
- Relay 模式仍处于设计阶段，网站接口只返回激活所需的数据，不返回硬编码命令字符串
- 前端访问节点时统一使用 `base_url`，与 [21-relay-mode.md](/Users/bixin/project/mindfs/design/21-relay-mode.md) 保持一致

---

## 2. 设计原则

- **极简**：主流程只有“创建节点、激活节点、打开节点、删除节点”
- **快速访问**：用户登录后直接进入节点列表
- **零配置默认值**：创建时自动生成名称
- **不把 CLI 细节耦合进网站**：网站展示激活码，不承担定义 CLI 语法的职责
- **状态简单**：节点状态只保留 `online` 和 `offline`
- **登录方式清晰**：同时支持邮箱登录、Google OAuth、GitHub OAuth

---

## 3. 首页

未登录用户先进入首页，首页提供产品说明和登录入口。

### 首页布局

```text
┌─────────────────────────────────────────────┐
│ MindFS Relay                     [登录]      │
├─────────────────────────────────────────────┤
│                                             │
│ 在浏览器中访问你的 MindFS                    │
│                                             │
│ [使用 Google 登录]                          │
│ [使用 GitHub 登录]                          │
│ [邮箱登录]                                  │
│                                             │
│ 说明：登录后可创建节点、复制激活码、打开节点。 │
└─────────────────────────────────────────────┘
```

### 首页职责

- 介绍 Relay 的用途
- 提供登录入口
- 对已登录用户可直接跳转到节点列表页

---

## 4. 登录页

Relay 网站需要一个独立登录页，支持 3 种认证方式：

- 邮箱 + 密码
- Google OAuth
- GitHub OAuth

### 登录页布局

```text
┌─────────────────────────────────────────────┐
│ MindFS Relay                                │
│                                             │
│ [使用 Google 登录]                          │
│ [使用 GitHub 登录]                          │
│                                             │
│ ───────────── 或使用邮箱 ─────────────       │
│                                             │
│ 邮箱                                        │
│ [____________________________]              │
│ 密码                                        │
│ [____________________________]              │
│                                             │
│ [登录]                                      │
└─────────────────────────────────────────────┘
```

### 约束

- 第一次使用 Google / GitHub 登录时，系统自动创建用户
- 同一邮箱的账号合并策略需要单独设计，本文先不展开
- 登录完成后统一回到节点列表页

---

## 5. 主界面

这是用户登录后的唯一主页面。

### 布局

```text
┌─────────────────────────────────────────────────────┐
│ [Logo] MindFS Relay                    [用户头像] [登出] │
├─────────────────────────────────────────────────────┤
│                                                     │
│  我的节点                    [部署指引] [+ 创建节点]  │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │ Online  home-server                         │  │
│  │         在线 2 小时                            │  │
│  │         https://relay.mindfs.com/n/node_abc  │  │
│  │         [打开] [复制链接] [删除]              │  │
│  └──────────────────────────────────────────────┘  │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │ Offline work-laptop                        │  │
│  │         离线 3 天前                            │  │
│  │         https://relay.mindfs.com/n/node_xyz  │  │
│  │         [获取新激活码] [删除]                │  │
│  └──────────────────────────────────────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### 节点卡片信息

每个节点显示：

- **状态**：`online | offline`
- **节点名称**：可直接重命名
- **最后在线信息**：`在线 2 小时` 或 `离线 3 天前`
- **base_url**：例如 `https://relay.mindfs.com/n/node_abc123`（第一阶段单一域名）
- **操作按钮**：
  - `online`：`[打开] [复制链接] [删除]`
  - `offline`：`[获取新激活码] [删除]`

### 状态定义

- `offline`：没有在线 connector
- `online`：节点当前可通过 relay 打开

### 重命名交互

点击节点名称进入编辑模式：

```text
┌──────────────────────────────────────────────┐
│ Online  [home-server___]                    │
│         在线 2 小时                            │
│         https://relay.mindfs.com/n/node_abc  │
│         [打开] [复制链接] [删除]              │
└──────────────────────────────────────────────┘
```

- 点击名称进入输入态
- 失焦自动保存
- `Enter` 保存
- `Esc` 取消

---

## 6. 创建节点

点击 `[+ 创建节点]` 后，系统直接创建节点并返回一个一次性激活码。

### 创建逻辑

- 自动生成节点名称：`node-1`、`node-2`、`node-3`
- 生成一次性 `activation_token`
- 激活码默认 24 小时过期
- 过期后尝试激活会返回明确错误，用户需要在网站上重新获取激活码

### 创建完成弹窗

```text
┌─────────────────────────────────────────────┐
│  节点已创建                                  │
├─────────────────────────────────────────────┤
│                                             │
│  节点名称：node-3                            │
│                                             │
│  激活码：                                    │
│  ┌─────────────────────────────────────┐   │
│  │ act_abc123def456...                 │   │
│  └─────────────────────────────────────┘   │
│                     [复制激活码] [完成]     │
│                                             │
│  说明：在设备上启动 MindFS 时传入该激活码。  │
│  具体 CLI 语法以实际实现为准。              │
│                                             │
└─────────────────────────────────────────────┘
```

这里不展示固定命令字符串，原因：

- 当前仓库里的 `mindfs` CLI 并不存在 `start` 子命令
- Relay/connector 还未实现，命令格式不是稳定接口
- 网站 API 的职责是发放激活数据，不是拼接 CLI 文本

---

## 7. 重新激活

当节点离线，或用户需要在新设备上重新绑定节点时，点击 `[获取新激活码]`。

### 重新激活弹窗

```text
┌─────────────────────────────────────────────┐
│  获取新激活码                                │
├─────────────────────────────────────────────┤
│                                             │
│  节点名称：work-laptop                       │
│                                             │
│  新激活码：                                  │
│  ┌─────────────────────────────────────┐   │
│  │ act_new_xyz789...                   │   │
│  └─────────────────────────────────────┘   │
│                     [复制激活码] [关闭]     │
│                                             │
│  说明：旧激活码立即失效；已保存的设备凭证    │
│  (device_token) 也会被吊销。重新激活后需要  │
│  在设备上使用新激活码重新绑定。              │
└─────────────────────────────────────────────┘
```

### 明确约束

- 重新激活的语义是”签发新的 `activation_token` 并吊销旧的 `device_token`”
- 它不等价于删除并重建节点
- 它不会改变 `node_id` 或 `base_url`
- 第一阶段约束：一个节点同时只能有一个活跃的 device_token
- 重新激活后，设备需要使用新激活码重新绑定

---

## 8. 部署指引页

点击 `[部署指引]`，跳转到独立页面。

```text
┌─────────────────────────────────────────────────────┐
│ ← 返回                                               │
├─────────────────────────────────────────────────────┤
│                                                     │
│  连接设备到 MindFS Relay                            │
│                                                     │
│  1. 安装 MindFS                                     │
│                                                     │
│  2. 在节点列表页点击 [+ 创建节点]                    │
│                                                     │
│  3. 复制激活码                                       │
│                                                     │
│  4. 在设备上启动 MindFS，并传入激活码                │
│     具体命令格式以实际 CLI 文档为准                 │
│                                                     │
│  5. 回到节点列表，等待节点进入 Online                │
│                                                     │
│  常见问题                                            │
│  • 如何在多台设备上使用？                            │
│    每台设备创建一个节点                              │
│  • 节点为什么还是 Offline？                         │
│    说明该设备还没有成功连上 relay                    │
│  • 重装系统后如何恢复？                              │
│    点击 [获取新激活码] 重新绑定                      │
└─────────────────────────────────────────────────────┘
```

---

## 9. 用户流程

### 首次激活

1. 用户登录 `relay.mindfs.com`
2. 点击 `[+ 创建节点]`
3. 复制 `activation_token`
4. 在设备上启动 MindFS，并传入该激活码
5. 设备连上 relay 后，节点状态变为 `online`
6. 用户点击 `[打开]` 访问该节点

### 后续启动

Connector 会将 device_token 持久化到本地配置文件。后续启动时：

1. Connector 自动读取已保存的 device_token
2. 使用该 token 连接 relay
3. 只有在 token 失效时才需要重新激活

用户无需每次启动都输入激活码。

### 重新绑定

1. 用户点击 `[获取新激活码]`
2. 网站签发新的 `activation_token`
3. 用户在设备上重新完成绑定
4. 节点恢复到 `online`

---

## 10. API 草案

这里只定义 Relay 网站需要的 control plane API，不把 CLI 文本当成 API 返回值。

### 认证

用途：邮箱密码登录。

```text
POST /api/auth/login
```

```json
{
  "email": "user@example.com",
  "password": "password"
}
```

```json
{
  "user_id": "user_123",
  "email": "user@example.com"
}
```

用途：发起 Google OAuth 登录。

```text
GET /api/auth/google
```

用途：处理 Google OAuth 回调，登录成功后建立会话。

```text
GET /api/auth/google/callback
```

用途：发起 GitHub OAuth 登录。

```text
GET /api/auth/github
```

用途：处理 GitHub OAuth 回调，登录成功后建立会话。

```text
GET /api/auth/github/callback
```

### 节点管理

用途：获取当前用户的节点列表，用于主页面首屏展示和轮询刷新。

```text
GET /api/nodes
```

```json
[
  {
    "id": "node_abc123",
    "name": "home-server",
    "status": "online",
    "last_seen_at": "2026-03-17T14:30:00Z",
    "base_url": "https://relay.mindfs.com/n/node_abc123",
    "created_at": "2026-03-10T08:00:00Z"
  }
]
```

用途：创建一个新节点，并立即返回首次绑定所需的激活码。

```text
POST /api/nodes
```

```json
{
  "id": "node_abc123",
  "name": "node-1",
  "status": "offline",
  "base_url": "https://relay.mindfs.com/n/node_abc123",
  "activation_token": "act_abc123def456",
  "activation_expires_at": "2026-03-18T14:30:00Z"
}
```

用途：修改节点名称。

```text
PATCH /api/nodes/:id
```

```json
{
  "name": "my-home-server"
}
```

```json
{
  "success": true
}
```

用途：给某个节点重新签发激活码，用于换设备或重新绑定。

```text
POST /api/nodes/:id/activation-tokens
```

```json
{
  "activation_token": "act_new_xyz789",
  "activation_expires_at": "2026-03-18T15:00:00Z"
}
```

用途：删除节点及其相关绑定。

```text
DELETE /api/nodes/:id
```

```json
{
  "success": true
}
```

### 首次激活

这个接口属于 connector/control-plane 交互，不是网站前端直接调用的核心路径，但网站设计需要知道它的存在。

用途：设备首次用一次性激活码换长期设备凭证，并拿到 relay 接入地址。

```text
POST /api/activate
```

```json
{
  "activation_token": "act_abc123def456"
}
```

```json
{
  "device_token": "dev_long_term_token",
  "node_id": "node_abc123",
  "endpoint": "wss://relay.mindfs.com/ws/connector"
}
```

---

## 11. 数据模型建议

网站只需要围绕“节点”和“激活码”做最小数据模型。

```sql
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT,
  google_id TEXT UNIQUE,
  github_id TEXT UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE nodes (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'offline',
  last_seen_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, name)
);

CREATE TABLE activation_tokens (
  token TEXT PRIMARY KEY,
  node_id TEXT NOT NULL REFERENCES nodes(id),
  expires_at TIMESTAMP NOT NULL,
  used_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

第一阶段明确约束：

- 一个节点同时只能有一个活跃的 device_token
- 重新激活会吊销旧 token 并签发新 token
- Device token 需要单独的表来管理：

```sql
CREATE TABLE device_tokens (
  token TEXT PRIMARY KEY,
  node_id TEXT NOT NULL REFERENCES nodes(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  revoked_at TIMESTAMP,
  UNIQUE(node_id)  -- 确保一个节点只有一个活跃 token
);
```

---

## 12. Relay 实现草案

根据 [21-relay-mode.md](/Users/bixin/project/mindfs/design/21-relay-mode.md) 的后端方案，relay 不放进现有 `mindfs` 项目里混写，而是单独一个项目目录：

```text
relay/
  web/
  server/
```

这样做的原因：

- relay 是独立部署的公网服务，不应和本地 `mindfs` server 代码耦合
- relay 有自己独立的认证、节点管理、网关和路由职责
- 前端站点和后端服务都要独立演进

### 12.1 三方模块图

```text
┌──────────────────────── 用户 ────────────────────────┐
│                                                      │
│  Browser                                              │
│    ├─ 首页 / 登录页                                   │
│    ├─ 节点列表页                                      │
│    └─ 打开节点后的 Web UI                             │
│                                                      │
└──────────────────────────────────────────────────────┘
                         │
                         │ HTTPS / WSS
                         ▼
┌──────────────────────── Relay ───────────────────────┐
│                                                      │
│  Caddy                                                │
│    ├─ /                -> relay/web                   │
│    ├─ /api/*           -> relay/server/controlplane  │
│    ├─ /n/{node_id}/*   -> relay/server/gateway       │
│    ├─ /ws/connector    -> relay/server/router        │
│                                                      │
│  relay/web                                             │
│    ├─ 首页                                            │
│    ├─ 登录页                                          │
│    ├─ 节点列表页                                      │
│    └─ guide 页                                        │
│                                                      │
│  relay/server                                          │
│    ├─ auth         -> session cookie / 节点授权       │
│    ├─ oauth        -> Google / GitHub OAuth          │
│    ├─ controlplane -> 用户/节点/激活码 API            │
│    ├─ gateway      -> /n/{node_id}/api/* 和 /ws      │
│    ├─ router       -> connector 长连接 / stream 路由  │
│    └─ store        -> DB                              │
│                                                      │
└──────────────────────────────────────────────────────┘
                         │
                         │ WSS + yamux
                         ▼
┌────────────────────── 用户设备 ──────────────────────┐
│                                                      │
│  mindfs-connector                                     │
│    ├─ activation / device_token                       │
│    ├─ 心跳 / 重连                                     │
│    └─ 转发 relay 请求到本地 MindFS                    │
│                                                      │
│  local mindfs server                                  │
│    ├─ /api/*                                          │
│    └─ /ws                                             │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 12.2 三方调用关系图

#### 登录和节点管理

```text
Browser
  -> Caddy
  -> relay/web
  -> relay/server/controlplane
     -> auth
     -> oauth
     -> store
```

#### 打开节点后的业务请求

```text
Browser
  -> Caddy
  -> relay/server/gateway
     -> auth
     -> router
        -> mindfs-connector
           -> local mindfs server
```

#### connector 接入 relay

```text
mindfs-connector
  -> Caddy
  -> relay/server/router
     -> store
```

### 12.3 项目结构

```text
relay/
  web/
    src/
      pages/
        login/
        nodes/
        guide/
      components/
      services/
        auth.ts
        nodes.ts
      router/
      styles/
    package.json

  server/
    cmd/
      relay-server/
        main.go
    internal/
      auth/
      controlplane/
      gateway/
      router/
      store/
      oauth/
    migrations/
    go.mod
```

### 12.4 `relay/web` 职责

`relay/web` 只负责浏览器端体验：

- 首页
- 登录页
- 节点列表页
- 部署指引页
- 调用 control plane API
- 打开 `base_url` 访问节点

前端不负责：

- 解析 connector 长连接
- 处理 node 到 connector 的路由
- 处理 `/n/{node_id}/ws` 转发

### 12.5 `relay/server` 职责

`relay/server` 是单独的 Go 服务，但内部按职责拆成 3 层，对应 21 号文档里的方案。

#### `controlplane`

负责：

- 用户认证
- OAuth 登录
- 节点 CRUD
- 激活码签发
- 设备首次激活
- 节点在线状态记录

建议 API：

- `/api/auth/*`
- `/api/nodes`
- `/api/nodes/:id`
- `/api/nodes/:id/activation-tokens`
- `/api/activate`

#### `gateway`

负责浏览器流量入口：

- 处理 `/n/{node_id}/api/*`
- 处理 `/n/{node_id}/ws`
- 校验当前用户是否有权访问该节点
- 把外部 HTTP / WebSocket 会话转交给 `router`

这里不直接执行业务逻辑，只做入口鉴权和协议转换。

#### `router`

负责设备连接和会话路由：

- 维护 `node_id -> connector session` 映射
- 接受 connector 的长期连接
- 在一条连接上复用多个 HTTP / WS stream
- 处理超时、断线、清理

第一阶段按 21 号文档建议：

- 外层协议：`WSS`
- 多路复用：`yamux`
- 由 connector 主动连入 relay

### 12.6 服务端内部目录建议

```text
relay/server/internal/
  auth/           # session、cookie、邮箱密码登录
  oauth/          # Google / GitHub OAuth
  controlplane/   # 节点、激活码、用户接口
  gateway/        # /n/{node_id}/api/* 和 /n/{node_id}/ws 入口
  router/         # connector 连接池、stream 路由、心跳与清理
  store/          # users/nodes/activation_tokens/device_tokens
```

### 12.7 小模块接口草案

下面的接口不是最终代码签名，而是模块边界草案。

#### `auth`

职责：

- 邮箱密码登录
- session cookie 签发和校验
- 当前用户提取
- 节点访问权限校验

核心接口：

```go
type AuthService interface {
    Login(ctx context.Context, email, password string) (User, error)
    LoadUser(ctx context.Context, sessionToken string) (User, error)
    CreateSession(ctx context.Context, userID string) (sessionToken string, err error)
    RequireNodeAccess(ctx context.Context, userID, nodeID string) error
}
```

对外被谁调用：

- `controlplane` 调用 `Login`、`CreateSession`
- `gateway` 调用 `LoadUser`、`RequireNodeAccess`

#### `oauth`

职责：

- 生成 Google / GitHub OAuth 跳转地址
- 处理回调 code
- 查找或创建本地用户

核心接口：

```go
type OAuthService interface {
    StartGoogle(ctx context.Context) (redirectURL string, err error)
    FinishGoogle(ctx context.Context, code string) (User, error)
    StartGitHub(ctx context.Context) (redirectURL string, err error)
    FinishGitHub(ctx context.Context, code string) (User, error)
}
```

对外被谁调用：

- `controlplane` 的认证 handler 调用
- 成功后再调用 `auth.CreateSession`

#### `controlplane`

职责：

- 用户认证 API
- 节点列表、创建、重命名、删除
- 激活码签发
- 首次激活换长期 `device_token`

核心接口：

```go
type ControlPlaneService interface {
    ListNodes(ctx context.Context, userID string) ([]Node, error)
    CreateNode(ctx context.Context, userID string) (NodeWithActivation, error)
    RenameNode(ctx context.Context, userID, nodeID, name string) error
    DeleteNode(ctx context.Context, userID, nodeID string) error
    IssueActivationToken(ctx context.Context, userID, nodeID string) (ActivationToken, error)
    ActivateNode(ctx context.Context, activationToken string) (DeviceBinding, error)
}
```

对外被谁调用：

- HTTP API handler 直接调用

内部依赖：

- `auth`
- `oauth`
- `store`
- `router` 的在线状态查询能力

#### `gateway`

职责：

- 处理浏览器到 relay 的业务访问入口
- 解析 `node_id`
- 鉴权
- 把外部 HTTP/WS 转给 `router`

核心接口：

```go
type Gateway interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
    ServeWebSocket(w http.ResponseWriter, r *http.Request)
}
```

内部依赖：

- `auth`
- `router`

它不依赖：

- `oauth`
- 数据库读写细节

#### `router`

职责：

- 管理 connector 长连接
- 维护 `node_id -> connector session`
- 代理 HTTP 请求
- 代理 WebSocket 会话
- 心跳、断线、超时清理

核心接口：

```go
type Router interface {
    RoundTripHTTP(ctx context.Context, req HTTPRequest) (*HTTPResponse, error)
    OpenWebSocket(ctx context.Context, req WSRequest) (WSBridge, error)
    RegisterConnector(ctx context.Context, conn ConnectorConn) error
    MarkNodeOnline(ctx context.Context, nodeID string)
    MarkNodeOffline(ctx context.Context, nodeID string)
    GetNodeStatus(ctx context.Context, nodeID string) (string, error)
}
```

内部依赖：

- `store` 用于持久化节点在线状态

#### `store`

职责：

- 数据库存取
- 不承载业务流程编排

核心接口：

```go
type Store interface {
    GetUserByEmail(ctx context.Context, email string) (User, error)
    GetOrCreateOAuthUser(ctx context.Context, provider, providerUserID, email string) (User, error)

    ListNodesByUser(ctx context.Context, userID string) ([]Node, error)
    CreateNode(ctx context.Context, userID, name string) (Node, error)
    UpdateNodeName(ctx context.Context, userID, nodeID, name string) error
    DeleteNode(ctx context.Context, userID, nodeID string) error
    SetNodeStatus(ctx context.Context, nodeID, status string, lastSeenAt time.Time) error

    CreateActivationToken(ctx context.Context, nodeID string, expiresAt time.Time) (ActivationToken, error)
    ConsumeActivationToken(ctx context.Context, token string) (Node, error)
    SaveDeviceToken(ctx context.Context, nodeID, deviceToken string) error
}
```

### 12.8 模块调用关系

#### 首页进入流程

未登录用户：

1. 浏览器访问 `/`
2. `relay/web` 展示首页
3. 用户点击首页上的登录入口
4. 进入登录页或直接跳第三方 OAuth

已登录用户：

1. 浏览器访问 `/`
2. `relay/web` 检查当前登录态
3. 已登录则直接跳转到节点列表页

#### 登录流程

邮箱登录：

1. HTTP handler -> `controlplane`
2. `controlplane` -> `auth.Login`
3. `controlplane` -> `auth.CreateSession`
4. handler 写 session cookie

Google / GitHub 登录：

1. HTTP handler -> `oauth.StartGoogle` / `oauth.StartGitHub`
2. 浏览器跳第三方授权页
3. OAuth callback handler -> `oauth.FinishGoogle` / `oauth.FinishGitHub`
4. callback handler -> `auth.CreateSession`
5. handler 写 session cookie

#### 节点管理流程

列出节点：

1. handler -> `auth.LoadUser`
2. handler -> `controlplane.ListNodes`
3. `controlplane` -> `store.ListNodesByUser`
4. `controlplane` 可选 -> `router.GetNodeStatus`

创建节点：

1. handler -> `auth.LoadUser`
2. handler -> `controlplane.CreateNode`
3. `controlplane` -> `store.CreateNode`
4. `controlplane` -> `store.CreateActivationToken`
5. 返回节点和激活码

重新签发激活码：

1. handler -> `auth.LoadUser`
2. handler -> `controlplane.IssueActivationToken`
3. `controlplane` -> `store.CreateActivationToken`

#### 设备首次激活流程

1. connector 请求 `POST /api/activate`
2. handler -> `controlplane.ActivateNode`
3. `controlplane` -> `store.ConsumeActivationToken`
4. `controlplane` -> `store.SaveDeviceToken`
5. 返回 `device_token` 和 relay endpoint

#### 浏览器访问节点流程

普通 HTTP：

1. 浏览器访问 `/n/{node_id}/api/...`
2. `gateway` -> `auth.LoadUser`
3. `gateway` -> `auth.RequireNodeAccess`
4. `gateway` 解析原生路径，去掉 `/n/{node_id}`
5. `gateway` -> `router.RoundTripHTTP`
6. `router` 找到对应 connector session
7. `router` 在 WSS/yamux 上开 stream
8. connector 转发给本地 MindFS
9. 响应回到 `gateway`
10. `gateway` 回写浏览器

WebSocket：

1. 浏览器访问 `/n/{node_id}/ws`
2. `gateway` 完成鉴权和 upgrade
3. `gateway` -> `router.OpenWebSocket`
4. `router` 打开内部 ws stream
5. 浏览器 frame <-> gateway <-> router <-> connector 双向转发

#### connector 长连接流程

1. connector 连到 `wss://relay.mindfs.com/ws/connector`
2. connector handler -> `router.RegisterConnector`
3. `router` 校验 `device_token`
4. `router` 建立 `node_id -> connector session`
5. `router` 定时收心跳并更新在线状态
6. 断线时 `router` 标记节点 `offline`

#### 用户请求鉴权链路

- 浏览器访问 relay 时，使用登录后的 session cookie 标识用户
- `controlplane` 和 `gateway` 先通过 `auth.LoadUser` 还原当前用户
- 如果是 `/n/{node_id}/api/*` 或 `/n/{node_id}/ws`，`gateway` 还要调用 `auth.RequireNodeAccess(userID, nodeID)`
- 只有在“用户已登录”且“该节点属于该用户”时，`gateway` 才会调用 `router`
- `router` 不负责浏览器用户鉴权，它只负责把请求转发到对应 connector
- connector 接入 relay 时使用的是 `device_token`，这套凭证只用于设备身份，不用于浏览器用户鉴权

### 12.9 启动方式

第一阶段建议最小部署形态：

1. `relay/web`
   - 构建成静态资源
2. `relay/server`
   - 提供 control plane API
   - 提供 connector 接入点
   - 提供 `/n/{node_id}` 网关入口
3. `Caddy`
   - 放在最前面
   - 负责 TLS、域名和反向代理

建议流量入口：

- `https://relay.mindfs.com/` -> `relay/web`
- `https://relay.mindfs.com/api/*` -> `relay/server` control plane
- `https://relay.mindfs.com/n/{node_id}/*` -> `relay/server` gateway
- `wss://relay.mindfs.com/ws/connector` -> `relay/server` connector/router

### 12.10 与当前 `mindfs` 仓库的关系

- `mindfs` 负责本地 server 和 connector（connector 作为 MindFS 的内置模块）
- `relay` 作为单独项目目录维护
- 两边通过协议交互，不共享运行时状态
- Connector 在 MindFS 启动时自动启动，将 device_token 持久化到本地配置文件
- 可以共享少量协议文档，但不要把 relay 的数据库和 Web 逻辑塞进 `mindfs/server`

---

## 13. 推荐开源依赖

下面是第一阶段比较合适的 Go 侧依赖选择。

| 模块 | 推荐依赖 | 用途 |
| --- | --- | --- |
| OAuth | `golang.org/x/oauth2` | Google / GitHub OAuth 登录 |
| WebSocket | `github.com/gorilla/websocket` | 浏览器侧和 connector 侧 WebSocket 处理 |
| 多路复用 | `github.com/hashicorp/yamux` | 在 connector 长连接上复用多个 HTTP / WS stream |
| HTTP 路由 | `github.com/go-chi/chi/v5` | `relay/server` 的路由和中间件 |
| 数据库访问 | `github.com/jackc/pgx/v5` | PostgreSQL 访问 |
| SQL 构建 | `github.com/Masterminds/squirrel` | 简单拼装 SQL，避免手写大量字符串 |
| Session | `github.com/alexedwards/scs/v2` | session 生命周期和 cookie 管理 |
| 密码哈希 | `golang.org/x/crypto/bcrypt` | 邮箱密码登录 |

说明：

- `Caddy` 建议以前置代理方式部署，不作为 Go SDK 集成
- GitHub OAuth 不必单独找专用 SDK，直接用 `x/oauth2` 即可
- 如果后面不想引入 `squirrel`，也可以直接用手写 SQL + `pgx`

---

## 14. 实现提示

- 登录态可以统一走 session cookie
- OAuth 回调成功后和邮箱登录一样落到同一套 session 体系
- 前端请求节点时统一使用 `base_url`
- 不把 `node_id` 暴露进 MindFS 自身业务 API
- 网站轮询或订阅的状态字段只用 `online | offline`
- 网站 API 返回结构里不要包含 `command`
- 部署指引页引用 CLI 文档，而不是复制一份可能过期的命令格式
