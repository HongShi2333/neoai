<div align="center">

![neoai](/app/public/logo.png)

# 🚀 NeoAI

#### Next Generation AIGC One-Stop Business Solution — forked from [coai](https://github.com/coaidev/coai)

#### *NeoAI = [Next Web](https://github.com/ChatGPTNextWeb/ChatGPTNext-Web) + [One API](https://github.com/songquanpeng/one-api) + Community Channels*

</div>

---

## 📖 简介

NeoAI 是基于 coai 的二开版本，针对**前后端分离部署**、**社区频道（类 Discord）**、**新型缓存后端**做了若干改进。后端用 Go + Gin + MySQL 编写，前端是 React + Vite + TypeScript SPA。

### 与原版 coai 的差异

| 特性 | 原版 coai | NeoAI |
| --- | --- | --- |
| 缓存后端 | 仅 Redis | Redis / Valkey / Dragonfly 任选（`cache.type` 切换） |
| 社区频道 | ❌ | ✅ 类 Discord，可见性 + 发送权限双层控制，WebSocket 实时推送 |
| 渠道模型输入 | 单个添加 | 逗号分隔批量输入（new-api 风格） |
| 渠道自动获取模型 | ❌ | ✅ 调上游 `/v1/models` 一键拉取 |
| 定价设置 | 仅可视化 | 可视化 + JSON 批量编辑（replace / merge） |
| 用户名修改 | ❌ | 管理员可改任意用户 + 用户可在个人中心改自己 |
| Pro / License / 广告页面 | 含 | 已移除 |
| 渠道模型刷新 bug | 添加渠道后 `/v1/models` 不刷新 | 已修复（mutation 后立即 `Load()`） |
| 健康检查 | ❌ | ✅ `/healthz` + `/ready` |
| 分端部署 | 单一镜像 | 提供 `Dockerfile.backend` / `Dockerfile.frontend` / `docker-compose.split.yaml` |

完整变更说明见 [DEPLOYMENT.md](./DEPLOYMENT.md)。

---

## ✨ 功能特性

- 🤖 **70+ 模型集成**：OpenAI / Claude / Gemini / 通义千问 / 讯飞星火 / 智谱 / 月之暗面 / 混元 / 360 / 百川 / 火山方舟 / DeepSeek / Coze / Dify / Midjourney / Bing 等
- 🌐 **社区频道**：类 Discord 的文本频道，管理员可精细控制谁可见、谁可发消息
- 💰 **灵活计费**：按 Token / 按次 / 免计费三种模式，支持 JSON 批量导入导出
- 👥 **多用户管理**：管理员后台完整 CRUD、用户配额、订阅、API Key 管理
- 🔌 **OpenAI 兼容 API**：`/v1/chat/completions`、`/v1/models`、`/v1/images/generations`、`/v1/videos`
- 🛡️ **限流 + 防滥用**：基于 IP 的请求频率限制，黑名单机制
- 📨 **公告 / 广播系统**：管理员可推送站点公告
- 🐳 **多种部署模式**：一体化、前后端分离、纯二进制
- 🔄 **缓存后端可选**：Redis / Valkey / Dragonfly 一键切换

---

## 🚀 快速开始

### 一体化部署（最简单，适合个人 / 小团队）

```bash
git clone https://github.com/your-org/neoai.git
cd neoai
docker compose up -d
```

打开 `http://localhost:8000`，默认管理员账号 `root` / 密码 `chatnio123456`。

### 前后端分离部署（推荐生产环境）

详见 [DEPLOYMENT.md](./DEPLOYMENT.md) 的「部署模式 B」章节。简而言之：

- **后端**：`docker compose -f docker-compose.split.yaml up -d backend`，或在 VPS 上直接跑二进制
- **前端**：构建后部署到 Cloudflare Pages / Vercel / Netlify / 任何静态托管

---

## 📦 部署指南

### 后端部署

支持三种方式，任选其一：

#### 方式一：Docker（推荐）

```bash
# 一体化（前端 + 后端打包）
docker compose up -d

# 仅后端（分端部署）
docker compose -f docker-compose.split.yaml up -d backend
```

#### 方式二：源码编译

```bash
# 需要 Go 1.20+
go build -o neoai .
./neoai
```

#### 方式三：下载预编译二进制

从 [Releases](../../releases) 下载对应平台的二进制，直接运行。

#### 配置文件

后端启动时会自动从 `config.example.yaml` 复制一份到 `config/config.yaml`。关键配置：

```yaml
mysql:
  host: localhost
  port: 3306
  user: root
  password: secret
  db: neoai

cache:                    # 三选一：redis / valkey / dragonfly
  type: valkey
  host: localhost
  port: 6379
  db: 0
  password: ""

secret: "<至少 32 字节的随机字符串>"
serve_static: false       # 分端部署时设为 false
server:
  port: 8094
```

环境变量等价覆盖（适合 docker / k8s）：

```
MYSQL_HOST=...
CACHE_TYPE=valkey
CACHE_HOST=...
SERVE_STATIC=false
ALLOW_ORIGINS=https://your-frontend.pages.dev
```

健康检查：`GET /healthz`（始终 200）、`GET /ready`（DB + Cache 都可达才 200）。

---

### 前端部署（Cloudflare Pages）

NeoAI 前端是纯静态 SPA，可直接部署到 Cloudflare Pages。详细步骤见 [DEPLOYMENT.md](./DEPLOYMENT.md) 的「Cloudflare Pages 部署」章节。简版：

1. 在 Cloudflare Pages 上连接你的 GitHub 仓库
2. 构建命令：`cd app && npm install --legacy-peer-deps && npm run build`
3. 输出目录：`app/dist`
4. 环境变量：
   - `VITE_BACKEND_ENDPOINT` = `https://api.your-domain.com`（**必填**，后端公网地址）
   - `VITE_APP_NAME` = `NeoAI`（可选）
5. 部署后 SPA 路由已自动 fallback 到 `index.html`

---

## 🗂️ 项目结构

```
neoai/
├── main.go                 # Go 后端入口
├── go.mod / go.sum         # Go 依赖（module 名：neoai）
├── config.example.yaml     # 配置模板
├── Dockerfile              # 一体化镜像
├── Dockerfile.backend      # 仅后端镜像
├── Dockerfile.frontend     # 仅前端镜像（nginx + SPA fallback）
├── docker-compose.yaml     # 一体化部署
├── docker-compose.split.yaml  # 前后端分离部署
├── nginx.conf              # 反向代理示例
│
├── adapter/                # 各 LLM 适配器（openai/claude/gemini/...）
├── admin/                  # 管理员 API（用户、邀请、兑换码、日志、统计）
├── auth/                   # 鉴权、JWT、用户、配额、订阅
├── channel/                # 渠道管理、计费规则、模型列表
├── community/              # ✨ 新增：社区频道（类 Discord）
│   ├── types.go            #   数据模型
│   ├── migration.go        #   自动建表
│   ├── store.go            #   DB 操作 + 权限判断
│   └── router.go           #   REST + WebSocket 路由
├── middleware/             # CORS、限流、鉴权、健康检查
├── manager/                # 聊天、对话、relay、broadcast
├── connection/             # MySQL + 缓存连接管理
├── addition/               # 文章生成、图片生成、搜索
├── globals/                # 全局常量、变量、工具函数
├── utils/                  # 通用工具
├── cli/                    # 命令行子命令
│
├── app/                    # 前端项目（React + Vite + TS）
│   ├── package.json
│   ├── vite.config.ts
│   ├── .env.example        # 前端环境变量模板
│   ├── public/             # 静态资源（favicon、logo、icon）
│   └── src/
│       ├── api/            # API 客户端
│       │   └── community.ts  # ✨ 新增：社区频道 API
│       ├── routes/
│       │   └── community/
│       │       └── Community.tsx  # ✨ 新增：社区频道页面
│       ├── components/
│       │   └── admin/
│       │       └── JSONChargeDialog.tsx  # ✨ 新增：JSON 定价编辑器
│       ├── admin/          # 后台类型与 API
│       ├── store/          # Redux store
│       └── ...
│
└── DEPLOYMENT.md           # 详细部署文档
```

---

## 🔌 主要 API 速查

### 社区频道（新增）

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/community/channels` | 可选 | 列出当前用户可见的频道 |
| POST | `/community/channels` | 管理员 | 创建频道 |
| GET | `/community/channels/:id` | 可选 | 获取单个频道 |
| POST | `/community/channels/:id` | 管理员 | 更新频道 |
| DELETE | `/community/channels/:id` | 管理员 | 删除频道（级联删除消息） |
| GET | `/community/channels/:id/messages` | 可选 | 拉取最近消息（`?limit=200`） |
| POST | `/community/channels/:id/messages` | 已登录 | 发送消息 |
| POST | `/community/messages/:id` | 已登录 | 编辑自己的消息 |
| DELETE | `/community/messages/:id` | 已登录 | 删除消息（作者或管理员） |
| GET | `/community/ws` | 已登录 | WebSocket 实时推送 |

### 用户名修改（新增）

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| POST | `/admin/user/username` | 管理员 | 改任意用户的用户名 |
| POST | `/profile/username` | 已登录 | 改自己的用户名 |

### 渠道自动获取模型（新增）

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| POST | `/admin/channel/fetch-models` | 管理员 | 调上游 `/v1/models` 拉取模型列表 |

### 定价 JSON 批量编辑（新增）

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/admin/charge/json` | 管理员 | 导出当前计费规则为 JSON |
| POST | `/admin/charge/json?mode=replace\|merge` | 管理员 | 批量导入计费规则 |

### 健康检查（新增）

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/healthz` | 公开 | 进程存活探针 |
| GET | `/ready` | 公开 | 就绪探针（DB + Cache 都可达才 200） |

---

## 🛠️ 开发

### 后端开发

```bash
go run .                    # 直接运行
go run . --debug            # 调试模式（刷新缓存、详细日志）
go build ./... && go vet ./...  # 类型检查
```

CLI 子命令：

```bash
./neoai --admin --username foo --password bar  # 创建管理员
./neoai --invite --number 10 --quota 100        # 生成邀请码
./neoai --token --username foo                  # 生成 JWT
```

### 前端开发

```bash
cd app
cp .env.example .env
# 编辑 .env，将 VITE_BACKEND_ENDPOINT 设为后端地址（默认 http://localhost:8094）
npm install --legacy-peer-deps
npm run dev       # 开发模式（热更新，默认 5173）
npm run build     # 生产构建
```

### 国际化

支持简中、繁中、英文、日文、俄文。新增的社区频道字符串已添加到 `app/src/resources/i18n/*.json` 的 `community` 段。

---

## 🔒 安全建议

1. **生产环境务必修改 `secret`**，至少 32 字节随机字符串（`openssl rand -hex 32`）
2. **管理员密码立即修改**：登录后到「账户」页改密码，或 CLI `./neoai --admin --username root --password <新密码>`
3. **CORS 白名单**：分端部署时设置 `ALLOW_ORIGINS=https://your-frontend.pages.dev`
4. **数据库密码**：docker-compose 默认 `neoai123456!` 仅作示例，生产环境请改强密码
5. **HTTPS**：前端走 Cloudflare Pages 自动 HTTPS，后端建议套 Caddy / nginx + Let's Encrypt

---

## 📄 License

继承自原项目 coai 的 [LICENSE](./LICENSE)（Apache-2.0）。

## 🙏 鸣谢

- [coai](https://github.com/coaidev/coai) — 原项目
- [QuantumNous/new-api](https://github.com/QuantumNous/new-api) — 渠道编辑器与 JSON 定价风格参考
- [Discord](https://discord.com) — 社区频道设计灵感
