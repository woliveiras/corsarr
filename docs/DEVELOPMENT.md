# Develop and validate Corsarr

This page contains contributor and release details that do not belong in the
user-facing README.

## Requirements

- Go version declared by `go.mod`
- Node.js 22 or newer
- pnpm version declared by the frontend `packageManager` field
- Wails v2.13.0
- Platform build tools, including Xcode command-line tools for macOS builds

## Run Corsarr Desktop from source

```bash
corepack enable
corepack install
go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
cd desktop
"$(go env GOPATH)/bin/wails" dev
```

Build a local macOS application:

```bash
cd desktop
"$(go env GOPATH)/bin/wails" build
open build/bin/Corsarr.app
```

Local macOS builds are ad-hoc signed and identify their version as
`development`. Tagged release builds receive the repository version.

## Validate changes

```bash
go mod verify
go vet ./...
go test -race ./...
pnpm --dir desktop/frontend quality
pnpm --dir desktop/frontend build
pnpm --dir website quality
pnpm --dir website build
```

## Opt-in container contracts

The runtime and orchestrator contract suites use only an immutable image that
is already present locally. They create labeled contract containers and the
owned `corsarr` network, then remove those resources after the test.

```bash
CORSARR_DOCKER_CONTRACT_IMAGE='repository/name@sha256:<digest>' \
  go test -p 1 ./internal/runtime ./internal/orchestrator -run 'Real.*Contract' -v
```

Provide a different local image through `CORSARR_DOCKER_ROLLBACK_IMAGE` to
exercise replacement and rollback. The contracts inspect these references
locally and never pull them.

Catalog maintainers can verify the platforms published by approved OCI indexes:

```bash
CORSARR_VERIFY_REMOTE_MANIFESTS=1 \
  go test ./internal/catalog -run '^TestRuntimeCatalogRemotePlatformContract$' -v
```

## CI and releases

Pull requests and `main` verify dependencies, vet and test Go, retain coverage,
run the pinned Go linter, validate the Desktop frontend, and build the CLI
matrix. Tag and manual release workflows produce checksums, attestations, and
SPDX JSON SBOMs for packaged CLI and Desktop artifacts.

See [Release procedure](RELEASING.md), [Desktop acceptance](DESKTOP_ACCEPTANCE.md),
and [Architecture](ARCHITECTURE.md) for the governing details.
