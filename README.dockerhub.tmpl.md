<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [dockerize](#dockerize)
  - [Quick start](#quick-start)
  - [Image variants](#image-variants)
  - [Core concepts](#core-concepts)
  - [Notable releases](#notable-releases)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# dockerize

Utility to simplify running applications in docker containers: renders config file templates
from environment variables, tails log files to stdout/stderr, and waits for dependent services
(TCP/HTTP/AMQP/file/unix) before starting your app's entrypoint. Also ships a hardened
[Chainguard](https://www.chainguard.dev/chainguard-images)-based image, including a FIPS 140-3
compliant crypto build (`-chainguard-fips`).

📖 **Full documentation for this version (`__TAG__`):**
https://github.com/ezintz/dockerize/blob/__TAG__/README.md

## Quick start

```Dockerfile
FROM ezintz/dockerize:__DOCKER_TAG__
...
ENTRYPOINT dockerize ...
```

## Image variants

| Tag suffix          | Base                        | Notes                                              |
|----------------------|------------------------------|-----------------------------------------------------|
| *(none)*             | Alpine                       | default; has a shell and package manager            |
| `-chainguard`        | `cgr.dev/chainguard/static`  | hardened, no shell, near-zero CVEs, `linux/amd64` + `linux/arm64` only |
| `-chainguard-fips`   | `cgr.dev/chainguard/static`  | same base, built with `GOFIPS140=latest` for FIPS 140-3 crypto |

## Core concepts

- **Templates** -- render config files from Go `text/template` + Sprig funcs using container env vars
- **Tail** -- forward arbitrary log files to stdout/stderr so `docker logs` picks them up
- **Wait** -- block startup until dependent services (tcp/http/amqp/file/unix) are reachable

Also published to [GHCR](https://ghcr.io/ezintz/dockerize) (source of truth for image
provenance/SLSA attestations) and built from the [GitHub repo](https://github.com/ezintz/dockerize).

## Notable releases

<!--
  Docker Hub has no per-version docs, so this is a manually curated list of
  releases that changed something worth knowing before you pick an image tag
  -- not every release. Add an entry here when a release changes what's in
  this description (new image variant, new verification method, etc).
-->

- **[v0.18.0](https://github.com/ezintz/dockerize/blob/v0.18.0/README.md)** -- added the
  Chainguard hardened image and the FIPS 140-3 compliant `-chainguard-fips` build
- **[v0.17.0](https://github.com/ezintz/dockerize/blob/v0.17.0/README.md)** -- migrated releases
  to GoReleaser: SLSA3 provenance, cosign-signed checksums, SBOM

Full history: [GitHub Releases](https://github.com/ezintz/dockerize/releases)
