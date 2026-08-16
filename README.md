[English](README.md) | [简体中文](README_ZH.md) | [日本語](README_JA.md) | [한국어](README_KO.md) | [Español](README_ES.md)

# YouTube Uploader MCP

A local MCP server for uploading, inspecting, auditing, and completing supported post-upload actions on YouTube through the YouTube Data API v3.

It gives MCP clients a channel-bound workflow for OAuth, private-first uploads, scheduling, thumbnails, captions, playlist insertion, and API-readable verification. Metadata can be prepared by the MCP client, but this server only sends the values it receives.

> [!IMPORTANT]
> This README documents the current `main` branch, which registers nine tools. The latest tagged release is older, and the bundled install scripts still target the upstream repository. Build from source to use the features documented here.

## What it does

- Authenticates YouTube channels locally with OAuth 2.0 and supports multiple cached channels.
- Verifies the live OAuth channel before channel-scoped operations.
- Uploads videos with title, description, tags, category, language, privacy, schedule, audience, synthetic-media, and subscriber-notification inputs.
- Defaults new uploads to `private` unless another valid status is explicitly supplied.
- Adds an uploaded video to a playlist, uploads a thumbnail under 2 MiB, and inserts a caption track.
- Reads owner-visible video metadata with `get_video`.
- Runs read-only, expectation-driven checks with `audit_video`, including caption resources and exact playlist membership or order.
- Safely updates the supported audience and synthetic-media declarations with ETag conflict protection.

## Requirements

- Go `1.24.4` or newer.
- A Google Cloud project with the YouTube Data API v3 enabled.
- A Google OAuth desktop client JSON file.
- Your Google account added as a test user while the OAuth app is in testing.
- YouTube API quota for the operations you run.
- An MCP client that supports local stdio servers.

See [youtube_oauth2_setup.md](youtube_oauth2_setup.md) for the Google Cloud setup walkthrough.

## Build from source

```bash
git clone https://github.com/ai-martin-lau/youtube-uploader-mcp.git
cd youtube-uploader-mcp
go mod download
go build -trimpath -o youtube-uploader-mcp .
```

Create private directories for the token cache and logs, and keep the OAuth client file private:

```bash
mkdir -p /absolute/path/private-state /absolute/path/private-logs
chmod 700 /absolute/path/private-state /absolute/path/private-logs
chmod 600 /absolute/path/client_secret.json
```

## Configure your MCP client

Use absolute paths. The configuration format may be named differently by your MCP client, but the server entry is the same:

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

Restart the MCP client after saving its configuration.

## How to use

### 1. Connect a channel

Ask your MCP client:

```text
Start YouTube authentication. Use the redirect URI configured in my Google OAuth client.
```

The client should call `authenticate` and return a Google authorization URL. Complete the Google consent flow, then give the one-time `code` query parameter from the localhost redirect URL to the local `accesstoken` tool. Treat that code as sensitive.

Next, verify the locally cached channel:

```text
List my authenticated YouTube channels and show the exact channel ID.
```

### 2. Upload privately

```text
Upload /absolute/path/night-drive.mp4 to channel UCxxxxxxxx as private.

Title: Night Drive Practice
Description: An original bass practice video.
Tags: practice,bass,original
Category ID: 10
Language: en
Made for kids: false
Contains synthetic media: false
Notify subscribers: false

Do not publish it or add it to a playlist yet. Return the YouTube video ID and the values actually submitted.
```

Keep the returned video ID. To schedule a video, provide an RFC 3339 `publish_at` value in the original `upload_video` request; scheduled uploads are submitted as private until YouTube publishes them.

### 3. Verify before post-upload actions

```text
Read video VIDEO_ID from channel UCxxxxxxxx. Confirm the owner channel, privacy status, title, category, language, audience declaration, and synthetic-media declaration. Do not change anything.
```

For an expectation-based check:

```text
Audit video VIDEO_ID on channel UCxxxxxxxx. Expect private visibility, category 10, language en, made for kids false, and synthetic media false. Do not modify the video.
```

### 4. Complete supported post-upload actions

```text
For video VIDEO_ID on channel UCxxxxxxxx:
- upload thumbnail /absolute/path/thumbnail.jpg
- add it to playlist PLAYLIST_ID
- upload /absolute/path/captions.srt as English captions

Report each action separately. If playlist insertion times out, audit the live playlist before retrying.
```

`update_video` can partially succeed. Always inspect the result for each requested action.

## Tool reference

| Tool | Purpose |
| --- | --- |
| `authenticate` | Creates the Google OAuth authorization URL. |
| `accesstoken` | Exchanges the one-time authorization code, verifies the live channel, and caches the token locally. |
| `channels` | Lists channels found in the local token cache. It is not a live YouTube channel search. |
| `refreshtoken` | Manually refreshes the cached token for one channel. Channel-scoped tools also refresh near-expiry tokens automatically. |
| `upload_video` | Uploads a local video and sends the supplied metadata, declarations, privacy status, schedule, and notification choice. |
| `get_video` | Reads API-visible metadata for an owner-visible video and verifies its channel ownership. |
| `audit_video` | Compares live API-readable values against caller-supplied expectations without changing YouTube. |
| `update_video_metadata` | Updates `self_declared_made_for_kids` and/or `contains_synthetic_media` using read-merge-write and ETag protection. |
| `update_video` | Inserts the video into a playlist, uploads a thumbnail, and/or inserts a caption track. |

## Defaults and important boundaries

- `upload_video` defaults to `private`.
- `publish_at` must be RFC 3339; scheduled videos are uploaded as private.
- `tags` is a comma-separated string.
- The thumbnail file must be smaller than 2 MiB.
- `made_for_kids`, `contains_synthetic_media`, and `notify_subscribers` are optional booleans. An explicit `false` is sent as `false`.
- YouTube does not expose `notify_subscribers` for later readback. A successful upload confirms that the request was accepted, but a later audit cannot prove that individual value and reports it as `unverifiable`.
- Playlist insertion is not idempotent. After an error or timeout, inspect or audit the live playlist before retrying to avoid duplicates.
- Caption insertion is supported; caption deletion and replacement are not.
- Playlist creation, item removal, reordering, playlist covers, and playlist language are not supported.
- Post-upload changes to title, description, tags, category, language, privacy, schedule, and most other fields are not currently exposed by this server.
- YouTube Studio-only controls such as automatic chapters, automatic places or concepts, Shorts remixing, comment moderation presets, cards, end screens, and caption certification cannot be reliably read or changed through this MCP. `audit_video` reports such expectations as `unverifiable` instead of claiming success.

## Security and privacy

- Use only a Google OAuth app that you control.
- The server requests `youtube.upload`, `youtube.readonly`, and the sensitive `youtube.force-ssl` scope. The last scope is required for caption uploads.
- Access and refresh tokens are cached locally in `<working_dir>/.youtube_uploader_channels_cache` with restricted file permissions and are masked in tool output.
- The one-time OAuth authorization code passes through the MCP client. Current request logging can record it, so keep the log directory private, never share logs, and delete old logs when they are no longer needed.
- Never commit `client_secret.json`, the channel cache, or MCP logs.
- The current OAuth implementation uses a fixed state value and does not use PKCE. Run it only on a trusted local machine and avoid concurrent or unsolicited authorization flows.

## Project origin

This project is based on [anwerj/youtube-uploader-mcp](https://github.com/anwerj/youtube-uploader-mcp) and remains available under the MIT License. This repository keeps the original copyright notice while extending the server with channel-bound verification, explicit upload declarations, owner-visible reads, read-only policy audits, and conflict-aware metadata updates.

## Contributing

Issues and focused pull requests are welcome. Please describe the YouTube API behavior being changed, keep unrelated refactors out of the patch, and add or update tests when behavior changes.

## License

[MIT](LICENSE)
