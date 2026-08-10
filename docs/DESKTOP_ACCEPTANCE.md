# Corsarr Desktop acceptance

This checklist defines the evidence required before Corsarr Desktop can be
considered ready for a non-technical user on a personal computer. It separates
repeatable development checks from actions that require the user, a release
identity, or another operating system.

Passing this checklist does not publish a release. The current Desktop remains
an unreleased development build.

## First Mac evidence snapshot

The first acceptance environment is an Apple Silicon Mac running macOS. On
2026-08-11, development validation used Docker Engine 29.6.2 through Docker
Desktop 4.83.0. These versions record the tested environment; they are not a
permanent product requirement.

The following evidence has passed on that Mac:

| Area | Evidence | Result |
| --- | --- | --- |
| Go backend | Unit, integration, race, vet, and golangci-lint checks | Passed |
| Frontend | Biome formatting/lint, TypeScript checks, and production build | Passed |
| Desktop bundle | Wails v2 macOS arm64 build and local self-signing | Passed |
| Clean-state interface | Home and license screens rendered; runtime and storage states were accurate; mutating actions remained disabled without consent and storage | Passed |
| Docker lifecycle | Owned network/container creation, inspection, logs, start, stop, restart, and cleanup using a local immutable image | Passed |
| Docker installation | Transactional container creation and HTTP readiness using a local immutable image | Passed |
| Bind mounts | Host storage and container mount contract using Docker on macOS | Passed |
| Update rollback | Failed readiness restored the previous running image; a healthy replacement completed | Passed |
| Image platforms | Every approved remote OCI index advertised Linux AMD64 and ARM64 without downloading layers | Passed |
| Cleanup | No Corsarr Desktop state, managed container, or managed network remained after validation | Passed |

The Docker contracts require immutable images already present locally and do
not authorize media image downloads. Run them serially because they share the
owned network:

```bash
CORSARR_DOCKER_CONTRACT_IMAGE='repository/name@sha256:<digest>' \
  CORSARR_DOCKER_ROLLBACK_IMAGE='repository/other@sha256:<digest>' \
  go test -p 1 ./internal/runtime ./internal/orchestrator \
    -run 'Real.*Contract' -count=1 -v
```

Catalog maintainers can query approved indexes without pulling their layers:

```bash
CORSARR_VERIFY_REMOTE_MANIFESTS=1 \
  go test ./internal/catalog \
    -run '^TestRuntimeCatalogRemotePlatformContract$' -count=1 -v
```

## Human acceptance on the first Mac

These checks deliberately require the user. An automated agent must not accept
runtime or application terms on the user's behalf.

- [ ] Launch the clean Desktop build and review the runtime/application terms
  before giving explicit consent.
- [ ] Choose a storage folder with at least 10 GiB available and confirm that
  the displayed folder, capacity, and hardlink result are understandable.
- [ ] Review the recommended movie/TV selection and any optional applications,
  then approve the real digest-pinned image downloads.
- [ ] Install the selected stack and confirm that progress identifies the
  current application without exposing credentials or runtime logs.
- [ ] Confirm that qBittorrent, the selected Arr applications, Prowlarr,
  Bazarr, Jellyfin, and Seerr finish their applicable automated provisioning.
- [ ] Open each installed application through its Corsarr shortcut and verify
  that its local web interface is ready.
- [ ] Verify stop, start, and restart for an application without using Docker
  Desktop or a terminal.
- [ ] Update an application and confirm that health verification completes; a
  deliberately induced failure may be used in development to confirm rollback.
- [ ] Remove an application while preserving its configuration, then confirm
  dependency protection prevents removal in the wrong order.
- [ ] After removing its container, archive its configuration separately and
  confirm that Corsarr shows the approximate size and recoverable destination.
- [ ] If Jellyfin LAN access is enabled, open the advertised private address
  from a TV or mobile device on the same network; administrative apps must
  remain local-only.
- [ ] Enable start at login, approve it in macOS when requested, then verify
  recovery after logout/login or reboot without reinstalling applications.
- [ ] Interrupt or disconnect the network during a development installation
  and confirm that the error is actionable and retry does not duplicate owned
  resources.
- [ ] Export diagnostics and verify that they contain useful versions and
  health facts but no password, cookie, API key, application log, or runtime
  socket.

Record the macOS version, Mac architecture, runtime version, selected apps,
storage filesystem, and result for every failed item. A failed human check is a
product defect or accepted limitation; an unchecked item is simply unverified.

## External release gates

The following work cannot be completed solely by the first-Mac development
loop:

- Apple Developer ID signing, notarization, and Gatekeeper validation require
  an authorized Apple developer identity. Local self-signing is not equivalent.
- Windows onboarding, elevation/reboot recovery, WSL setup, and its native
  secure credential store require Windows test machines.
- Linux distribution onboarding, rootless permissions, SELinux behavior, and
  its native secure credential store require the supported Linux matrix.
- Podman promotion requires the empirical matrix recorded in
  [ADR 0002](decisions/0002-docker-first-runtime-with-podman-gate.md), including
  volumes, reboot recovery, update/rollback, and removal on Windows, macOS, and
  Linux.
- A published release requires a separate authorization and release process;
  implementation and local commits do not grant permission to publish.

## Governing documentation

- [Desktop RFC](rfcs/0001-corsarr-desktop.md)
- [Architecture](ARCHITECTURE.md)
- [Wails decision](decisions/0001-use-wails-v2-for-desktop.md)
- [Runtime strategy](decisions/0002-docker-first-runtime-with-podman-gate.md)
- [Troubleshooting](TROUBLESHOOTING.md)

