package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/selfstat"
	"golang.org/x/oauth2"
)

// GitHub - plugin main structure
type GitHub struct {
	Repositories      []string          `toml:"repositories"`
	AccessToken       string            `toml:"access_token"`
	EnterpriseBaseURL string            `toml:"enterprise_base_url"`
	HTTPTimeout       internal.Duration `toml:"http_timeout"`
	githubClient      *github.Client

	obfuscatedToken string

	RateLimit       selfstat.Stat
	RateLimitErrors selfstat.Stat
	RateRemaining   selfstat.Stat
}

const sampleConfig = `
  ## List of repositories to monitor.
  repositories = [
	  "influxdata/telegraf",
	  "influxdata/influxdb"
  ]

  ## Github API access token.  Unauthenticated requests are limited to 60 per hour.
  # access_token = ""

  ## Github API enterprise url. Github Enterprise accounts must specify their base url.
  # enterprise_base_url = ""

  ## Timeout for HTTP requests.
  # http_timeout = "5s"
`

// SampleConfig returns sample configuration for this plugin.
func (g *GitHub) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (g *GitHub) Description() string {
	return "Gather repository information from GitHub hosted repositories."
}

// Create GitHub Client
func (g *GitHub) createGitHubClient(ctx context.Context) (*github.Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: g.HTTPTimeout.Duration,
	}

	g.obfuscatedToken = "Unauthenticated"

	if g.AccessToken != "" {
		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: g.AccessToken},
		)
		oauthClient := oauth2.NewClient(ctx, tokenSource)
		_ = context.WithValue(ctx, oauth2.HTTPClient, oauthClient)

		g.obfuscatedToken = g.AccessToken[0:4] + "..." + g.AccessToken[len(g.AccessToken)-3:]

		return g.newGithubClient(oauthClient)
	}

	return g.newGithubClient(httpClient)
}

func (g *GitHub) newGithubClient(httpClient *http.Client) (*github.Client, error) {
	if g.EnterpriseBaseURL != "" {
		return github.NewEnterpriseClient(g.EnterpriseBaseURL, "", httpClient)
	}
	return github.NewClient(httpClient), nil
}

// Gather GitHub Metrics
func (g *GitHub) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	if g.githubClient == nil {
		githubClient, err := g.createGitHubClient(ctx)

		if err != nil {
			return err
		}

		g.githubClient = githubClient

		tokenTags := map[string]string{
			"access_token": g.obfuscatedToken,
		}

		g.RateLimitErrors = selfstat.Register("github", "rate_limit_blocks", tokenTags)
		g.RateLimit = selfstat.Register("github", "rate_limit_limit", tokenTags)
		g.RateRemaining = selfstat.Register("github", "rate_limit_remaining", tokenTags)
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
				g.RateLimitErrors.Incr(1)
			}

			if err != nil {
				acc.AddError(err)
				return
			}

			g.RateLimit.Set(int64(response.Rate.Limit))
			g.RateRemaining.Set(int64(response.Rate.Remaining))

			now := time.Now()
			tags := getTags(repositoryInfo)
			fields := getFields(repositoryInfo)

			acc.AddFields("github_repository", fields, tags, now)
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

func getLicense(rI *github.Repository) string {
	if licenseName := rI.GetLicense().GetName(); licenseName != "" {
		return licenseName
	}

	return "None"
}

func getTags(repositoryInfo *github.Repository) map[string]string {
	return map[string]string{
		"owner":    repositoryInfo.GetOwner().GetLogin(),
		"name":     repositoryInfo.GetName(),
		"language": repositoryInfo.GetLanguage(),
		"license":  getLicense(repositoryInfo),
	}
}

func getFields(repositoryInfo *github.Repository) map[string]interface{} {
	return map[string]interface{}{
		"stars":       repositoryInfo.GetStargazersCount(),
		"subscribers": repositoryInfo.GetSubscribersCount(),
		"watchers":    repositoryInfo.GetWatchersCount(),
		"networks":    repositoryInfo.GetNetworkCount(),
		"forks":       repositoryInfo.GetForksCount(),
		"open_issues": repositoryInfo.GetOpenIssuesCount(),
		"size":        repositoryInfo.GetSize(),
	}
}

func init() {
	inputs.Add("github", func() telegraf.Input {
		return &GitHub{
			HTTPTimeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
