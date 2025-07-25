#!/bin/sh
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "${SCRIPT_DIR}/.."

PROJECT_NAME=${COMPOSE_PROJECT_NAME:-$(basename "$SCRIPT_DIR")}
VOLUME_NAME="${PROJECT_NAME}_dev-elasticsearch"
_renew_api_key=false

echo "Checking for Docker volume: $VOLUME_NAME..."
if ! docker volume ls -q | grep -q "^${VOLUME_NAME}$"; then
    echo "Volume not found. A new API key will be generated."
    # Remove the old key from .env if it exists
    sed -i'.bak' '/^ES_LOCAL_API_KEY=/d' .env && rm .env.bak
    _renew_api_key=true
fi

PROFILE="dev"
if [ "${CI:-false}" = "true" ]; then
    echo "CI environment detected. Using 'test' profile."
    PROFILE="test"
    if [ "$(uname -m)" != "arm64" ]; then
        echo "Removing arm64 architecture from .env file."
        sed -i'.bak' '/^ES_LOCAL_VERSION=/s/-arm64//g' .env && rm .env.bak
    fi
fi

set -o allexport
# shellcheck source=./.env
[ -f .env ] && . ./.env
set +o allexport

echo "Starting services with profile: '$PROFILE'..."
docker compose --profile "$PROFILE" up --remove-orphans --wait --wait-timeout 300

# --- API Key Generation Logic ---
if { [ "$_renew_api_key" = true ] || ! grep -q "ES_LOCAL_API_KEY" .env; }; then
    echo "Renewing API key..."
    api_key=$(curl -s -u "elastic:$ES_LOCAL_PASSWORD" -X POST "http://localhost:${ES_LOCAL_PORT}/_security/api_key" -H "Content-Type: application/json" -d "{\"name\": \"es-local\"}" | grep -Eo '"encoded":"[A-Za-z0-9+/=]+' | sed 's/"encoded":"//')
    if [ -n "$api_key" ]; then
        echo "ES_LOCAL_API_KEY=$api_key" >>.env
        echo "API Key generated and saved to .env"
    else
        echo "Warning: Failed to generate API key."
    fi
fi
