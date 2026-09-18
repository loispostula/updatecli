package json

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (j *Json) Source(_ context.Context, workingDir string, resultSource *result.Source) error {

	if len(j.contents) > 1 {
		return errors.New("source only supports one file")
	}

	content := j.contents[0]

	sourceOutput := ""
	if err := content.Read(workingDir); err != nil {
		return fmt.Errorf("reading json file: %w", err)
	}

	query := j.spec.Key
	switch j.engine {
	case ENGINEDASEL_V3:
		logrus.Debugf("Using engine %q", j.engine)
		queryResults, err := content.QueryV3(j.spec.Key)

		if err != nil {
			if strings.Contains(err.Error(), "map key not found") {
				err := fmt.Errorf("%s cannot find value for path %q from file %q",
					result.FAILURE,
					j.spec.Key,
					content.FilePath)
				return err
			}
			return fmt.Errorf("running query %q: %w", j.spec.Key, err)
		}

		j.foundVersion, err = j.versionFilter.Search(queryResults)
		if err != nil {
			return fmt.Errorf("filtering information: %w", err)
		}
		sourceOutput = j.foundVersion.GetVersion()

	default:
		return fmt.Errorf("engine %q not supported", j.engine)
	}

	resultSource.Information = sourceOutput
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("value %q, found in file %q, for key %q",
		sourceOutput,
		content.FilePath,
		query)

	return nil
}
