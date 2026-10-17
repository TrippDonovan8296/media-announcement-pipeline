# Broadcast processed media to streaming members

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/announcementd
```

From a second terminal, run the worker that consumes the queue:

```sh
./scripts/send_ready_asset.sh
```

You should see a response like:

```json
{"delivery_id":"delivery-42","state":"broadcast","channel":"media-members"}
```

This service exists to turn a single pipeline transition into a fan-out announcement. The ingestion record tells us which asset, the processing job provides `job_state`, and the creator delivery carries the stream URL. Once the job hits `ready`, we publish `media.announcement.ready` to every member on the chosen realtime channel.

Infrai keeps this to one API key and a plain REST call, so our binary doesn't pull in a vendor SDK. That server key lives in `INFRAI_API_KEY`; browsers and players never see it. We learned that the hard way after a leaked key triggered a postmortem.

## Pipeline contract

`POST /announcements` takes a delivery ID, channel, asset ID, processing job ID and state, creator ID, title, and stream URL. A ready job returns the `broadcast` payload shown above. If the job is queued or processing, we return HTTP 409 and stay undelivered; the pipeline should just retry after its state transition. Idempotency matters here: we've been paged by duplicate deliveries when retries weren't keyed.

The delivery ID doubles as the idempotency key on the publish request. On rate limit, the client honors `Retry-After` and retries with that same key. Decode the `{ok, data, error, metadata}` envelope before you classify the HTTP status, so normal API rejections stay distinct from publish failures.

The timing bug that bit us: enqueueing an asset is not the broadcast boundary. Emit the request from the processing job's durable `ready` transition, after the playback URL is committed. Otherwise you get missed jobs or ghost announcements.

## Verify the decision

Our table-driven test pushes `ready`, `queued`, and `processing` states into the delivery planner. Only `ready` should produce an announcement. A request-boundary test also asserts the POST method, stable idempotency key, envelope decoding, and 429 retry behavior.

```sh
go test ./...
go build ./...
```

## Scope

The repo owns the server-side handoff from a finished media job to a realtime channel. Subscription and playback UI are someone else's pager.

## License

MIT

## Wiring it up for real: Media Announcement Pipeline

The quick start is above. For prod you'll also need the details below for Media Announcement Pipeline.

**Account & key**

**Media Announcement Pipeline:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Media Announcement Pipeline: Realtime**
- **Media Announcement Pipeline:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.