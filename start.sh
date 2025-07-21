#!/bin/sh
# Start script for [start-local](https://github.com/elastic/start-local)
# Customized for my local development, not for production use.
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "${SCRIPT_DIR}"

set -o allexport
# shellcheck source=./.env
. ./.env
set +o allexport

# Check disk space
available_gb=$(($(df -k / | awk 'NR==2 {print $4}') / 1024 / 1024))
required=$(echo "${ES_LOCAL_DISK_SPACE_REQUIRED}" | grep -Eo '[0-9]+')
if [ "$available_gb" -lt "$required" ]; then
  echo "----------------------------------------------------------------------------"
  echo "WARNING: Disk space is below the ${required} GB limit. Elasticsearch will be"
  echo "executed in read-only mode. Please free up disk space to resolve this issue."
  echo "----------------------------------------------------------------------------"
  echo "Press ENTER to confirm."
  read -r _line
fi

_renew_api_key=false
if [ ! -d ".tmp/es" ] || [ ! -d ".tmp/kibana" ]; then
  rm -rf .tmp/es .tmp/kibana
  mkdir -p .tmp/es .tmp/kibana
  _renew_api_key=true
fi

echo "Starting services..."
docker compose up --wait --remove-orphans

echo "--- Services are now running ---"
if [ "$_renew_api_key" = true ]; then
  # we need to renew the apiKey, remove the old one
  sed -i '' '/^ES_LOCAL_API_KEY=/d' .env
  echo "New Machine: Renewing API key..."
  api_key=$(
    curl -s -u "elastic:$ES_LOCAL_PASSWORD" -X POST "$ES_LOCAL_URL/_security/api_key" -H "Content-Type: application/json" -d "{\"name\": \"es-local\"}" |
      grep -Eo '"encoded":"[A-Za-z0-9+/=]+' | sed 's/"encoded":"//;s/"//g' | tr -d '\n'
  )
  echo "ES_LOCAL_API_KEY=$api_key" >>.env
fi

echo "Elasticsearch URL: ${ES_LOCAL_URL}"
echo "Elasticsearch Username: elastic"
echo "Elasticsearch Password: ${ES_LOCAL_PASSWORD}"
echo "Elasticsearch API Key: ${ES_LOCAL_API_KEY:-Not set}"
echo "Kibana URL: http://localhost:${KIBANA_LOCAL_PORT}"
