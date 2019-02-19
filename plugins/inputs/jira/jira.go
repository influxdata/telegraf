package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type (
	Jira struct {
		Servers []string `toml:"hosts"`
		Auth    Auth     `toml:"authentication"`
		Fields  []string `toml:"fields"`
		Tags    []string `toml:"tag_fields"`
		Jql     []Jql    `toml:"jql"`
		client  *http.Client
	}
	Jql struct {
		Key   string `toml:"name"`
		Value string `toml:"jql"`
	}
	Auth struct {
		Username string `toml:"username"`
		Password string `toml:"password"`
		Email    string `toml:"email"`
		Token    string `toml:"token"`
	}
)

type (
	field struct {
		Id    int64  `json:"id"`
		Field string `json:"_"`
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	issues struct {
		Id     string           `json:"id"`
		Uri    string           `json:"self"`
		Key    string           `json:"key"`
		Fields map[string]field `json:"fields"`
	}

	Stats struct {
		StartAt    int64    `json:"startAt"`
		MaxResults int64    `json:"maxResults"`
		Total      int64    `json:"total"`
		Issues     []issues `json:"issues"`
	}
)

func (j *Jira) Description() string {
	return "Jira ticket statistics"
}

func (j *Jira) SampleConfig() string {
	return `
# Works with multiple Atlassian Jira instances
[[inputs.jira]]
  # Multiple Hosts from which to read ticket stats
  hosts = ["http://jira:8080/"]

  # Give here all fields to be selected. Each field will be counted grouped by the hosts, tags and JQLs below
  fields = ["priority", "custom_field_1234"]

  # Create tags based on these fields values
  tag_fields = ["customfield_666"]


# Define here the preffered authentication values
# You should prefere the API-Token and leave username and password clear
[[inputs.jira.authentication]]
  # Username amd Password for BasicAuth - this may be deprecated in your Jira-Installation
  username = MyUser
  password = MyPass

  # If you're using the new API-Token, fill in these values
  email = myjira@example.com
  token = my-generated-api-token


# ${DATE} will be replaced with the current date on every request
# Define as much JQLs as you need and give them each a name for having statistics on the count of issues
[[inputs.jira.jql]]
  name = "new"
  jql = "Team in (DevTeam, TestingTeam) AND issuetype = Bug AND status changed to \"Ready for develope\" on ${DATE}"

[[inputs.jira.jql]]
  name = "closed"
  jql = "Team in (DevTeam, TestingTeam) AND issuetype = Bug AND status changed to (Closed, Resolved) on ${DATE} AND status was QA on ${DATE}"

[[inputs.jira.jql]]
  name = "total"
  jql = "Team in (DevTeam, TestingTeam) AND issuetype = Bug AND status was in (\"Ready for develope\", Development, QA) on ${DATE}"
`
}

func (j *Jira) Gather(accumulator telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, h := range j.Servers {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			if err := j.fetchAndProcess(accumulator, host); err != nil {
				accumulator.AddError(fmt.Errorf("[host=%s]: %s", host, err))
			}
		}(h)
	}

	wg.Wait()
	return nil
}

func (j *Jira) fetchAndProcess(accumulator telegraf.Accumulator, host string) error {
	if j.client == nil {
		j.client = &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: time.Duration(3 * time.Second),
			},
			Timeout: time.Duration(4 * time.Second),
		}
	}

	time := time.Now()
	cache := make(map[string]map[string]int64)
	uri := host + "/rest/api/latest/search"

	for _, jql := range j.Jql {
		name := jql.Key
		filter := strings.Replace(jql.Value, "${DATE}", strconv.Itoa(time.Year())+"-"+strconv.Itoa(int(time.Month()))+"-"+strconv.Itoa(time.Day()), -1)

		var post = []byte(`{"jql": "` + strings.Replace(filter, `"`, `\"`, -1) + `",
    "startAt": 0,
    "maxResults": 9999,
    "fields": ["` + strings.Join(append(j.Fields, j.Tags...)[:], `","`) + `"]
  }`)
		request, error := http.NewRequest("POST", uri, bytes.NewBuffer(post))
		request.Header.Add("Content-Type", "application/json")
		if len(j.Auth.Username) > 0 {
			request.SetBasicAuth(j.Auth.Username, j.Auth.Password)
		} else if len(j.Auth.Email) > 0 {
			request.SetBasicAuth(j.Auth.Email, j.Auth.Token)
		}
		response, error := j.client.Do(request)
		if error != nil {
			return error
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			return fmt.Errorf("Failed to get stats from jira: HTTP responded %d", response.StatusCode)
		}

		stats := Stats{}
		decoder := json.NewDecoder(response.Body)
		decoder.Decode(&stats)

		for _, issue := range stats.Issues {
			tags := j.getTagValues(issue.Fields)
			if _, ok := cache[tags]; !ok {
				cache[tags] = make(map[string]int64)
			}

			var fieldName string
			for _, fn := range j.Fields {
				if field, ok := issue.Fields[fn]; ok {
					fieldName = name + "_" + fn
					if len(field.Name) > 0 {
						fieldName += "_" + field.Name
					} else if len(field.Value) > 0 {
						fieldName += "_" + field.Value
					}
					cache[tags][fieldName] += 1
				}
			}
		}
	}

	for tag, values := range cache {
		tags := map[string]string{
			"server": host,
		}

		for _, part := range strings.Split(tag, ",") {
			var p = strings.Split(part, "=")
			if len(p) == 2 {
				tags[p[0]] = p[1]
			}
		}

		fields := map[string]interface{}{}
		for n, v := range values {
			fields[n] = v
		}
		accumulator.AddFields("jira", fields, tags, time.UTC())
	}
	return nil
}

func (j *Jira) getTagValues(fields map[string]field) string {
	re := regexp.MustCompile("[\\s,.]+")
	var tags []string
	for _, tag := range j.Tags {
		if field, ok := fields[tag]; ok {
			var value string
			if len(field.Value) > 0 {
				value = field.Value
			} else if len(field.Name) > 0 {
				value = field.Name
			}
			tags = append(tags, tag+"="+re.ReplaceAllString(value, ""))
		}
	}
	return strings.Join(tags[:], ",")
}

func init() {
	inputs.Add("jira", func() telegraf.Input {
		return &Jira{
			client: &http.Client{
				Transport: &http.Transport{
					ResponseHeaderTimeout: time.Duration(3 * time.Second),
				},
				Timeout: time.Duration(4 * time.Second),
			},
		}
	})
}
