# paved-path

One entry point for everything a developer does in this repo: clone, build,
test, scan, image, sign, release — for every service in the repo. If it works
with `make`, it passes in CI: the local path and the pipeline are the same
path.

A small Go fleet (gateway, billing, worker) behind one shared image policy,
one CI template pair, and one release script. The services are deliberately
boring; the point is the tooling around them, and how it holds up when you
add a fourth service (it doesn't take — that's the test).

## Quickstart

```sh
make bootstrap   # one-time: checks Go, installs govulncheck
make test        # run all tests (all services)
make build       # build every service into bin/
make up          # run the gateway locally (docker compose)
curl localhost:8090/healthz
```

Everything else is in `make help`. Per-service targets take `SERVICE=`:
`make image SERVICE=billing`, `make release SERVICE=gateway VERSION=1.2.3`.

## What's inside

| Piece | Where | Why it exists |
|---|---|---|
| Makefile (single entry point) | `Makefile` | One vocabulary for local and CI; fleet-aware (`SERVICE=` or all) |
| Shared image policy | `Dockerfile` | ONE Dockerfile for every service, parameterized by `SERVICE`; pinned base, non-root distroless |
| Local dev environment | `.devcontainer/` | Same toolchain for everyone, including agents |
| CI gates (GitLab) | `.gitlab/ci/base.yml` | lint → test → build → scan; service matrix — new services join every pipeline automatically |
| Release promotion (GitLab) | `.gitlab/ci/release.yml` | tag → promote SHA image, sign, SBOM; nothing else ships |
| CI gates (GitHub Actions) | `.github/workflows/ci.yml` | same gates, second platform — templates stay in lockstep |
| Release promotion (GitHub Actions) | `.github/workflows/release.yml` | same tag pipeline as GitLab: promote, sign (keyless), SBOM to GHCR |
| Local release path | `scripts/release.sh` | per-service SBOM + cosign signature + manifest, and a fleet manifest for `--all` |
| Setup-time measurement | `scripts/setup-time.sh`, `scripts/fleet-metrics.sh` | onboarding as a metric: per-service, with fleet rollup (median/worst) |
| Agent operating rules | `AGENTS.md` | how coding agents work in this repo, versioned with the code |
| Scaling design doc | `docs/SCALE.md` | what changes at 40 services: template repo, runner fleet as access control, telemetry, adoption |

## Design decisions

**One Dockerfile for every service.**
When many services share an image policy, you want one file to review and one
place to change the base pin. The `SERVICE` build arg keeps the per-service
surface to exactly what differs: which main package. Per-service Dockerfiles
mean N review surfaces for one decision.

**The CI matrix is the scaling mechanism.**
Adding a service to the repo adds it to every pipeline with zero per-service
CI editing. Copy-pasted pipelines drift; matrices can't.

**The image is tagged by commit SHA and promoted, never rebuilt.**
Every environment runs the exact artifact that was tested. "Rebuild it for
prod" is how bits change between staging and production.

**Local and CI run the same commands.**
The Makefile targets are what the pipelines invoke. If a gate exists only in
CI, developers meet it for the first time on their PR — the worst possible
place to learn about a rule.

**Releases are a script, not a ritual.**
`scripts/release.sh` and the tag pipeline do the same thing: build, SBOM via
syft, keyless signature via cosign, write a manifest. `--all` adds a fleet
manifest so "what is running in prod" has one answer. A release that depends
on someone remembering the steps is not a release process.

**Setup time is a metric, and it has a fleet rollup.**
`make setup-time` measures fresh-clone → first passing test for one service;
`make fleet-metrics` does it per-service and reports median/worst. It's the
number that tells you whether onboarding got better or worse after any
tooling change — and the fleet rollup is what you show leadership when a new
service lands.

## Steal this

Fork it, replace the Go services with yours, swap the CI platform if you
must. Keep the invariants: one entry point, same gates everywhere, promoted
artifacts, measured onboarding. Then read `docs/SCALE.md` for what changes
when "your repo" becomes "the company's repos." The tools are the easy part;
the invariants and the ownership model are the paved path.

MIT-licensed (see `LICENSE`), so "steal it" is literal: fork, copy the
patterns, or lift whole files. No attribution required, though it's welcome.
