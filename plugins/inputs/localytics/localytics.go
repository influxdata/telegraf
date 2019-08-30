package localytics

import (
	"context"
	"net/http"
	"time"

	localytics "github.com/Onefootball/go-localytics"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/selfstat"
)

// Localytics - plugin main structure
type Localytics struct {
	AccessKey   string            `toml:"access_key"`
	SecretKey   string            `toml:"secret_key"`
	HTTPTimeout internal.Duration `toml:"http_timeout"`

	RateLimit selfstat.Stat

	client *localytics.Client
}

const sampleConfig = `
  ## Localytics API access key.
  # access_token = ""

  ## Localytics API secret key.
  # secret_key = ""

  ## Timeout for HTTP requests.
  # http_timeout = "5s"
`

// SampleConfig returns sample configuration for this plugin.
func (l *Localytics) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (l *Localytics) Description() string {
	return "Gather app usage statistics from Localytics."
}

// Create Localytics Client
func (l *Localytics) createLocalyticsClient(ctx context.Context) *localytics.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: l.HTTPTimeout.Duration,
	}

	return l.newLocalyticsClient(httpClient)
}

func (l *Localytics) newLocalyticsClient(httpClient *http.Client) *localytics.Client {
	return localytics.New(httpClient, localytics.Auth(l.AccessKey, l.SecretKey))
}

// Gather Localytics Metrics
func (l *Localytics) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	// reuse the client, if possible
	if l.client == nil {
		l.client = l.createLocalyticsClient(ctx)
	}

	if err := l.gatherApps(acc); err != nil {
		return err
	}

	return nil
}

func (l *Localytics) gatherApps(acc telegraf.Accumulator) error {
	apps, err := l.client.Apps()

	if err == localytics.ErrRateLimitExeceed {
		l.RateLimit.Incr(1)
	}

	if err != nil {
		acc.AddError(err)

		return err
	}

	for _, app := range apps {
		addApp(app, acc)
	}

	return nil
}

func getTags(app *localytics.App) map[string]string {
	return map[string]string{
		"name": app.Name,
		"id":   app.AppID,
	}
}

func getFields(app *localytics.App) map[string]interface{} {
	return map[string]interface{}{
		"sessions": app.Stats.Sessions,
		"closes":   app.Stats.Closes,
		"users":    app.Stats.Users,
		"events":   app.Stats.Events,
	}
}

func addApp(app *localytics.App, acc telegraf.Accumulator) {
	fields := getFields(app)
	tags := getTags(app)

	now := time.Now()
	acc.AddFields("localytics", fields, tags, now)
}

func init() {
	inputs.Add("localytics", func() telegraf.Input {
		return &Localytics{
			HTTPTimeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
