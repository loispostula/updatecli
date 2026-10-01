package githubsearch

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
)

// discoveredBranch is a branch matching the branch filter, in a repository returned by the search.
type discoveredBranch struct {
	owner      string
	repository string
	branch     string
}

// discoveryKey holds every setting that changes the outcome of a discovery.
type discoveryKey struct {
	url      string
	username string
	token    string
	app      app.Spec
	search   string
	branch   string
	limit    int
}

var (
	/*
		discoveries holds the result of each discovery done by the current process.
		Every manifest using a githubsearch scm runs a discovery when it is loaded, and the
		manifests of a compose file or a policy usually share the same search, which costs
		one GraphQL query per repository found.
	*/
	discoveries     = map[discoveryKey][]discoveredBranch{}
	discoveriesLock sync.Mutex
)

// ScmsGenerator generates GitHub SCM specs based on the search query and branch filter.
func (g GitHubSearch) ScmsGenerator(ctx context.Context) (results []github.Spec, err error) {

	branches, err := g.discover(ctx)
	if err != nil {
		return nil, err
	}

	results = make([]github.Spec, 0, len(branches))

	for _, b := range branches {
		results = append(results, github.Spec{
			App:                    g.spec.App,
			Branch:                 b.branch,
			CommitMessage:          g.spec.CommitMessage,
			CommitUsingAPI:         g.spec.CommitUsingAPI,
			Directory:              g.spec.Directory,
			Depth:                  g.spec.Depth,
			Email:                  g.spec.Email,
			Force:                  g.spec.Force,
			GPG:                    g.spec.GPG,
			Owner:                  b.owner,
			Repository:             b.repository,
			Submodules:             g.spec.Submodules,
			Token:                  g.spec.Token,
			URL:                    g.spec.URL,
			Username:               g.spec.Username,
			User:                   g.spec.User,
			WorkingBranch:          g.spec.WorkingBranch,
			WorkingBranchPrefix:    g.spec.WorkingBranchPrefix,
			WorkingBranchSeparator: g.spec.WorkingBranchSeparator,
		})
	}

	return results, nil
}

// discover returns the branches matching the branch filter in the repositories returned by
// the search, querying GitHub only the first time a process runs a given discovery.
func (g GitHubSearch) discover(ctx context.Context) ([]discoveredBranch, error) {
	key := discoveryKey{
		url:      g.spec.URL,
		username: g.spec.Username,
		token:    g.spec.Token,
		search:   g.search,
		branch:   g.branch,
		limit:    g.limit,
	}
	if g.spec.App != nil {
		key.app = *g.spec.App
	}

	discoveriesLock.Lock()
	defer discoveriesLock.Unlock()

	if branches, ok := discoveries[key]; ok {
		logrus.Debugf("Reusing the GitHub repositories already discovered for %q", g.search)
		return branches, nil
	}

	branches, err := g.queryBranches(ctx)
	if err != nil {
		return nil, err
	}

	discoveries[key] = branches

	return branches, nil
}

// queryBranches queries GitHub for the branches matching the branch filter in the
// repositories returned by the search.
func (g GitHubSearch) queryBranches(ctx context.Context) ([]discoveredBranch, error) {
	results := []discoveredBranch{}

	repositories, err := github.SearchRepositories(g.client, g.search, 0, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed generating spec: %w", err)
	}

	re, err := regexp.Compile(g.branch)
	if err != nil {
		return nil, fmt.Errorf("invalid branch filter %q: %w", g.branch, err)
	}

	for _, repo := range repositories {
		logrus.Debugf("Processing GitHub repository: %s", repo)

		repositoryParts := strings.Split(repo, "/")
		if len(repositoryParts) != 2 {
			return nil, fmt.Errorf("invalid repository format: %s", repo)
		}

		branches, err := github.ListBranches(g.client, repositoryParts[0], repositoryParts[1], 0, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed generating GitHub scm: %w", err)
		}

		for _, b := range branches {
			if !re.MatchString(b) {
				continue
			}

			results = append(results, discoveredBranch{
				owner:      repositoryParts[0],
				repository: repositoryParts[1],
				branch:     b,
			})

			if g.limit > 0 && len(results) >= g.limit {
				return results, nil
			}
		}
	}

	return results, nil
}
