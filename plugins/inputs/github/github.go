package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	githubLib "github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/selfstat"
)

// GitHub - plugin main structure
type GitHub struct {
	Repositories      []string        `toml:"repositories"`
	AccessToken       string          `toml:"access_token"`
	AdditionalFields  []string        `toml:"additional_fields"`
	EnterpriseBaseURL string          `toml:"enterprise_base_url"`
	HTTPTimeout       config.Duration `toml:"http_timeout"`
	githubClient      *githubLib.Client

	obfuscatedToken string

	RateLimit       selfstat.Stat
	RateLimitErrors selfstat.Stat
	RateRemaining   selfstat.Stat
}

// Create GitHub Client
func (g *GitHub) createGitHubClient(ctx context.Context) (*githubLib.Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(g.HTTPTimeout),
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

func (g *GitHub) newGithubClient(httpClient *http.Client) (*githubLib.Client, error) {
	if g.EnterpriseBaseURL != "" {
		return githubLib.NewEnterpriseClient(g.EnterpriseBaseURL, "", httpClient)
	}
	return githubLib.NewClient(httpClient), nil
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

					for k, v := range addFields {
						fields[k] = v
					}
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

func (g *GitHub) handleRateLimit(response *githubLib.Response, err error) {
	if err == nil {
		g.RateLimit.Set(int64(response.Rate.Limit))
		g.RateRemaining.Set(int64(response.Rate.Remaining))
	} else if _, ok := err.(*githubLib.RateLimitError); ok {
		g.RateLimitErrors.Incr(1)
	}
}

func splitRepositoryName(repositoryName string) (owner string, repository string, err error) {
	splits := strings.SplitN(repositoryName, "/", 2)

	if len(splits) != 2 {
		return "", "", fmt.Errorf("%v is not of format 'owner/repository'", repositoryName)
	}

	return splits[0], splits[1], nil
}

func getLicense(rI *githubLib.Repository) string {
	if licenseName := rI.GetLicense().GetName(); licenseName != "" {
		return licenseName
	}

	return "None"
}

func getTags(repositoryInfo *githubLib.Repository) map[string]string {
	return map[string]string{
		"owner":    repositoryInfo.GetOwner().GetLogin(),
		"name":     repositoryInfo.GetName(),
		"language": repositoryInfo.GetLanguage(),
		"license":  getLicense(repositoryInfo),
	}
}

func getFields(repositoryInfo *githubLib.Repository) map[string]interface{} {
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

func (g *GitHub) getPullRequestFields(ctx context.Context, owner, repo string) (map[string]interface{}, error) {
	options := githubLib.SearchOptions{
		TextMatch: false,
		ListOptions: githubLib.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	classes := []string{"open", "closed"}
	fields := make(map[string]interface{})
	for _, class := range classes {
		q := fmt.Sprintf("repo:%s/%s is:pr is:%s", owner, repo, class)
		searchResult, response, err := g.githubClient.Search.Issues(ctx, q, &options)
		g.handleRateLimit(response, err)
		if err != nil {
			return fields, err
		}

		f := fmt.Sprintf("%s_pull_requests", class)
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
