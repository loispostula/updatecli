package gitbranch

import (
	"fmt"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// listRemoteURLBranches lists the branches of the remote repository, sorted by name, without their dates.
func (gb *GitBranch) listRemoteURLBranches() ([]gitgeneric.DatedBranch, error) {
	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{gb.spec.URL},
	})

	listOptions := &git.ListOptions{}
	if gb.spec.Username != "" && gb.spec.Password != "" {
		listOptions.Auth = &http.BasicAuth{
			Username: gb.spec.Username,
			Password: gb.spec.Password,
		}
	}

	refs, err := remote.List(listOptions)
	if err != nil {
		return nil, err
	}

	branches := []gitgeneric.DatedBranch{}
	for _, ref := range refs {
		if !ref.Name().IsBranch() {
			continue
		}
		branches = append(branches, gitgeneric.DatedBranch{
			Name: ref.Name().Short(),
			Hash: ref.Hash().String(),
		})
	}

	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Name < branches[j].Name
	})

	return branches, nil
}

// remoteBranchCondition checks that the branch exists on the remote repository.
func (gb *GitBranch) remoteBranchCondition(source string) (bool, string, error) {
	gb.branch = source
	if gb.spec.Branch != "" {
		gb.branch = gb.spec.Branch
	}

	branches, err := gb.listRemoteURLBranches()
	if err != nil {
		return false, "", fmt.Errorf("listing remote branches: %w", err)
	}

	for _, b := range branches {
		if b.Name == gb.branch {
			return true, fmt.Sprintf("git branch %q matching", gb.branch), nil
		}
	}

	return false, fmt.Sprintf("git branch %q not found", gb.branch), nil
}
