package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Validate(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		wantErrMessage string
		wantConfig     Config
	}{
		{
			name:           "Failing case with missing 'Kind' and 'scmid'",
			config:         Config{},
			wantErrMessage: `missing value for parameter(s) ["kind,scmid"]`,
		},
		{
			name:           "Reject mixed-case kind",
			wantErrMessage: `kind value "GitHub/PullRequest" must be lowercase`,
			config: Config{
				Kind:  "GitHub/PullRequest",
				ScmID: "default",
			},
			wantConfig: Config{
				Kind:  "github/pullrequest",
				ScmID: "default",
			},
		},
		{
			name:           "Reject removed github action kind",
			wantErrMessage: `action kind "github" was removed in v1; use github/pullrequest`,
			config: Config{
				Kind:  "github",
				ScmID: "default",
			},
			wantConfig: Config{
				Kind:  "github/pullrequest",
				ScmID: "default",
			},
		},
		{
			name:           "Reject removed gitea action kind",
			wantErrMessage: `action kind "gitea" was removed in v1; use gitea/pullrequest`,
			config: Config{
				Kind:  "gitea",
				ScmID: "default",
			},
			wantConfig: Config{
				Kind:  "gitea/pullrequest",
				ScmID: "default",
			},
		},
		{
			name: "Passing case with 'azuredevops/pullrequest'",
			config: Config{
				Kind:  "azuredevops/pullrequest",
				ScmID: "default",
			},
			wantConfig: Config{
				Kind:  "azuredevops/pullrequest",
				ScmID: "default",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := tt.config
			gotErr := sut.Validate()
			if tt.wantErrMessage != "" {
				require.Error(t, gotErr)
				assert.Equal(t, tt.wantErrMessage, gotErr.Error())
				return
			}
			require.NoError(t, gotErr)
			// sut can be mutated so we check the new state
			assert.Equal(t, tt.wantConfig, sut)
		})
	}
}
