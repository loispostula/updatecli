package config

import (
	"fmt"
	"strings"
)

func validateRemovedSettings(manifest map[string]interface{}) error {
	check := func(values map[string]interface{}, path string, removed map[string]string) error {
		for _, key := range sortedKeys(values) {
			if replacement, ok := removed[key]; ok {
				return fmt.Errorf("%s%s was removed in v1; %s", path, key, replacement)
			}
		}
		return nil
	}
	mapping := func(value interface{}) map[string]interface{} {
		m, _ := value.(map[string]interface{})
		return m
	}
	if err := check(manifest, "", map[string]string{
		"title": "use name", "pullrequests": "use actions",
	}); err != nil {
		return err
	}
	if err := check(mapping(manifest["autodiscovery"]), "autodiscovery.", map[string]string{
		"pullrequestid": "use actionid",
	}); err != nil {
		return err
	}
	for _, stage := range []string{"sources", "conditions", "targets", string(sectionActions), "scms"} {
		resources := mapping(manifest[stage])
		for _, id := range sortedKeys(resources) {
			resource := mapping(resources[id])
			path := stage + "." + id + "."
			if err := check(resource, path, map[string]string{
				"scmID": "use scmid", "sourceID": "use sourceid", "depends_on": "use dependson",
				"conditionids": "use dependson entries prefixed with condition# and disableconditions: true",
			}); err != nil {
				return err
			}
			transformers, _ := resource["transformers"].([]interface{})
			for i, transformer := range transformers {
				values := mapping(transformer)
				prefix := fmt.Sprintf("%stransformers[%d].", path, i)
				if err := check(values, prefix, map[string]string{
					"addPrefix": "use addprefix", "addSuffix": "use addsuffix",
					"trimPrefix": "use trimprefix", "trimSuffix": "use trimsuffix",
					"semverInc": "use semverinc", "findSubMatch": "use findsubmatch with pattern and captureindex",
				}); err != nil {
					return err
				}
				if err := check(mapping(values["findsubmatch"]), prefix+"findsubmatch.", map[string]string{
					"captureIndex": "use captureindex",
				}); err != nil {
					return err
				}
			}
			kind, _ := resource["kind"].(string)
			if kind != strings.ToLower(kind) {
				return fmt.Errorf("%skind %q must be lowercase", path, kind)
			}
			if stage == string(sectionActions) && (kind == "github" || kind == "gitea") {
				return fmt.Errorf("%skind %q was removed in v1; use %s/pullrequest", path, kind, kind)
			}
			spec := lowercaseKeys(mapping(resource["spec"])).(map[string]interface{})
			removed := map[string]string{}
			switch stage {
			case "sources", "conditions", "targets":
				switch kind {
				case "json", "toml", "csv":
					removed = map[string]string{"query": "use key with Dasel v3 syntax", "multiple": "use key with a Dasel v3 selector"}
				case "maven":
					removed["url"] = "include the full URL in repository"
				case "cargopackage":
					removed["indexurl"] = "use registry.url"
				}
			case string(sectionActions):
				if kind == "github/pullrequest" {
					removed["automerge"] = "use merge.strategy: auto or manual"
				}
			}
			if err := check(spec, path+"spec.", removed); err != nil {
				return err
			}
			if stage == "scms" {
				if err := check(mapping(spec["commitmessage"]), path+"spec.commitmessage.", map[string]string{
					"title": "set the target name to control the commit title",
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
