# Mailchimp Input Plugin

This plugin gathers metrics from the [Mailchimp][mailchimp] service using the
[Mailchimp API][api].

⭐ Telegraf v0.2.4
🏷️ cloud, web
💻 all

[mailchimp]: https://mailchimp.com
[api]: https://developer.mailchimp.com/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gathers metrics from the /3.0/reports MailChimp API
[[inputs.mailchimp]]
  ## MailChimp API key
  ## get from https://admin.mailchimp.com/account/api/
  api_key = "" # required

  ## Reports for campaigns sent more than days_old ago will not be collected.
  ## 0 means collect all and is the default value.
  days_old = 0

  ## Campaign ID to get, if empty gets all campaigns, this option overrides days_old
  # campaign_id = ""
```

## Metrics

- mailchimp
  - tags:
    - id
    - campaign_title
  - fields:
    - emails_sent (integer, emails)
    - abuse_reports (integer, reports)
    - unsubscribed (integer, unsubscribes)
    - hard_bounces (integer, emails)
    - soft_bounces (integer, emails)
    - syntax_errors (integer, errors)
    - forwards_count (integer, emails)
    - forwards_opens (integer, emails)
    - opens_total (integer, emails)
    - unique_opens (integer, emails)
    - open_rate (double, percentage)
    - clicks_total (integer, clicks)
    - unique_clicks (integer, clicks)
    - unique_subscriber_clicks (integer, clicks)
    - click_rate (double, percentage)
    - facebook_recipient_likes (integer, likes)
    - facebook_unique_likes (integer, likes)
    - facebook_likes (integer, likes)
    - industry_type (string, type)
    - industry_open_rate (double, percentage)
    - industry_click_rate (double, percentage)
    - industry_bounce_rate (double, percentage)
    - industry_unopen_rate (double, percentage)
    - industry_unsub_rate (double, percentage)
    - industry_abuse_rate (double, percentage)
    - list_stats_sub_rate (double, percentage)
    - list_stats_unsub_rate (double, percentage)
    - list_stats_open_rate (double, percentage)
    - list_stats_click_rate (double, percentage)

## Example Output

```text
mailchimp,campaign_title=Freddie's\ Jokes\ Vol.\ 1,id=42694e9e57 abuse_reports=0i,click_rate=42,clicks_total=42i,emails_sent=200i,facebook_likes=42i,facebook_recipient_likes=5i,facebook_unique_likes=8i,forwards_count=0i,forwards_opens=0i,hard_bounces=0i,industry_abuse_rate=0.00021111996110887,industry_bounce_rate=0.0063767751251474,industry_click_rate=0.027431311866951,industry_open_rate=0.17076777144396,industry_type="Social Networks and Online Communities",industry_unopen_rate=0.82285545343089,industry_unsub_rate=0.001436957032815,list_stats_click_rate=42,list_stats_open_rate=42,list_stats_sub_rate=10,list_stats_unsub_rate=20,open_rate=42,opens_total=186i,soft_bounces=2i,syntax_errors=0i,unique_clicks=400i,unique_opens=100i,unique_subscriber_clicks=42i,unsubscribed=2i 1741188555526302348
```
