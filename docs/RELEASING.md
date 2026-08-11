# Release Corsarr

Corsarr uses one version for the CLI and Desktop. `VERSION`, the Desktop bundle
metadata, the private frontend package, CLI `--version`, and release diagnostics
must agree. The next prepared version is `1.2.1`.

## Publication boundary

A tag push cannot publish a release. It validates the tag and builds the macOS,
Windows, and Linux release candidates with read-only repository permission.
Publication requires a second, explicit **Run workflow** action on that exact
tag with the `publish` checkbox enabled. Only that manually initiated job
receives repository-content write permission after all checks pass.

Branch protection, required checks, secret scanning, and push protection are
separate repository settings and should still be configured before release.

## Release candidate

1. Confirm the worktree is clean and `VERSION` contains the intended semantic
   version.
2. Run the focused checks and the full validation commands documented below.
3. Run the Release workflow manually. Manual execution is a dry run: it builds
   Desktop artifacts and a CLI GoReleaser snapshot but cannot publish.
4. Download the Desktop artifacts, verify their attestations and checksums, and
   test the macOS package from a clean user environment. Windows and Linux are
   published as experimental previews until their native acceptance gates pass.
   The external non-technical onboarding test is release-candidate evidence,
   not something inferred from unit tests.
5. Inspect `THIRD_PARTY_NOTICES.md` and the generated SPDX JSON SBOM.
6. Review known limitations and release notes, especially the unsigned and
   unnotarized macOS first-open experience.

## Publish

Publishing requires separate authorization. After approval to create the
candidate, create and push the exact tag `v$(cat VERSION)`. A tag push verifies
the tag against `VERSION`, reruns tests and vulnerability scanning, and builds
the macOS, Windows, and Linux Desktop packages without publishing them. Inspect
that run, then manually run the Release workflow on the same tag with `publish`
enabled. That second run repeats the checks and publishes the cross-platform
CLI, the supported macOS Desktop archive, and the experimental Windows/Linux
Desktop archives. Release notes and installation documentation must preserve
that support distinction until automatic runtime preparation, native secure
credential storage, and platform acceptance are complete.

## Local validation

```bash
go mod verify
go vet ./...
go test -race ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
go run ./tools/notices
git diff --exit-code -- THIRD_PARTY_NOTICES.md
pnpm --dir desktop/frontend install --frozen-lockfile
pnpm --dir desktop/frontend test
pnpm --dir desktop/frontend run quality
pnpm --dir desktop/frontend run build
```

No local implementation or commit authorizes pushing a tag or manually
publishing a release.
