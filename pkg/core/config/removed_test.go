package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemovedManifestSettings(t *testing.T) {
	tests := []struct{ name, manifest, message string }{
		{"title", "title: old", "title was removed"},
		{"pullrequests", "pullrequests: {}", "pullrequests was removed"},
		{"autodiscovery", "autodiscovery: {pullrequestid: default}", "autodiscovery.pullrequestid was removed"},
		{"scmid", "sources: {version: {kind: shell, scmID: default}}", "sources.version.scmID was removed"},
		{"sourceid", "targets: {version: {kind: file, sourceID: version}}", "targets.version.sourceID was removed"},
		{"depends_on", "conditions: {check: {kind: shell, depends_on: []}}", "conditions.check.depends_on was removed"},
		{"conditions", "targets: {version: {kind: file, conditionids: []}}", "conditionids was removed"},
		{"action kind", "actions: {pr: {kind: github}}", "use github/pullrequest"},
		{"gitea kind", "actions: {pr: {kind: gitea}}", "use gitea/pullrequest"},
		{"kind case", "sources: {version: {kind: Shell}}", "must be lowercase"},
		{"transformer", "sources: {version: {kind: shell, transformers: [{addPrefix: v}]}}", "use addprefix"},
		{"capture index", "sources: {version: {kind: shell, transformers: [{findsubmatch: {pattern: x, captureIndex: 0}}]}}", "use captureindex"},
		{"query", "sources: {version: {kind: json, spec: {query: .version}}}", "use key with Dasel v3 syntax"},
		{"multiple false", "targets: {version: {kind: toml, spec: {multiple: false}}}", "multiple was removed"},
		{"plugin case", "targets: {version: {kind: csv, spec: {Multiple: false}}}", "multiple was removed"},
		{"automerge false", "actions: {pr: {kind: github/pullrequest, spec: {automerge: false}}}", "use merge.strategy"},
		{"commit title", "scms: {repo: {kind: git, spec: {commitmessage: {title: old}}}}", "commitmessage.title was removed"},
		{"maven", "sources: {version: {kind: maven, spec: {url: https://example.com}}}", "include the full URL in repository"},
		{"cargo", "sources: {version: {kind: cargopackage, spec: {indexurl: https://example.com}}}", "use registry.url"},
		{"alias", "old: &old {kind: json, spec: {query: .version}}\nsources: {version: *old}", "query was removed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var specs []Spec
			err := unmarshalConfigSpec([]byte(tt.manifest), &specs)
			require.ErrorContains(t, err, tt.message)
			require.Empty(t, specs)
		})
	}
}

func TestRemovedSettingsDoNotRejectPluginPayloads(t *testing.T) {
	var specs []Spec
	err := unmarshalConfigSpec([]byte(`
name: Current manifest
sources:
  version:
    kind: json
    spec:
      file: package.json
      key: version
      versionfilter:
        kind: semver
        pattern: '*'
actions:
  pr:
    kind: gitlab/mergerequest
    title: Release
    spec:
      automerge: false
conditions:
  api:
    kind: http
    spec:
      headers:
        title: permitted
        scmID: permitted
`), &specs)
	require.NoError(t, err)
	require.Len(t, specs, 1)
}

func TestRemovedSettingsInLaterDocument(t *testing.T) {
	var specs []Spec
	err := unmarshalConfigSpec([]byte(fmt.Sprintf("name: valid\n---\n%s", "title: removed")), &specs)
	require.ErrorContains(t, err, "title was removed")
}
