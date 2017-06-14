package main

// fork from https://github.com/tcnksm/ghr/blob/master/github.go
import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
)

var (
	RelaseNotFound = errors.New("release is not found")
)

type GitHub interface {
	CreateRelease(ctx context.Context, req *github.RepositoryRelease) (*github.RepositoryRelease, error)
	GetRelease(ctx context.Context, tag string) (*github.RepositoryRelease, error)
	DeleteRelease(ctx context.Context, releaseID int) error
	DeleteTag(ctx context.Context, tag string) error
}

type GitHubClient struct {
	Owner, Repo string
	*github.Client
}

func NewGitHubClient(owner, repo, token string, urlStr string) (*GitHubClient, error) {
	if len(owner) == 0 {
		return nil, errors.New("missing GitHub repository owner")
	}

	if len(owner) == 0 {
		return nil, errors.New("missing GitHub repository name")
	}

	if len(token) == 0 {
		return nil, errors.New("missing GitHub API token")
	}

	if len(urlStr) == 0 {
		return nil, errors.New("missgig GitHub API URL")
	}

	baseURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Github API URL")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)
	client.BaseURL = baseURL

	return &GitHubClient{
		Owner:  owner,
		Repo:   repo,
		Client: client,
	}, nil
}

func (c *GitHubClient) CreateRelease(ctx context.Context, req *github.RepositoryRelease) (*github.RepositoryRelease, error) {

	release, res, err := c.Repositories.CreateRelease(ctx, c.Owner, c.Repo, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a release")
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("create release: invalid status: %s", res.Status)
	}

	return release, nil
}

func (c *GitHubClient) GetRelease(ctx context.Context, tag string) (*github.RepositoryRelease, error) {
	// Check Release is already exist or not
	release, res, err := c.Repositories.GetReleaseByTag(ctx, c.Owner, c.Repo, tag)
	if err != nil {
		if res == nil {
			return nil, errors.Wrapf(err, "failed to get release tag: %s", tag)
		}

		// TODO(tcnksm): Handle invalid token
		if res.StatusCode != http.StatusNotFound {
			return nil, errors.Wrapf(err,
				"get release tag: invalid status: %s", res.Status)
		}

		return nil, RelaseNotFound
	}

	return release, nil
}

func (c *GitHubClient) DeleteRelease(ctx context.Context, releaseID int) error {
	res, err := c.Repositories.DeleteRelease(ctx, c.Owner, c.Repo, releaseID)
	if err != nil {
		return errors.Wrap(err, "failed to delete release")
	}

	if res.StatusCode != http.StatusNoContent {
		return errors.Errorf("delete release: invalid status: %s", res.Status)
	}

	return nil
}

func (c *GitHubClient) DeleteTag(ctx context.Context, tag string) error {
	ref := fmt.Sprintf("tags/%s", tag)
	res, err := c.Git.DeleteRef(ctx, c.Owner, c.Repo, ref)
	if err != nil {
		return errors.Wrapf(err, "failed to delete tag: %s", ref)
	}

	if res.StatusCode != http.StatusNoContent {
		return errors.Errorf("delete tag: invalid status: %s", res.Status)
	}

	return nil
}
