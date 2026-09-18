package toml

import (
	"errors"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

var (
	// ErrDaselFailedParsingTOMLByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingTOMLByteFormat error = errors.New("fail to parse Toml data")
)

// Toml stores configuration about the file and the key value which needs to be updated.
type Toml struct {
	spec     Spec
	contents []dasel.FileContent
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// engine defines the engine used to manipulate the toml file
	// If not set, the default engine is dasel/v3
	engine string
}

func New(spec interface{}) (*Toml, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	err = newSpec.Validate()

	if err != nil {
		return nil, err
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	engine := ENGINEDEFAULT
	if newSpec.Engine != nil {
		// Resolve aliases (e.g. bare "dasel" -> latest engine) so the rest of the
		// code only ever deals with explicit engine versions.
		engine = resolveEngine(*newSpec.Engine)
	}

	j := Toml{
		spec:          newSpec,
		versionFilter: newFilter,
		engine:        engine,
	}

	// Init currentContents
	switch len(j.spec.File) > 0 {
	case true:
		j.contents = append(
			j.contents, dasel.FileContent{
				DataType:         "toml",
				FilePath:         j.spec.File,
				ContentRetriever: &text.Text{},
			})

	case false:
		for i := range j.spec.Files {
			j.contents = append(
				j.contents, dasel.FileContent{
					DataType:         "toml",
					FilePath:         j.spec.Files[i],
					ContentRetriever: &text.Text{},
				})
		}
	}

	return &j, err
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (s *Toml) ReportConfig() interface{} {
	return Spec{
		File:             s.spec.File,
		Files:            s.spec.Files,
		Key:              s.spec.Key,
		Value:            s.spec.Value,
		VersionFilter:    s.spec.VersionFilter,
		CreateMissingKey: s.spec.CreateMissingKey,
		Engine:           s.spec.Engine,
	}
}
