package csv

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (c *CSV) Source(_ context.Context, workingDir string, resultSource *result.Source) error {

	if len(c.contents) > 1 {
		return errors.New("source only supports one file")
	}

	content := c.contents[0]

	sourceOutput := ""

	if err := content.Read(workingDir); err != nil {
		return fmt.Errorf("reading csv file: %w", err)
	}

	query := c.spec.Key
	switch c.engine {
	case ENGINEDASEL_V3:
		logrus.Debugf("Using engine %q", c.engine)
		queryResults, err := content.QueryV3(c.spec.Key)
		if err != nil {
			if strings.Contains(err.Error(), "map key not found") {
				return fmt.Errorf("cannot find value for path %q from file %q",
					c.spec.Key,
					content.FilePath)
			}
			return fmt.Errorf("running query %q: %w", c.spec.Key, err)
		}

		c.foundVersion, err = c.versionFilter.Search(queryResults)
		if err != nil {
			return fmt.Errorf("filtering version: %w", err)
		}
		sourceOutput = c.foundVersion.GetVersion()

	default:
		return fmt.Errorf("engine %q not supported", c.engine)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = sourceOutput
	resultSource.Description = fmt.Sprintf("csv value %q, found in file %q, for path %q",
		sourceOutput,
		content.FilePath,
		query)

	return nil

}
