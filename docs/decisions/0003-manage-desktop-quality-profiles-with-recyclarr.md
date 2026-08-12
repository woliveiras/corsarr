# Manage Desktop quality profiles with versioned Recyclarr presets

- Status: accepted
- Date: 2026-08-12
- Decision makers: @woliveiras
- Consulted: Codex research using official Recyclarr, TRaSH Guides, and Profilarr sources
- Supersedes: none

## Context and Problem Statement

Corsarr Desktop can install and connect Radarr, Sonarr, download clients,
indexers, Jellyfin, and Seerr, but a first-time user still has to understand
quality profiles, quality definitions, custom formats, and scoring before media
requests behave predictably. The CLI is aimed at advanced users and does not
own post-install configuration, so extending this automation to both products
would blur their current responsibilities.

Recyclarr already provides a headless synchronization engine for TRaSH Guides.
Profilarr provides a dedicated interface for users who want to design and
manage profiles themselves. Corsarr should compose those tools rather than
implementing another profile synchronization engine.

## Decision Drivers

- Give a first-time Desktop user useful quality defaults without teaching Arr
  internals.
- Preserve an explicit way to leave profiles unmanaged.
- Keep the exact applied policy reproducible and reviewable.
- Never persist Arr API keys in the Recyclarr configuration or frontend state.
- Avoid two controllers managing the same Arr profiles.
- Keep upstream changes from silently changing a working media server.
- Preserve the CLI as an advanced, manually configured workflow.

## Considered Options

- Maintain a native Corsarr quality-profile synchronization engine.
- Run Recyclarr continuously with the latest TRaSH Guides data.
- Use Recyclarr for Corsarr presets and keep Profilarr as a mutually exclusive
  advanced option.
- Install Profilarr as the only profile-management path.
- Keep every quality profile manual.

## Decision Outcome

Chosen option: **use Recyclarr as the headless engine for versioned Corsarr
presets, with an unmanaged path for advanced users and Profilarr as a separate,
mutually exclusive advanced application**.

When Radarr or Sonarr is selected, Desktop adds one final quality step before
installation completes. The step offers:

- `economy`: WEB-focused 1080p with lower storage use.
- `balanced-1080p`: balanced HD Bluray and WEB behavior; the default.
- `high-1080p`: Remux and WEB 1080p with larger files.
- `4k-hdr`: 2160p and HDR, only after explicit selection.
- `unmanaged`: Corsarr does not alter quality profiles.

The user-facing names are stable Corsarr concepts. Their mapping to TRaSH
Guides identifiers is versioned independently as the Corsarr preset catalog.
The first catalog version is `2026-08-12.1`.

For a managed preset, Desktop:

1. Finishes installing and provisioning the selected applications.
2. Reads the generated Radarr/Sonarr API keys only in the Go backend.
3. Writes a private Recyclarr configuration containing environment-variable
   references, never key values.
4. Pulls and executes the digest-pinned Recyclarr image as an ephemeral
   container on the Corsarr network.
5. Runs `sync --preview`; any failure stops the onboarding attempt.
6. Applies the same configuration only after the user's install action has
   provided consent.
7. Synchronizes the guide's quality definitions, automatically associated
   custom-format groups, and recommended media naming for the chosen profile.
8. Reconciles Seerr afterward so its reserved Arr connection selects the
   resulting `Corsarr - <preset>` profile.

The exact engine image is Recyclarr `8.7.0` at
`sha256:2d6107f758d882a59fe9d646aa54fa8a5a4fb7a40995125fade575652a3f7871`.
The resource provider replaces Recyclarr's default TRaSH Guides source with
commit `0943b2677a0454d7d69fc9697a8ddcdb2eebd8d9`. These pins and the Corsarr
preset-catalog version are changed only through a reviewed Corsarr release.

Automatic periodic synchronization is off by default. A future scheduled sync
must be opt-in, reuse these authority and secret boundaries, preview before
applying, and expose the exact policy update to the user.

Profilarr is not installed by the first Recyclarr increment. It remains the
accepted advanced application direction, but requires a separate catalog and
lifecycle implementation. Selecting it must set Corsarr quality management to
`unmanaged`; Corsarr and Profilarr must never concurrently own the same
profiles.

The CLI does not gain this onboarding or implicit post-install synchronization.
Its users continue to configure Recyclarr, Profilarr, or Arr directly.

### Consequences

- Good: New Desktop installations receive useful, named defaults with one
  decision.
- Good: Recyclarr and TRaSH Guides remain independently attributable upstream
  components instead of copied logic.
- Good: Engine, guide data, and Corsarr mappings are independently pinned.
- Good: API keys stay process-scoped and do not cross the Wails boundary.
- Good: No background controller can silently change profiles by default.
- Bad: A failed apply can leave part of a multi-Arr synchronization complete;
  Recyclarr does not make the operation transactional. Onboarding remains
  incomplete and retryable instead of claiming rollback.
- Bad: Updating a preset requires maintaining and testing three version pins.
- Neutral: Users choosing `unmanaged` retain full responsibility for profiles
  and Seerr defaults.

## Confirmation

The decision is confirmed when deterministic tests prove that:

1. The quality step appears only for Radarr or Sonarr.
2. Managed configurations use only known preset IDs and pinned upstream data.
3. API keys occur in neither generated files nor command arguments.
4. Preview precedes apply and a preview failure prevents apply.
5. Recyclarr runs ephemerally on the Corsarr network.
6. Seerr prefers a Corsarr-owned profile when one exists.
7. The legal catalog attributes both Recyclarr and TRaSH Guides.
8. No scheduled synchronization or CLI behavior is enabled by default.

Real-host acceptance must additionally execute a fresh Radarr/Sonarr install,
inspect the created profiles and custom formats, add media through Seerr, and
verify that retrying the same preset is idempotent.

## More Information

- [Corsarr Desktop RFC](../rfcs/0001-corsarr-desktop.md)
- [Recyclarr features](https://recyclarr.dev/guide/features/)
- [Recyclarr sync command and preview](https://recyclarr.dev/cli/sync/)
- [Recyclarr resource providers](https://recyclarr.dev/reference/settings/resource-providers/)
- [Recyclarr quality profiles](https://recyclarr.dev/reference/configuration/quality-profiles/)
- [TRaSH Guides Guide Sync](https://trash-guides.info/Guide-Sync/)
- [Profilarr source and license](https://github.com/Dictionarry-Hub/profilarr)
