#!/bin/sh
set -eu

curl --request POST http://localhost:8080/announcements \
  --header 'Content-Type: application/json' \
  --data '{
    "delivery_id": "delivery-42",
    "channel": "media-members",
    "asset_id": "asset-17",
    "job_id": "job-9",
    "job_state": "ready",
    "creator_id": "creator-3",
    "title": "Studio update",
    "stream_url": "https://media.example/asset-17/master.m3u8"
  }'
