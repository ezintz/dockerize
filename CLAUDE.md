# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`dockerize` is a small statically-compiled Go CLI (single `main` package at repo root, no
subpackages) that wraps a container's entrypoint command to: render config files from Go
`text/template` templates using env vars, tail log files to stdout/stderr, and wait for
dependent services (tcp/http/amqp/file/unix) before starting the wrapped process.

## Commands

```sh
go generate ./...              # installs pinned tool versions into .gobincache (see tools.go)
scripts/test                   # full CI check: go generate, hadolint, shellcheck, golangci-lint, gotestsum -race ./...
scripts/cover                  # coverage run, opens HTML report locally (skipped under $CI)
goreleaser release --snapshot --clean --skip=sign,sbom  # local dry-run of the full release (binaries + both Docker image variants)
goreleaser check               # validate .goreleaser.yaml

go test ./...                  # run all tests directly
go test -run TestName ./...    # run a single test
go build .                     # build the dockerize binary
```

`scripts/*` add `.gobincache` to `PATH` and expect tools installed via `go generate` (which reads
the `//go:generate` directives in `tools.go`, a `generate`-build-tag-only file). Always run
`go generate` before invoking golangci-lint/hadolint/shellcheck/gotestsum directly.

Go version is pinned in `go.mod` (single source of truth — CI's `setup-go` reads `go-version-file:
go.mod`, and there's no separate Docker build stage to keep in sync since release images are built
from prebuilt binaries, not compiled in-container).

Releases (tag pushes matching `v*`) are built by [GoReleaser](https://goreleaser.com)
(`.goreleaser.yaml`) via `.github/workflows/ci-cd.yml`: cross-compiles the 11 platform/arch
binaries, publishes them + a SLSA3 provenance attestation + cosign-signed checksums + SBOM to the
GitHub Release, and builds two container image variants (`Dockerfile` = Alpine default,
`Dockerfile.chainguard` = hardened Chainguard `static` base, tagged with a `-chainguard` suffix)
pushed to both Docker Hub and GHCR, with SLSA3 image provenance attached to the GHCR copies via
`slsa-framework/slsa-github-generator`. main-branch pushes only run the `test` job — nothing is
published outside of tagged releases (GoReleaser's snapshot mode, the only way to build without a
version tag, disables all publishing).

## Architecture

Everything lives in flat files at the repo root, one concern per file:

- **main.go** — flag definitions (`init()`), a single `cfg` global struct holding all parsed
  options, and `main()`'s linear pipeline: validate flag combinations → load CA cert → load INI
  defaults into env → render templates → wait for URLs → start tailing → exec the wrapped command
  (or block with `select{}` if only tailing was requested, no command given).
- **flag.go** — custom `flag.Value` implementations (`stringsFlag`, `urlsFlag`, `httpHeadersFlag`,
  `statusCodesFlag`, `delimsFlag`) used to support repeatable/composite CLI flags.
- **wait.go** — `waitForURLs` fans out one goroutine per `-wait` URL (by scheme: file/tcp/tcp4/tcp6/
  unix/http/https/amqp/amqps), each retrying on `cfg.wait.delay` until ready or until the shared
  `context.WithTimeout` (`-timeout`) expires; results collected over a channel.
- **template.go** — walks `-template src:dst` args (file or directory, recursive), rendering each
  with Go's `text/template` + Sprig funcs + custom funcs (`exists`, `parseUrl`, `isTrue`,
  `jsonQuery`, `readFile`). Preserves source file mode/uid/gid on the destination.
- **ini.go** — optional `-env` INI file (local path or http(s) URL) whose keys seed default env
  vars (`setDefaultEnv` in env.go only sets vars not already present in the environment).
- **tail.go** — wraps `github.com/powerman/tail` to follow a file and copy new content to
  stdout/stderr.
- **exec.go** / **exec_linux.go** / **exec_other.go** — runs the wrapped command, forwards signals
  to it, and (Linux only, via `Pdeathsig`) ensures it's killed if `dockerize` itself dies.
- **tls.go** — loads a custom CA cert (`-cacert`) into the system pool, shared by the `-wait`
  https/amqps and `-env` http(s) code paths.

Tests: `main_test.go` (flag/pipeline behavior) and `init_test.go` use `github.com/powerman/check`
and `github.com/powerman/gotest`; fixtures for template/env-injection scenarios live in `testdata/`
and `examples/`.

**Backward-compat note:** this fork's template functions (from Sprig) intentionally diverge from
upstream `jwilder/dockerize` — `default`, `contains`, `split`, `replace` behave differently and
`loop` was removed (use `untilStep`). See README's "Using Templates" section before changing
template function wiring.

## Conventions

- New CLI flags: add the `flag.Var`/`flag.XVar` call in `main.go`'s `init()`, wire validation into
  the `switch` in `main()` (pattern: one `case` per invalid combination, reporting via
  `fatalFlagValue`), and document it in the README's Command-line Options / Usage sections
  (required by `.github/CONTRIBUTING.md`).
- Linting is strict (`.golangci.yml` enables most golangci-lint linters); existing `//nolint`
  comments always include a reason — follow that pattern rather than blanket-disabling a linter.
