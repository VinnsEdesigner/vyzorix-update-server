# Vyzorix Update Server - Manual Setup Guide

This guide walks you through setting up all third-party services required to run the Vyzorix Update Server in production.

---

## Table of Contents

1. [Google Cloud Console - OAuth Setup](#1-google-cloud-console---oauth-setup)
2. [Resend - Email Service Setup](#2-resend---email-service-setup)
3. [Render - Deployment Setup](#3-render---deployment-setup)
4. [Environment Variables Reference](#4-environment-variables-reference)

---

## 1. Google Cloud Console - OAuth Setup

Google OAuth allows users to sign in with their Google account instead of creating a new password.

### 1.1 Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Sign in with your Google account
3. Click **"Select a project"** at the top left
4. Click **"New Project"**
5. **Project name**: `vyzorix-update-server`
6. **Location**: Leave as "No organization" or select your organization
7. Click **Create**
8. Wait for the project to be created (notification appears)

### 1.2 Enable the Google+ API

1. In the left sidebar, click **APIs & Services** → **Library**
2. In the search bar, type: `Google+ API`
3. Click on **Google+ API** in the results
4. Click **Enable**
5. If prompted to create credentials, you can skip - we'll do that manually in step 1.4

> **Note**: If Google+ API is not available, search for "Identity Platform" - this provides the same OAuth functionality.

### 1.3 Configure the OAuth Consent Screen

The OAuth consent screen is what users see when they click "Sign in with Google".

1. Go to **APIs & Services** → **OAuth consent screen**
2. Select **External** as the user type
3. Click **Create**
4. Fill in the required fields:

| Field | Value |
|-------|-------|
| **App name** | `Vyzorix Update Server` |
| **User support email** | `your-email@example.com` |
| **Developer contact email** | `your-email@example.com` |

5. Click **Save and Continue**
6. On the **Scopes** page:
   - Click **Add or Remove Scopes**
   - Check these scopes:
     - [OK] `../auth/userinfo.email` - View your email address
     - [OK] `../auth/userinfo.profile` - View your profile info
   - Click **Update**
7. Click **Save and Continue**
8. On the **Test users** page (optional):
   - Click **Add Users**
   - Enter your email for testing
   - Click **Add**
9. Click **Save and Continue** → **Back to Dashboard**

> **Publishing Status**: For testing, "Testing" status is fine. For production, you'll need to submit for verification. Testing allows up to 100 users until you publish.

### 1.4 Create OAuth Credentials

1. Go to **APIs & Services** → **Credentials**
2. Click **+ Create Credentials** → **OAuth client ID**
3. **Application type**: Select **Web application**
4. **Name**: `Vyzorix Web Client`
5. **Authorized JavaScript origins**: Add your domains

   ```
   https://vyzorix-update-server.onrender.com   (production)
   http://localhost:5173                        (local development)
   ```

6. **Authorized redirect URIs**: Add callback endpoint

   ```
   https://vyzorix-update-server.onrender.com/v1/auth/google/callback
   http://localhost:3000/v1/auth/google/callback
   ```

   > [WARN]️ Important: The redirect URI must match exactly what your app expects. The path `/v1/auth/google/callback` is hardcoded.

7. Click **Create**
8. A modal will appear with your credentials:

   ```
   Your Client ID: 
   1234567890-abcdefghijklmnop.apps.googleusercontent.com

   Your Client Secret:
   GOCSPX-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   ```

9. **Copy and save these values** - you won't be able to see the secret again!

### 1.5 Verify Your Setup

Test that your credentials work:

1. Open your browser to:
   ```
   https://accounts.google.com/o/oauth2/v2/auth?
   client_id=YOUR_CLIENT_ID&
   redirect_uri=https://vyzorix-update-server.onrender.com/v1/auth/google/callback&
   response_type=code&
   scope=email%20profile
   ```
   (Replace `YOUR_CLIENT_ID` with your actual client ID)

2. You should see Google's OAuth consent screen
3. If it works, your setup is correct!

---

## 2. Resend - Email Service Setup

Resend is an email API service that makes it easy to send transactional emails (verification, password reset, etc.).

### 2.1 Create a Resend Account

1. Go to [Resend.com](https://resend.com)
2. Click **Sign Up** (use GitHub or Google to sign up for fastest access)
3. Verify your email address

### 2.2 Create an API Key

1. In the Resend dashboard, click **API Keys** in the left sidebar
2. Click **Create API Key**
3. **Name**: `vyzorix-production` (or similar)
4. **Permissions**: Leave as default (Full Access)
5. Click **Create**
6. **Copy the API key** - it looks like:
   ```
   re_xxxxxxxxxxxxxxxxxxxxxxxxxxxx
   ```

   > [WARN]️ Important: This is the only time you'll see this key. Copy it now!

### 2.3 Add a Domain (Recommended for Production)

For better email deliverability, add your domain:

1. Go to **Domains** in the left sidebar
2. Click **Add Domain**
3. Enter your domain: `vyzorix.app` (or your actual domain)
4. Click **Add Domain**
5. Resend will show you DNS records to add:

   | Type | Name | Value |
   |------|------|-------|
   | TXT | `resend` | `v=spf1 include:spf.resend.io ~all` |
   | DKIM | `resend._domainkey` | `p=...` (long string) |
   | MX | `@` | `feedback-smtp.us-east-1.amazonses.com` |

6. Add these DNS records to your domain registrar
7. Click **Verify** in Resend once DNS is propagated

### 2.4 Update Your Sender Email

1. Go to **Domains** → select your domain
2. Click **Add Recipients** to add sender emails
3. Add an email like `noreply@vyzorix.app`

> For testing without a domain, Resend gives you a sandbox domain automatically. Check your dashboard for the verification email to approve sending.

### 2.5 Test Your Resend Setup

```bash
# Using curl to test your API key
curl -X POST https://api.resend.com/emails \
  -H "Authorization: Bearer re_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "from": "onboarding@resend.dev",
    "to": "your-email@example.com",
    "subject": "Test Email",
    "html": "<p>This is a test!</p>"
  }'
```

---

## 3. Render - Deployment Setup

Render is a cloud platform that hosts our Go backend and provides the server infrastructure.

### 3.1 Create a Render Account

1. Go to [Render.com](https://render.com)
2. Click **Sign Up**
3. Sign up with GitHub (easiest) or email
4. Authorize Render to access your repositories

### 3.2 Connect Your Repository

1. In the Render dashboard, click **New** → **Blueprint**
2. If you haven't created a `render.yaml` file, click **Create a Blueprint**
3. Connect your GitHub account if not already connected
4. Select your repository: `vyzorix-update-server`
5. Click **Connect**

### 3.3 Configure Environment Variables

On the Blueprint settings page, add these environment variables:

| Variable | Value | Notes |
|----------|-------|-------|
| `GOOGLE_OAUTH_CLIENT_ID` | `xxx.apps.googleusercontent.com` | From Google Cloud Console |
| `GOOGLE_OAUTH_CLIENT_SECRET` | `xxx` | From Google Cloud Console |
| `RESEND_API_KEY` | `re_xxx` | From Resend |
| `BASE_URL` | `https://vyzorix-update-server.onrender.com` | Your Render URL |
| `EMAIL_FROM` | `noreply@vyzorix.app` | Your verified domain |
| `EMAIL_FROM_NAME` | `Vyzorix` | Display name |
| `JWT_SECRET` | (generate strong random string) | Use `openssl rand -hex 32` to generate |
| `DATABASE_URL` | `./data/vyzorix.db` | Default for SQLite |

#### Generate a Strong JWT_SECRET:

```bash
# Run this command to generate a secure secret
openssl rand -hex 32

# Example output:
# 8f4e2b1a9c3d7e6f5a4b8c9d2e1f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1
```

### 3.4 Create a Web Service

If not using Blueprint, manually create a web service:

1. Click **New** → **Web Service**
2. Connect your GitHub repository
3. Configure:

| Setting | Value |
|---------|-------|
| **Name** | `vyzorix-update-server` |
| **Region** | Oregon (or closest to you) |
| **Branch** | `main` |
| **Runtime** | `Go` |
| **Build Command** | `go build -o server ./cmd/mockserver` |
| **Start Command** | `./server` |
| **Plan** | Free (for testing) or Starter ($7/month for production) |

4. Add the environment variables from section 3.3
5. Click **Create Web Service**

### 3.5 Configure Health Checks

Render supports health checks to monitor your service:

1. After creating the service, go to **Health Checks** tab
2. Set **Path** to `/health`
3. Set **Interval** to `30` seconds
4. Set **Threshold** to `3`

### 3.6 Custom Domain (Optional)

To use your own domain:

1. Go to your Web Service → **Settings**
2. Scroll to **Custom Domains**
3. Enter your domain: `api.vyzorix.app`
4. Click **Add Domain**
5. Add the DNS records Render provides

### 3.7 Environment Variables in Render

To update environment variables after deployment:

1. Go to your Web Service → **Environment**
2. Click **Add Environment Variable**
3. Enter key-value pairs
4. Changes trigger a redeploy automatically

### 3.8 Monitor Logs

To view application logs:

1. Go to your Web Service
2. Click **Logs** tab
3. Use the search box to filter logs
4. Click **Logs Dashboard** for real-time monitoring

---

## 4. Environment Variables Reference

All environment variables needed for production:

### Authentication & Security

| Variable | Example | Required | Description |
|----------|---------|----------|-------------|
| `JWT_SECRET` | `8f4e2b1a...` | [OK] Yes | Secret key for signing JWTs. Generate with `openssl rand -hex 32` |
| `JWT_DURATION_HOURS` | `168` | No | JWT expiry in hours. Default: `168` (7 days) |

### Google OAuth

| Variable | Example | Required | Description |
|----------|---------|----------|-------------|
| `GOOGLE_OAUTH_CLIENT_ID` | `123456789-xxx.apps.googleusercontent.com` | [OK] Yes | From Google Cloud Console |
| `GOOGLE_OAUTH_CLIENT_SECRET` | `GOCSPX-xxx` | [OK] Yes | From Google Cloud Console |

### Email (Resend)

| Variable | Example | Required | Description |
|----------|---------|----------|-------------|
| `RESEND_API_KEY` | `re_xxx` | [OK] Yes | From Resend dashboard |
| `EMAIL_FROM` | `noreply@vyzorix.app` | No | Default: `noreply@vyzorix.app` |
| `EMAIL_FROM_NAME` | `Vyzorix` | No | Default: `Vyzorix` |

### Security - Token Expiry

| Variable | Example | Required | Description |
|----------|---------|----------|-------------|
| `EMAIL_VERIFY_TOKEN_EXPIRY_HOURS` | `24` | No | Default: `24` hours |
| `PASSWORD_RESET_TOKEN_EXPIRY_MINUTES` | `60` | No | Default: `60` minutes |

### URLs

| Variable | Example | Required | Description |
|----------|---------|----------|-------------|
| `BASE_URL` | `https://vyzorix-api.onrender.com` | [OK] Yes | Your deployment URL |
| `FRONTEND_URL` | `https://vyzorix-app.onrender.com` | No | Your frontend URL |

### Database

| Variable | Example | Required | Description |
|----------|---------|----------|-------------|
| `DATABASE_URL` | `./data/vyzorix.db` | No | SQLite database file. Default: `./data/vyzorix.db` |

### Server

| Variable | Example | Required | Description |
|----------|---------|----------|-------------|
| `PORT` | `3000` | No | Server port. Default: `3000` |
| `NODE_ENV` | `production` | No | Set to `production` for production |
| `ALLOWED_ORIGINS` | `https://vyzorix.app` | No | CORS origins (comma-separated) |

---

## Quick Setup Checklist

Use this checklist to ensure everything is configured:

### Google Cloud Console
- [ ] Created Google Cloud project
- [ ] Enabled Google+ API
- [ ] Configured OAuth consent screen
- [ ] Created Web application credentials
- [ ] Added redirect URIs
- [ ] Copied Client ID and Client Secret

### Resend
- [ ] Created Resend account
- [ ] Created API key
- [ ] Verified domain (or using sandbox)
- [ ] Tested API key with curl

### Render
- [ ] Connected GitHub repository
- [ ] Configured build command: `go build -o server ./cmd/mockserver`
- [ ] Configured start command: `./server`
- [ ] Added all environment variables
- [ ] Verified deployment successful
- [ ] Tested `/health` endpoint

---

## Troubleshooting

### Google OAuth Not Working

**Problem**: Users see "redirect_uri_mismatch" error.

**Solution**:
1. Check that redirect URI in Google Cloud Console matches exactly
2. Include both `http://localhost:3000` for local and your production URL
3. Remove trailing slashes

### Resend Emails Not Sending

**Problem**: Emails fail to send.

**Solution**:
1. Verify your API key is correct
2. Check that `EMAIL_FROM` is a verified domain in Resend
3. Check Resend dashboard for any errors
4. Verify your domain DNS records are correct

### Render Deployment Fails

**Problem**: Build or startup fails on Render.

**Solution**:
1. Check Build Logs for error details
2. Verify `GO_BUILD_COMMAND` and `START_COMMAND` are correct
3. Ensure all environment variables are set
4. Check that your Go version is compatible (1.21+)

### Database Issues

**Problem**: "database is locked" or connection errors.

**Solution**:
1. Use SQLite with WAL mode (default in this project)
2. For production, consider migrating to PostgreSQL
3. Check that `DATABASE_URL` points to a writable path

---

## Support & Resources

- **Google Cloud Console**: https://console.cloud.google.com/
- **Resend Documentation**: https://resend.com/docs
- **Render Documentation**: https://render.com/docs
- **Project Repository**: https://github.com/VinnsEdesigner/vyzorix-update-server

---

*Last updated: 2024*