package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunActions runs all actions defined in the configuration.
func (e *Engine) runActions(ctx context.Context) error {

	errs := []string{}

	logrus.Infof("\n\n%s\n", strings.ToTitle("Actions"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Actions")+1))

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if len(pipeline.Actions) > 0 {
			if err := pipeline.RunActions(ctx); err != nil {
				errs = append(errs, err.Error())
				pipeline.Report.Result = result.FAILURE
				logrus.Errorf("action stage:\t%q", err.Error())
				continue
			}
		}
	}

	errs = append(errs, e.cleanActions(ctx)...)

	if len(errs) > 0 {
		return fmt.Errorf(
			"errors occurred while running actions:\n\t* %s",
			strings.Join(errs, "\n\t* "))
	}

	return nil
}

// cleanActions cleans up the actions published by previous executions, such as pull requests
// that don't carry any change anymore, and returns the errors it met.
func (e *Engine) cleanActions(ctx context.Context) []string {
	errs := []string{}

	logrus.Infof("Cleaning up actions published by previous executions")

	/*
		Pipelines sharing a remote object, such as the pull request of every pipeline an
		autodiscovery crawler generated for one repository, only need one cleanup.
		Keys published during this execution are recorded first: a pipeline can be visited
		before the one that published to the same remote object, and must not clean it.
	*/
	handled := map[string]bool{}
	for _, pipeline := range e.Pipelines {
		for _, a := range pipeline.Actions {
			if key := a.CleanupKey(); a.Published && key != "" {
				handled[key] = true
			}
		}
	}

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if len(pipeline.Actions) > 0 {
			if err := pipeline.RunCleanActions(ctx, handled); err != nil {
				errs = append(errs, "cleaning: "+err.Error())
				pipeline.Report.Result = result.FAILURE
				logrus.Errorf("cleaning action stage:\t%q", err.Error())
				continue
			}
		}
	}

	return errs
}
