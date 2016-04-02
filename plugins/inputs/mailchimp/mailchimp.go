package mailchimp

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type MailChimp struct {
	api *ChimpAPI

	ApiKey     string
	DaysOld    int
	CampaignId string
}

var sampleConfig = `
  ## MailChimp API key
  ## get from https://admin.mailchimp.com/account/api/
  api_key = "" # required
  ## Reports for campaigns sent more than days_old ago will not be collected.
  ## 0 means collect all.
  days_old = 0
  ## Campaign ID to get, if empty gets all campaigns, this option overrides days_old
  # campaign_id = ""
`

func (m *MailChimp) SampleConfig() string {
	return sampleConfig
}

func (m *MailChimp) Description() string {
	return "Gathers metrics from the /3.0/reports MailChimp API"
}

func (m *MailChimp) Gather(acc telegraf.Accumulator) error {
	if m.api == nil {
		m.api = NewChimpAPI(m.ApiKey)
	}
	m.api.Debug = false

	if m.CampaignId == "" {
		since := ""
		if m.DaysOld > 0 {
			now := time.Now()
			d, _ := time.ParseDuration(fmt.Sprintf("%dh", 24*m.DaysOld))
			since = now.Add(-d).Format(time.RFC3339)
		}

		reports, err := m.api.GetReports(ReportsParams{
			SinceSendTime: since,
		})
		if err != nil {
			return err
		}
		now := time.Now()

		for _, report := range reports.Reports {
			gatherReport(acc, report, now)
		}
	} else {
		report, err := m.api.GetReport(m.CampaignId)
		if err != nil {
			return err
		}
		now := time.Now()
		gatherReport(acc, report, now)
	}

	return nil
}

func gatherReport(acc telegraf.Accumulator, report Report, now time.Time) {
	tags := make(map[string]string)
	tags["id"] = report.ID
	tags["campaign_title"] = report.CampaignTitle
	fields := map[string]interface{}{
		"emails_sent":              report.EmailsSent,
		"abuse_reports":            report.AbuseReports,
		"unsubscribed":             report.Unsubscribed,
		"hard_bounces":             report.Bounces.HardBounces,
		"soft_bounces":             report.Bounces.SoftBounces,
		"syntax_errors":            report.Bounces.SyntaxErrors,
		"forwards_count":           report.Forwards.ForwardsCount,
		"forwards_opens":           report.Forwards.ForwardsOpens,
		"opens_total":              report.Opens.OpensTotal,
		"unique_opens":             report.Opens.UniqueOpens,
		"open_rate":                report.Opens.OpenRate,
		"clicks_total":             report.Clicks.ClicksTotal,
		"unique_clicks":            report.Clicks.UniqueClicks,
		"unique_subscriber_clicks": report.Clicks.UniqueSubscriberClicks,
		"click_rate":               report.Clicks.ClickRate,
		"facebook_recipient_likes": report.FacebookLikes.RecipientLikes,
		"facebook_unique_likes":    report.FacebookLikes.UniqueLikes,
		"facebook_likes":           report.FacebookLikes.FacebookLikes,
		"industry_type":            report.IndustryStats.Type,
		"industry_open_rate":       report.IndustryStats.OpenRate,
		"industry_click_rate":      report.IndustryStats.ClickRate,
		"industry_bounce_rate":     report.IndustryStats.BounceRate,
		"industry_unopen_rate":     report.IndustryStats.UnopenRate,
		"industry_unsub_rate":      report.IndustryStats.UnsubRate,
		"industry_abuse_rate":      report.IndustryStats.AbuseRate,
		"list_stats_sub_rate":      report.ListStats.SubRate,
		"list_stats_unsub_rate":    report.ListStats.UnsubRate,
		"list_stats_open_rate":     report.ListStats.OpenRate,
		"list_stats_click_rate":    report.ListStats.ClickRate,
	}
	acc.AddFields("mailchimp", fields, tags, now)
}

func init() {
	inputs.Add("mailchimp", func() telegraf.Input {
		return &MailChimp{}
	})
}
