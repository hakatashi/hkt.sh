# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Project Overview

hkt.sh is a serverless short URL/link sharing service deployed on AWS. It provides a public directory of shortened links, a redirect service with access tracking, and an admin panel protected by Google OAuth via Amazon Cognito.

## Build Commands

```bash
make deps       # Download AWS Lambda build tools
make build      # Build all three Lambda functions into .zip deployment packages
make deploy     # Build, deploy via AWS SAM, and sync assets to S3
make clean      # Remove built binaries and .zip files
```

Deployment requires AWS credentials and the SAM CLI. CI/CD is handled by GitHub Actions (`.github/workflows/deploy.yml`) — it builds on every push/PR and deploys only on the master branch.

## Architecture

Three independent Go Lambda functions, each in its own directory with a `main.go`:

- **`home/`** — Serves the root page (`GET /`). Scans DynamoDB for all non-unlisted entries and renders an HTML page using `home/home.html.tpl`. Also handles subdomain redirects (`*.hkt.sh`, `*.hkt.si`). Includes an admin panel with Google OAuth login.
- **`entry/`** — Handles individual short link redirects (`GET /{name}`). Fetches the entry from DynamoDB, increments the access counter, and returns HTTP 301. Returns 404 if not found.
- **`put-entry/`** — Admin endpoint (`PUT /admin/entry`) protected by Cognito authorization. Creates new short URL entries in DynamoDB.

**AWS Infrastructure** (defined in `template.yaml`):
- API Gateway routing to the three Lambda functions
- DynamoDB table `hkt-sh-entries` (partition key: `Name`) storing links
- S3 bucket for frontend assets (CSS, JS, favicon from `assets/`)
- Cognito User Pool with Google OAuth identity provider for admin auth
- X-Ray tracing enabled on all functions

**Environment variables** injected by SAM at runtime:
- `AUTH_USER_POOL_CLIENT_ID` — Cognito client ID
- `ASSETS_WEBSITE_DOMAIN_NAME` — S3 bucket domain for CSS/JS/favicon

## Key Details

- Go module: `github.com/hakatashi/hkt.sh`, requires Go 1.14+
- AWS region: `ap-northeast-1` (Tokyo)
- Authorized admin email is hardcoded in `home/home.html.tpl`
- Alternative domain `hkt.si` is used for platforms that don't auto-link `.sh` domains
- Cache headers on redirects: `private, max-age=90`
- Frontend uses vanilla JS with localStorage for JWT token persistence (OAuth implicit flow)
