package csv

import (
	"errors"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

var (
	// ErrDaselFailedParsingJSONByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingJSONByteFormat error = errors.New("fail to parse Json data")
)

const (
	// DEFAULTSEPARATOR defines the default csv separator
	DEFAULTSEPARATOR rune = ','
	// DEFAULTCOMMENT defines the default comment character
	DEFAULTCOMMENT rune = '#'
)

// CSV stores configuration about the file and the key value which needs to be updated.
type CSV struct {
	spec     Spec
	contents []csvContent
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// engine defines the engine used to manipulate the csv file
	// If not set, the default engine is dasel/v3
	engine string
}

func New(spec interface{}) (*CSV, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	comma := DEFAULTSEPARATOR
	comment := DEFAULTCOMMENT
	var emptyRune rune
	if newSpec.Comma != emptyRune {
		comma = newSpec.Comma
	}

	if newSpec.Comment != emptyRune {
		comment = newSpec.Comment
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	if err := newSpec.Validate(); err != nil {
		return nil, err
	}

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

	c := CSV{
		spec:          newSpec,
		versionFilter: newFilter,
		engine:        engine,
	}

	// Init currentContents
	switch len(c.spec.File) > 0 {
	case true:
		c.contents = append(
			c.contents, csvContent{
				comma:   comma,
				comment: comment,
				FileContent: dasel.FileContent{
					DataType:         "csv",
					FilePath:         c.spec.File,
					ContentRetriever: &text.Text{},
				},
			})

	case false:
		for i := range c.spec.Files {
			c.contents = append(
				c.contents, csvContent{
					comma:   comma,
					comment: comment,
					FileContent: dasel.FileContent{
						DataType:         "csv",
						FilePath:         c.spec.Files[i],
						ContentRetriever: &text.Text{},
					},
				})
		}
	}

	return &c, err
}

// ReportConfig returns a cleaned version of the configuration
// to identify the resource without any sensitive information or context specific data.
func (c *CSV) ReportConfig() interface{} {
	return Spec{
		File:   c.spec.File,
		Files:  c.spec.Files,
		Key:    c.spec.Key,
		Value:  c.spec.Value,
		Engine: c.spec.Engine,
	}
}
