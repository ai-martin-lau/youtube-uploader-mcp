[English](README.md) | [简体中文](README_ZH.md) | [日本語](README_JA.md) | [한국어](README_KO.md) | [Español](README_ES.md)

![YouTube Uploader MCP](assets/repository-cover.png)

# YouTube Uploader MCP

一个通过 YouTube Data API v3 在本地运行的 MCP 服务器，用于上传、检查和审计 YouTube 视频，以及执行其支持的上传后操作。

它为 MCP 客户端提供绑定到指定频道的工作流，涵盖 OAuth、优先私密上传、定时发布、缩略图、字幕、播放列表插入和 API 可读信息验证。MCP 客户端可以准备元数据，但本服务器只会发送它收到的值。

> [!IMPORTANT]
> 本 README 说明当前 `main` 分支，其中注册了九个工具。最新标签版本较旧，随附的安装脚本仍指向上游仓库。若要使用本文所述功能，请从源码构建。

## 功能

- 通过 OAuth 2.0 在本地认证 YouTube 频道，并支持缓存多个频道。
- 在涉及频道的操作前，验证 OAuth 当前绑定的实际频道。
- 上传视频，并接收标题、说明、标签、类别、语言、隐私状态、发布时间、受众、合成媒体声明和订阅者通知设置。
- 除非明确提供其他有效状态，否则新上传默认设为 `private`。
- 将已上传的视频添加到播放列表、上传小于 2 MiB 的缩略图，以及插入字幕轨道。
- 使用 `get_video` 读取所有者可见的视频元数据。
- 使用 `audit_video` 按给定预期执行只读检查，包括字幕资源，以及精确的播放列表成员或顺序。
- 使用 ETag 冲突保护，安全更新受支持的受众和合成媒体声明。

## 要求

- Go `1.24.4` 或更高版本。
- 已启用 YouTube Data API v3 的 Google Cloud 项目。
- Google OAuth 桌面客户端 JSON 文件。
- OAuth 应用处于测试阶段时，已将你的 Google 账号添加为测试用户。
- 足以执行相关操作的 YouTube API 配额。
- 支持本地 stdio 服务器的 MCP 客户端。

Google Cloud 配置步骤请参阅 [youtube_oauth2_setup.md](youtube_oauth2_setup.md)。

## 从源码构建

```bash
git clone https://github.com/ai-martin-lau/youtube-uploader-mcp.git
cd youtube-uploader-mcp
go mod download
go build -trimpath -o youtube-uploader-mcp .
```

为令牌缓存和日志创建私有目录，并妥善保护 OAuth 客户端文件：

```bash
mkdir -p /absolute/path/private-state /absolute/path/private-logs
chmod 700 /absolute/path/private-state /absolute/path/private-logs
chmod 600 /absolute/path/client_secret.json
```

## 配置 MCP 客户端

请使用绝对路径。不同 MCP 客户端对配置格式的命名可能不同，但服务器条目相同：

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

保存配置后，重启 MCP 客户端。

## 使用方法

### 1. 连接频道

向 MCP 客户端提出以下请求：

```text
开始 YouTube 身份验证。使用我在 Google OAuth 客户端中配置的重定向 URI。
```

客户端应调用 `authenticate` 并返回 Google 授权 URL。完成 Google 同意流程后，将 localhost 重定向 URL 中的一次性 `code` 查询参数交给本地 `accesstoken` 工具。请将该代码视为敏感信息。

接着验证本地缓存的频道：

```text
列出我已认证的 YouTube 频道，并显示准确的频道 ID。
```

### 2. 以私密状态上传

```text
将 /absolute/path/night-drive.mp4 以私密状态上传到频道 UCxxxxxxxx。

标题：Night Drive Practice
说明：一段原创贝斯练习视频。
标签：practice,bass,original
类别 ID：10
语言：en
面向儿童：false
包含合成媒体：false
通知订阅者：false

暂时不要发布，也不要将其加入播放列表。返回 YouTube 视频 ID 和实际提交的值。
```

请保留返回的视频 ID。如需定时发布，请在最初的 `upload_video` 请求中提供 RFC 3339 格式的 `publish_at` 值；定时发布的视频会先以私密状态提交，直至 YouTube 将其发布。

### 3. 在执行上传后操作前验证

```text
读取频道 UCxxxxxxxx 中的视频 VIDEO_ID。确认所有者频道、隐私状态、标题、类别、语言、受众声明和合成媒体声明。不要更改任何内容。
```

如需按预期值检查：

```text
审计频道 UCxxxxxxxx 中的视频 VIDEO_ID。预期隐私状态为 private、类别为 10、语言为 en、面向儿童为 false、合成媒体为 false。不要修改视频。
```

### 4. 完成受支持的上传后操作

```text
对于频道 UCxxxxxxxx 中的视频 VIDEO_ID：
- 上传缩略图 /absolute/path/thumbnail.jpg
- 将其加入播放列表 PLAYLIST_ID
- 将 /absolute/path/captions.srt 作为英语字幕上传

分别报告每项操作。如果插入播放列表时超时，请在重试前审计实际播放列表。
```

`update_video` 可能只完成部分操作。请始终逐项检查每个请求操作的结果。

## 工具参考

| 工具 | 用途 |
| --- | --- |
| `authenticate` | 创建 Google OAuth 授权 URL。 |
| `accesstoken` | 交换一次性授权代码、验证实际频道，并在本地缓存令牌。 |
| `channels` | 列出本地令牌缓存中的频道。它不会实时搜索 YouTube 频道。 |
| `refreshtoken` | 手动刷新某个频道的缓存令牌。涉及频道的工具也会自动刷新即将到期的令牌。 |
| `upload_video` | 上传本地视频，并发送所提供的元数据、声明、隐私状态、发布时间和通知选项。 |
| `get_video` | 读取所有者可见视频的 API 可见元数据，并验证其频道所有权。 |
| `audit_video` | 在不更改 YouTube 的情况下，将 API 当前可读值与调用方提供的预期值进行比较。 |
| `update_video_metadata` | 通过读取、合并、写入流程和 ETag 保护更新 `self_declared_made_for_kids` 和／或 `contains_synthetic_media`。 |
| `update_video` | 将视频插入播放列表、上传缩略图和／或插入字幕轨道。 |

## 默认值与重要边界

- `upload_video` 默认使用 `private`。
- `publish_at` 必须采用 RFC 3339 格式；定时发布的视频会以私密状态上传。
- `tags` 是以逗号分隔的字符串。
- 缩略图文件必须小于 2 MiB。
- `made_for_kids`、`contains_synthetic_media` 和 `notify_subscribers` 是可选布尔值。明确提供 `false` 时会发送 `false`。
- YouTube 不提供 `notify_subscribers` 的事后回读。上传成功说明请求已被接受，但后续审计无法证明这一单独值，只会将其报告为 `unverifiable`。
- 播放列表插入并非幂等操作。遇到错误或超时后，请在重试前检查或审计实际播放列表，以免产生重复项。
- 支持插入字幕；不支持删除和替换字幕。
- 不支持创建播放列表、移除播放列表项目、重新排序、设置播放列表封面和播放列表语言。
- 本服务器目前不提供上传后修改标题、说明、标签、类别、语言、隐私状态、发布时间及大多数其他字段的功能。
- 自动章节、自动地点或概念、Shorts 混剪、评论审核预设、信息卡、片尾画面和字幕认证等仅限 YouTube Studio 的设置，无法通过本 MCP 可靠读取或更改。对于这类预期，`audit_video` 会报告 `unverifiable`，而不会声称成功。

## 安全与隐私

- 只使用由你控制的 Google OAuth 应用。
- 服务器会请求 `youtube.upload`、`youtube.readonly` 和敏感的 `youtube.force-ssl` 权限范围。最后一项是上传字幕所必需的。
- 访问令牌和刷新令牌会以受限文件权限缓存在 `<working_dir>/.youtube_uploader_channels_cache` 中，并在工具输出中隐藏。
- 一次性 OAuth 授权代码会经过 MCP 客户端。当前的请求日志可能记录该代码，因此请将日志目录设为私有、切勿分享日志，并在不再需要旧日志时将其删除。
- 切勿提交 `client_secret.json`、频道缓存或 MCP 日志。
- 当前 OAuth 实现使用固定的 state 值，并且不使用 PKCE。请只在可信的本地计算机上运行，并避免并发或未经请求的授权流程。

## 项目来源

本项目基于 [anwerj/youtube-uploader-mcp](https://github.com/anwerj/youtube-uploader-mcp)，并继续以 MIT 许可证发布。本仓库保留原始版权声明，同时扩展了绑定频道的验证、明确的上传声明、所有者可见信息读取、只读策略审计和具备冲突感知能力的元数据更新。

## 参与贡献

欢迎提交 Issue 和范围明确的 Pull Request。请说明要更改的 YouTube API 行为，避免在补丁中混入无关重构，并在行为变化时添加或更新测试。

## 许可证

[MIT](LICENSE)
