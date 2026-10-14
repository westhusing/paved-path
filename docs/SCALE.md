# Scaling the paved path: what actually changes

This repo demonstrates a paved path at n=3 services. This document is the
answer to the question every interviewer asks next: *how does this work when
it's 40 services, 6 teams, and a runner fleet?*

The short version: **the invariants don't change; the ownership model does.**
A paved path that scales is not a bigger version of the same repo — it's the
same invariants with different owners at each layer.

## Layer 1: The repo (what this demonstrates)

One team, one repo, N services. Everything is in-repo: the shared Dockerfile
(one image policy for all services, parameterized by `SERVICE`), the CI
templates (service matrix — adding a service adds it to every pipeline with
zero per-service editing), the release script (per-service manifests plus a
fleet manifest), and fleet metrics (per-service setup time with a rollup).

The design choices that matter at this layer:

- **One Dockerfile, parameterized.** When 40 services share an image policy,
  you want one file to review and one place to change the base pin. A
  per-service Dockerfile means 40 review surfaces for what is really one
  decision. The `SERVICE` build arg keeps the per-service surface to exactly
  what differs: which main package.
- **Matrix, not copy-paste.** The CI matrix is the difference between "add a
  service" being a one-line change and a per-service CI edit. Copy-pasted
  pipelines drift; matrices can't.
- **Fleet manifest.** A release is not "each service shipped its own tag."
  The fleet manifest records which version of every service was released
  together, so "what is running in prod" has one answer.

## Layer 2: The platform (where the repo stops being enough)

At ~10–20 services across several teams, three things stop working in-repo:

**1. The templates need a home of their own.**
In-repo CI templates work until team B copies them and starts diverging. The
fix is a **template repo** (or GitLab's include-from-registry) with its own
versioning and release notes. Services `include` a pinned template version;
upgrading the template is an explicit, reviewable change per service — which
is exactly what you want, because a template upgrade is a supply-chain event.

**2. Runners stop being "the CI machine."**
A self-hosted runner fleet needs: capacity planning (how many concurrent
jobs, what happens when the queue exceeds it), isolation (which projects can
use which runner tags — a shared `docker`-capable runner is a privilege
boundary), and lifecycle (OS patching, Docker upgrades, disk cleanup). The
paved path at this layer is **runner tags as an access-control model**: a
service declares what it needs (`docker`, `gpu`, `arm64`) and the platform
team owns what those tags mean.

**3. Metrics move from script to telemetry.**
`fleet-metrics.sh` is the right tool at n=3. At n=40 you want pipeline
duration, queue wait, and setup time as **continuous telemetry** (OpenTelemetry
into your existing observability stack), with dashboards and alerts — because
the question changes from "what's the number?" to "is the number *trending*
bad, and which service caused it?" The script's JSON output is deliberately
shaped like the telemetry record, so the migration is a sink change, not a
schema change.

## Layer 3: The organization (where most paved paths actually fail)

This is the part that isn't in any repo, and where I'd spend most of a
Staff-level interview:

**Adoption is the product.** The paved path succeeds or fails on whether
teams *choose* it. The mechanism is making the alternative measurably worse:
onboarding a service on the paved path takes an afternoon; off it, it takes
a week of CI archaeology. You don't mandate the paved path — you make the
off-path cost visible (the fleet metrics dashboard shows both) and let teams
feel it. Mandates without a working alternative produce compliance, not
adoption; the moment the path breaks for a team, they leave and don't come back.

**The paved path is a product with users.** That means: a changelog and
release notes for template changes, a migration guide when you change the
base image (with a deadline and an owner), a way for teams to file friction
that gets triaged like bugs, and — the part everyone forgets — **a sunset
process for what you stop supporting**. A paved path that never deprecates
anything becomes a museum.

**Escalation without authority.** When team C says "the paved path doesn't
work for our batch workloads," you don't have their org chart. You do have:
their pipeline data (queue wait, failure rate), the cost of their workaround
(measured), and a proposal that changes the template for everyone (reviewed).
The paved path team's leverage is *evidence plus a working alternative*, not
rank.

**What I would NOT do at scale:**
- Build an internal developer platform (IDP) before the paved path is
  adopted. The IDP is a UI over the paved path; without adoption underneath,
  it's a portal nobody opens.
- Centralize the build cache before standardizing the base image. Caching a
  fleet of divergent images is caching entropy — you get hit rates that look
  fine and builds that are still slow.
- Make the paved path team own service code reviews. The moment platform
  engineers approve feature PRs, they're a bottleneck with opinions; the
  paved path team should own *the path*, and teams own what runs on it.

## The one-paragraph version (for the interview)

The invariants — one entry point, same gates everywhere, promoted artifacts,
measured onboarding — are scale-invariant. What changes is ownership: the
templates graduate to a versioned template repo, runners become an
access-controlled fleet with tags as the privilege model, metrics move from a
script to continuous telemetry, and the paved path team becomes a product
team with adoption as its north-star metric. The failure mode I watch for is
organizational, not technical: paved paths die when they're mandated instead
of adopted, or when the team that owns them becomes a bottleneck for the code
that runs on them.
