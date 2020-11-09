#!/bin/bash
set -a

HOST=:9001

DB_HOST=localhost
DB_PORT=26257
DB_USER=root
DB_PASS=example
DB_NAME=codecoach_dev

GITHUB_CLIENT_ID=$CC_GITHUB_CLIENT_ID
GITHUB_CLIENT_SECRET=$CC_GITHUB_CLIENT_SECRET
GITHUB_REDIRECT_URI=http://localhost:9001/api/v1/oauth/callback

set +a
