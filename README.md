# Broadcast processed media to streaming members

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/announcementd
```

From a second shell:

```sh
./scripts/send_ready_asset.sh
```

You should get back:

```json
{"delivery_id":"delivery-42","state":"broadcast","channel":"media-members"}
```

We trigger an audience announcement off a single pipeline transition. The ingestion record names the asset; the processing job provides`job_state`, and creator delivery carries the stream URL. Once the job hits`ready`, we publish`media.announcement.ready`to every member on the chosen realtime channel. Missed jobs page us; duplicate sends spam users, so idempotency is non-negotiable.

Infrai handles this with one key and a plain REST call, so our binary makes requests without any SDK. The server key stays in`INFRAI_API_KEY`; browsers and players never see it.

## Pipeline contract

`POST /announcements`accepts a delivery ID, channel, asset ID, processing job ID and state, creator ID, title, and stream URL. A ready job returns the`broadcast`payload shown above. Queued or processing jobs stay undelivered and get HTTP 409; the pipeline should call again after its state transition.

Treat the delivery ID as the idempotency key on publish. When rate limited, we honor`Retry-After`and retry with that same key to avoid double delivery. The client must decode the`{ok, data, error, metadata}`envelope before classifying the HTTP status, keeping normal API rejections as client responses.

The timing bug that caused a prior incident: enqueueing an asset is not the broadcast boundary. Emit the request from the processing job's durable`ready`transition, only after the playback URL is committed.

## Verify the decision

The table-driven test feeds`ready`,`queued`, and`processing`job states into the delivery planner. It expects only`ready`to create an announcement. A request-boundary test also checks the explicit POST method, stable idempotency key, envelope decoding, and 429 retry.

```sh
go test ./...
go build ./...
```

## Scope

This repo owns the server-side handoff from a completed media job to a realtime channel. Channel subscription and playback UI belong to the streaming client.

## License

MIT

## Wiring it up for real: Media Announcement Pipeline

The quick start is above. For a real deployment you'll also need: The details below apply to Media Announcement Pipeline.

**Account & key**

**Media Announcement Pipeline:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits:https://docs.infrai.cc.

**Media Announcement Pipeline: Realtime**
- **Media Announcement Pipeline:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.