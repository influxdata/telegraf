package github

import (
	"fmt"
	"reflect"
	"testing"

	gh "github.com/google/go-github/github"
)

func TestSplitRepositoryNameWithWorkingExample(t *testing.T) {
	owner, repository, _ := splitRepositoryName("influxdata/influxdb")

	if owner != "influxdata" {
		t.Errorf("Owner from 'influxdata/influxdb' should return 'influxdata'")
	}

	if repository != "influxdb" {
		t.Errorf("Repository from 'influxdata/influxdb' should return 'influxdb'")
	}
}

func TestSplitRepositoryNameWithNoSlash(t *testing.T) {
	_, _, error := splitRepositoryName("influxdata-influxdb")

	if error == nil {
		t.Errorf("Repository name without a slash should return an error")
	}
}

func TestGetLicenseWhenExists(t *testing.T) {
	licenseName := "MIT"
	license := gh.License{Name: &licenseName}
	repository := gh.Repository{License: &license}

	getLicenseReturn := getLicense(&repository)

	if getLicenseReturn != "MIT" {
		t.Errorf("GetLicense doesn't return the correct variable")
	}
}

func TestGetLicenseWhenMissing(t *testing.T) {
	repository := gh.Repository{}

	getLicenseReturn := getLicense(&repository)

	if getLicenseReturn != "None" {
		t.Errorf("GetLicense doesn't return 'None' when no license exists")
	}
}

func TestGetTags(t *testing.T) {
	licenseName := "MIT"
	license := gh.License{Name: &licenseName}

	ownerName := "influxdata"
	owner := gh.User{Login: &ownerName}

	fullName := "influxdata/influxdb"
	repositoryName := "influxdb"

	language := "Go"

	repository := gh.Repository{FullName: &fullName, Name: &repositoryName, License: &license, Owner: &owner, Language: &language}

	getTagsReturn := getTags(&repository)

	correctTagsReturn := map[string]string{
		"full_name": fullName,
		"owner":     ownerName,
		"name":      repositoryName,
		"language":  language,
		"license":   licenseName,
	}

	if !reflect.DeepEqual(getTagsReturn, correctTagsReturn) {
		t.Errorf("GetTags doesn't return the correct values")
	}
}

func TestGetFields(t *testing.T) {
	stars := 1
	forks := 2
	open_issues := 3
	size := 4

	repository := gh.Repository{StargazersCount: &stars, ForksCount: &forks, OpenIssuesCount: &open_issues, Size: &size}

	getFieldsReturn := getFields(&repository)

	correctFieldReturn := make(map[string]interface{})

	correctFieldReturn["stars"] = 1
	correctFieldReturn["forks"] = 2
	correctFieldReturn["open_issues"] = 3
	correctFieldReturn["size"] = 4

	fmt.Printf("%v\n\n", getFieldsReturn)
	fmt.Printf("%v\n\n", correctFieldReturn)

	if !reflect.DeepEqual(getFieldsReturn, correctFieldReturn) {
		t.Errorf("GetFields doesn't return the correct values")
	}
}
