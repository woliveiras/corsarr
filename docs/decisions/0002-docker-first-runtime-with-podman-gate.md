# Use Docker first and gate Podman promotion on a cross-platform spike

- Status: accepted
- Date: 2026-08-10
- Decision makers: @woliveiras
- Consulted: Codex research using official Docker, Microsoft, and Podman sources
- Supersedes: none

## Context and Problem Statement

Corsarr currently generates Docker Compose configurations. Corsarr Desktop must
install and manage media services without requiring users to understand a
container runtime. Podman can provide the required container, image, network,
volume, secret, and machine operations without Compose, but its cross-platform
path and virtual-machine behavior has not yet been validated with the Corsarr
stack.

The decision is how to deliver the first desktop MVP without hard-coding the
product domain to Docker or prematurely making Podman the default.

## Decision Drivers

- Reach a useful MVP by reusing the current Docker Compose behavior.
- Keep Docker, Compose, WSL, Podman, and virtual machines out of the primary user
  vocabulary.
- Make desired application state belong to Corsarr rather than to Compose YAML.
- Preserve the ability to control Podman directly without Compose.
- Require evidence for Windows, macOS, and Linux path, permission, restart,
  update, rollback, and deletion behavior.
- Keep runtime sockets and equivalent privileged APIs behind the Go backend.

## Considered Options

- Docker-first MVP behind a runtime interface, followed by a Podman spike.
- Podman as the immediate and only runtime.
- Docker as a permanent product-wide dependency.
- Build and maintain a custom Linux VM and containerd distribution.

## Decision Outcome

Chosen option: **Docker-first MVP behind a runtime interface, with Podman
promotion gated by a cross-platform spike**.

The desktop implementation controls Docker containers directly and does not use
Compose. New product behavior depends on a runtime-neutral desired-state model
and a narrow `Manager` interface, not on Compose files or Docker-specific
types.

Podman is a candidate for the default runtime before the stable desktop release.
It will be promoted only if the validation gate in this record passes.

### Consequences

- Good: Existing Compose generation accelerates the MVP.
- Good: Podman can later be implemented with its native API or official Go
  bindings, without Compose.
- Good: Runtime selection remains an infrastructure decision rather than a user
  concept.
- Bad: A temporary Docker implementation and a later Podman implementation may
  both need maintenance during migration.
- Bad: Docker Desktop installation and license consent remain part of the first
  Windows/macOS onboarding path.
- Neutral: Windows and macOS require a Linux virtualization layer for either
  Docker or Podman.

### Implementation checkpoint: direct Podman adapter

The first direct `PodmanManager` adapter was added on 2026-08-10. It consumes
the same validated `ContainerSpec` contract as Docker and uses fixed Podman CLI
arguments for immutable image pulls, one labeled bridge network, independent
labeled containers, bind mounts, ports, lifecycle operations, bounded logs, and
inspection. It does not generate Compose YAML, create a shared pod, expose the
Podman API, or make Podman selectable in the desktop UI.

This is implementation evidence only. It does not satisfy the promotion gate:
Podman is not installed by Corsarr, no Podman Machine lifecycle has shipped,
and the adapter has not yet run the Corsarr workloads across the required host
and filesystem matrix. Docker remains the default until those results are
recorded here.

## Pros and Cons of the Options

### Docker-first with a Podman gate

- Good: Reuses shipped Corsarr behavior and known images.
- Good: Allows a controlled comparison with real workloads.
- Bad: Requires discipline to prevent Compose concepts from leaking into the
  product model.

### Podman immediately

- Good: Open-source runtime with native management APIs and no Docker Desktop
  commercial licensing model.
- Good: Does not require Compose.
- Bad: Cross-platform path, mount, machine, and restart behavior is not yet
  proven for the Corsarr workloads.

### Docker permanently

- Good: Lowest near-term implementation cost.
- Bad: Makes product architecture and onboarding dependent on Docker Desktop and
  its terms even if Podman proves to be a better fit.

### Custom VM and containerd

- Good: Complete control of the runtime experience.
- Bad: Makes Corsarr responsible for a Linux guest, networking, filesystem
  sharing, upgrades, security patches, recovery, and host integration.

## Podman promotion gate

The spike must run at least on:

- Windows 11 Home with WSL 2.
- macOS on Apple Silicon.
- Ubuntu Linux.

Intel macOS and Windows on ARM are follow-up compatibility targets unless they
are declared supported for the first public release.

The spike must prove:

1. Automated runtime detection and installation with explicit OS consent.
2. Idempotent `podman machine init/start/inspect` on Windows and macOS.
3. One Corsarr-owned network with DNS aliases between independent containers.
4. One base bind mount containing downloads and libraries.
5. Working hardlinks or a clearly detected and explained fallback.
6. Install, start, stop, restart, logs, health, and browser opening.
7. Recovery after host reboot and user login.
8. Version-pinned update with backup, health verification, and image rollback.
9. Removal that preserves data by default.
10. Separately confirmed deletion of only Corsarr-owned application data.
11. Runtime/client version compatibility across a Podman Machine update.
12. No runtime socket, named pipe, or arbitrary command channel exposed to the
    Wails frontend or any managed container.

Promotion requires all critical scenarios to pass on all three platforms. A
documented fallback is acceptable for hardlinks on a platform only if it does
not risk data loss and the user is informed before installation.

## Confirmation

The Docker portion is confirmed when the desktop MVP completes the lifecycle in
the governing RFC through the runtime interface. Podman is promoted through a
new decision update only after the spike records versions, host filesystems,
commands/API versions, results, and known limitations.

## More Information

- [Corsarr Desktop RFC](../rfcs/0001-corsarr-desktop.md)
- [Docker Engine SDK](https://docs.docker.com/reference/api/engine/sdk/)
- [Docker Desktop Windows installer](https://docs.docker.com/desktop/setup/install/windows-install/)
- [Microsoft WSL commands](https://learn.microsoft.com/windows/wsl/basic-commands)
- [Podman installation](https://podman.io/docs/installation)
- [Podman Machine](https://docs.podman.io/en/latest/markdown/podman-machine.1.html)
- [Podman REST API](https://docs.podman.io/en/stable/_static/api.html)
- [Podman Go bindings](https://pkg.go.dev/github.com/containers/podman/v5/pkg/bindings)
- [Podman system service security boundary](https://docs.podman.io/en/stable/markdown/podman-system-service.1.html)
- [Podman license](https://github.com/containers/podman/blob/main/LICENSE)
