# ADR-0026: Serve several base stations at once

- Status: Accepted
- Date: 2026-11-22
- Partially superseded: [ADR-0027](0027-live-wall-layout.md)

## Context

Until then the dashboard held exactly one workspace. Loading the cameras of a
second base station replaced the first view. Anyone with two hubs could never see
both at once, even though the stream manager had long allowed one stream per DID
and channel (ADR-0011) and the API has no limit across DIDs. The goal is to play
the streams of all loaded base stations in parallel.

## Decision

- Each loaded base station gets its own workspace (`.workspace[data-hub-key]`)
  below `#workspaces-section`. Loading another station no longer replaces an
  existing workspace; the sections stay visible below each other.
- The workspace header names the base station (model · DID) and its camera count.
  The "Start all"/"Stop all" actions stay per base station.
- The step 2 header also gets the global actions "Start all base stations" and
  "Stop all base stations". They start and stop every channel of every loaded hub.
- Inventory and motion actions are locked while **any** stream runs, matching the
  API, which returns `409 conflict` for active streams. After stopping, loading and
  discovery are available again.
- Security codes stay per DID in browser RAM (ADR-0025). The global actions ask for
  missing codes one after another (single flight per DID, one dialog at a time).
- A language switch updates both workspaces in place without interrupting streams;
  tiles without a device name use translatable fallback keys.

## Alternatives considered

- **Tabs per hub:** would switch exclusively again and contradicts the goal of
  seeing all streams at once.
- **One shared camera wall without hub grouping:** loses the camera-to-base-station
  mapping and with it the unambiguous assignment of code and stream ID.
- **Global actions only:** prevents starting a single hub on purpose.

## Consequences

- `web/index.html`: the static `#workspace` block is replaced by
  `#workspaces-section` (later replaced by the shared wall in ADR-0027) with global
  actions and the `#workspaces` container.
- `web/app.js`: workspace per hub, global start and stop functions, inventory lock
  across all hubs, prompt chain for several hubs, fallback titles through i18n.
- `web/style.css`: spacing and scroll offset for stacked workspaces.
- `web/locales/{en,de}.json`: `workspace.startAllHubs`,
  `workspace.stopAllHubs`, `notice.channelsLoaded` names the hub.
- Tests: new Playwright suite `workspaces.spec.ts` (two hubs, start and stop all,
  inventory lock); credentials and branding suites adjusted to the new structure.
- No API, OpenAPI or SBOM change: the API already works per DID.

## Addendum 2026-11-23

ADR-0027 replaces the workspace sections with a shared live wall. Each loaded hub
appears as a filter chip and in the hub manager, no longer as its own
`.workspace` section. The invariants of this ADR stay valid:

- all loaded hubs can stream at the same time,
- loading another hub does not replace an existing stream,
- global actions start and stop all hubs, and each hub stays individually
  controllable,
- inventory and motion actions are locked while streams run,
- codes stay per DID in browser RAM.

The concrete wall, popover and mobile implementation is in ADR-0027.

## References

- ADR-0011: one stream per DID and channel
- ADR-0025: security code per base station
- `specs/product.md`, `specs/architecture.md`, `specs/testing.md`
