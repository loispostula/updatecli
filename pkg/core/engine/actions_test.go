package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// mockCleanHandler counts the cleanups an action handler receives.
type mockCleanHandler struct {
	cleanCounter int
}

func (m *mockCleanHandler) CreateAction(ctx context.Context, report *reports.Action, resetDescription bool) error {
	return nil
}

func (m *mockCleanHandler) CleanAction(ctx context.Context, report *reports.Action) error {
	m.cleanCounter++
	return nil
}

func (m *mockCleanHandler) CheckActionExist(ctx context.Context, report *reports.Action) error {
	return nil
}

func TestCleanActions(t *testing.T) {
	type pipelineSpec struct {
		workingBranch string
		published     bool
	}

	testdata := []struct {
		name      string
		pipelines []pipelineSpec
		// expectedCleanCounter reports how many cleanups reached the remote
		expectedCleanCounter int
	}{
		{
			name: "pipelines sharing a pull request clean it once",
			pipelines: []pipelineSpec{
				{workingBranch: "updatecli_main_a"},
				{workingBranch: "updatecli_main_a"},
				{workingBranch: "updatecli_main_a"},
			},
			expectedCleanCounter: 1,
		},
		{
			name: "pipelines with their own pull request clean each of them",
			pipelines: []pipelineSpec{
				{workingBranch: "updatecli_main_a"},
				{workingBranch: "updatecli_main_b"},
			},
			expectedCleanCounter: 2,
		},
		{
			/*
				The pipeline which published comes last, so the cleanup must know about it
				before visiting the pipelines sharing its pull request.
			*/
			name: "a pull request published by any pipeline of this execution is not cleaned",
			pipelines: []pipelineSpec{
				{workingBranch: "updatecli_main_a"},
				{workingBranch: "updatecli_main_a"},
				{workingBranch: "updatecli_main_a", published: true},
			},
			expectedCleanCounter: 0,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			handler := mockCleanHandler{}
			e := Engine{}

			for _, spec := range tt.pipelines {
				var scmHandler scm.ScmHandler = &mockPushScm{
					url:           "https://github.com/updatecli/updatecli.git",
					workingBranch: spec.workingBranch,
					targetBranch:  "main",
				}

				e.Pipelines = append(e.Pipelines, &pipeline.Pipeline{
					Name: "test",
					Targets: map[string]target.Target{
						"default": {},
					},
					Actions: map[string]action.Action{
						"default": {
							Config:    action.Config{Kind: "github/pullrequest"},
							Scm:       &scm.Scm{Handler: scmHandler},
							Handler:   &handler,
							Published: spec.published,
						},
					},
				})
			}

			errs := e.cleanActions(context.Background())

			assert.Empty(t, errs)
			assert.Equal(t, tt.expectedCleanCounter, handler.cleanCounter)
		})
	}
}
