package toml

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Toml) Source(_ context.Context, workingDir string, resultSource *result.Source) error {

	if len(t.contents) > 1 {
		return errors.New("source only supports one file")
	}

	content := t.contents[0]

	sourceOutput := ""

	if err := content.Read(workingDir); err != nil {
		return fmt.Errorf("reading toml file: %w", err)
	}

	query := t.spec.Key
	switch t.engine {
	case ENGINEDASEL_V3:
		logrus.Debugf("Using engine %q", t.engine)
		queryResults, err := content.QueryV3(t.spec.Key)
		if err != nil {
			if strings.Contains(err.Error(), "map key not found") {
				return fmt.Errorf("cannot find value for path %q from file %q",
					t.spec.Key,
					content.FilePath)
			}
			return fmt.Errorf("running query %q: %w", t.spec.Key, err)
		}

		t.foundVersion, err = t.versionFilter.Search(queryResults)
		if err != nil {
			return fmt.Errorf("filtering result: %w", err)
		}
		sourceOutput = t.foundVersion.GetVersion()

	default:
		return fmt.Errorf("engine %q not supported", t.engine)
	}

	resultSource.Information = sourceOutput
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("value %q, found in file %q, for key %q'",
		sourceOutput,
		content.FilePath,
		query)

	return nil
}
