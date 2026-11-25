# ADR-0022: App icon, favicon, license notices and SBOM

- Status: Accepted
- Date: 2026-11-22
- Changed: 2026-11-23 (addendum, ADR-0024/ADR-0027)

> **Updated 2026-11-23:** the footer, copyright line and separate license page
> were replaced by the scrollable license dialog behind the help icon; see the
> addendum at the end.
>
> The decision text below describes the state on 2026-11-22. The **License
> notices** section and its consequences were replaced by ADR-0024 and ADR-0027
> and remain only as history. The addendum is authoritative.

## Context

The dashboard used a single text letter "A" as its brand mark. There was no
favicon, no overview of the bundled third-party components and no software bill
of materials (SBOM). For a locally operated product, visible license notices and
a machine-readable SBOM are prerequisites for redistribution and operation.

The constraints are strict: no cloud, no CDN, no persistence of credentials. The
icon must ship locally and follow the light and dark theme. The SBOM must be
generable without external SaaS services.

## Decision

### Icon and favicon

- The brand mark is the Font Awesome camera icon ("camera", solid) as a locally
  bundled SVG: `web/icon-light.svg` and `web/icon-dark.svg`.
- The dashboard renders the same icon inline as SVG in `.brand-mark`. The color
  comes from the theme variables through `currentColor` (`--accent`,
  `--accent-text`); a theme switch changes the icon without a reload.
- The favicon is the same camera icon. Two variants (light and dark) load through
  `<link id="favicon" rel="icon">`. The inline theme bootstrap in `<head>` and
  `applyTheme()` in `app.js` swap the `href` to match the active theme.
- There is no CDN reference. The icon files sit under `web/` and ship from the
  same static file server as the rest of the dashboard.
- The icon license (Font Awesome Free, icons under CC BY 4.0) ships as
  `web/FONTAWESOME-LICENSE.txt`.

### License notices (historical state, replaced on 2026-11-23)

- ~~The footer carried the copyright notice "© André Raabe" and a "Third-party
  licenses" link to a separate, simple page `web/licenses.html`.~~
- ~~The license page does not collapse.~~ The license texts for hls.js, Font
  Awesome, Inter Tight and Swagger UI sit in the web root as
  `web/HLS-LICENSE.txt`, `web/FONTAWESOME-LICENSE.txt`,
  `web/INTERTIGHT-LICENSE.txt` and `web/swagger/LICENSE`. The project license
  stays `web/LICENSE.txt` (MIT).
- **Current (ADR-0024, ADR-0027):** an info icon at the far right of the top bar
  opens the help popup. Its "Licenses" button opens the scrollable license
  dialog (`#licenses-dialog`). It lists the MIT project license and the
  third-party components in the runtime image: hls.js (Apache-2.0), Font Awesome
  Free (icons: CC BY 4.0), Inter Tight (OFL-1.1), Swagger UI (Apache-2.0),
  FFmpeg (GPL-2.0-or-later), Alpine Linux (various) and Go (BSD-3-Clause).
- There is **no** page footer and **no** `web/licenses.html`.
- All visible text runs through the i18n catalogs (`help.*`, `licenses.*` and
  `footer.licenseFile` in `web/locales/en.json` and `web/locales/de.json`,
  ADR-0021).

### SBOM

- The SBOM is generated from the built runtime image with
  [Syft](https://github.com/anchore/syft), not from the source tree. It therefore
  contains the operating system packages actually shipped, including FFmpeg and
  Alpine.
- Both formats are generated: CycloneDX (JSON) and SPDX (JSON).
- `scripts/sbom.sh` pins the Syft image (`SYFT_IMAGE`, default
  `anchore/syft:v1.57.2`), exports the image with `docker save` and scans the
  archive. The Docker socket is not needed.
- The generated files sit as a snapshot under
  `sbom/akkucam-gateway.cyclonedx.json` and `sbom/akkucam-gateway.spdx.json`.
  `sbom/README.md` describes scope and regeneration.
- CI regenerates the SBOM on every run and uploads it as an artifact. The
  committed snapshot is a readable reference, not an automatically updated file.
- The web assets bundled in the image (hls.js, Font Awesome, Inter Tight, Swagger
  UI) are not package-manager dependencies and therefore do not appear in the
  SBOM automatically. The license dialog and `sbom/README.md` document them
  explicitly.

## Reasoning

- **Cloud-free:** a locally bundled SVG avoids CDN calls and extra network
  dependencies. An SVG is also smaller than an icon font and sharp at every size.
- **Theme consistency:** `currentColor` and theme variables keep the icon in sync
  with the dashboard automatically. The favicon follows the same theme through
  the two variants.
- **License clarity:** the footer section makes redistribution more transparent
  without overloading the interface. The license texts sit in the same web root
  and work offline.
- **Reproducibility:** an SBOM from the image describes the shipped artifact, not
  just the source tree. Two formats serve different consumers (CycloneDX for
  security tools, SPDX for compliance).
- **No new runtime dependency:** Syft runs only while generating the SBOM in a
  container. The gateway itself stays unchanged.

## Alternatives considered

- **Bundle the Font Awesome web font:** more payload, an extra font file, more
  complicated theme handling. Rejected.
- **Collapsible license list in the footer:** would burden the footer with a
  second disclosure level. A separate, simple license page is clearer and
  directly linkable. Rejected. *(Replaced on 2026-11-23 by the license dialog in
  the help popup under ADR-0024/ADR-0027.)*
- **Icon as a PNG or favicon package:** no theme support without several raster
  variants. Rejected.
- **CDN for the icon or font:** violates the cloud-free rule. Rejected.
- **Only `go version -m` or the `go mod` graph:** captures no operating system
  packages and no bundled web assets. Rejected.
- **SBOM in CycloneDX only:** SPDX is often required for license compliance. Syft
  produces both without extra work.
- **Generate the SBOM in the Docker build:** the build would need Syft and would
  get slower. Rejected; the SBOM stays a separate step.

## Consequences

- New files: `web/icon-light.svg`, `web/icon-dark.svg`,
  `web/FONTAWESOME-LICENSE.txt`, `scripts/sbom.sh`, `sbom/README.md`,
  `sbom/akkucam-gateway.cyclonedx.json`, `sbom/akkucam-gateway.spdx.json`.
  (`web/licenses.html` was removed on 2026-11-23.)
- New i18n keys: `help.open`, `help.eyebrow`, `help.title`, `help.text`,
  `help.dismiss`, `help.wallText`, `licenses.open`, `licenses.eyebrow`,
  `licenses.heading`, `licenses.intro`, `licenses.projectLicense`,
  `licenses.thirdPartyHeading`, `footer.apiDocs`, `footer.licenseFile`,
  `footer.license.various`.
- New tests: `internal/gateway/static_test.go` checks icon and asset references;
  `tests/sbom_test.go` checks both SBOM formats; `web/e2e/branding.spec.ts`
  checks the icon, favicon, brand typography, help and license dialogs and the
  German translation.
- Changing a dependency in the image requires regenerating the SBOM with
  `./scripts/sbom.sh` and updating the snapshot. CI builds the image and also
  generates a fresh SBOM artifact on every run (`.github/workflows/test.yml`).
- The license dialog lists only components contained in the image. Development-only
  dependencies (for example Playwright) do not appear there.

- The icon is a Font Awesome work; changes to the icon must keep the attribution
  in `web/FONTAWESOME-LICENSE.txt`.

## Addendum 2026-11-23

The footer, copyright line, MIT footer link and the separate `licenses.html` are
gone. License information opens through the info/help icon as a scrollable
license dialog. It lists the MIT project license and all third-party components;
the API documentation is linked there. The settings dialog no longer contains
license links. The license texts stay local in the web root.

## Addendum 2026-11-24

The license dialog names seven groups: the project (MIT), the bundled web assets
(hls.js, Font Awesome, Inter Tight, Swagger UI), FFmpeg, Alpine Linux and Go. The
SBOM lists around 125 individual Alpine packages. Both remain valid:
`product.md` requires naming all third-party components, and the dialog groups
the operating system packages under "Alpine Linux". A complete package list stays
in the SBOM.

## References

- ADR-0003: local, no cloud fallback
- ADR-0013: responsive dashboard and theme
- ADR-0021: bilingual UI
- `specs/sources.md`, `specs/architecture.md`, `specs/testing.md`
- `scripts/sbom.sh`, `sbom/README.md`

## Addendum 2026-11-24

The dialog checks the versions it shows against the shipped artifacts
(`TestLicenseVersionsMatchArtifacts`): hls.js against `web/hls.min.js`, FFmpeg,
Alpine and Go against the committed SBOM and the base images in the `Dockerfile`.
Changing a version requires regenerating the SBOM (`make sbom`); otherwise the
test fails.

The Font Awesome version disappears from the dialog. There is no machine-readable
source for it (neither license file nor SVG), and an unchecked number is worse
than none. The license (CC BY 4.0) stays named.
