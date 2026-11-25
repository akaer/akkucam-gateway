# Proposal: Multi-camera dashboard

## Problem

The base station has several coupled cameras, but the interface showed them only
through a static channel selector and only one after another. The interface carried
redundant diagnostic actions and did not scale well to large screens or mobile
devices.

## Goal

- load coupled channels dynamically from the hub
- stream several channels at the same time
- single and bulk actions
- fullscreen per tile
- responsive dashboard layout
- light and dark theme
- technical details without burdening the main flow
- available device and firmware metadata without generic placeholders
- switch motion detection per camera
- lock conflicting inventory and settings actions while streams run

## Success criteria

- the real hub camera list returns two tiles with device names
- both HLS streams run in parallel and carry their own status metric
- no code is stored persistently
- the interface works from 320 px width and uses several columns on large displays
- camera MAC and other unavailable values create no UI fields
- motion changes use a complete read-modify-write and are verified

## Status

Implemented.
