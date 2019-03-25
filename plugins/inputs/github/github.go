package github

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	gh "github.com/google/go-github/github"
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
	githubClient *gh.Client

	RateLimit     selfstat.Stat
	RateRemaining selfstat.Stat
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

// NewGitHub returns a new instance of the GitHub input plugin
func NewGitHub() *GitHub {
	return &GitHub{
		HTTPTimeout: internal.Duration{Duration: time.Second * 5},
	}
}

// SampleConfig returns sample configuration for this plugin.
func (github *GitHub) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (github *GitHub) Description() string {
	return "Read repository information from GitHub, including forks, stars, and more."
}

// Create HTTP Client
func (github *GitHub) createGitHubClient() (*gh.Client, error) {
	var githubClient *gh.Client

	if github.AccessToken != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: github.AccessToken},
		)
		tc := oauth2.NewClient(ctx, ts)

		githubClient = gh.NewClient(tc)
	} else {
		githubClient = gh.NewClient(nil)
	}

	return githubClient, nil
}

// Gather GitHub Metrics
func (github *GitHub) Gather(acc telegraf.Accumulator) error {
	if github.githubClient == nil {
		githubClient, err := github.createGitHubClient()

		if err != nil {
			return err
		}

		github.githubClient = githubClient
	}

	var wg sync.WaitGroup
	wg.Add(len(github.Repositories))

	for _, repository := range github.Repositories {
		go func(repositoryName string, acc telegraf.Accumulator) {
			defer wg.Done()

			ctx := context.Background()

			owner, repository, err := splitRepositoryName(repositoryName)
			if err != nil {
				log.Printf("E! [github]: %v", err)
				return
			}

			repositoryInfo, response, err := github.githubClient.Repositories.Get(ctx, owner, repository)

			if _, ok := err.(*gh.RateLimitError); ok {
				log.Printf("E! [github]: %v of %v requests remaining", response.Rate.Remaining, response.Rate.Limit)
				return
			}

			now := time.Now()
			tags := getTags(repositoryInfo)
			fields := getFields(repositoryInfo)

			acc.AddFields("github_repository", fields, tags, now)

			github.RateLimit = selfstat.Register("github", "rate_limit", tags)
			github.RateLimit.Set(int64(response.Rate.Limit))

			github.RateRemaining = selfstat.Register("github", "rate_remaining", tags)
			github.RateRemaining.Set(int64(response.Rate.Remaining))
		}(repository, acc)
	}

	wg.Wait()
	return nil
}

func splitRepositoryName(repositoryName string) (string, string, error) {
	splits := strings.Split(repositoryName, "/")

	if len(splits) != 2 {
		return "", "", fmt.Errorf("%v is not of format 'owner/repository'", repositoryName)
	}

	return splits[0], splits[1], nil
}

func getLicense(repositoryInfo *gh.Repository) string {
	if repositoryInfo.GetLicense() != nil {
		return *repositoryInfo.License.Name
	}

	return "None"
}

func getTags(repositoryInfo *gh.Repository) map[string]string {
	return map[string]string{
		"full_name": *repositoryInfo.FullName,
		"owner":     *repositoryInfo.Owner.Login,
		"name":      *repositoryInfo.Name,
		"language":  *repositoryInfo.Language,
		"license":   getLicense(repositoryInfo),
	}
}

func getFields(repositoryInfo *gh.Repository) map[string]interface{} {
	return map[string]interface{}{
		"stars":       *repositoryInfo.StargazersCount,
		"forks":       *repositoryInfo.ForksCount,
		"open_issues": *repositoryInfo.OpenIssuesCount,
		"size":        *repositoryInfo.Size,
	}
}

func init() {
	inputs.Add("github", func() telegraf.Input {
		return &GitHub{}
	})
}
