package github

import (
	"reflect"
	"testing"

	gh "github.com/google/go-github/github"
	"github.com/stretchr/testify/require"
)

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

			require.NotNil(t, err)
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

	repository := gh.Repository{
		StargazersCount: &stars,
		ForksCount:      &forks,
		OpenIssuesCount: &openIssues,
		Size:            &size,
	}

	getFieldsReturn := getFields(&repository)

	correctFieldReturn := make(map[string]interface{})

	correctFieldReturn["stars"] = 1
	correctFieldReturn["forks"] = 2
	correctFieldReturn["open_issues"] = 3
	correctFieldReturn["size"] = 4

	require.Equal(t, true, reflect.DeepEqual(getFieldsReturn, correctFieldReturn))
}
