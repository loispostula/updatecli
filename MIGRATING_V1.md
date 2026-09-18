# Migrating to v1

The `v1` branch prepares `1.0.0-rc.1` by removing deprecated interfaces from
`main`. No release is published yet. Continue v0 fixes on `main` and bring them
into this branch during RC testing.

Resolve deprecation warnings on a recent v0 release, migrate the settings below,
then build and test this branch:

```sh
go build -o ./bin/updatecli ./main.go
./bin/updatecli manifest show --config path/to/manifest.yaml
./bin/updatecli pipeline diff --config path/to/manifest.yaml
```

Review the diff before running `pipeline apply`. Removed settings fail validation.
Release builds obtain their version from the tag; this build has no release version.

## Removed interfaces

| Previous interface | Replacement |
| --- | --- |
| `updatecli apply`, `diff`, `prepare` | `updatecli pipeline apply`, `diff`, `prepare` |
| `updatecli show` | `updatecli manifest show` |
| Default compose filename `update-compose.yaml` | Rename to `updatecli-compose.yaml` |
| Manifest `title` | `name` |
| Manifest `pullrequests` | `actions` with an explicit action `kind` |
| `autodiscovery.pullrequestid` | `autodiscovery.actionid` |
| Resource/action `scmID`, resource `sourceID`, `depends_on` | `scmid`, `sourceid`, `dependson` |
| Target `conditionids` | `dependson: [condition#id]` and `disableconditions: true` to select specific conditions |
| Mixed-case resource, action, or SCM `kind` | Lowercase `kind` |
| Action kinds `github`, `gitea` | `github/pullrequest`, `gitea/pullrequest` |
| GitHub action `spec.automerge: true` | `spec.merge.strategy: auto` |
| GitHub action `spec.automerge: false` | Omit `merge`, or use `spec.merge.strategy: manual` |
| SCM `spec.commitmessage.title` | Set the target `name` used for the commit title |
| Transformer `addPrefix`, `addSuffix`, `trimPrefix`, `trimSuffix`, `semverInc`, `findSubMatch` | `addprefix`, `addsuffix`, `trimprefix`, `trimsuffix`, `semverinc`, `findsubmatch` |
| Transformer `findsubmatch.captureIndex` | `findsubmatch.captureindex` |
| Maven `spec.url` | Put the complete URL in `spec.repository` |
| Cargo package `spec.indexurl` | `spec.registry.url` |
| GitHub release `spec.key: name` or `hash` | `tagname` or `taghash` |
| JSON/TOML/CSV `spec.query`, `spec.multiple` | `spec.key` with a Dasel v3 selector |
| JSON/TOML/CSV `engine: dasel/v1` or `dasel/v2` | `engine: dasel/v3` |
| YAML keys without `$`, or escaped dots such as `foo\.bar` | Explicit JSONPath, for example `$.version` or `$.'foo.bar'` |

An explicitly supplied compose filename still works with `--file`.
GitLab's `automerge` option is unrelated to the removed GitHub setting and remains
supported. A GitHub release `key: title` also remains supported.

## Dasel v3

JSON, TOML, and CSV default to `dasel/v3`; `dasel` aliases v3. Migrate selectors
as well as the engine setting. Test source selection and target writes.

| Selection | `key` |
| --- | --- |
| JSON/TOML property | `version` |
| Property containing dots | `get("field.with.dots")` |
| All versions in an array | `map(version)...` |
| First CSV row's version | `$this[0].version` |

TOML `createmissingkey: true` is unsupported. Create the key before updating it.

## Release configuration

RCs skip GitHub's latest release, stable container tags, Homebrew/AUR uploads,
announcements, and downstream notifications. Branch CI runs short tests;
integration tests requiring upstream credentials are skipped on forks.

Publishing still requires configuring the fork's destinations and credentials;
the release workflow retains upstream registries. Creating a tag is a separate
publishing step.

For opt-in adoption, pin a v0 CLI version and constrain update policies to
`<1.0.0`. Test published RCs by explicit version. The Updatecli Action's CLI
default and update policy need separate changes in that repository; this branch
does not prevent automatic adoption of the final v1 release by those consumers.
