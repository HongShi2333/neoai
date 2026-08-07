# NeoAI — 二开部署指南

NeoAI 是 [coai](https://github.com/coaidev/coai) 的二开版本，针对独立前后端部署、社区频道功能与新型缓存后端做了若干改进。

## 与原版 coai 的差异

1. **缓存后端可选 Valkey / Dragonfly / Redis**  
   `config.yaml` 中 `cache.type` 字段（`redis` / `valkey` / `dragonfly`）任选其一。三者都讲 RESP 协议，底层客户端一致，仅日志与健康检查会区分类型。

2. **新增「社区频道」功能（类 Discord）**  
   - 路由 `/community`，菜单 `bar.community`。
   - 管理员可创建频道，分别设置「可见性」（公开 / 仅登录 / 按角色 / 白名单）与「发送权限」（所有人 / 仅管理员 / 白名单）。
   - 消息落 MySQL（`community_message` 表），同时通过 WebSocket（`/api/community/ws`）实时推送给所有订阅者。
   - 表结构通过 `community.Migrate(db)` 在启动时自动创建。

3. **渠道添加支持逗号分隔模型 + 自动拉取**  
   - 渠道编辑器的「添加自定义模型」输入框现在按 `,` 或空格切分（new-api 风格）。
   - 新增「自动获取模型」按钮：调用上游 `/v1/models` 端点（OpenAI 兼容渠道才支持，其他类型会返回空列表并提示手动添加）。
   - 后端 endpoint：`POST /admin/channel/fetch-models`。

4. **定价支持 JSON 批量编辑**  
   - 收费页面新增「JSON Pricing」按钮，弹窗中可粘贴/编辑 JSON 文档，一键 `replace` 或 `merge`。
   - 后端 endpoint：`POST /admin/charge/json?mode=replace|merge`，`GET /admin/charge/json` 导出。

5. **管理员可改用户名，用户也可在个人中心改用户名**  
   - 管理员侧：`POST /admin/user/username`，前端在用户列表的「操作」下拉中。
   - 用户自助：`POST /profile/username`，前端在「账户」页面用户名旁的「Edit username」按钮。
   - 改名后会自动清空对应 `nio:user:*` 缓存键；现有 JWT 仍然有效。

6. **移除 Pro / License / 广告相关页面**  
   - 管理后台菜单删除了 `License / Warmup / Payment / Record` 等 Pro-only 入口。
   - 路由表中也删除了对应子路由。
   - 保留数据模型（`subscription.level`、`ProType` 等）以兼容已有数据库与按角色限流逻辑。

7. **修复渠道模型不刷新的 bug**  
   原仓库 `Manager.CreateChannel/UpdateChannel/DeleteChannel/Activate/Deactivate` 仅调用 `SaveConfig`，未触发 `Load()`，导致 `globals.V1ListModels` 一直停在启动时的快照——这就是「添加渠道后模型不在市场/价格页面显示」的根因。NeoAI 在所有渠道变更后调用 `reloadAfterMutation()`，立即刷新内存模型列表。

8. **新增健康检查与分离部署支持**  
   - `GET /healthz` 进程存活探针（始终 200）。
   - `GET /ready` 就绪探针（DB 与 Cache 都可达时才 200，并附 `backend` 字段标识当前缓存类型）。
   - 新增 `Dockerfile.backend`、`Dockerfile.frontend`、`docker-compose.split.yaml`，专门用于前后端独立部署。

---

## 部署模式

### A. 一体化部署（与原版相同）

```bash
docker compose up -d
```

使用 `docker-compose.yaml`，前端 SPA 与 Go 后端打包在同一镜像，由后端在 `:8094` 一并提供 `/api`、`/v1`、`/community` 与静态资源。

### B. 前后端分离部署（推荐生产环境）

1. 修改 `docker-compose.split.yaml` 中的：
   - `VITE_BACKEND_ENDPOINT=https://api.your-domain.com`（构建时注入到前端）
   - `ALLOW_ORIGINS=https://app.your-domain.com`（后端 CORS 白名单）
2. 启动：

   ```bash
   docker compose -f docker-compose.split.yaml up -d
   ```

3. 在你的反向代理（nginx / Caddy / Cloudflare）上：
   - `app.your-domain.com` → 前端容器（默认 8080）
   - `api.your-domain.com` → 后端容器（8094）

4. 后端 `config.yaml` 必须设置 `serve_static: false`（split compose 已经通过环境变量覆盖），并且 `system.general.backend` 建议填 `https://api.your-domain.com`，以便邮件、分享等链接生成正确的外链。

### C. 手动构建

```bash
# 后端
go build -o neoai .
SERVE_STATIC=false ./neoai

# 前端
cd app
cp .env.example .env
# 编辑 .env，将 VITE_BACKEND_ENDPOINT 设为后端公网地址
pnpm install && pnpm run build
# 把 dist/ 部署到任意静态托管服务
```

---

## D. Cloudflare Pages 部署前端（详细步骤）

NeoAI 前端是纯静态 SPA，Cloudflare Pages 是免费且全球 CDN 加速的最佳选择。

### D.1 准备工作

1. **后端先就绪**：必须有一个公网可访问的后端地址，例如 `https://api.your-domain.com`。如果你还没部署后端，请先按上面 A/B/C 任一方式部署后端并通过反向代理暴露到公网。

2. **GitHub 仓库**：把 NeoAI 源码推到 GitHub（public 或 private 都行）。

### D.2 在 Cloudflare Pages 创建项目

1. 登录 [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. 左侧菜单 → **Workers & Pages** → **Create application** → **Pages** 标签 → **Connect to Git**
3. 选择你的 GitHub 账号并授权，选中 NeoAI 仓库
4. 配置构建设置：

   | 字段 | 值 |
   | --- | --- |
   | Project name | `neoai`（或任意，会成为 `neoai.pages.dev` 子域） |
   | Production branch | `main`（或你的默认分支） |
   | Framework preset | `None` |
   | Build command | `cd app && npm install --legacy-peer-deps && npm run build` |
   | Build output directory | `app/dist` |
   | Root directory | `/`（仓库根，留空） |

5. 展开下方的 **Environment variables (advanced)**，添加：

   | 变量名 | 值 | 说明 |
   | --- | --- | --- |
   | `VITE_BACKEND_ENDPOINT` | `https://api.your-domain.com` | **必填**，后端公网地址（无尾斜杠） |
   | `VITE_APP_NAME` | `NeoAI` | 可选，浏览器标题栏 |
   | `VITE_APP_LOGO` | `/favicon.ico` | 可选 |
   | `VITE_DOCS_ENDPOINT` | `https://your-docs.com` | 可选 |
   | `VITE_BLOB_ENDPOINT` | `https://blob.your-domain.com` | 可选，文件上传端点 |
   | `VITE_USE_DEEPTRAIN` | `false` | 可选，是否启用 DeepTrain SSO |

6. 点 **Save and Deploy**，等 3-5 分钟构建完成

### D.3 SPA 路由 fallback（重要）

Cloudflare Pages 默认能识别 `index.html` 作为 SPA 入口，但有时刷新子路由会 404。保险起见，在 `app/public/` 下添加 `_redirects` 文件：

```bash
echo '/*    /index.html   200' > app/public/_redirects
```

Vite 构建时会自动把 `app/public/` 内容复制到 `app/dist/`，所以 `_redirects` 会被一起发布。Cloudflare Pages 看到这个文件就会按里面的规则重写所有路径到 `index.html`，SPA 路由就能正常工作了。

### D.4 后端 CORS 配置

前端跑在 `https://neoai.pages.dev` 或自定义域名，后端必须放行该来源。编辑后端 `config/config.yaml`：

```yaml
allow_origins:
  - https://neoai.pages.dev
  - https://app.your-domain.com   # 如果绑定了自定义域名
```

或环境变量（适合 docker / k8s）：

```
ALLOW_ORIGINS=https://neoai.pages.dev,https://app.your-domain.com
```

改完后重启后端。

### D.5 WebSocket 路径（社区频道实时推送）

NeoAI 社区频道用 WebSocket 推送消息。前端会通过 `wss://api.your-domain.com/api/community/ws` 连接，需要在你的反向代理上确保 `Upgrade` 和 `Connection` 头被透传。

**nginx 配置示例**：

```nginx
location / {
    proxy_pass http://127.0.0.1:8094;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_read_timeout 86400s;
}
```

**Caddy 配置示例**（自动 HTTPS）：

```
api.your-domain.com {
    reverse_proxy 127.0.0.1:8094
}
```

Caddy 默认支持 WebSocket，无需额外配置。

### D.6 绑定自定义域名

1. 在 Cloudflare Pages 项目 → **Custom domains** → **Set up a custom domain**
2. 输入 `app.your-domain.com`
3. 按提示添加 CNAME 记录指向 `<your-project>.pages.dev`
4. Cloudflare 自动签发 HTTPS 证书（如果你的域名也在 Cloudflare 托管，全程自动；否则需要按提示添加 TXT 验证）
5. 域名生效后，记得把后端 `ALLOW_ORIGINS` 加上自定义域名

### D.7 验证部署

部署完成后，访问你的 Pages 域名，应能看到 NeoAI 首页。检查：

1. **静态资源加载**：F12 → Network，应看到 `assets/index-*.js`、`assets/index-*.css` 返回 200
2. **API 连通**：在登录页输入账号密码登录，F12 → Network 看 `/api/login` 请求应返回 200，无 CORS 报错
3. **WebSocket**：登录后进入 `/community` 页面，F12 → Network → WS，应看到 `wss://api.your-domain.com/api/community/ws` 连接为 101 Switching Protocols
4. **健康检查**：浏览器访问 `https://api.your-domain.com/healthz` 应返回 `{"status":"ok","service":"neoai"}`

### D.8 常见问题

**Q: 构建失败 `Could not resolve dependency`**

A: NeoAI 前端依赖某些 peer 包版本不严格，构建命令里务必用 `--legacy-peer-deps`（已包含在上面的 Build command 中）。

**Q: 部署后访问页面空白，F12 看到 CORS 错误**

A: 后端没放行 Pages 域名。检查 `ALLOW_ORIGINS` 是否包含 `https://<your-project>.pages.dev` 和自定义域名。

**Q: 社区频道进入后看不到消息，WS 连接一直 pending**

A: 反向代理没透传 WebSocket 升级头。按 D.5 章节检查 nginx / Caddy 配置。

**Q: 刷新 `/community` 路由 404**

A: SPA fallback 没生效。按 D.3 章节添加 `app/public/_redirects` 文件后重新部署。

**Q: 想更换后端地址怎么办**

A: 在 Cloudflare Pages 项目设置 → Environment variables 中改 `VITE_BACKEND_ENDPOINT`，然后触发一次重新部署（Deployments → Retry deployment）。

---

## 缓存后端切换

`config.yaml`（推荐使用新块）：

```yaml
cache:
  type: valkey          # redis | valkey | dragonfly
  host: localhost
  port: 6379
  db: 0
  password: ""
```

或保留旧的 `redis:` 块（向后兼容，两者都存在时以 `cache:` 为准）。

环境变量等价写法（适合 docker / k8s）：

```
CACHE_TYPE=valkey
CACHE_HOST=cache
CACHE_PORT=6379
CACHE_PASSWORD=
CACHE_DB=0
```

切换后端不需要重启代码，只需改配置 + 重启进程。

---

## 社区频道 API 速查

所有路径前缀为 `/api`（一体化）或后端公网根路径（分离部署）。

| 方法   | 路径                                | 鉴权     | 说明                     |
| ------ | ----------------------------------- | -------- | ------------------------ |
| GET    | `/community/channels`              | 可选     | 列出当前用户可见的频道   |
| POST   | `/community/channels`               | 管理员   | 创建频道                 |
| GET    | `/community/channels/:id`           | 可选     | 获取单个频道             |
| POST   | `/community/channels/:id`           | 管理员   | 更新频道                 |
| DELETE | `/community/channels/:id`           | 管理员   | 删除频道（级联删除消息） |
| GET    | `/community/channels/:id/messages`  | 可选     | 拉取最近消息（?limit=）  |
| POST   | `/community/channels/:id/messages`  | 已登录   | 发送消息                 |
| POST   | `/community/messages/:id`           | 已登录   | 编辑自己的消息           |
| DELETE | `/community/messages/:id`           | 已登录   | 删除消息（作者或管理员） |
| GET    | `/community/ws`                     | 已登录   | WebSocket 实时推送       |

WebSocket 协议：
- 客户端发送：`{"action":"subscribe","channel_id":1}` / `{"action":"send","channel_id":1,"content":"hi"}`
- 服务端推送：`{"type":"message","data":{...}}` / `{"type":"delete","message_id":123}` / `{"type":"error","message":"..."}`

---

## 用户名修改 API

| 方法 | 路径                  | 鉴权   | 说明                          |
| ---- | --------------------- | ------ | ----------------------------- |
| POST | `/admin/user/username` | 管理员 | 改任意用户的用户名            |
| POST | `/profile/username`   | 已登录 | 改自己的用户名                |

请求体均为 `{ "username": "<new>" }`。改名后会清空 `nio:user:<old>` 与 `nio:user:<new>` 两个缓存键；现有 JWT 不失效。
