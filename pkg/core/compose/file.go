package compose

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

var (
	// DefaultComposeFilename is the default name of the Updatecli compose file
	DefaultComposeFilename = "updatecli-compose.yaml"
)

// GetDefaultComposeFilename returns the default compose filename.
func GetDefaultComposeFilename() string {
	return DefaultComposeFilename
}

// LoadFile loads an Updatecli compose file into a compose Spec
func LoadFile(filename string) (*Spec, error) {
	var composeSpec Spec

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening Updatecli compose file %q: %s", filename, err)
	}
	defer f.Close()

	composeFileByte, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading Updatecli compose file %q: %s", filename, err)
	}

	err = yaml.Unmarshal(composeFileByte, &composeSpec)
	if err != nil {
		return nil, fmt.Errorf("parsing Updatecli compose file %q: %s", filename, err)
	}

	// Reported before the spec is post-processed, so that a problem points at what the
	// compose file actually says rather than at a value Updatecli defaulted.
	if ValidateSchema {
		validateSchema(filename, composeFileByte)
	}

	composeSpec.Env_files = relativePathToFile(filename, composeSpec.Env_files)

	for i := range composeSpec.Policies {
		if composeSpec.Policies[i].Name == "" {
			composeSpec.Policies[i].Name = fmt.Sprintf("policy-%d - local", i)
			if composeSpec.Policies[i].Policy != "" {
				composeSpec.Policies[i].Name = fmt.Sprintf("policy-%d- %s", i, composeSpec.Policies[i].Policy)
			}
		}
		composeSpec.Policies[i].Config = relativePathToFile(filename, composeSpec.Policies[i].Config)
		composeSpec.Policies[i].Values = relativePathToFile(filename, composeSpec.Policies[i].Values)
		composeSpec.Policies[i].Secrets = relativePathToFile(filename, composeSpec.Policies[i].Secrets)
	}

	return &composeSpec, nil
}

func relativePathToFile(configFileName string, paths []string) []string {
	for i := range paths {
		if filepath.IsAbs(paths[i]) {
			continue
		}
		paths[i] = filepath.Join(filepath.Dir(configFileName), paths[i])
	}

	return paths
}
