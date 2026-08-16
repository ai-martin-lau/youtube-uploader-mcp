[English](README.md) | [简体中文](README_ZH.md) | [日本語](README_JA.md) | [한국어](README_KO.md) | [Español](README_ES.md)

# YouTube Uploader MCP

YouTube Data API v3를 통해 YouTube 동영상을 업로드하고, 조회하고, 감사하며, 지원되는 업로드 후 작업을 수행하는 로컬 MCP 서버입니다.

MCP 클라이언트에 OAuth, 비공개 우선 업로드, 예약 게시, 썸네일, 자막, 재생목록 추가, API로 확인 가능한 항목의 검증을 위한 채널 단위 워크플로를 제공합니다. 메타데이터는 MCP 클라이언트가 준비할 수 있지만, 이 서버는 전달받은 값만 전송합니다.

> [!IMPORTANT]
> 이 README는 9개의 도구를 등록하는 현재 `main` 브랜치를 설명합니다. 최신 태그 릴리스는 더 오래된 버전이며, 포함된 설치 스크립트는 여전히 업스트림 저장소를 대상으로 합니다. 여기에 설명된 기능을 사용하려면 소스에서 빌드하세요.

## 주요 기능

- OAuth 2.0으로 YouTube 채널을 로컬에서 인증하고 캐시된 여러 채널을 지원합니다.
- 채널 단위 작업 전에 실제 OAuth 채널을 검증합니다.
- 제목, 설명, 태그, 카테고리, 언어, 공개 상태, 예약 시간, 아동용 여부, 합성 미디어 여부, 구독자 알림 입력값을 사용해 동영상을 업로드합니다.
- 유효한 다른 상태를 명시하지 않으면 새 업로드의 기본값은 `private`입니다.
- 업로드한 동영상을 재생목록에 추가하고, 2 MiB 미만의 썸네일을 업로드하며, 자막 트랙을 삽입합니다.
- `get_video`로 소유자에게 표시되는 동영상 메타데이터를 읽습니다.
- `audit_video`로 자막 리소스와 정확한 재생목록 포함 여부 또는 순서를 포함한 기대값 기반 읽기 전용 검사를 실행합니다.
- ETag 충돌 방지 기능으로 지원되는 아동용 및 합성 미디어 선언을 안전하게 업데이트합니다.

## 요구 사항

- Go `1.24.4` 이상.
- YouTube Data API v3가 활성화된 Google Cloud 프로젝트.
- Google OAuth 데스크톱 클라이언트 JSON 파일.
- OAuth 앱이 테스트 상태인 동안 테스트 사용자로 추가된 Google 계정.
- 실행할 작업에 필요한 YouTube API 할당량.
- 로컬 stdio 서버를 지원하는 MCP 클라이언트.

Google Cloud 설정 절차는 [youtube_oauth2_setup.md](youtube_oauth2_setup.md)를 참고하세요.

## 소스에서 빌드

```bash
git clone https://github.com/ai-martin-lau/youtube-uploader-mcp.git
cd youtube-uploader-mcp
go mod download
go build -trimpath -o youtube-uploader-mcp .
```

토큰 캐시와 로그를 위한 비공개 디렉터리를 만들고 OAuth 클라이언트 파일도 비공개로 유지하세요.

```bash
mkdir -p /absolute/path/private-state /absolute/path/private-logs
chmod 700 /absolute/path/private-state /absolute/path/private-logs
chmod 600 /absolute/path/client_secret.json
```

## MCP 클라이언트 설정

절대 경로를 사용하세요. MCP 클라이언트에 따라 설정 형식의 이름은 다를 수 있지만, 서버 항목은 동일합니다.

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

설정을 저장한 뒤 MCP 클라이언트를 다시 시작하세요.

## 사용 방법

### 1. 채널 연결

MCP 클라이언트에 다음과 같이 요청하세요.

```text
YouTube 인증을 시작하세요. 내 Google OAuth 클라이언트에 설정된 리디렉션 URI를 사용하세요.
```

클라이언트는 `authenticate`를 호출하고 Google 인증 URL을 반환해야 합니다. Google 동의 절차를 완료한 뒤 localhost 리디렉션 URL의 일회용 `code` 쿼리 매개변수를 로컬 `accesstoken` 도구에 전달하세요. 이 코드는 민감한 정보로 취급하세요.

다음으로 로컬에 캐시된 채널을 확인하세요.

```text
인증된 YouTube 채널을 나열하고 정확한 채널 ID를 표시하세요.
```

### 2. 비공개로 업로드

```text
/absolute/path/night-drive.mp4를 채널 UCxxxxxxxx에 비공개로 업로드하세요.

제목: Night Drive Practice
설명: 오리지널 베이스 연습 동영상입니다.
태그: practice,bass,original
카테고리 ID: 10
언어: en
아동용: false
합성 미디어 포함: false
구독자 알림: false

아직 게시하거나 재생목록에 추가하지 마세요. YouTube 동영상 ID와 실제로 전송된 값을 반환하세요.
```

반환된 동영상 ID를 보관하세요. 동영상 게시를 예약하려면 원래 `upload_video` 요청에 RFC 3339 형식의 `publish_at` 값을 제공하세요. 예약 업로드는 YouTube가 게시할 때까지 비공개 상태로 제출됩니다.

### 3. 업로드 후 작업 전에 확인

```text
채널 UCxxxxxxxx의 동영상 VIDEO_ID를 읽으세요. 소유 채널, 공개 상태, 제목, 카테고리, 언어, 아동용 선언, 합성 미디어 선언을 확인하세요. 아무것도 변경하지 마세요.
```

기대값을 기준으로 검사하려면 다음과 같이 요청하세요.

```text
채널 UCxxxxxxxx의 동영상 VIDEO_ID를 감사하세요. 공개 상태는 private, 카테고리는 10, 언어는 en, 아동용은 false, 합성 미디어는 false여야 합니다. 동영상을 수정하지 마세요.
```

### 4. 지원되는 업로드 후 작업 완료

```text
채널 UCxxxxxxxx의 동영상 VIDEO_ID에 대해 다음을 실행하세요.
- 썸네일 /absolute/path/thumbnail.jpg 업로드
- 재생목록 PLAYLIST_ID에 추가
- /absolute/path/captions.srt를 영어 자막으로 업로드

각 작업의 결과를 별도로 보고하세요. 재생목록 삽입 시간이 초과되면 재시도하기 전에 실제 재생목록을 감사하세요.
```

`update_video`는 일부 작업만 성공할 수 있습니다. 요청한 각 작업의 결과를 항상 개별적으로 확인하세요.

## 도구 참조

| 도구 | 용도 |
| --- | --- |
| `authenticate` | Google OAuth 인증 URL을 생성합니다. |
| `accesstoken` | 일회용 인증 코드를 교환하고 실제 채널을 검증한 뒤 토큰을 로컬에 캐시합니다. |
| `channels` | 로컬 토큰 캐시에서 찾은 채널을 나열합니다. 실시간 YouTube 채널 검색이 아닙니다. |
| `refreshtoken` | 한 채널의 캐시된 토큰을 수동으로 갱신합니다. 채널 단위 도구도 만료가 임박한 토큰을 자동으로 갱신합니다. |
| `upload_video` | 로컬 동영상을 업로드하고 제공된 메타데이터, 선언, 공개 상태, 예약 시간, 알림 선택값을 전송합니다. |
| `get_video` | 소유자에게 표시되는 동영상의 API 노출 메타데이터를 읽고 채널 소유권을 검증합니다. |
| `audit_video` | YouTube를 변경하지 않고 실제 API 조회 가능 값을 호출자가 제공한 기대값과 비교합니다. |
| `update_video_metadata` | 읽기-병합-쓰기 방식과 ETag 보호를 사용해 `self_declared_made_for_kids` 및/또는 `contains_synthetic_media`를 업데이트합니다. |
| `update_video` | 동영상을 재생목록에 삽입하고, 썸네일을 업로드하며, 자막 트랙을 삽입합니다. |

## 기본값 및 중요 제한 사항

- `upload_video`의 기본값은 `private`입니다.
- `publish_at`은 RFC 3339 형식이어야 하며, 예약 동영상은 비공개로 업로드됩니다.
- `tags`는 쉼표로 구분된 문자열입니다.
- 썸네일 파일은 2 MiB보다 작아야 합니다.
- `made_for_kids`, `contains_synthetic_media`, `notify_subscribers`는 선택적 불리언 값입니다. 명시적인 `false`는 `false`로 전송됩니다.
- YouTube는 나중에 다시 읽을 수 있도록 `notify_subscribers` 값을 제공하지 않습니다. 업로드 성공은 요청이 수락되었음을 확인하지만, 이후 감사에서는 해당 개별 값을 증명할 수 없으므로 `unverifiable`로 보고합니다.
- 재생목록 삽입은 멱등적이지 않습니다. 오류나 시간 초과가 발생한 뒤에는 중복을 피하기 위해 재시도 전에 실제 재생목록을 조회하거나 감사하세요.
- 자막 삽입은 지원하지만 자막 삭제와 교체는 지원하지 않습니다.
- 재생목록 생성, 항목 제거, 순서 변경, 재생목록 커버, 재생목록 언어는 지원하지 않습니다.
- 업로드 후 제목, 설명, 태그, 카테고리, 언어, 공개 상태, 예약 시간 및 대부분의 다른 필드를 변경하는 기능은 현재 이 서버에서 제공하지 않습니다.
- 자동 챕터, 자동 장소 또는 개념, Shorts 리믹스, 댓글 검토 사전 설정, 카드, 최종 화면, 자막 인증과 같은 YouTube Studio 전용 설정은 이 MCP를 통해 안정적으로 읽거나 변경할 수 없습니다. `audit_video`는 성공을 주장하는 대신 이러한 기대값을 `unverifiable`로 보고합니다.

## 보안 및 개인정보 보호

- 본인이 관리하는 Google OAuth 앱만 사용하세요.
- 서버는 `youtube.upload`, `youtube.readonly`, 민감한 `youtube.force-ssl` 범위를 요청합니다. 마지막 범위는 자막 업로드에 필요합니다.
- 액세스 토큰과 갱신 토큰은 제한된 파일 권한으로 `<working_dir>/.youtube_uploader_channels_cache`에 로컬 캐시되며 도구 출력에서는 마스킹됩니다.
- 일회용 OAuth 인증 코드는 MCP 클라이언트를 통과합니다. 현재 요청 로깅에 이 코드가 기록될 수 있으므로 로그 디렉터리를 비공개로 유지하고, 로그를 공유하지 말며, 더 이상 필요하지 않은 오래된 로그는 삭제하세요.
- `client_secret.json`, 채널 캐시 또는 MCP 로그를 커밋하지 마세요.
- 현재 OAuth 구현은 고정된 state 값을 사용하며 PKCE를 사용하지 않습니다. 신뢰할 수 있는 로컬 컴퓨터에서만 실행하고 동시 또는 요청하지 않은 인증 흐름을 피하세요.

## 프로젝트 출처

이 프로젝트는 [anwerj/youtube-uploader-mcp](https://github.com/anwerj/youtube-uploader-mcp)를 기반으로 하며 MIT License에 따라 계속 제공됩니다. 이 저장소는 원래 저작권 고지를 유지하면서 채널 단위 검증, 명시적 업로드 선언, 소유자에게 표시되는 데이터 조회, 읽기 전용 정책 감사, 충돌을 인식하는 메타데이터 업데이트 기능을 확장했습니다.

## 기여

이슈와 범위가 명확한 풀 리퀘스트를 환영합니다. 변경되는 YouTube API 동작을 설명하고, 패치에 관련 없는 리팩터링을 포함하지 않으며, 동작이 변경될 때 테스트를 추가하거나 업데이트해 주세요.

## 라이선스

[MIT](LICENSE)
