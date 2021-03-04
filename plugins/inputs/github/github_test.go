package github

import (
	"net/http"
	"reflect"
	"testing"

	gh "github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/require"
)

func TestNewGithubClient(t *testing.T) {
	httpClient := &http.Client{}
	g := &GitHub{}
	client, err := g.newGithubClient(httpClient)
	require.NoError(t, err)
	require.Contains(t, client.BaseURL.String(), "api.github.com")
	g.EnterpriseBaseURL = "api.example.com/"
	enterpriseClient, err := g.newGithubClient(httpClient)
	require.NoError(t, err)
	require.Contains(t, enterpriseClient.BaseURL.String(), "api.example.com")
}

func TestSplitRepositoryNameWithWorkingExample(t *testing.T) {
	var validRepositoryNames = []struct {
		fullName   string
		owner      string
		repository string
	}{
		{"influxdata/telegraf", "influxdata", "telegraf"},
		{"influxdata/influxdb", "influxdata", "influxdb"},
		{"rawkode/saltstack-dotfiles", "rawkode", "saltstack-dotfiles"},
	}

	for _, tt := range validRepositoryNames {
		t.Run(tt.fullName, func(t *testing.T) {
			owner, repository, _ := splitRepositoryName(tt.fullName)

			require.Equal(t, tt.owner, owner)
			require.Equal(t, tt.repository, repository)
		})
	}
}

func TestSplitRepositoryNameWithNoSlash(t *testing.T) {
	var invalidRepositoryNames = []string{
		"influxdata-influxdb",
	}

	for _, tt := range invalidRepositoryNames {
		t.Run(tt, func(t *testing.T) {
			_, _, err := splitRepositoryName(tt)

			require.Error(t, err)
		})
	}
}

func TestGetLicenseWhenExists(t *testing.T) {
	licenseName := "MIT"
	license := gh.License{Name: &licenseName}
	repository := gh.Repository{License: &license}

	getLicenseReturn := getLicense(&repository)

	require.Equal(t, "MIT", getLicenseReturn)
}

func TestGetLicenseWhenMissing(t *testing.T) {
	repository := gh.Repository{}

	getLicenseReturn := getLicense(&repository)

	require.Equal(t, "None", getLicenseReturn)
}

func TestGetTags(t *testing.T) {
	licenseName := "MIT"
	license := gh.License{Name: &licenseName}

	ownerName := "influxdata"
	owner := gh.User{Login: &ownerName}

	fullName := "influxdata/influxdb"
	repositoryName := "influxdb"

	language := "Go"

	repository := gh.Repository{
		FullName: &fullName,
		Name:     &repositoryName,
		License:  &license,
		Owner:    &owner,
		Language: &language,
	}

	getTagsReturn := getTags(&repository)

	correctTagsReturn := map[string]string{
		"owner":    ownerName,
		"name":     repositoryName,
		"language": language,
		"license":  licenseName,
	}

	require.Equal(t, true, reflect.DeepEqual(getTagsReturn, correctTagsReturn))
}

func TestGetFields(t *testing.T) {
	stars := 1
	forks := 2
	openIssues := 3
	size := 4
	subscribers := 5
	watchers := 6

	repository := gh.Repository{
		StargazersCount:  &stars,
		ForksCount:       &forks,
		OpenIssuesCount:  &openIssues,
		Size:             &size,
		NetworkCount:     &forks,
		SubscribersCount: &subscribers,
		WatchersCount:    &watchers,
	}

	getFieldsReturn := getFields(&repository)

	correctFieldReturn := make(map[string]interface{})

	correctFieldReturn["stars"] = 1
	correctFieldReturn["forks"] = 2
	correctFieldReturn["networks"] = 2
	correctFieldReturn["open_issues"] = 3
	correctFieldReturn["size"] = 4
	correctFieldReturn["subscribers"] = 5
	correctFieldReturn["watchers"] = 6

	require.Equal(t, true, reflect.DeepEqual(getFieldsReturn, correctFieldReturn))
}
