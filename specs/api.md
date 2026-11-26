# HTTP API

## General

- Default address: `http://<gateway-host>:8080`
- Format: JSON, UTF-8
- Current state: no HTTP authentication
- Scope: trusted local network

## Error format

Error responses are flat and backward compatible:

```json
{"error":"valid ip, did and securityCode are required","code":"validation_failed"}
```

`error` carries a human-readable message. `code` is stable and meant for clients.
Unknown errors return `internal_error` with HTTP 500 and a generic message; details
stay in the log.

The API stays English. The web interface translates known codes through
`error.code.<code>` from `web/locales/*.json` (ADR-0021); other clients can use the
same code mechanism. New codes must be added to both catalogs.

| Code | Typical HTTP status | Meaning |
|---|---:|---|
| `invalid_json` | 400 | Request body is not valid JSON |
| `validation_failed` | 400 | Required field missing or value outside the allowed range |
| `method_not_allowed` | 405 | Path exists, method does not |
| `not_found` | 404 | API path does not exist |
| `conflict` | 409 | Action is blocked while a stream is active |
| `unauthorized` | 401 | Base station rejected the security code |
| `upstream_timeout` | 504 | Base station does not answer in time |
| `upstream_error` | 502 | Base station reports an error or the session breaks |
| `internal_error` | 500 | Unexpected internal error |

## Machine-readable contract

- OpenAPI document: `GET /api/openapi.json`
- Swagger UI: `GET /api/docs`

The Swagger UI ships locally and loads no CDN resources. The contract lives in
`internal/openapi/openapi.json`; a contract test checks that the route table and the
document match. The spec uses the relative server URL `/` so "Try it out" sends
requests to the same origin, even when the dashboard is opened through a LAN IP or a
reverse proxy. A fixed `127.0.0.1` URL would point at the client in the browser and
trigger `NetworkError`.

## `GET /api/health`

Example response:

```json
{"status":"ok","mode":"local-only","live":"hls"}
```

## `GET /api/version`

Returns the semantic version and short Git commit of the running binary. If the
build has no commit, the field is omitted.

```json
{"version":"0.2.1","commit":"bfe827d"}
```

## `POST /api/discover`

Starts WAI and PPPP LAN discovery. The request body is empty.

```json
{
  "devices": [
    {
      "ip": "192.0.2.10",
      "did": "CMAG-000001-ABCDE",
      "model": "SMH-T1R",
      "ppcsPort": 17500,
      "mac": "02:00:00:00:00:10",
      "local": true
    }
  ],
  "durationMs": 4001
}
```

The PPCS port is dynamic and must not be stored permanently. `mac` is omitted when
the address could not be read from the local ARP table.

## `POST /api/session/probe`

Opens a local session up to `P2P_RDY` and closes it immediately.

```json
{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE"}
```

The endpoint sends no credentials or camera commands.

## `POST /api/session/authenticate`

Checks the security code and closes the session after successful authentication.

```json
{
  "ip":"192.0.2.10",
  "did":"CMAG-000001-ABCDE",
  "securityCode":"example-code"
}
```

The code appears in neither responses nor logs. Transport timeouts are retried up to
three times; an explicit rejection is not.

## `POST /api/live/start`

Starts or replaces the active stream for the same DID and channel.

```json
{
  "ip":"192.0.2.10",
  "did":"CMAG-000001-ABCDE",
  "securityCode":"example-code",
  "cameraIndex":0,
  "quality":1
}
```

Success response:

```json
{
  "id":"CMAG-000001-ABCDE-ch1",
  "playlistUrl":"/hls/CMAG-000001-ABCDE-ch1/index.m3u8"
}
```

The endpoint confirms the session and the FFmpeg start. Camera wake-up can add
several seconds before the first HLS playlist appears.

The stream ID carries the channel number, for example `CMAG-000001-ABCDE-ch2`.
Several channels of the same DID can therefore run at the same time.

## `POST /api/channels`

Authenticates locally and reads the coupled camera list from the base station.

```json
{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"example-code"}
```

```json
{
  "did":"CMAG-000001-ABCDE",
  "hub": {
    "model":"SMH-T1R",
    "vendor":"ExampleVendor",
    "firmware":"1.0.0.12",
    "mcuVersion":"1.2.3",
    "mac":"02:00:00:00:00:10",
    "sdCardStatus":"normal",
    "sdTotalBytes":32000000000,
    "sdFreeBytes":8000000000
  },
  "channels":[
    {
      "index":0,
      "name":"Einfahrt",
      "model":"SAPP-TAS",
      "fwVersion":"0.0.36.0",
      "mcuVersion":"0.5.2.0",
      "fwVersionRaw":9216,
      "mcuVersionRaw":328192,
      "updateable":false,
      "motionEnabled":true,
      "pirMode":1,
      "pirSensitivityDay":65
    }
  ]
}
```

Unavailable metadata fields are omitted entirely. A camera MAC is not part of the
known hub replies and is therefore not returned. While a stream is active, the
endpoint answers with `409 Conflict`.

## `POST /api/motion`

Changes PIR motion detection for a coupled channel. The gateway first reads the
complete parameter block, changes only `pirMode` and verifies the setting by reading
it again.

```json
{
  "ip":"192.0.2.10",
  "did":"CMAG-000001-ABCDE",
  "securityCode":"example-code",
  "channel":0,
  "enabled":false
}
```

```json
{"index":0,"motionEnabled":false,"pirMode":0,"pirSensitivityDay":65}
```

`enabled` must be present explicitly even for `false`. While a stream is active, the
endpoint answers with `409 Conflict`.

## `GET /api/live/status?id=<id>`

```json
{
  "id":"CMAG-000001-ABCDE-ch2",
  "running":true,
  "channel":1,
  "bitrateBytesPerSecond":105469,
  "framesPerSecond":25,
  "battery":{"capacity":87,"charging":false},
  "uptimeSeconds":42
}
```

`channel` is zero-based. `battery` is omitted until a valid reply exists.
`framesPerSecond` is derived from the incoming H.264 payload rate per second and is
omitted until a measurement exists.

## `POST /api/live/stop`

```json
{"id":"CMAG-000001-ABCDE-ch2"}
```

Stops the PPCS session and the FFmpeg process. Stopping repeatedly is idempotent.

## HLS

```text
GET /hls/<id>/index.m3u8
GET /hls/<id>/indexN.ts
```

HLS files are transient runtime data and not a recording interface.
