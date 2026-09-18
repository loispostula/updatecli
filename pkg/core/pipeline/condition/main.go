package condition

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrWrongConfig is returned when a condition spec has missing attributes which are mandatory
	ErrWrongConfig = errors.New("wrong condition configuration")
)

// Condition defines which condition needs to be met
// in order to update targets based on the source output
type Condition struct {
	// Result stores the condition result after a condition run.
	Result *result.Condition
	// Config defines condition input parameters
	Config Config
	// Scm stores scm information
	Scm *scm.ScmHandler
}

// Config defines conditions input parameters
type Config struct {
	resource.ResourceConfig `yaml:",inline,omitempty"`
	// sourceid specifies which "source", based on its ID, is used to retrieve the default value.
	SourceID string `yaml:",omitempty"`
	// disablesourceinput disable the mechanism to retrieve a default value from a source.
	DisableSourceInput bool `yaml:",omitempty"`
	// FailWhen allows to reverse a condition expected result from true to false.
	FailWhen bool `yaml:",omitempty"`
}

// Run tests if a specific condition is true
func (c *Condition) Run(ctx context.Context, source string) (err error) {

	var consoleOutput bytes.Buffer
	// By default logrus logs to stderr, so I guess we want to keep this behavior...
	logrus.SetOutput(io.MultiWriter(os.Stdout, &consoleOutput))
	/*
		The last defer will be executed first,
		so in this case we want to first save the console output
		before setting back the logrus output to stdout.
	*/
	// By default logrus logs to stdout and we want to keep this behavior.
	defer logrus.SetOutput(os.Stdout)
	defer c.Result.SetConsoleOutput(&consoleOutput)

	c.Result.Result = result.FAILURE

	condition, err := resource.New(c.Config.ResourceConfig)
	if err != nil {
		return err
	}

	if len(c.Config.Transformers) > 0 {
		source, err = c.Config.Transformers.Apply(source)
		if err != nil {
			return err
		}
	}

	var s scm.ScmHandler
	if c.Scm != nil {
		// If scm is defined then clone the repository
		s = *c.Scm
		if err != nil {
			return err
		}

		err = s.Checkout()
		if err != nil {
			return err
		}
	}

	ok, message, err := condition.Condition(ctx, source, s)
	if ok {
		c.Result.Result = result.SUCCESS
		c.Result.Pass = true
	} else {
		c.Result.Result = result.FAILURE
		c.Result.Pass = false
	}
	c.Result.Description = message
	if err != nil {
		return err
	}

	// FailWhen is used to reverse the expected condition value
	// If failwhen is set to true, then we expected a condition returning "true" would be considered as a failure
	if c.Config.FailWhen {
		logrus.Debugf("Expected successful condition result to be %v", !c.Config.FailWhen)
		if c.Result.Pass {
			c.Result.Result = result.FAILURE
			c.Result.Pass = false
		} else {
			c.Result.Result = result.SUCCESS
			c.Result.Pass = true
		}
	}

	logrus.Infof("%s %s", c.Result.Result, c.Result.Description)

	return nil
}

// JSONSchema implements the json schema interface to generate the "condition" jsonschema.
func (c Config) JSONSchema() *jschema.Schema {

	type configAlias Config
	anyOfSpec := resource.GetResourceMapping()

	return jsonschema.AppendOneOfToJsonSchema(configAlias{}, anyOfSpec)
}

// Validate checks if a condition configuration is valid
func (c *Config) Validate() error {
	gotError := false
	missingParameters := []string{}

	// Validate that kind is set
	if len(c.Kind) == 0 {
		missingParameters = append(missingParameters, "kind")
	}

	// Ensure kind is lowercase
	if c.Kind != strings.ToLower(c.Kind) {
		return fmt.Errorf("kind value %q must be lowercase", c.Kind)
	}

	err := c.Transformers.Validate()
	if err != nil {
		return err
	}

	if len(c.SourceID) > 0 && c.DisableSourceInput {
		logrus.Errorln("disablesourceinput is incompatible with sourceid, ignoring the latter")
		gotError = true
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing value for parameter(s) [%q]", strings.Join(missingParameters, ","))
		gotError = true
	}

	if gotError {
		return ErrWrongConfig
	}

	return nil
}
