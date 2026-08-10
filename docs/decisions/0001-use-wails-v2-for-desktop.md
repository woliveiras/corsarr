# Use Wails v2 for Corsarr Desktop

- Status: accepted
- Date: 2026-08-10
- Decision makers: @woliveiras
- Consulted: Codex research using official project documentation
- Supersedes: none

## Context and Problem Statement

Corsarr is implemented in Go and currently ships as a CLI. Corsarr Desktop needs
a cross-platform graphical interface without duplicating the existing domain
logic or introducing a second privileged backend. The decision is which desktop
application framework should host the UI and connect it to Go operations.

## Decision Drivers

- Preserve Go as the owner of orchestration, runtime access, filesystem access,
  secrets, validation, and service adapters.
- Reuse the existing Go module and progressively extract CLI logic into reusable
  packages.
- Support Windows, macOS, and Linux.
- Keep the privileged runtime boundary out of the frontend.
- Use ordinary web UI technologies without requiring an embedded Node.js server
  in production.
- Prefer a stable framework over a pre-release generation.

## Considered Options

- Wails v2 with a Go backend and web frontend.
- Electron with a Node.js main process.
- Tauri with a Rust backend.
- Keep the CLI as the only interface.

## Decision Outcome

Chosen option: **Wails v2**, because it preserves Go as the application backend,
supports a web frontend, generates TypeScript bindings for Go methods, uses the
platform webview, and supports the three target desktop operating systems.
Wails v2 is the stable line; Wails v3 is still pre-release and is not selected
for the MVP.

### Consequences

- Good: Existing Go logic can be reused by both the CLI and desktop entrypoints.
- Good: Runtime sockets, credentials, and privileged operations remain behind Go
  methods instead of being exposed to browser code.
- Good: The repository can remain a single product repository with Go and a
  TypeScript/HTML/CSS frontend.
- Bad: The project gains a frontend toolchain and platform-specific packaging
  requirements.
- Bad: Linux desktop builds depend on the webview libraries expected by Wails.
- Neutral: The frontend is not an independent service and must not directly call
  Docker, Podman, or service administration APIs.

## Pros and Cons of the Options

### Wails v2

- Good: Native fit with the existing Go codebase.
- Good: Generated JavaScript and TypeScript bindings for Go methods.
- Good: Cross-platform build and packaging support.
- Bad: Adds Wails-specific lifecycle and packaging behavior.

### Electron

- Good: Mature ecosystem and broad frontend familiarity.
- Bad: Introduces a Node.js privileged process beside the Go core or requires an
  additional IPC/service boundary.
- Bad: Bundles a browser runtime that Corsarr does not otherwise need.

### Tauri

- Good: Small desktop bundles and a deliberate permission model.
- Bad: Makes Rust a second backend language for a product whose orchestration is
  already implemented in Go.

### Keep the CLI only

- Good: No new UI toolchain.
- Bad: Does not solve the product problem for non-technical users.

## Confirmation

This decision is confirmed when a Wails v2 application:

1. Builds on Windows, macOS, and Linux.
2. Calls a reusable Go application service shared with the CLI.
3. Opens an installed service in the default browser through a Go-bound method.
4. Proves that the frontend cannot access runtime sockets or arbitrary command
   execution.
5. Produces signed/packageable artifacts for the supported targets.

Revisit the decision if Wails v2 cannot satisfy packaging, accessibility, or
security requirements in the cross-platform proof of concept. Do not adopt
Wails v3 until it is stable and a migration is separately evaluated.

## More Information

- [Corsarr Desktop RFC](../rfcs/0001-corsarr-desktop.md)
- [Wails repository and version status](https://github.com/wailsapp/wails)
- [Wails v2 installation and supported platforms](https://wails.io/docs/v2.12.0/gettingstarted/installation/)
- [How Wails binds Go and frontend code](https://wails.io/docs/howdoesitwork/)
