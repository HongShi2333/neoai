<div align="center">

![neoai](/app/public/logo.png)

# 🥳 NeoAI

#### 🚀 基于 NeoAI 二次开发的下一代 AIGC 一站式商业解决方案

#### *"NeoAI = [NeoAI](https://github.com/coaidev/coai) 内核 + 社区频道 + 多缓存驱动 + 增强定价"*

[English](./README.md) · 简体中文 · [日本語](./README_ja-JP.md)

<img alt="NeoAI Preview" src="./screenshot/coai.png" width="100%" style="border-radius: 8px">

</div>

## 📌 项目说明

**NeoAI** 是基于开源项目 [NeoAI](https://github.com/coaidev/coai)（Apache-2.0 协议）进行的二次开发项目，旨在继承 NeoAI 优秀的 AIGC 商业能力的同时，针对社区互动、缓存架构与渠道定价三大方向进行功能增强。

> 🙏 特别感谢 [NeoAI.Dev](https://coai.dev) 开源社区提供的优秀基础项目，本项目在其基础上进行了功能扩展与定制化改进。

### ✨ NeoAI 相对原版的增强功能

1. 🗄️ **多缓存驱动支持**: 在原版仅支持 Redis 的基础上，新增 **Valkey** 与 **Dragonfly** 两种高性能兼容缓存数据库的可选支持。仅需修改一项配置即可无缝切换，三种驱动均使用 Redis 协议，连接参数完全兼容。
2. 💬 **社区频道系统（类似 Discord）**: 全新引入社区频道功能，管理员可创建多个频道，并精细化设置：
   - **频道可见性**: 指定哪些用户组、哪些成员可以看到该频道
   - **发言权限**: 控制哪些用户可以在频道内发送消息
   - **实时消息同步**: 基于 WebSocket 的实时消息广播，所有消息同步写入 MySQL 数据库持久化
   - **历史消息加载**: 进入频道自动拉取历史消息记录
3. 🎛️ **增强的渠道与定价管理（NewAPI 风格）**:
   - **批量模型添加**: 渠道编辑器和定价编辑器均支持 **逗号 / 空格 / 换行分隔** 批量添加模型，告别逐个输入
   - **JSON 定价编辑器**: 在原有可视化定价设置的基础上，新增 **JSON 格式批量编辑器**，支持一键导出当前定价规则、批量编辑、一键导入覆盖，适合大规模定价调整

## 📝 继承自 NeoAI 的核心功能

1. 🤖️ **丰富模型支持**: 多模型服务商支持 (OpenAI / Anthropic / Gemini / Midjourney 等十余种格式兼容 & 私有化 LLM 支持)
2. 🤯 **美观 UI 设计**: UI 兼容 PC / Pad / 移动三端，遵循 [Shadcn UI](https://ui.shadcn.com) & [Tremor Charts](https://blocks.tremor.so) 设计规范，丰富美观的界面设计和后台仪表盘
3. 🎃 **完整 Markdown 支持**: 支持 **LaTeX 公式** / **Mermaid 思维导图** / 表格渲染 / 代码高亮 / 图表绘制 / 进度条等进阶 Markdown 语法支持
4. 👀 **多主题支持**: 支持多种主题切换，包含亮色主题的**明亮模式**和暗色主题的**深色模式**。
5. 📚 **国际化支持**: 支持国际化，支持多语言切换 🇨🇳 🇺🇸 🇯🇵 🇷🇺
6. 🎨 **文生图支持**: 支持多种文生图模型: **OpenAI DALL-E**✅ & **Midjourney** (支持 **U/V/R** 操作)✅ & Stable Diffusion✅ 等
7. 📡 **强大对话同步**: **用户 0 成本对话跨端同步支持**，支持**对话分享** (支持链接分享 & 保存为图片 & 分享管理)
8. 🎈 **模型市场 & 预设系统**: 支持后台可自定义的模型市场，同时支持预设系统，包含 **自定义预设** 和 **云端同步** 功能。
9. 📖 **丰富文件解析**: **开箱即用**，支持**所有模型**的文件解析 (PDF / Docx / Pptx / Excel / 图片等格式解析)，支持 OCR 图片识别
10. 🌏 **全模型联网搜索**: 基于 [SearXNG](https://github.com/searxng/searxng) 开源引擎，支持 Google / Bing / DuckDuckGo 等丰富搜索引擎
11. 💕 **渐进式 Web 应用 (PWA)**: 支持 PWA 应用 & 支持桌面端 (桌面端基于 [Tauri](https://github.com/tauri-apps/tauri))
12. 🤩 **齐全后台管理**: 支持美观丰富的仪表盘，公告 & 通知管理，用户管理，订阅管理，礼品码 & 兑换码管理，价格设定，订阅设定等功能
13. 🤑 **多种计费方式**: 支持 💴 **订阅制** 和 💴 **弹性计费** 两种计费方式，弹性计费支持 次数计费 / Token 计费 / 不计费 / 可匿名调用
14. 🎉 **创新模型缓存**: 支持开启模型缓存，同一个请求入参 Hash 下，如果之前已请求过，将直接返回缓存结果
15. 😎 **优秀渠道管理**: 支持⚡ **多渠道管理**，支持🥳**优先级**、🥳**权重**、🥳**用户分组**、🥳**失败自动重试**、🥳**模型重定向**等企业级功能
16. ⭐ **OpenAI API 分发 & 中转系统**: 支持以 **OpenAI API** 标准格式调用各种大模型
17. 👌 **快速同步上游**: 渠道设置、模型市场、价格设定等设置都可快速同步上游站点
18. 👋 **SEO 优化**: 支持自定义站点名称、站点 Logo 等 SEO 优化设置
19. 🎫 **多种兑换码体系**: 支持礼品码和兑换码，支持批量生成
20. 🥰 **商用友好协议**: 继承 **Apache-2.0** 开源协议，商用二开 & 分发友好

## 🔨 支持模型
1. OpenAI & Azure OpenAI *(✅ Vision ✅ Function Calling)*
2. Anthropic Claude *(✅ Vision ✅ Function Calling)*
3. Google Gemini & PaLM2 *(✅ Vision)*
4. Midjourney *(✅ Mode Toggling ✅ U/V/R Actions)*
5. 讯飞星火 SparkDesk *(✅ Vision ✅ Function Calling)*
6. 智谱清言 ChatGLM *(✅ Vision)*
7. 通义千问 Tongyi Qwen
8. 腾讯混元 Tencent Hunyuan
9. 百川大模型 Baichuan AI
10. 月之暗面 Moonshot AI (👉 OpenAI)
11. 深度求索 DeepSeek AI (👉 OpenAI)
12. 字节云雀 ByteDance Skylark *(✅ Function Calling)*
13. Groq Cloud AI
14. OpenRouter (👉 OpenAI)
15. 360 GPT
16. LocalAI / Ollama (👉 OpenAI)

## 👻 中转 OpenAI 兼容 API
   - [x] Chat Completions _(/v1/chat/completions)_
   - [x] Image Generation _(/v1/images)_
   - [x] Model List _(/v1/models)_
   - [x] Dashboard Billing _(/v1/billing)_


## 📦 部署方式
> [!TIP]
> **部署成功后, 管理员账号为 `root`, 密码默认为 `chatnio123456`**

### ⚡ Docker Compose 安装 (推荐)
> [!NOTE]
> 运行成功后, 宿主机映射地址为 `http://localhost:8000`

 ```shell
 git clone --depth=1 https://github.com/your-org/neoai.git
 cd neoai
 docker-compose up -d # 运行服务
```

#### 切换缓存驱动

NeoAI 支持 Redis / Valkey / Dragonfly 三种缓存驱动，三种均使用 Redis 协议，连接参数完全兼容。仅需修改 `docker-compose.yaml` 中两项即可切换：

```yaml
  redis:
    # 切换到 valkey/valkey:latest 或 dragonflydb/dragonfly:latest 即可使用对应驱动
    image: redis:latest            # 或 valkey/valkey:latest 或 dragonflydb/dragonfly:latest

  chatnio:
      environment:
          REDIS_TYPE: redis        # 可选值: redis (默认) / valkey / dragonfly
```

版本更新：
```shell
docker-compose down 
docker-compose pull
docker-compose up -d
```

> - MySQL 数据库挂载目录项目 ~/**db**
> - 缓存数据库挂载目录项目 ~/**redis**
> - 配置文件挂载目录项目 ~/**config**

### ⚡ Docker 安装 (轻量运行时, 常用于外置 _MYSQL/RDS_ 服务)
> [!NOTE]
> 运行成功后, 宿主机地址为 `http://localhost:8094`。

```shell
docker run -d --name neoai \
   --network host \
   -v ~/config:/config \
   -v ~/logs:/logs \
   -v ~/storage:/storage \
   -e MYSQL_HOST=localhost \
   -e MYSQL_PORT=3306 \
   -e MYSQL_DB=chatnio \
   -e MYSQL_USER=root \
   -e MYSQL_PASSWORD=chatnio123456 \
   -e REDIS_HOST=localhost \
   -e REDIS_PORT=6379 \
   -e REDIS_TYPE=redis \
   -e SECRET=secret \
   -e SERVE_STATIC=true \
   neoai:latest
```

> - `REDIS_TYPE`: 缓存驱动类型，可选 `redis` (默认) / `valkey` / `dragonfly`
> - `SECRET`: JWT 密钥, 自行生成随机字符串修改
> - `SERVE_STATIC`: 是否启用静态文件服务 (正常情况下不需要更改此项)

### ⚒ 编译安装
> [!NOTE]
> 部署成功后, 默认端口为 **8094**, 访问地址为 `http://localhost:8094`
> 
> Config 配置项 (~/config/**config.yaml**) 可以使用环境变量进行覆盖, 如 `MYSQL_HOST` 环境变量可覆盖 `mysql.host` 配置项

```shell
git clone https://github.com/your-org/neoai.git
cd neoai

cd app
npm install -g pnpm
pnpm install
pnpm build

cd ..
go build -o neoai

# 使用 nohup 后台运行（也可使用 systemd 或其他服务管理器）
nohup ./neoai > output.log &
```

#### 配置文件示例 (`config.yaml`)

```yaml
mysql:
  db: chatnio
  host: localhost
  password: chatnio123456
  port: 3306
  user: root
  tls: false

redis:
  # 缓存驱动: "redis" (默认), "valkey" 或 "dragonfly"
  # valkey 和 dragonfly 均兼容 redis 协议, 以下连接参数对三者通用
  type: redis
  host: localhost
  port: 6379
  db: 0
  password: ""

secret: secret
serve_static: true
server:
  port: 8094
```

## 🆕 新功能使用指南

### 多缓存驱动

如上所述，通过 `redis.type` 配置项或 `REDIS_TYPE` 环境变量即可切换 Redis / Valkey / Dragonfly。后端启动时会在日志中输出当前生效的缓存驱动类型，方便排查。

### 社区频道

1. **管理员创建频道**: 登录后台 → 左侧菜单「社区频道」→ 点击「新建频道」，填写频道名称、描述，并设置可见用户组、可见成员、可发言成员。
2. **用户进入频道**: 用户登录后，在导航栏点击「社区」即可看到自己有权限查看的频道列表，点击频道进入聊天界面。
3. **消息同步**: 所有发送的消息会实时通过 WebSocket 推送给频道内在线用户，并同时写入数据库。重新进入频道会自动加载历史消息。

### 增强的渠道与定价管理

1. **批量添加模型**: 在渠道编辑器或定价编辑器的「添加自定义模型」输入框中，可输入 `gpt-4, gpt-4o  gpt-4-turbo`（逗号、空格、换行均可作为分隔符），点击添加即可一次性加入多个模型。
2. **JSON 定价编辑器**: 在「定价管理」页面点击「JSON 定价编辑器」按钮：
   - 打开时会自动将当前所有定价规则导出为格式化的 JSON
   - 可直接在编辑器中修改 JSON 内容（编辑器会实时校验 JSON 格式合法性）
   - 点击「导入并覆盖」即可将编辑后的定价规则覆盖写入数据库

## ❓ 常见问题 Q&A
1. **为什么我部署后的站点可以访问页面, 可以登录注册, 但是无法使用聊天 (一直在转圈)？**
   - 聊天等此类功能通过 websocket 进行通信, 请确保你的服务支持 websocket。
   - 如果你使用了 Nginx, Apache 等反向代理, 请确保已配置 websocket 支持。
2. **Valkey / Dragonfly 与 Redis 有什么区别？我该选哪个？**
   - Valkey 是 Redis 的开源分支，由 Linux Foundation 维护；Dragonfly 是兼容 Redis 协议但采用多线程架构的高性能缓存。三者对应用层完全兼容，NeoAI 中可按需自由切换。一般场景下默认使用 Redis 即可；追求更高吞吐量时可尝试 Dragonfly；关注开源治理可持续性时可选择 Valkey。
3. **社区频道的消息会占用数据库吗？**
   - 会的。所有频道消息都会持久化到 MySQL 的 `community_message` 表中，便于历史追溯和审计。如需清理可在数据库中手动删除对应记录。
4. **JSON 定价编辑器导入会覆盖原有定价吗？**
   - 是的。为避免混淆，JSON 编辑器的导入采用覆盖模式。导入前会自动导出当前定价规则作为参考，建议在编辑前先备份。
5. **我配置的 Midjourney Proxy 格式的渠道一直转圈或报错 `please provide available notify url`**
   - 请确保你的 Midjourney Proxy 服务已正常运行, 并且渠道类型选择 Midjourney 而不是 OpenAI。
   - 请查看系统设置中的**后端域名**是否已经配置并配置正确。
6. **如何修改 Root 默认密码？**
   - 点击右上角头像进入后台管理, 点击系统设置下常规设置操作栏的「修改 Root 密码」进行修改。

## 📦 技术栈
- 🥗 前端: React + Redux + Radix UI + Tailwind CSS
- 🍎 后端: Golang + Gin + Redis/Valkey/Dragonfly + MySQL
- 🍒 应用技术: PWA + WebSocket

## 📄 开源协议

本项目继承自 [NeoAI](https://github.com/coaidev/coai)，沿用 **Apache-2.0** 开源协议。商用二开 & 分发友好，请遵守 Apache-2.0 协议的规定，请勿用于违法用途。

## ❤ 致谢与捐助

- 🙏 感谢 [NeoAI.Dev](https://coai.dev) 及其贡献者提供的优秀开源基础项目
- 如果您觉得这个项目对您有所帮助, 您可以点个 Star 支持一下！
