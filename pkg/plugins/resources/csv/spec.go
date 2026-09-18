package csv

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type Spec struct {
	// engine defines the engine used to manipulate the csv file.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   * "dasel/v3" is the default engine used to manipulate csv files
	//
	// accepted values:
	//   * "dasel/v3" for dasel v3 engine
	//   * "dasel" for the latest dasel engine which is currently dasel v3
	Engine *string `yaml:",omitempty"`
	// [s][c][t] File specifies the csv file
	File string `yaml:",omitempty"`
	// [c][t] Files specifies a list of Json file to manipulate
	Files []string `yaml:",omitempty"`
	// [s][c][t] Key specifies the csv query
	Key string `yaml:",omitempty"`
	// [s][c][t] Key specifies the csv value, default to source output
	Value string `yaml:",omitempty"`
	// [s][c][t] Comma specifies the csv separator character, default ","
	Comma rune `yaml:",omitempty"`
	// [s][c][t] Comma specifies the csv comment character, default "#"
	Comment rune `yaml:",omitempty"`
	// [s]VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

var (
	ErrSpecFileUndefined       = errors.New("csv file undefined")
	ErrSpecKeyUndefined        = errors.New("csv key undefined")
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

	if engine != ENGINEDASEL_V3 {
		errs = append(errs, fmt.Errorf("engine %q is not supported; use %q", engine, ENGINEDASEL_V3))
	}

	if len(errs) > 0 {
		for i := range errs {
			logrus.Errorln(errs[i])
		}
		return ErrWrongSpec
	}

	return nil
}
