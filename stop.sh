#!/bin/sh
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "${SCRIPT_DIR}"

if [ "${CI:-false}" = "true" ]; then
  echo "CI environment detected. Tearing down 'test' profile..."
  docker compose --profile test down -v --remove-orphans
else
  echo "Stopping 'dev' profile services..."
  docker compose --profile dev stop
fi