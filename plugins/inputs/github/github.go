package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/selfstat"
	"golang.org/x/oauth2"
)

// GitHub - plugin main structure
type GitHub struct {
	Repositories []string          `toml:"repositories"`
	AccessToken  string            `toml:"access_token"`
	HTTPTimeout  internal.Duration `toml:"http_timeout"`
	githubClient *github.Client

	RateLimit       selfstat.Stat
	RateLimitErrors selfstat.Stat
	RateRemaining   selfstat.Stat
}

const sampleConfig = `
  ## Specify a list of repositories.
  # eg: repositories = ["influxdata/influxdb"]
  repositories = []

  ## API Key for GitHub API requests.
  api_key = ""

  ## Timeout for GitHub API requests.
  http_timeout = "5s"
`

// SampleConfig returns sample configuration for this plugin.
func (g *GitHub) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (g *GitHub) Description() string {
	return "Read repository information from GitHub, including forks, stars, and more."
}

// Create GitHub Client
func (g *GitHub) createGitHubClient(ctx context.Context) (*github.Client, error) {
	var githubClient *github.Client

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: g.HTTPTimeout.Duration,
	}

	if g.AccessToken != "" {
		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: g.AccessToken},
		)
		oauthClient := oauth2.NewClient(ctx, tokenSource)
		ctx = context.WithValue(ctx, oauth2.HTTPClient, oauthClient)

		return github.NewClient(oauthClient), nil
	}

	return github.NewClient(httpClient), nil
}

// Gather GitHub Metrics
func (g *GitHub) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	if g.githubClient == nil {
		githubClient, err := g.createGitHubClient()

		if err != nil {
			return err
		}

		g.githubClient = githubClient
	}

	var wg sync.WaitGroup
	wg.Add(len(g.Repositories))

	for _, repository := range g.Repositories {
		go func(repositoryName string, acc telegraf.Accumulator) {
			defer wg.Done()

			owner, repository, err := splitRepositoryName(repositoryName)
			if err != nil {
				acc.AddError(err)
				return
			}

			repositoryInfo, response, err := g.githubClient.Repositories.Get(ctx, owner, repository)

			if _, ok := err.(*github.RateLimitError); ok {
				g.RateLimitErrors = selfstat.Register("github", "rate_limit_blocks", map[string]string{})
				g.RateLimitErrors.Incr(1)
			}

			if err != nil {
				acc.AddError(err)
				return
			}

			now := time.Now()
			tags := getTags(repositoryInfo)
			fields := getFields(repositoryInfo)

			acc.AddFields("github_repository", fields, tags, now)

			g.RateLimit = selfstat.Register("github", "rate_limit_limit", tags)
			g.RateLimit.Set(int64(response.Rate.Limit))

			g.RateRemaining = selfstat.Register("github", "rate_limit_remaining", tags)
			g.RateRemaining.Set(int64(response.Rate.Remaining))
		}(repository, acc)
	}

	wg.Wait()
	return nil
}

func splitRepositoryName(repositoryName string) (string, string, error) {
	splits := strings.SplitN(repositoryName, "/", 2)

	if len(splits) != 2 {
		return "", "", fmt.Errorf("%v is not of format 'owner/repository'", repositoryName)
	}

	return splits[0], splits[1], nil
}

func getLicense(repositoryInfo *github.Repository) string {
	if repositoryInfo.GetLicense() != nil {
		return *repositoryInfo.License.Name
	}

	return "None"
}

func getTags(repositoryInfo *github.Repository) map[string]string {
	return map[string]string{
		"full_name": *repositoryInfo.FullName,
		"owner":     *repositoryInfo.Owner.Login,
		"name":      *repositoryInfo.Name,
		"language":  *repositoryInfo.Language,
		"license":   getLicense(repositoryInfo),
	}
}

func getFields(repositoryInfo *github.Repository) map[string]interface{} {
	return map[string]interface{}{
		"stars":       *repositoryInfo.StargazersCount,
		"forks":       *repositoryInfo.ForksCount,
		"open_issues": *repositoryInfo.OpenIssuesCount,
		"size":        *repositoryInfo.Size,
	}
}

func init() {
	inputs.Add("github", func() telegraf.Input {
		return &GitHub{
			HTTPTimeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
