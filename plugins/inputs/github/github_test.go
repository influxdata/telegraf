package github

import (
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
