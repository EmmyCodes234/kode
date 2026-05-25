#!/usr/bin/env bash
set -euo pipefail

# Deploy the Kode landing page to trykode.xyz
# Usage: ./deploy.sh [--prod]
#
# Prerequisites:
#   - wrangler CLI installed (for Cloudflare Pages)
#   - OR configured S3/static host

MODE="${1:-preview}"

echo "Building Kode landing page..."

# The site is a single static file - just ensure it's valid
if [ ! -f "web/index.html" ]; then
  echo "Error: web/index.html not found"
  exit 1
fi

echo "  ✓ web/index.html ready ($(wc -c < web/index.html) bytes)"

if command -v wrangler &>/dev/null; then
  echo "Deploying to Cloudflare Pages..."
  
  # Create a temporary directory with just the files we need
  TMPDIR=$(mktemp -d)
  cp web/index.html "$TMPDIR/"
  
  if [ "$MODE" = "--prod" ]; then
    npx wrangler pages deploy "$TMPDIR" --project-name kode --branch main
  else
    npx wrangler pages deploy "$TMPDIR" --project-name kode
  fi
  
  rm -rf "$TMPDIR"
else
  echo "Open web/index.html in your browser to preview locally."
  echo "To deploy, install wrangler: npm install -g wrangler"
fi
