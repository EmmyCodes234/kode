#!/usr/bin/env bash
set -euo pipefail

DEPLOY_MSG="${1:-"deploy: update landing page"}"

if ! command -v netlify &>/dev/null; then
  echo "Installing Netlify CLI..."
  npm install -g netlify-cli
fi

echo "Deploying Kode landing page to Netlify..."
echo "  Source: web/index.html + netlify.toml + _redirects"
echo "  Message: $DEPLOY_MSG"

netlify deploy \
  --dir . \
  --functions .netlify/functions \
  --message "$DEPLOY_MSG" \
  --prod
