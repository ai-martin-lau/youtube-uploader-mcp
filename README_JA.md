[English](README.md) | [简体中文](README_ZH.md) | [日本語](README_JA.md) | [한국어](README_KO.md) | [Español](README_ES.md)

![YouTube Uploader MCP](assets/repository-cover.png)

# YouTube Uploader MCP

YouTube Data API v3 を通じて、YouTube へのアップロード、確認、監査、および対応しているアップロード後の操作を行うローカル MCP サーバーです。

OAuth、非公開を起点とするアップロード、公開予約、サムネイル、字幕、再生リストへの追加、API で読み取り可能な設定の検証を、MCP クライアントからチャンネルに紐付けて実行できます。メタデータは MCP クライアント側で作成できますが、このサーバーが送信するのは受け取った値だけです。

> [!IMPORTANT]
> この README は、9 個のツールを登録する現在の `main` ブランチについて説明しています。最新のタグ付きリリースはそれより古く、同梱のインストールスクリプトも上流リポジトリを参照しています。ここに記載された機能を使うには、ソースからビルドしてください。

## 機能

- YouTube チャンネルをローカルで OAuth 2.0 認証し、複数チャンネルの認証情報をキャッシュできます。
- チャンネル単位の操作前に、OAuth で認証された実際のチャンネルを検証します。
- タイトル、説明、タグ、カテゴリ、言語、公開設定、公開予約、子ども向け設定、合成メディア申告、登録者通知を指定して動画をアップロードします。
- 有効な別の公開設定を明示しない限り、新規アップロードは `private` になります。
- アップロード済み動画を再生リストに追加し、2 MiB 未満のサムネイルをアップロードし、字幕トラックを挿入します。
- `get_video` で、所有者権限で確認できる動画メタデータを読み取ります。
- `audit_video` で、字幕リソースや再生リストへの正確な登録状況・順序を含む、期待値に基づく読み取り専用の確認を実行します。
- 対応している子ども向け設定と合成メディア申告を、ETag による競合防止付きで安全に更新します。

## 必要なもの

- Go `1.24.4` 以降。
- YouTube Data API v3 を有効にした Google Cloud プロジェクト。
- Google OAuth デスクトップクライアントの JSON ファイル。
- OAuth アプリがテスト中の場合は、Google アカウントをテストユーザーとして追加しておくこと。
- 実行する操作に必要な YouTube API クォータ。
- ローカルの stdio サーバーに対応した MCP クライアント。

Google Cloud の設定手順は [youtube_oauth2_setup.md](youtube_oauth2_setup.md) を参照してください。

## ソースからビルド

```bash
git clone https://github.com/ai-martin-lau/youtube-uploader-mcp.git
cd youtube-uploader-mcp
go mod download
go build -trimpath -o youtube-uploader-mcp .
```

トークンキャッシュとログ用の非公開ディレクトリを作成し、OAuth クライアントファイルも非公開に保ちます。

```bash
mkdir -p /absolute/path/private-state /absolute/path/private-logs
chmod 700 /absolute/path/private-state /absolute/path/private-logs
chmod 600 /absolute/path/client_secret.json
```

## MCP クライアントの設定

絶対パスを使用してください。設定形式の名称は MCP クライアントによって異なる場合がありますが、サーバーの登録内容は同じです。

```json
{
  "mcpServers": {
    "youtube-uploader-mcp": {
      "command": "/absolute/path/youtube-uploader-mcp",
      "args": [
        "-client_secret_file",
        "/absolute/path/client_secret.json",
        "-working_dir",
        "/absolute/path/private-state"
      ],
      "env": {
        "YOUTUBE_UPLOADER_MCP_LOG_DIR": "/absolute/path/private-logs"
      }
    }
  }
}
```

設定を保存したら、MCP クライアントを再起動してください。

## 使い方

### 1. チャンネルを接続する

MCP クライアントに次のように依頼します。

```text
YouTube の認証を開始してください。Google OAuth クライアントに設定したリダイレクト URI を使用してください。
```

クライアントは `authenticate` を呼び出し、Google の認証 URL を返します。Google の同意フローを完了したら、localhost のリダイレクト URL に含まれる一度限りの `code` クエリパラメータを、ローカルの `accesstoken` ツールに渡します。このコードは機密情報として扱ってください。

次に、ローカルにキャッシュされたチャンネルを確認します。

```text
認証済みの YouTube チャンネル一覧と、正確なチャンネル ID を表示してください。
```

### 2. 非公開でアップロードする

```text
/absolute/path/night-drive.mp4 をチャンネル UCxxxxxxxx に private でアップロードしてください。

タイトル: Night Drive Practice
説明: オリジナルのベース練習動画です。
タグ: practice,bass,original
カテゴリ ID: 10
言語: en
子ども向け: false
合成メディアを含む: false
登録者に通知: false

まだ公開せず、再生リストにも追加しないでください。YouTube の動画 ID と、実際に送信した値を返してください。
```

返された動画 ID を保管してください。公開を予約する場合は、最初の `upload_video` リクエストで RFC 3339 形式の `publish_at` を指定します。予約投稿は、YouTube が公開するまで非公開として送信されます。

### 3. アップロード後の操作前に確認する

```text
チャンネル UCxxxxxxxx の動画 VIDEO_ID を読み取ってください。所有チャンネル、公開設定、タイトル、カテゴリ、言語、子ども向け設定、合成メディア申告を確認し、何も変更しないでください。
```

期待値に基づいて確認する場合：

```text
チャンネル UCxxxxxxxx の動画 VIDEO_ID を監査してください。公開設定は private、カテゴリは 10、言語は en、子ども向けは false、合成メディアは false を期待します。動画は変更しないでください。
```

### 4. 対応しているアップロード後の操作を完了する

```text
チャンネル UCxxxxxxxx の動画 VIDEO_ID に対して、次を実行してください。
- サムネイル /absolute/path/thumbnail.jpg をアップロード
- 再生リスト PLAYLIST_ID に追加
- /absolute/path/captions.srt を英語字幕としてアップロード

各操作の結果を個別に報告してください。再生リストへの追加がタイムアウトした場合は、再試行前に実際の再生リストを監査してください。
```

`update_video` は一部の操作だけ成功する場合があります。依頼した各操作の結果を必ず確認してください。

## ツール一覧

| ツール | 用途 |
| --- | --- |
| `authenticate` | Google OAuth 認証 URL を作成します。 |
| `accesstoken` | 一度限りの認証コードを交換し、実際のチャンネルを検証して、トークンをローカルにキャッシュします。 |
| `channels` | ローカルのトークンキャッシュにあるチャンネルを一覧表示します。YouTube 上のチャンネルをリアルタイム検索するツールではありません。 |
| `refreshtoken` | 1 つのチャンネルのキャッシュ済みトークンを手動更新します。チャンネル単位のツールも、有効期限が近いトークンを自動更新します。 |
| `upload_video` | ローカル動画をアップロードし、指定されたメタデータ、申告、公開設定、公開予約、通知設定を送信します。 |
| `get_video` | 所有者権限で確認できる動画について、API から参照可能なメタデータを読み取り、チャンネルの所有関係を検証します。 |
| `audit_video` | YouTube を変更せずに、API から読み取れる実際の値と呼び出し側が指定した期待値を比較します。 |
| `update_video_metadata` | read-merge-write と ETag 保護を使い、`self_declared_made_for_kids` と `contains_synthetic_media` の一方または両方を更新します。 |
| `update_video` | 動画を再生リストに挿入し、サムネイルや字幕トラックをアップロードします。 |

## デフォルト値と重要な制約

- `upload_video` のデフォルトは `private` です。
- `publish_at` は RFC 3339 形式で指定する必要があります。予約投稿の動画は非公開としてアップロードされます。
- `tags` はカンマ区切りの文字列です。
- サムネイルファイルは 2 MiB 未満である必要があります。
- `made_for_kids`、`contains_synthetic_media`、`notify_subscribers` は省略可能な真偽値です。明示した `false` は `false` として送信されます。
- YouTube は、後から読み取れる値として `notify_subscribers` を公開していません。アップロード成功はリクエストが受理されたことを示しますが、後の監査ではこの個別の値を証明できないため、`unverifiable` として報告されます。
- 再生リストへの挿入は冪等ではありません。エラーやタイムアウトの後は、重複を避けるため、再試行前に実際の再生リストを確認または監査してください。
- 字幕の挿入には対応していますが、削除と置換には対応していません。
- 再生リストの作成、項目の削除、並べ替え、再生リストのカバー、再生リストの言語には対応していません。
- タイトル、説明、タグ、カテゴリ、言語、公開設定、公開予約、そのほか大半の項目をアップロード後に変更する機能は、現在このサーバーでは公開されていません。
- 自動チャプター、自動設定される場所やコンセプト、Shorts のリミックス、コメント管理プリセット、カード、終了画面、字幕認定など、YouTube Studio 専用の設定は、この MCP では確実に読み取りまたは変更できません。`audit_video` は成功したと主張せず、こうした期待値を `unverifiable` として報告します。

## セキュリティとプライバシー

- 自分で管理する Google OAuth アプリのみを使用してください。
- サーバーは `youtube.upload`、`youtube.readonly`、機密スコープの `youtube.force-ssl` を要求します。最後のスコープは字幕のアップロードに必要です。
- アクセストークンとリフレッシュトークンは、制限されたファイル権限で `<working_dir>/.youtube_uploader_channels_cache` にローカル保存され、ツールの出力ではマスクされます。
- 一度限りの OAuth 認証コードは MCP クライアントを経由します。現在のリクエストログに記録される可能性があるため、ログディレクトリを非公開にし、ログを共有せず、不要になった古いログは削除してください。
- `client_secret.json`、チャンネルキャッシュ、MCP ログは絶対にコミットしないでください。
- 現在の OAuth 実装では固定の state 値を使用し、PKCE は使用していません。信頼できるローカルマシン上でのみ実行し、同時実行または意図しない認証フローを避けてください。

## プロジェクトの由来

このプロジェクトは [anwerj/youtube-uploader-mcp](https://github.com/anwerj/youtube-uploader-mcp) を基にしており、MIT License の下で引き続き公開されています。このリポジトリは元の著作権表示を維持しながら、チャンネルに紐付けた検証、明示的なアップロード申告、所有者権限で確認できる情報の読み取り、読み取り専用のポリシー監査、競合を考慮したメタデータ更新を追加しています。

## コントリビューション

Issue と焦点を絞った Pull Request を歓迎します。変更する YouTube API の挙動を説明し、無関係なリファクタリングをパッチに含めず、挙動を変更する場合はテストを追加または更新してください。

## ライセンス

[MIT](LICENSE)
