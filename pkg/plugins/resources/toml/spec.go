package toml

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type Spec struct {
	// engine defines the engine used to manipulate the toml file.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   * "dasel/v3" is the default engine used to manipulate toml files
	//
	// accepted values:
	//   * "dasel/v3" for dasel v3 engine
	//   * "dasel" for the latest dasel engine which is currently dasel v3
	Engine *string `yaml:",omitempty"`
	// [s][c][t] File specifies the toml file to manipulate
	File string `yaml:",omitempty"`
	// [c][t] Files specifies a list of Json file to manipulate
	Files []string `yaml:",omitempty"`
	// [s][c][t] Key specifies the query to retrieve an information from a toml file
	Key string `yaml:",omitempty"`
	// [s][c][t] Value specifies the value for a specific key. Default to source output
	Value string `yaml:",omitempty"`
	// [s] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// CreateMissingKey is reserved for validation and must be false.
	// Dasel v3 requires target keys to exist before updating them.
	CreateMissingKey bool `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("toml file undefined")
	// ErrSpecKeyUndefined is returned if a key wasn't specified
	ErrSpecKeyUndefined = errors.New("toml key undefined")
	// ErrSpecFileAndFilesDefines when we both spec File and Files have been specified
	ErrSpecFileAndFilesDefined = errors.New("parameter \"file\" and \"files\" are mutually exclusive")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

const (
	ENGINEDASEL_V3 = "dasel/v3"
	// ENGINEDASEL_LATEST is an alias resolving to the latest dasel engine.
	ENGINEDASEL_LATEST = "dasel"
	ENGINEDEFAULT      = ENGINEDASEL_V3
)

// resolveEngine normalizes a user-provided engine value, resolving the "dasel"
// alias to the latest supported engine. An empty value resolves to the default.
func resolveEngine(engine string) string {
	switch engine {
	case "":
		return ENGINEDEFAULT
	case ENGINEDASEL_LATEST:
		return ENGINEDASEL_V3
	default:
		return engine
	}
}

func (s *Spec) Validate() error {
	var errs []error

	if len(s.File) == 0 && len(s.Files) == 0 {
		errs = append(errs, ErrSpecFileUndefined)
	}
	if len(s.Key) == 0 {
		errs = append(errs, ErrSpecKeyUndefined)
	}

	if len(s.File) > 0 && len(s.Files) > 0 {
		errs = append(errs, ErrSpecFileAndFilesDefined)
	}

	engine := ENGINEDEFAULT
	if s.Engine != nil {
		engine = resolveEngine(*s.Engine)
	}

	// The dasel v3 engine cannot create missing keys (its API resolves the key
	// before setting a value), so "createmissingkey" is incompatible with it.
	if engine == ENGINEDASEL_V3 && s.CreateMissingKey {
		errs = append(errs, fmt.Errorf("engine %q does not support the parameter \"createmissingkey\"", engine))
	}

	if engine != ENGINEDASEL_V3 {
		errs = append(errs, fmt.Errorf("engine %q is not supported; use %q", engine, ENGINEDASEL_V3))
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}
