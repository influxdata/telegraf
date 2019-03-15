package github

import "testing"

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
