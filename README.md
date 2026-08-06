<div align="center">

![neoai](/app/public/logo.png)

# 🥳 NeoAI

#### 🚀 Next-Generation AIGC One-Stop Business Solution (Based on CoAI)

#### *"NeoAI = [CoAI](https://github.com/coaidev/coai) Core + Community Channels + Multi-Cache Driver + Enhanced Pricing"*

[English](./README.md) · [简体中文](./README_zh-CN.md) · [日本語](./README_ja-JP.md)

<img alt="NeoAI Preview" src="./screenshot/coai.png" width="100%" style="border-radius: 8px">

</div>

## 📌 About This Project

**NeoAI** is a secondary development project based on the open-source project [CoAI](https://github.com/coaidev/coai) (Apache-2.0 license). It inherits CoAI's excellent AIGC commercial capabilities while enhancing three key areas: community interaction, cache architecture, and channel pricing.

> 🙏 Special thanks to the [CoAI.Dev](https://coai.dev) open-source community for providing the outstanding foundational project. NeoAI extends and customizes it further.

### ✨ NeoAI Enhancements Over the Original

1. 🗄️ **Multi-Cache Driver Support**: While the original only supports Redis, NeoAI adds optional support for **Valkey** and **Dragonfly** — two high-performance Redis-compatible cache databases. Switch seamlessly with a single config change; all three drivers use the Redis protocol and share identical connection parameters.
2. 💬 **Community Channel System (Discord-like)**: A brand-new community channel feature where admins can create multiple channels with fine-grained controls:
   - **Channel Visibility**: Specify which user groups and members can see each channel
   - **Posting Permissions**: Control which users are allowed to send messages in a channel
   - **Real-time Message Sync**: WebSocket-based real-time message broadcasting; all messages are persisted to MySQL
   - **History Loading**: Historical messages are automatically loaded when entering a channel
3. 🎛️ **Enhanced Channel & Pricing Management (NewAPI-style)**:
   - **Batch Model Input**: Both the channel editor and pricing editor support **comma / space / newline separated** batch model input — no more adding models one by one
   - **JSON Pricing Editor**: In addition to the original visual pricing editor, a **JSON batch editor** is now available. Export current pricing rules, edit in bulk, and import to overwrite — ideal for large-scale pricing adjustments

## 📝 Core Features Inherited from CoAI

1. 🤖️ **Rich Model Support**: Multi-provider support (OpenAI / Anthropic / Gemini / Midjourney and 10+ compatible formats & private LLMs)
2. 🤯 **Beautiful UI Design**: Responsive UI for PC / Pad / Mobile, following [Shadcn UI](https://ui.shadcn.com) & [Tremor Charts](https://blocks.tremor.so) design standards
3. 🎃 **Complete Markdown Support**: **LaTeX formulas** / **Mermaid mind maps** / tables / code highlighting / charts / progress bars
4. 👀 **Multi-Theme Support**: Light mode and dark mode with customizable color schemes
5. 📚 **Internationalization Support**: Multi-language switching 🇨🇳 🇺🇸 🇯🇵 🇷🇺
6. 🎨 **Text-to-Image Support**: **OpenAI DALL-E**✅ & **Midjourney** (with **U/V/R** actions)✅ & Stable Diffusion✅
7. 📡 **Powerful Conversation Sync**: Zero-cost cross-device conversation sync with link sharing & save-as-image
8. 🎈 **Model Market & Preset System**: Customizable model market and preset system with cloud sync
9. 📖 **Rich File Parsing**: Out-of-the-box file parsing for all models (PDF / Docx / Pptx / Excel / images) with OCR support
10. 🌏 **Full Internet Search**: Based on [SearXNG](https://github.com/searxng/searxng), supporting Google / Bing / DuckDuckGo and more
11. 💕 **PWA Support**: Progressive Web App with desktop support via [Tauri](https://github.com/tauri-apps/tauri)
12. 🤩 **Comprehensive Backend**: Dashboard, announcements, user management, subscriptions, gift codes, pricing, SMTP, and more
13. 🤑 **Multiple Billing Methods**: Subscription and elastic billing (per-request / token / free / anonymous calls)
14. 🎉 **Model Caching**: Return cached results for identical request parameter hashes (cache hits are not billed)
15. 😎 **Channel Management**: Multi-channel with priority, weight, user grouping, auto-retry, model redirection
16. ⭐ **OpenAI API Distribution & Proxy**: Call all major LLMs via the standard OpenAI API format
17. 👌 **Upstream Sync**: Quickly sync channel settings, model market, and pricing from upstream sites
18. 👋 **SEO Optimization**: Custom site name, logo, and other SEO settings
19. 🎫 **Redemption Code System**: Gift codes and redemption codes with batch generation
20. 🥰 **Business-Friendly License**: Inherits the **Apache-2.0** open-source license

## 🔨 Supported Models
1. OpenAI & Azure OpenAI *(✅ Vision ✅ Function Calling)*
2. Anthropic Claude *(✅ Vision ✅ Function Calling)*
3. Google Gemini & PaLM2 *(✅ Vision)*
4. Midjourney *(✅ Mode Toggling ✅ U/V/R Actions)*
5. iFlytek SparkDesk *(✅ Vision ✅ Function Calling)*
6. Zhipu ChatGLM *(✅ Vision)*
7. Alibaba Tongyi Qwen
8. Tencent Hunyuan
9. Baichuan AI
10. Moonshot AI (👉 OpenAI)
11. DeepSeek AI (👉 OpenAI)
12. ByteDance Skylark *(✅ Function Calling)*
13. Groq Cloud AI
14. OpenRouter (👉 OpenAI)
15. 360 GPT
16. LocalAI / Ollama (👉 OpenAI)

## 👻 OpenAI Compatible API Proxy
   - [x] Chat Completions _(/v1/chat/completions)_
   - [x] Image Generation _(/v1/images)_
   - [x] Model List _(/v1/models)_
   - [x] Dashboard Billing _(/v1/billing)_


## 📦 Deployment
> [!TIP]
> **After successful deployment, the admin account is `root`, with the default password `chatnio123456`**

### ⚡ Docker Compose Installation (Recommended)
> [!NOTE]
> After successful execution, the host machine mapping address is `http://localhost:8000`

```shell
git clone --depth=1 https://github.com/your-org/neoai.git
cd neoai
docker-compose up -d # Run the service
```

#### Switching the Cache Driver

NeoAI supports three cache drivers — Redis / Valkey / Dragonfly — all using the Redis protocol with identical connection parameters. Just modify two entries in `docker-compose.yaml` to switch:

```yaml
  redis:
    # Switch to valkey/valkey:latest or dragonflydb/dragonfly:latest to use the corresponding driver
    image: redis:latest            # or valkey/valkey:latest or dragonflydb/dragonfly:latest

  chatnio:
      environment:
          REDIS_TYPE: redis        # Allowed values: redis (default) / valkey / dragonfly
```

Version update:
```shell
docker-compose down 
docker-compose pull
docker-compose up -d
```

> - MySQL database mount directory: ~/**db**
> - Cache database mount directory: ~/**redis**
> - Configuration file mount directory: ~/**config**

### ⚡ Docker Installation (Lightweight runtime, commonly used with external _MYSQL/RDS_ services)
> [!NOTE]
> After successful execution, the host machine address is `http://localhost:8094`.

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

> - `REDIS_TYPE`: Cache driver type. Allowed values: `redis` (default) / `valkey` / `dragonfly`
> - `SECRET`: JWT secret key — generate a random string and modify accordingly
> - `SERVE_STATIC`: Whether to enable static file serving (normally no need to change)

### ⚒ Compile and Install
> [!NOTE]
> After successful deployment, the default port is **8094**, accessible at `http://localhost:8094`
>
> Config settings (~/config/**config.yaml**) can be overridden using environment variables. For example, the `MYSQL_HOST` environment variable overrides the `mysql.host` config item

```shell
git clone https://github.com/your-org/neoai.git
cd neoai

cd app
npm install -g pnpm
pnpm install
pnpm build

cd ..
go build -o neoai

# e.g. using nohup (you can also use systemd or other service managers)
nohup ./neoai > output.log &
```

#### Configuration Example (`config.yaml`)

```yaml
mysql:
  db: chatnio
  host: localhost
  password: chatnio123456
  port: 3306
  user: root
  tls: false

redis:
  # cache driver: "redis" (default), "valkey" or "dragonfly"
  # valkey and dragonfly are wire-compatible with the redis protocol,
  # so the connection options below apply to all three.
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

## 🆕 New Features Guide

### Multi-Cache Driver

As described above, switch between Redis / Valkey / Dragonfly via the `redis.type` config item or the `REDIS_TYPE` environment variable. The backend logs the active cache driver on startup for easy troubleshooting.

### Community Channels

1. **Admin creates a channel**: Log in to the backend → "Community Channels" in the left sidebar → click "New Channel", fill in the name and description, then set the visible user groups, visible members, and posting members.
2. **Users enter a channel**: After logging in, users click "Community" in the navigation bar to see the list of channels they have permission to view, then click a channel to enter the chat.
3. **Message sync**: All sent messages are pushed in real time to online users in the channel via WebSocket and simultaneously written to the database. Re-entering a channel automatically loads historical messages.

### Enhanced Channel & Pricing Management

1. **Batch model input**: In the channel editor or pricing editor's "Add custom model" input box, you can type `gpt-4, gpt-4o  gpt-4-turbo` (commas, spaces, and newlines all work as separators), then click add to insert multiple models at once.
2. **JSON pricing editor**: On the "Pricing Management" page, click the "JSON Pricing Editor" button:
   - On open, all current pricing rules are automatically exported as formatted JSON
   - Edit the JSON directly in the editor (real-time JSON format validation)
   - Click "Import & Overwrite" to write the edited pricing rules back to the database

## ❓ FAQ
1. **Why can I access pages and log in after deployment, but chat keeps loading?**
   - Chat and similar features communicate via WebSocket. Please ensure your service supports WebSocket.
   - If you use Nginx, Apache, or other reverse proxies, make sure WebSocket support is configured.
2. **What's the difference between Valkey / Dragonfly and Redis? Which should I choose?**
   - Valkey is the open-source fork of Redis, maintained by the Linux Foundation; Dragonfly is a Redis-protocol-compatible cache with a multi-threaded architecture for higher throughput. All three are fully compatible at the application layer in NeoAI — switch freely as needed. The default Redis is fine for most cases; try Dragonfly for higher throughput; choose Valkey if you care about open-source governance sustainability.
3. **Do community channel messages consume database space?**
   - Yes. All channel messages are persisted to the `community_message` table in MySQL for historical tracing and auditing. To clean up, manually delete the corresponding records in the database.
4. **Does the JSON pricing editor import overwrite existing pricing?**
   - Yes. To avoid ambiguity, the JSON editor import uses overwrite mode. It automatically exports the current pricing rules as a reference before import — we recommend backing up before editing.
5. **My Midjourney Proxy channel keeps loading or reports `please provide available notify url`**
   - Make sure your Midjourney Proxy service is running and that the channel type is set to Midjourney, not OpenAI.
   - Check that the **backend domain** in system settings is correctly configured.
6. **How do I change the default Root password?**
   - Click your avatar in the top-right to enter the backend, then click "Change Root Password" under System Settings → General Settings.

## 📦 Tech Stack
- 🥗 Frontend: React + Redux + Radix UI + Tailwind CSS
- 🍎 Backend: Golang + Gin + Redis/Valkey/Dragonfly + MySQL
- 🍒 Application Technology: PWA + WebSocket

## 📄 License

This project is derived from [CoAI](https://github.com/coaidev/coai) and inherits the **Apache-2.0** open-source license. Commercial secondary development and distribution are welcome — please comply with the Apache-2.0 license and do not use it for illegal purposes.

## ❤ Acknowledgements & Donations

- 🙏 Thanks to [CoAI.Dev](https://coai.dev) and its contributors for the excellent open-source foundational project
- If you find this project helpful, please give it a Star to show your support!
