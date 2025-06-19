package gh

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)


type Client struct {
	*github.Client
}


func NewClient(ctx context.Context, token string) *Client {
	var tc *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc = oauth2.NewClient(ctx, ts)
	}
	return &Client{
		Client: github.NewClient(tc),
	}
}


func (c *Client) GetLatestRelease(ctx context.Context, owner, repo string, includePrereleases bool) (*github.RepositoryRelease, error) {
	if includePrereleases {
		releases, _, err := c.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: 1})
		if err != nil {
			return nil, fmt.Errorf("could not fetch releases: %w", err)
		}
		if len(releases) == 0 {
			return nil, fmt.Errorf("no releases found for %s/%s", owner, repo)
		}
		return releases[0], nil
	}

	release, _, err := c.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		
		if e, ok := err.(*github.ErrorResponse); ok && e.Response.StatusCode == 404 {
			return nil, fmt.Errorf("no stable releases found (latest may be a pre-release)")
		}
		return nil, fmt.Errorf("could not fetch latest release: %w", err)
	}
	return release, nil
}


func (c *Client) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, error) {
	release, _, err := c.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("could not get release by tag '%s': %w", tag, err)
	}
	return release, nil
}


func (c *Client) ListReleases(ctx context.Context, owner, repo string, limit int) ([]*github.RepositoryRelease, error) {
	releases, _, err := c.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: limit})
	if err != nil {
		return nil, fmt.Errorf("could not list releases: %w", err)
	}
	return releases, nil
}


func (c *Client) SearchRepos(ctx context.Context, query string, limit int) (*github.RepositoriesSearchResult, error) {
	opts := &github.SearchOptions{
		Sort:        "stars",
		Order:       "desc",
		ListOptions: github.ListOptions{PerPage: limit},
	}
	result, _, err := c.Search.Repositories(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("could not search repositories: %w", err)
	}
	return result, nil
}


func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	repository, _, err := c.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("could not get repo info: %w", err)
	}
	return repository, nil
}
