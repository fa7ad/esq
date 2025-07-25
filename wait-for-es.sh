#!/bin/sh
set -eu

# This script requires the ES_LOCAL_PASSWORD to be set in the environment.
if [ -z "${ES_LOCAL_PASSWORD}" ]; then
  echo "❌ Error: ES_LOCAL_PASSWORD environment variable is not set."
  exit 1
fi

# 1. Wait for the Elasticsearch API to become available.
until curl -s -u "elastic:${ES_LOCAL_PASSWORD}" "http://localhost:9200" >/dev/null; do
  echo "Waiting for Elasticsearch API at http://localhost:9200..."
  sleep 5
done
echo "✅ Elasticsearch API is up!"


# 2. Wait for the data from makelogs to be indexed and searchable.
echo "⏳ Waiting for data to be indexed in makelogs-*..."
retries=30
count=0
# Loop until the document count is 10000 or more.
until [ "${count:-0}" -ge 10000 ]; do
  count_json=$(curl -s -u "elastic:${ES_LOCAL_PASSWORD}" "http://localhost:9200/makelogs-*/_count")
  
  # Check if the response contains a count field and extract it.
  if echo "$count_json" | grep -q '"count"'; then
    count=$(echo "$count_json" | grep -o '"count":[0-9]*' | sed 's/"count"://')
  else
    count=0 # The index might not exist yet.
  fi

  echo "Current document count: ${count:-0}. Waiting for 10000..."
  retries=$((retries - 1))
  if [ "$retries" -le 0 ]; then
    echo "❌ Timed out waiting for documents to be indexed."
    exit 1
  fi
  sleep 5
done

echo "✅ Data is ready! Found ${count} documents."