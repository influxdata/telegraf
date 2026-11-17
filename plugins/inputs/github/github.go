//go:generate ../../../tools/readme_config_includer/generator
package github

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v93/github"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/selfstat"
)

//go:embed sample.conf
var sampleConfig string

type GitHub struct {
	Repositories      []string        `toml:"repositories"`
	AccessToken       string          `toml:"access_token"`
	AdditionalFields  []string        `toml:"additional_fields"`
	EnterpriseBaseURL string          `toml:"enterprise_base_url"`
	HTTPTimeout       config.Duration `toml:"http_timeout"`

	githubClient    *github.Client
	obfuscatedToken string

	rateLimit       selfstat.Stat
	rateLimitErrors selfstat.Stat
	rateRemaining   selfstat.Stat
}

func (*GitHub) SampleConfig() string {
	return sampleConfig
}

func (g *GitHub) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	if g.githubClient == nil {
		githubClient, err := g.createGitHubClient()
		if err != nil {
			return err
		}

		g.githubClient = githubClient

		tokenTags := map[string]string{
			"access_token": g.obfuscatedToken,
		}

		g.rateLimitErrors = selfstat.Register("github", "rate_limit_blocks", tokenTags)
		g.rateLimit = selfstat.Register("github", "rate_limit_limit", tokenTags)
		g.rateRemaining = selfstat.Register("github", "rate_limit_remaining", tokenTags)
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
			g.handleRateLimit(response, err)
			if err != nil {
				acc.AddError(err)
				return
			}

			now := time.Now()
			tags := getTags(repositoryInfo)
			fields := getFields(repositoryInfo)

			for _, field := range g.AdditionalFields {
				switch field {
				case "pull-requests":
					// Pull request properties
					addFields, err := g.getPullRequestFields(ctx, owner, repository)
					if err != nil {
						acc.AddError(err)
						continue
					}

					maps.Copy(fields, addFields)
				default:
					acc.AddError(fmt.Errorf("unknown additional field %q", field))
					continue
				}
			}

			acc.AddFields("github_repository", fields, tags, now)
		}(repository, acc)
	}

	wg.Wait()
	return nil
}

func (g *GitHub) createGitHubClient() (*github.Client, error) {
	opts := []github.ClientOptionsFunc{
		github.WithTimeout(time.Duration(g.HTTPTimeout)),
		github.WithEnvProxy(),
	}

	g.obfuscatedToken = "Unauthenticated"
	if g.AccessToken != "" {
		g.obfuscatedToken = g.AccessToken[0:4] + "..." + g.AccessToken[len(g.AccessToken)-3:]
		opts = append(opts, github.WithAuthToken(g.AccessToken))
	}

	if g.EnterpriseBaseURL != "" {
		opts = append(opts, github.WithEnterpriseURLs(g.EnterpriseBaseURL, g.EnterpriseBaseURL))
	}

	return github.NewClient(opts...)
}

func (g *GitHub) handleRateLimit(response *github.Response, err error) {
	if err == nil {
		g.rateLimit.Set(int64(response.Rate.Limit))
		g.rateRemaining.Set(int64(response.Rate.Remaining))
	} else if _, ok := errors.AsType[*github.RateLimitError](err); ok {
		g.rateLimitErrors.Incr(1)
	}
}

func splitRepositoryName(repositoryName string) (owner, repository string, err error) {
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

func getFields(repositoryInfo *github.Repository) map[string]any {
	return map[string]any{
		"stars":       repositoryInfo.GetStargazersCount(),
		"subscribers": repositoryInfo.GetSubscribersCount(),
		"watchers":    repositoryInfo.GetWatchersCount(),
		"networks":    repositoryInfo.GetNetworkCount(),
		"forks":       repositoryInfo.GetForksCount(),
		"open_issues": repositoryInfo.GetOpenIssuesCount(),
		"size":        repositoryInfo.GetSize(),
	}
}

func (g *GitHub) getPullRequestFields(ctx context.Context, owner, repo string) (map[string]any, error) {
	options := github.SearchOptions{
		TextMatch: false,
		PerPage:   100,
		Page:      1,
	}

	classes := []string{"open", "closed"}
	fields := make(map[string]any)
	for _, class := range classes {
		q := fmt.Sprintf("repo:%s/%s is:pr is:%s", owner, repo, class)
		searchResult, response, err := g.githubClient.Search.Issues(ctx, q, &options)
		g.handleRateLimit(response, err)
		if err != nil {
			return fields, err
		}

		f := class + "_pull_requests"
		fields[f] = searchResult.GetTotal()
	}

	return fields, nil
}

func init() {
	inputs.Add("github", func() telegraf.Input {
		return &GitHub{
			HTTPTimeout: config.Duration(time.Second * 5),
		}
	})
}
