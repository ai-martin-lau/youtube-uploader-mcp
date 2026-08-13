<p align="center"> <img src="https://github.com/user-attachments/assets/21a9baa2-06e8-4af4-9bcd-1dbce52a2733"/> </p>


# YouTube Uploader MCP
[![Trust Score](https://archestra.ai/mcp-catalog/api/badge/quality/anwerj/youtube-uploader-mcp)](https://archestra.ai/mcp-catalog/anwerj__youtube-uploader-mcp)
[![Tests](https://github.com/anwerj/youtube-uploader-mcp/actions/workflows/tests.yaml/badge.svg)](https://github.com/anwerj/youtube-uploader-mcp/actions/workflows/tests.yaml)

AI‑powered YouTube uploader—no CLI, no YouTube Studio, and no secrets ever shared with LLMs or third‑party apps and all free of cost! It includes OAuth2 authentication, token management, and video upload functionality.

## Features

* **Direct Uploads**: Upload videos to YouTube from Claude, Cursor, VS Code, or any other MCP client.
* **AI-Assisted Metadata**: Automatically generate titles, descriptions, and tags via your MCP client.
* **OAuth2 Authentication**: Secure local login, multi-channel support, and auto-refreshing.
* **Metadata & Settings**: Category tags, optional language settings, explicit audience and altered/synthetic-media declarations, and subscriber-notification control.
* **Privacy & Scheduling**: Support for public, private, or unlisted statuses, and scheduled publish times.
* **Post-Upload Configs (`update_video`)**: Add videos to playlists, upload custom thumbnails (<2MB), and attach subtitles.
* **Safe Metadata Audits and Updates**: Read owner-visible video settings with `get_video` and update supported declarations with a read-merge-write workflow.
* **Policy Audits (`audit_video`)**: Compare live owner-visible settings, caption tracks, playlist membership, and exact playlist order against explicit expectations without changing YouTube.

## Single Command Installation

### For Mac and Linux
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/anwerj/youtube-uploader-mcp/master/scripts/install.sh)"
```


### For Windows(Powershell)
```Powershell
Invoke-WebRequest -UseBasicParsing "https://raw.githubusercontent.com/anwerj/youtube-uploader-mcp/master/scripts/install.ps1" -OutFile "$env:TEMP\install.ps1"; PowerShell -NoProfile -ExecutionPolicy Bypass -File "$env:TEMP\install.ps1"
```
### Expected result

This single command will

1. Help in downloading oAuth client secret files, if not downloaded,
2. Download the MCP server,
3. Set minimum required permission to run mcp server,
4. Auto update **Cluade Desktop config** with youtube-uploader-mcp server and
5. At last print exact MCP config for any other clients **VS Code/Cursor/AnythingLLM etc**.

## Demo
### Setup and Demo Video
<p align="center"> <a href="https://youtu.be/fcywz5FIUpM" target="_blank"><img src="https://img.youtube.com/vi/fcywz5FIUpM/0.jpg"/></a> </p>

![output](https://github.com/user-attachments/assets/f8c2c303-ef77-4fa9-99a6-5de7f120ffac)

## Manual Installation
Please check [Single Command Installation](#single-command-installation), proceed if you prefer manual installation.

Visit the [Releases](https://github.com/anwerj/youtube-uploader-mcp/releases) page and download the appropriate binary for your operating system:

- `youtube-uploader-mcp-linux-amd64`
- `youtube-uploader-mcp-darwin-arm64`
- `youtube-uploader-mcp-windows-amd64.exe`
- etc.

> You can use the latest versioned tag, e.g., `v1.0.0`.

---

### 2. Make it Executable (Linux/macOS)

```bash
chmod +x path/to/youtube-uploader-mcp-<os>-<arch>
```

### 3. Configure MCP (e.g., in Claude Desktop or Cursor)
```json
{
  "mcpServers": {
    "youtube-uploader-mcp": {
      "command": "/absolute/path/to/youtube-uploader-mcp-<os>-<arch>",
      "args": [
        "-client_secret_file",
        "/absolute/path/to/client_secret.json(See Below)"
      ]
    }
  }
}
```
### 4. Set Up Google OAuth 2.0
To upload to YouTube, you must configure OAuth and get a client_secret.json file from the Google Developer Console.

➡️ Follow the guide in [youtube_oauth2_setup.md](./youtube_oauth2_setup.md) for a step-by-step walkthrough.

> [!WARNING]
> **Important: Sensitive OAuth Scope & Re-Authentication**:
> Subtitle/caption upload relies on the `https://www.googleapis.com/auth/youtube.force-ssl` scope, which is required by the YouTube Data API to perform caption write operations. 
> * **Existing Users**: If you authenticated with an older version of this MCP, your saved token will lack this scope. Any calls attempting to upload subtitles will fail with a `403 Forbidden` API error. 
> * **How to Update**: You must delete the channel cache file (usually `.youtube_uploader_channels_cache` in your working directory) and rerun the `authenticate` tool to grant the new permissions.
> * **Google Console Warning**: Adding this sensitive scope will trigger a warning screen in Google Cloud Console. Since this is for your personal use, you can safely proceed past the "unverified app" warnings.

### Usage & Tool Orchestration

The MCP server registers the following tools:
1. `authenticate`: Generates the OAuth2 URL for authentication.
2. `accesstoken`: Exchanges the code for user credentials and channel info.
3. `channels`: Retrieves authenticated channels.
4. `refreshtoken`: Force-refreshes tokens.
5. `upload_video`: Uploads the video file and configures main details (title, description, tags, category, optional language, status, audience declaration, altered/synthetic-media disclosure, subscriber notifications, and scheduled publish).
6. `update_video`: Decoupled tool that manages post-upload configurations: adds the video to a playlist, uploads a custom thumbnail (must be <2MB), and attaches subtitle/caption tracks.
7. `get_video`: Reads owner-visible metadata for one video and verifies that it belongs to the selected channel.
8. `audit_video`: Performs a read-only policy audit. API-readable checks return `pass` or `fail`; upload-time and Studio-only settings return `unverifiable` rather than a false pass. It can also detect missing or duplicate playlist membership and compare an entire playlist's ordered video IDs.
9. `update_video_metadata`: Safely updates supported video declarations after first reading and preserving the video's other writable status fields.

Channel-scoped tools verify the OAuth token's live `mine=true` channel before
continuing. Metadata updates also send the video's current ETag with
`If-Match`, so a concurrent Studio edit aborts the update instead of being
silently overwritten.

`made_for_kids`, `contains_synthetic_media`, and `notify_subscribers` are
optional booleans. Omission preserves YouTube's default behavior; an explicit
`false` is sent as an explicit declaration.

YouTube Data API v3 does not expose reliable write operations for every
YouTube Studio control. Automatic chapters, automatic places, automatic
concepts, Shorts remixing, the per-video comment moderation preset, cards, and
end screens remain manual Studio settings.

### Defaults and automation boundary

The MCP itself defaults a new upload to `private`. Low End Atlas-style
automation should still explicitly send language, audience declaration,
altered/synthetic-media declaration, and `notify_subscribers` instead of
depending on account state. In particular, YouTube's `videos.insert` default
for `notifySubscribers` is `true`.

Some platform defaults are useful but are not substitutes for a live audit:

* The standard YouTube license is the upload default.
* Paid product placement is `false` unless declared.
* Shorts remixing is enabled by default, subject to rights restrictions.
* Automatic chapters are enabled for new uploads by default, so omission does
  not satisfy a policy that requires chapters to be off.
* YouTube Studio upload defaults apply only to browser uploads, not API
  uploads.

`audit_video` accepts only expectations supplied by the caller. This keeps the
server reusable while allowing a project manifest or agent workflow to run the
same checks automatically. For playlist writes, use
`expected_playlist_contents` with the complete ordered video-ID list before
deciding whether an insertion may be retried.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) to contribute.
