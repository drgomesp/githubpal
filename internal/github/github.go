package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v48/github"

	"golang.org/x/oauth2"
)

type KindSort int
type KindDirection int

const (
	CREATED KindSort = iota
	UPDATED
	PUSHED
	FULL_NAME
)

const (
	ASC KindDirection = iota
	DESC
)

func (k KindSort) String() string {
	return [...]string{"created", "updated", "pushed", "full_name"}[k]
}

func (k KindDirection) String() string {
	return [...]string{"asc", "desc"}[k]
}

type Github struct {
	client *github.Client
	cfg    *CfgParams
}

func NewGithub(opts []ConfigOpts) *Github {
	c := &CfgParams{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			panic(err)
		}
	}

	return &Github{
		cfg: c,
	}
}

func (g *Github) Auth(ctx context.Context) error {

	if g.cfg.GetToken() == "" {
		return fmt.Errorf("'GH_TOKEN' not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.cfg.GetToken()},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	g.client = client

	return nil
}

func (g *Github) GetRepo(ctx context.Context, org, repo string) (*github.Repository, error) {

	repos, _, err := g.client.Repositories.Get(ctx, org, repo)

	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (g *Github) GetRepos(ctx context.Context, sort KindSort,
	direct KindDirection, owner string) ([]*github.Repository, error) {
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 50},
		Sort:        sort.String(),
		Direction:   direct.String(),
	}
	var allRepos []*github.Repository

	for {
		repos, resp, err := g.client.Repositories.List(ctx, owner, opt)
		if err != nil {
			return allRepos, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (g *Github) GetRepoBranches(ctx context.Context, org,
	repo string) ([]*github.Branch, error) {

	branches, _, err := g.client.Repositories.ListBranches(ctx, org, repo, nil)

	if err != nil {
		return nil, err
	}

	return branches, nil
}

func (g *Github) GetRepoBranch(ctx context.Context, org,
	repo, branch string) (*github.Branch, error) {

	branches, _, err := g.client.Repositories.GetBranch(ctx, org, repo, branch, false)

	if err != nil {
		return nil, err
	}

	return branches, nil
}

func (g *Github) GetReposByNames(ctx context.Context, sort KindSort,
	direct KindDirection, owner string, repos []string) ([]*github.Repository, error) {

	var result []*github.Repository

	reposgh, err := g.GetRepos(ctx, sort, direct, owner)
	if err != nil {
		return nil, err
	}

	if repos[0] == "*" {
		result = append(result, reposgh...)
		return result, nil
	}

	for _, rr := range repos {
		for _, r := range reposgh {
			if strings.HasPrefix(r.GetName(), rr) {
				result = append(result, r)
			}
		}
	}

	return result, nil
}

func (g *Github) IsBranchExists(ctx context.Context, org,
	repo, branch string) (bool, error) {

	branches, err := g.GetRepoBranches(ctx, org, repo)
	if err != nil {
		return false, err
	}

	for _, brch := range branches {
		if brch.GetName() == branch {
			return true, nil
		}
	}

	brchck, _ := g.GetRepoBranch(ctx, org, repo, branch)

	if brchck.GetName() == branch {
		return true, nil
	}

	return false, nil
}

func (g *Github) GetLastCommitBranch(ctx context.Context, owner,
	repo, branch string) (*github.RepositoryCommit, error) {

	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	}

	commits, _, err := g.client.Repositories.ListCommits(ctx, owner, repo, opt)
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, nil
	}

	return commits[0], nil
}

func (g *Github) GetEmailFromUser(ctx context.Context, user string) (string, error) {

	u, _, err := g.client.Users.Get(ctx, user)
	if err != nil {
		return "", err
	}

	return u.GetEmail(), nil
}

func (g *Github) GetReposNotForkedAndArchived(ctx context.Context,
	owner string) ([]*github.Repository, error) {

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}

	var allRepos []*github.Repository

	for {
		repos, resp, err := g.client.Repositories.List(ctx, owner, opt)
		if err != nil {
			return allRepos, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	var result []*github.Repository

	for _, repo := range allRepos {
		if !repo.GetFork() && !repo.GetArchived() {
			result = append(result, repo)
		}
	}

	return result, nil
}

func (g *Github) UpdateFileFromRepository(ctx context.Context, owner,
	repo, branch, path, content string) error {

	// get last commit
	commit, err := g.GetLastCommitBranch(ctx, owner, repo, branch)
	if err != nil {
		return err
	}

	// get file
	file, _, _, err := g.client.Repositories.GetContents(ctx, owner, repo,
		path, &github.RepositoryContentGetOptions{
			Ref: commit.GetSHA(),
		})
	if err != nil {
		return err
	}

	// update file
	_, _, err = g.client.Repositories.UpdateFile(ctx, owner, repo,
		path, &github.RepositoryContentFileOptions{
			Message: github.String("Update file"),
			Content: []byte(content),
			SHA:     file.SHA,
			Branch:  github.String(branch),
		})
	if err != nil {
		return err
	}

	return nil
}

func (g *Github) SearchCommits(ctx context.Context, query string) ([]*github.CommitResult, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
		Sort: "author-date",
	}

	var allCommits []*github.CommitResult

	for {
		commits, resp, err := g.client.Search.Commits(ctx, query, opt)
		if err != nil {
			return allCommits, err
		}
		allCommits = append(allCommits, commits.Commits...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allCommits, nil
}
