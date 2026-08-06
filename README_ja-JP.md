<div align="center">

![neoai](/app/public/logo.png)

# 🥳 NeoAI

#### 🚀 次世代AIGCワンストップビジネスソリューション（CoAIベース）

#### *"NeoAI = [CoAI](https://github.com/coaidev/coai) コア + コミュニティチャンネル + マルチキャッシュドライバ + 拡張価格設定"*

[English](./README.md) · [简体中文](./README_zh-CN.md) · 日本語

<img alt="NeoAI Preview" src="./screenshot/coai.png" width="100%" style="border-radius: 8px">

</div>

## 📌 プロジェクト概要

**NeoAI** は、オープンソースプロジェクト [CoAI](https://github.com/coaidev/coai)（Apache-2.0ライセンス）をベースにした二次開発プロジェクトです。CoAIの優れたAIGC商用能力を引き継ぎつつ、コミュニティインタラクション、キャッシュアーキテクチャ、チャネル価格設定の3つの領域を強化しています。

> 🙏 [CoAI.Dev](https://coai.dev) オープンソースコミュニティに、優秀な基盤プロジェクトを提供していただき、特別に感謝申し上げます。NeoAIはこれを拡張・カスタマイズしています。

### ✨ NeoAIのオリジナルに対する拡張機能

1. 🗄️ **マルチキャッシュドライバ対応**: オリジナルではRedisのみ対応していましたが、NeoAIは **Valkey** と **Dragonfly** の2つの高性能な互換キャッシュデータベースを選択可能にしました。1つの設定を変更するだけでシームレスに切り替えられ、3つのドライバはすべてRedisプロトコルを使用し、接続パラメータは完全互換です。
2. 💬 **コミュニティチャンネルシステム（Discord風）**: 全く新しいコミュニティチャンネル機能。管理者は複数のチャンネルを作成し、きめ細かく設定できます：
   - **チャンネルの可視性**: どのユーザーグループとメンバーがチャンネルを閲覧できるかを指定
   - **投稿権限**: どのユーザーがチャンネル内でメッセージを送信できるかを制御
   - **リアルタイムメッセージ同期**: WebSocketベースのリアルタイムメッセージブロードキャスト、すべてのメッセージはMySQLに永続化
   - **履歴メッセージの読み込み**: チャンネルに入ると自動的に履歴メッセージを読み込み
3. 🎛️ **拡張チャネル・価格管理（NewAPI風）**:
   - **バッチモデル追加**: チャネルエディタと価格エディタの両方で **カンマ / スペース / 改行区切り** のバッチモデル入力をサポート。1つずつ入力する必要はありません
   - **JSON価格エディタ**: オリジナルの視覚的価格設定に加え、**JSON形式バッチエディタ** を新設。現在の価格ルールをエクスポート、一括編集、インポートで上書きが可能で、大規模な価格調整に最適

## 📝 CoAIから引き継いだコア機能

1. 🤖️ **豊富なモデルサポート**: マルチプロバイダ対応（OpenAI / Anthropic / Gemini / Midjourneyなど10種類以上の互換形式 & プライベートLLM対応）
2. 🤯 **美しいUIデザイン**: PC / タブレット / モバイルの3端末に対応、[Shadcn UI](https://ui.shadcn.com) & [Tremor Charts](https://blocks.tremor.so) デザイン規格に準拠
3. 🎃 **完全なMarkdownサポート**: **LaTeX数式** / **Mermaidマインドマップ** / テーブル / コードハイライト / チャート / プログレスバー
4. 👀 **マルチテーマ対応**: ライトモードとダークモード、カスタムカラースキーム対応
5. 📚 **国際化対応**: マルチ言語切り替え 🇨🇳 🇺🇸 🇯🇵 🇷🇺
6. 🎨 **テキストから画像生成**: **OpenAI DALL-E**✅ & **Midjourney**（**U/V/R**操作対応）✅ & Stable Diffusion✅
7. 📡 **強力な対話同期**: ゼロコストのクロスデバイス対話同期、リンク共有 & 画像保存対応
8. 🎈 **モデル市場 & プリセットシステム**: カスタマイズ可能なモデル市場とクラウド同期プリセットシステム
9. 📖 **豊富なファイル解析**: すべてのモデルでPDF / Docx / Pptx / Excel / 画像解析をサポート、OCR対応
10. 🌏 **フルインターネット検索**: [SearXNG](https://github.com/searxng/searxng) ベース、Google / Bing / DuckDuckGo等に対応
11. 💕 **PWAサポート**: [Tauri](https://github.com/tauri-apps/tauri) ベースのデスクトップ対応
12. 🤩 **充実のバックエンド管理**: ダッシュボード、お知らせ、ユーザー管理、サブスク、ギフトコード等
13. 🤑 **複数の課金方式**: サブスクと弾力課金（リクエスト単位 / トークン単位 / 無料 / 匿名呼び出し）
14. 🎉 **モデルキャッシュ**: 同一リクエストパラメータのハッシュに対してキャッシュ結果を返却（キャッシュヒットは課金なし）
15. 😎 **優秀なチャネル管理**: マルチチャネル、優先度、重み、ユーザーグループ、自動リトライ、モデルリダイレクト
16. ⭐ **OpenAI API配信 & プロキシ**: 標準OpenAI API形式で各種LLMを呼び出し可能
17. 👌 **上流同期**: チャネル設定、モデル市場、価格設定を上流サイトから素早く同期
18. 👋 **SEO最適化**: カスタムサイト名、ロコ等のSEO設定
19. 🎫 **引換コードシステム**: ギフトコードと引換コード、バッチ生成対応
20. 🥰 **商用フレンドリーライセンス**: **Apache-2.0** オープンソースライセンスを継承

## 🔨 対応モデル
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

## 👻 OpenAI互換APIプロキシ
   - [x] Chat Completions _(/v1/chat/completions)_
   - [x] Image Generation _(/v1/images)_
   - [x] Model List _(/v1/models)_
   - [x] Dashboard Billing _(/v1/billing)_


## 📦 デプロイ方法
> [!TIP]
> **デプロイ成功後、管理者アカウントは `root`、デフォルトパスワードは `chatnio123456`**

### ⚡ Docker Compose インストール（推奨）
> [!NOTE]
> 実行成功後、ホストマシンのマッピングアドレスは `http://localhost:8000`

```shell
git clone --depth=1 https://github.com/your-org/neoai.git
cd neoai
docker-compose up -d # サービスを実行
```

#### キャッシュドライバの切り替え

NeoAIは Redis / Valkey / Dragonfly の3種類のキャッシュドライバをサポートしています。3つともRedisプロトコルを使用し、接続パラメータは完全互換です。`docker-compose.yaml` の2箇所を変更するだけで切り替えられます：

```yaml
  redis:
    # valkey/valkey:latest または dragonflydb/dragonfly:latest に切り替えて対応ドライバを使用
    image: redis:latest            # または valkey/valkey:latest または dragonflydb/dragonfly:latest

  chatnio:
      environment:
          REDIS_TYPE: redis        # 許可値: redis (デフォルト) / valkey / dragonfly
```

バージョン更新：
```shell
docker-compose down 
docker-compose pull
docker-compose up -d
```

> - MySQLデータベースマウントディレクトリ: ~/**db**
> - キャッシュデータベースマウントディレクトリ: ~/**redis**
> - 設定ファイルマウントディレクトリ: ~/**config**

### ⚡ Docker インストール（軽量ランタイム、外部 _MYSQL/RDS_ サービスでよく使用）
> [!NOTE]
> 実行成功後、ホストマシンアドレスは `http://localhost:8094`。

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

> - `REDIS_TYPE`: キャッシュドライバタイプ。許可値: `redis` (デフォルト) / `valkey` / `dragonfly`
> - `SECRET`: JWTシークレットキー — ランダム文字列を生成して変更してください
> - `SERVE_STATIC`: 静的ファイルサービングを有効にするか（通常は変更不要）

### ⚒ コンパイルしてインストール
> [!NOTE]
> デプロイ成功後、デフォルトポートは **8094**、アクセスアドレスは `http://localhost:8094`
>
> Config設定 (~/config/**config.yaml**) は環境変数で上書き可能。例: `MYSQL_HOST` 環境変数で `mysql.host` 設定を上書き

```shell
git clone https://github.com/your-org/neoai.git
cd neoai

cd app
npm install -g pnpm
pnpm install
pnpm build

cd ..
go build -o neoai

# nohupでバックグラウンド実行（systemd等のサービスマネージャも使用可能）
nohup ./neoai > output.log &
```

#### 設定例 (`config.yaml`)

```yaml
mysql:
  db: chatnio
  host: localhost
  password: chatnio123456
  port: 3306
  user: root
  tls: false

redis:
  # キャッシュドライバ: "redis" (デフォルト), "valkey" または "dragonfly"
  # valkey と dragonfly は redis プロトコルと互換性があり、
  # 以下の接続オプションは3つすべてに適用されます。
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

## 🆕 新機能ガイド

### マルチキャッシュドライバ

上記の通り、`redis.type` 設定項目または `REDIS_TYPE` 環境変数で Redis / Valkey / Dragonfly を切り替えられます。バックエンドは起動時に有効なキャッシュドライバをログに出力し、トラブルシューティングを容易にします。

### コミュニティチャンネル

1. **管理者がチャンネルを作成**: バックエンドにログイン → 左サイドバーの「コミュニティチャンネル」→「新規チャンネル」をクリック、名称と説明を入力し、閲覧可能なユーザーグループ、メンバー、投稿可能なメンバーを設定。
2. **ユーザーがチャンネルに入る**: ログイン後、ナビゲーションバーの「コミュニティ」をクリックすると、閲覧権限のあるチャンネルリストが表示され、チャンネルをクリックしてチャットに入ります。
3. **メッセージ同期**: 送信されたすべてのメッセージはWebSocket経由でチャンネル内のオンラインユーザーにリアルタイムでプッシュされ、同時にデータベースに書き込まれます。チャンネルに再入すると自動的に履歴メッセージが読み込まれます。

### 拡張チャネル・価格管理

1. **バッチモデル追加**: チャネルエディタまたは価格エディタの「カスタムモデル追加」入力ボックスに `gpt-4, gpt-4o  gpt-4-turbo`（カンマ、スペース、改行いずれも区切り文字として使用可能）と入力し、追加をクリックすると複数モデルを一括追加できます。
2. **JSON価格エディタ**: 「価格管理」ページで「JSON価格エディタ」ボタンをクリック：
   - 開く時に現在のすべての価格ルールがフォーマット済みJSONとして自動エクスポート
   - エディタで直接JSONを編集（リアルタイムJSON形式バリデーション）
   - 「インポートして上書き」をクリックすると編集後の価格ルールがデータベースに上書き保存

## ❓ よくある質問
1. **デプロイ後にページにアクセスしてログイン登録はできるが、チャットが使えない（ずっとローディング）のはなぜ？**
   - チャット等の機能はWebSocketで通信します。サービスがWebSocketをサポートしていることを確認してください。
   - Nginx、Apache等のリバースプロキシを使用している場合、WebSocketサポートが設定されていることを確認してください。
2. **Valkey / Dragonfly と Redis の違いは？どれを選ぶべき？**
   - Valkey は Redis のオープンソースフォークで Linux Foundation が管理；Dragonfly は Redis プロトコル互換のマルチスレッドアーキテクチャキャッシュ。3つとも NeoAI ではアプリケーション層で完全互換 — 必要に応じて自由に切り替え可能。ほとんどのケースではデフォルトの Redis で問題ありません；より高いスループットを求める場合は Dragonfly を；オープンソースガバナンスの持続性を重視する場合は Valkey を選択してください。
3. **コミュニティチャンネルのメッセージはデータベースを消費しますか？**
   - はい。すべてのチャンネルメッセージはMySQLの `community_message` テーブルに永続化され、履歴追跡と監査に使用されます。クリーンアップする場合はデータベースで該当レコードを手動削除してください。
4. **JSON価格エディタのインポートは既存の価格を上書きしますか？**
   - はい。混同を避けるため、JSONエディタのインポートは上書きモードを採用しています。インポート前に現在の価格ルールを参考として自動エクスポートします — 編集前にバックアップすることをお勧めします。
5. **Midjourney Proxy チャネルがずっとローディングまたは `please provide available notify url` エラー**
   - Midjourney Proxy サービスが実行中で、チャネルタイプが OpenAI ではなく Midjourney に設定されていることを確認してください。
   - システム設定の**バックエンドドメイン**が正しく設定されているか確認してください。
6. **デフォルトの Root パスワードを変更するには？**
   - 右上のアバターをクリックしてバックエンドに入り、システム設定 → 一般設定の「Root パスワードを変更」をクリックしてください。

## 📦 技術スタック
- 🥗 フロントエンド: React + Redux + Radix UI + Tailwind CSS
- 🍎 バックエンド: Golang + Gin + Redis/Valkey/Dragonfly + MySQL
- 🍒 アプリケーション技術: PWA + WebSocket

## 📄 ライセンス

本プロジェクトは [CoAI](https://github.com/coaidev/coai) から派生し、**Apache-2.0** オープンソースライセンスを継承します。商用二次開発・配布歓迎 — Apache-2.0 ライセンスの規定を遵守し、違法目的で使用しないでください。

## ❤ 謝辞と寄付

- 🙏 [CoAI.Dev](https://coai.dev) とそのコントリビューターの皆様に、優秀なオープンソース基盤プロジェクトを提供していただき感謝いたします
- このプロジェクトがお役に立ちましたら、Star をお願いします！
