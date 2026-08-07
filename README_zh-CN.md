<div align="center">

![neoai](/app/public/logo.png)

# 🚀 NeoAI

#### 新一代 AIGC 一站式业务解决方案 — 基于 [coai](https://github.com/coaidev/coai) 二开

#### *NeoAI = [Next Web](https://github.com/ChatGPTNextWeb/ChatGPTNext-Web) + [One API](https://github.com/songquanpeng/one-api) + 社区频道*

</div>

---

## 📖 简介

NeoAI 是 coai 的二开版本，针对**前后端分离部署**、**社区频道（类 Discord）**、**新型缓存后端**做了若干改进。后端 Go + Gin + MySQL，前端 React + Vite + TypeScript SPA。

### 与原版 coai 的差异

| 特性 | 原版 coai | NeoAI |
| --- | --- | --- |
| 缓存后端 | 仅 Redis | Redis / Valkey / Dragonfly 任选 |
| 社区频道 | ❌ | ✅ 类 Discord，可见性 + 发送权限双层控制 + WebSocket |
| 渠道模型输入 | 单个添加 | 逗号分隔批量输入（new-api 风格） |
| 渠道自动获取模型 | ❌ | ✅ 调上游 `/v1/models` 一键拉取 |
| 定价设置 | 仅可视化 | 可视化 + JSON 批量编辑（replace / merge） |
| 用户名修改 | ❌ | 管理员 + 用户自助 |
| Pro / License / 广告页面 | 含 | 已移除 |
| 渠道模型刷新 bug | 添加渠道后不刷新 | 已修复 |
| 健康检查 | ❌ | ✅ `/healthz` + `/ready` |
| 分端部署 | 单一镜像 | 提供 `Dockerfile.backend` / `Dockerfile.frontend` / `docker-compose.split.yaml` |

完整变更说明见 [DEPLOYMENT.md](./DEPLOYMENT.md)。

---

## 🚀 快速开始

### 一体化部署

```bash
git clone https://github.com/your-org/neoai.git
cd neoai
docker compose up -d
```

打开 `http://localhost:8000`，默认管理员 `root` / `chatnio123456`。

### 前后端分离部署（推荐生产）

- **后端**：`docker compose -f docker-compose.split.yaml up -d backend`，或直接跑二进制
- **前端**：构建后部署到 Cloudflare Pages / Vercel / Netlify

详细步骤见 [DEPLOYMENT.md](./DEPLOYMENT.md)。

---

## 📦 后端部署

### 方式一：Docker（推荐）

```bash
docker compose up -d                          # 一体化
docker compose -f docker-compose.split.yaml up -d backend   # 仅后端
```

### 方式二：源码编译

```bash
# 需要 Go 1.20+
go build -o neoai .
./neoai
```

### 配置文件

后端启动时自动从 `config.example.yaml` 复制到 `config/config.yaml`。关键配置：

```yaml
mysql:
  host: localhost
  port: 3306
  user: root
  password: secret
  db: neoai

cache:                    # redis / valkey / dragonfly 三选一
  type: valkey
  host: localhost
  port: 6379
  db: 0
  password: ""

secret: "<至少 32 字节随机字符串>"
serve_static: false       # 分端部署时设为 false
server:
  port: 8094
```

环境变量覆盖（适合 docker / k8s）：

```
MYSQL_HOST=...
CACHE_TYPE=valkey
CACHE_HOST=...
SERVE_STATIC=false
ALLOW_ORIGINS=https://your-frontend.pages.dev
```

健康检查：`GET /healthz`（始终 200）、`GET /ready`（DB + Cache 都可达才 200）。

---

## ☁️ 前端部署（Cloudflare Pages）

NeoAI 前端是纯静态 SPA，可直接部署到 Cloudflare Pages。

### 步骤一：Fork / 上传仓库到 GitHub

### 步骤二：在 Cloudflare Pages 创建项目

1. 登录 [Cloudflare Dashboard](https://dash.cloudflare.com/) → Workers & Pages → Create application → Pages → Connect to Git
2. 选择你的 GitHub 仓库
3. 配置构建设置：
   - **Framework preset**: `None`
   - **Build command**: `cd app && npm install --legacy-peer-deps && npm run build`
   - **Build output directory**: `app/dist`
   - **Root directory**: `/`（仓库根）
4. 添加环境变量（Settings → Environment variables）：
   - `VITE_BACKEND_ENDPOINT` = `https://api.your-domain.com`（**必填**，后端公网地址）
   - `VITE_APP_NAME` = `NeoAI`（可选）
   - `VITE_APP_LOGO` = `/favicon.ico`（可选）
5. Save and Deploy

### 步骤三：配置 SPA 路由 fallback

Cloudflare Pages 会自动识别 `dist/index.html` 作为 SPA 入口，刷新子路由不会 404。若仍遇到 404，在仓库根添加 `public/_redirects` 文件：

```
/*    /index.html   200
```

并在 `app/vite.config.ts` 中确保 `public` 目录被复制（vite 默认行为）。

### 步骤四：自定义域名

在 Cloudflare Pages 项目设置 → Custom domains 中绑定你自己的域名（如 `app.your-domain.com`），Cloudflare 会自动签发 HTTPS 证书。

### 后端 CORS 配置

由于前端跑在 `*.pages.dev` 或自定义域名，后端必须放行该来源：

```yaml
# config.yaml
allow_origins:
  - https://app.your-domain.com
  - https://your-project.pages.dev
```

或环境变量：

```
ALLOW_ORIGINS=https://app.your-domain.com,https://your-project.pages.dev
```

---

## 🗂️ 项目结构

```
neoai/
├── main.go                 # 后端入口
├── community/              # ✨ 新增：社区频道（类 Discord）
├── middleware/             # CORS、限流、鉴权、健康检查
├── Dockerfile              # 一体化镜像
├── Dockerfile.backend      # 仅后端镜像
├── Dockerfile.frontend     # 仅前端镜像
├── docker-compose.yaml     # 一体化部署
├── docker-compose.split.yaml  # 前后端分离部署
├── nginx.conf              # 反向代理示例
├── app/                    # 前端项目
│   ├── .env.example        # 前端环境变量模板
│   ├── src/api/community.ts       # ✨ 新增：社区频道 API
│   ├── src/routes/community/      # ✨ 新增：社区频道页面
│   └── src/components/admin/JSONChargeDialog.tsx  # ✨ 新增：JSON 定价编辑器
└── DEPLOYMENT.md           # 详细部署文档
```

---

## 🔌 主要 API 速查

### 社区频道（新增）

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/community/channels` | 可选 | 列出可见频道 |
| POST | `/community/channels` | 管理员 | 创建频道 |
| GET | `/community/channels/:id/messages` | 可选 | 拉取消息 |
| POST | `/community/channels/:id/messages` | 已登录 | 发送消息 |
| GET | `/community/ws` | 已登录 | WebSocket 实时推送 |

### 用户名修改（新增）

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| POST | `/admin/user/username` | 管理员 | 改任意用户名 |
| POST | `/profile/username` | 已登录 | 改自己的用户名 |

### 健康检查（新增）

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/healthz` | 公开 | 进程存活探针 |
| GET | `/ready` | 公开 | 就绪探针 |

---

## 🛠️ 开发

### 后端

```bash
go run .                    # 直接运行
go run . --debug            # 调试模式
go build ./... && go vet ./...  # 类型检查
```

### 前端

```bash
cd app
cp .env.example .env
npm install --legacy-peer-deps
npm run dev       # 开发模式（热更新）
npm run build     # 生产构建
```

---

## 🔒 安全建议

1. **生产环境务必修改 `secret`**，至少 32 字节（`openssl rand -hex 32`）
2. **管理员密码立即修改**：登录后到「账户」页改，或 CLI `./neoai --admin --username root --password <新密码>`
3. **CORS 白名单**：分端部署时设置 `ALLOW_ORIGINS=https://your-frontend.pages.dev`
4. **数据库密码**：docker-compose 默认 `neoai123456!` 仅作示例
5. **HTTPS**：前端走 Cloudflare Pages 自动 HTTPS，后端建议套 Caddy / nginx + Let's Encrypt

---

## 📄 License

继承自原项目 coai 的 [LICENSE](./LICENSE)（Apache-2.0）。

## 🙏 鸣谢

- [coai](https://github.com/coaidev/coai) — 原项目
- [QuantumNous/new-api](https://github.com/QuantumNous/new-api) — 渠道编辑器与 JSON 定价风格参考
- [Discord](https://discord.com) — 社区频道设计灵感
