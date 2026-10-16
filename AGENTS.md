# AGENTS.md — operating rules for coding agents in this repo

This file is the contract between humans and agents working in this
repository. It is versioned with the code on purpose: when the rules change,
the change is reviewable like any other diff.

## Before you commit

1. `make test` and `make lint` must pass. No exceptions, no "it's a small
   change."
2. If you touched `Dockerfile`, run `make image` and confirm it builds.
3. If you touched a CI template, update **both** `.gitlab/ci/` and
   `.github/workflows/` in the same commit. The two platforms are a pair;
   they drift when they're edited separately. Then run `make ci-lint` — it
   lints the workflow files (actionlint + shellcheck, same as CI) and checks
   the GitLab templates parse. There's no `.gitlab-ci.yml` to feed `glab ci
   lint` because this repo *is* the template; the YAML check is the local
   stand-in.

## How the CI is shaped (don't break this)

- Branches/PRs run `ci.yml` / `base.yml`: lint -> test (+race) -> build+scan.
- Tags run the release pipeline (`release.yml` on both platforms): the same
  gates, then promotion. Promotion retags the SHA image that passed; it
  never rebuilds. Tag rules match stable versions (`v1.2.3`) and release
  candidates (`v1.2.3-rc.1`); the GitLab regex and the GitHub filter
  patterns must stay in sync.
- Tool versions in CI (trivy, cosign, syft, govulncheck) are matched across
  both platforms on purpose. If you bump one, bump the other in the same
  commit.
- The Go version is declared in `go.mod` and must match the toolchain image
  pins (`Dockerfile`, `.gitlab/ci/base.yml`). Bumping Go is what keeps the
  image scan clean: older toolchains embed stdlib CVEs in static binaries,
  which is why the scan gate has no ignore file.
- Actions are pinned by commit SHA with a version comment; base images and
  CI tool images are pinned by digest. Don't "tidy" these into tags.

## Long-running processes: fully detach them

A watcher, server, or daemon started from a shell that waits for its output
will hang the session that started it. Start background processes like this:

```sh
(nohup cmd </dev/null >> /tmp/cmd.log 2>&1 &) ; sleep 2; tail -5 /tmp/cmd.log
```

The subshell, the stdin redirect, and the output redirects are all required.
Poll logs with plain `tail`; never `tail -f`.

## Searches: scope them

This repo is small, but the machine it runs on is not. Never sweep a
filesystem to find something you could name directly: search the repo, not
`/`. A full-filesystem scan on a machine with large data directories can
stall for minutes and blocks everyone behind it.

## Decide, don't deliberate

When a question comes up mid-task that a human can't answer until you stop:
answer it yourself with the best available assumption, state the assumption
in your summary, and keep moving. Prefer the smallest reversible step when
options tie. A wrong-but-stated assumption costs one correction; an open
question costs the whole task.

## What you must not do without a human saying so

- Push to `main` or create tags (tags trigger the release pipeline)
- Change base image pins in `Dockerfile` (supply-chain decision, human call)
- Add dependencies to `go.mod` beyond what the task requires
- Edit this file (it's the contract; changes to it get human review first)

## When you're done

Summarize: what changed, which `make` targets you ran and their results, any
assumptions you made, and anything a human should look at twice.
