# YouTube metadata controls design

## Goal

Extend the MCP with the smallest safe set of controls needed for repeatable
Low End Atlas uploads while keeping the server useful for other channels.

The upgrade must:

- record the channel owner's audience declaration correctly;
- disclose altered or synthetic media explicitly when requested;
- let callers choose whether a new upload notifies subscribers;
- expose a read-only video metadata inspection tool; and
- update existing video metadata without silently clearing unrelated fields.

## Selected approach

Keep upload settings on `upload_video`, add a read-only `get_video` tool, and
add a separate `update_video_metadata` tool. Do not add metadata mutations to
the existing `update_video` tool because that tool can also perform
non-idempotent playlist insertion and file uploads.

### Upload behavior

- Retain the public argument name `made_for_kids` for compatibility, but map it
  to YouTube's writable `status.selfDeclaredMadeForKids` field.
- Preserve argument presence. Omitted booleans leave YouTube defaults alone;
  explicit `false` values are sent using the Go client's `ForceSendFields`.
- Add optional `contains_synthetic_media` and send it with the same presence
  semantics.
- Add optional `notify_subscribers` and apply it to the `videos.insert` call.
  Omission preserves YouTube's default; callers can explicitly send `false`.

### Read behavior

`get_video` requires `channel_id` and `video_id`. It authenticates with the
selected channel, reads the owner-visible metadata, verifies that the returned
video belongs to the requested channel, and returns a sanitized DTO rather than
the raw Google API object.

The DTO includes the fields needed to audit uploads: title, description, tags,
category, language, privacy and schedule, license and embedding flags, audience
declaration, altered/synthetic media disclosure, paid product placement, and
recording date.

### Update behavior

`update_video_metadata` requires `channel_id`, `video_id`, and at least one
supported patch field. The first release supports:

- `self_declared_made_for_kids`
- `contains_synthetic_media`

The tool performs a read, verifies channel ownership, constructs a clean
status payload containing the current writable status values, merges only the
requested patch, updates the `status` part, then reads again and returns the
live result.

This read-merge-update-read sequence is required because `videos.update` is a
replacement operation for each included part, not a JSON PATCH operation.

## Validation and safety

- Reject missing videos and channel-owner mismatches before writing.
- Reject an empty metadata patch.
- Preserve explicit `false` values in both insert and update requests.
- Do not return OAuth credentials or token-cache contents.
- Cover requests and responses with deterministic HTTP fixtures.
- Run the full Go test suite and build the Darwin arm64 binary.

## Deliberately unsupported Studio settings

YouTube Data API v3 does not expose reliable write operations for automatic
chapters, automatic places, automatic concepts, Shorts remixing, the per-video
comment moderation preset, cards, or end screens. These remain a documented
manual Studio checklist rather than being implemented through browser
automation.

## Release and repository plan

- Preserve the upstream Git history and publish the changes from a feature
  branch in the public `ai-martin-lau/youtube-uploader-mcp` fork.
- Keep the project-local installed binary pinned and checksummed.
- Replace the installed binary only after tests and a reproducible build pass.
- Record the fork commit and local binary checksum in the Low End Atlas MCP
  installation notes.
