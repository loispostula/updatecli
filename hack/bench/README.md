# Read-only org sweep benchmark

Runs `updatecli compose apply --push=false --commit=false --labels <label>` against a real
compose file, traces every HTTP call into a local OpenTelemetry collector, and compares builds.

`--push=false --commit=false` never commits, pushes or opens a pull request, but the clean stage
still runs and closes pull requests without changes. Build every binary with `patches/` applied:
0001 logs instead of closing, 0002 marks the clean stage in the log.

```shell
git apply hack/bench/patches/*.diff      # on top of the branch to measure
go build -o hack/bench/bin/updatecli-<variant> .
```

Credentials: `UPDATECLI_GITHUB_TOKEN`, or updatecli's `UPDATECLI_GITHUB_APP_*` variables. Every
run waits until the credentials' GraphQL budget has `MIN_BUDGET` points left (default 7000, set it
close to your hourly limit) so no variant inherits another's rate-limit waits.

```shell
COMPOSE_DIR=/path/to/compose/checkout nix shell nixpkgs#opentelemetry-collector-contrib nixpkgs#uv \
  nixpkgs#pnpm nixpkgs#nodejs -c hack/bench/run-one.sh <variant> ecosystem:docker
python3 hack/bench/analyze.py
```
