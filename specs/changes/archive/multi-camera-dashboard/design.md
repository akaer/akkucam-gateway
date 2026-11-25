# Design: Multi-camera dashboard

## Backend

- new endpoint `POST /api/channels`
- IO-Control `40962` reads the count and 60-byte camera entries
- stream ID carries the DID and the 1-based channel number
- the stream manager replaces only identical DID/channel combinations
- control handshakes stay serialized, media loops run in parallel

## Frontend

- security code dialog, code only in a JavaScript variable
- dynamic tiles from the hub reply
- "Start all" starts channels sequentially; parallel operation follows
- status polling per stream ID
- one HLS instance per video tile
- native fullscreen or `webkitEnterFullscreen` on iOS
- diagnostics area with a manual channel fallback
- CSS custom properties for themes; theme preference in LocalStorage
- metadata fields render based on value; missing values create no placeholder rows
- metadata and live metrics are collapsed `details/summary` areas by default
  (progressive disclosure)

## Error behavior

- if the camera list fails, no channel is invented
- the diagnostics area opens and offers manual slots 1 to 4
- a channel error does not stop the other streams
