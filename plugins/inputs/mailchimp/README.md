# Mailchimp Input

Pulls campaign reports from the [Mailchimp API](https://developer.mailchimp.com/).

### Configuration

```toml
[[inputs.mailchimp]]
  ## MailChimp API key
  ## get from https://admin.mailchimp.com/account/api/
  api_key = "" # required
  ## Reports for campaigns sent more than days_old ago will not be collected.
  ## 0 means collect all.
  days_old = 0
  ## Campaign ID to get, if empty gets all campaigns, this option overrides days_old
  # campaign_id = ""
```

### Metrics

- mailchimp
  - id
  - campaign_title
    - emails_sent
    - abuse_reports
    - unsubscribed
    - hard_bounces
    - soft_bounces
    - syntax_errors
    - forwards_count
    - forwards_opens
    - opens_total
    - unique_opens
    - open_rate
    - clicks_total
    - unique_clicks
    - unique_subscriber_clicks
    - click_rate
    - facebook_recipient_likes
    - facebook_unique_likes
    - facebook_likes
    - industry_type
    - industry_open_rate
    - industry_click_rate
    - industry_bounce_rate
    - industry_unopen_rate
    - industry_unsub_rate
    - industry_abuse_rate
    - list_stats_sub_rate
    - list_stats_unsub_rate
    - list_stats_open_rate
    - list_stats_click_rate
