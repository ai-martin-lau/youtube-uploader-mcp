# 🎥 YouTube OAuth2 Setup Guide

To allow `youtube-uploader-mcp` to upload videos to your YouTube account, you'll need to create OAuth 2.0 credentials on Google Cloud and download a `client_secret.json` file.

---

## ✅ Step 1: Create a Project on Google Cloud Console

1. Go to [Google Cloud Console](https://console.cloud.google.com/).
2. From the top bar, click on the **project dropdown**, then click **New Project**.
3. Name the project something like `YouTubeUploader`, and click **Create**.

---

## ✅ Step 2: Enable the YouTube Data API

1. With your new project selected, go to: [YouTube Data API v3](https://console.cloud.google.com/apis/library/youtube.googleapis.com)
2. Click **Enable**.

---

## ✅ Step 3: Create OAuth 2.0 Credentials

1. Navigate to: [APIs & Services > Credentials](https://console.cloud.google.com/apis/credentials)
2. Click **Create Credentials** > **OAuth client ID**.
3. If prompted to configure consent screen, do that first:
   - Select **External**.
   - Enter your app name, email, and scroll down to save.
   - You don’t need to manually add scopes here; the app will request them during authentication.
4. Now choose:
   - **Application Type**: `Desktop App`
   - **Name**: e.g., `YouTubeUploaderMCP`
5. Click **Create**

This app requests these YouTube scopes during OAuth:

- `https://www.googleapis.com/auth/youtube.upload`
- `https://www.googleapis.com/auth/youtube.readonly`
- `https://www.googleapis.com/auth/youtube.force-ssl` (required for subtitle/caption upload via API)

## ✅ Step 4: Add your email to Test Audience
This step is necessary to get Google authenticating your application.
1. Navigate to [Audience](https://console.cloud.google.com/auth/audience)
2. Add your email(with youtube account) to **Test Users** section.

---

## ✅ Step 5: Download `client_secret.json`

1. After creating the OAuth client, you'll see a dialog with the **Client ID** and **Client Secret**.
2. Click **Download JSON**.
3. Save it to a known and safe location, for example: client_secret_***********.json
