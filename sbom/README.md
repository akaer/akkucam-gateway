# Software Bill of Materials

Generated snapshots of the gateway container image in two formats:

| File | Format | Spec |
|---|---|---|
| `akkucam-gateway.cyclonedx.json` | CycloneDX JSON | CycloneDX 1.7 |
| `akkucam-gateway.spdx.json` | SPDX JSON | SPDX 2.3 |

The SBOM is produced from the runtime image with
[Syft](https://github.com/anchore/syft). It covers the Alpine base image,
the installed APK packages (including FFmpeg and its libraries), and the Go
binary. Regenerate after changing the image:

```sh
./scripts/sbom.sh                 # builds akkucam-gateway:sbom if needed
./scripts/sbom.sh akkucam-gateway:test
```

The script exports the image with `docker save` and scans the archive, so the
scan container does not need access to the Docker socket. The committed files
are snapshots; CI regenerates them and uploads them as build artifacts.

## Vendored web assets

Syft scans packages. The dashboard and API docs ship vendored assets that
have no package manifest, so they are tracked here and on the licenses page instead:

| Asset | Version | License | Notice |
|---|---|---|---|
| hls.js | 1.5.18 | Apache-2.0 | `web/HLS-LICENSE.txt` |
| Font Awesome Free (camera icon) | 6.7.2 | Icons: CC BY 4.0 | `web/FONTAWESOME-LICENSE.txt` |
| Inter Tight (brand name and step headings, 400/700) | – | OFL-1.1 | `web/INTERTIGHT-LICENSE.txt` |
| Swagger UI | – | Apache-2.0 | `web/swagger/LICENSE` |

See `specs/sources.md` for the complete provenance list.
