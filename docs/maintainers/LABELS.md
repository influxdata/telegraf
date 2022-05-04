# Labels

This page describes the meaning of the various
[labels](https://github.com/influxdata/telegraf/labels) we use on the Github
issue tracker.

## Categories

New issues are automatically labeled `feature request`, `bug`, or `support`.
If you are unsure what problem the author is proposing, you can use the `need more info` label
and if there is another issue you can add the `closed/duplicate` label and close the
new issue.

New pull requests are usually labeled one of `enhancement`, `bugfix` or `new
plugin`.

## Additional Labels

Apply any of the `area/*` labels that match.  If an area doesn't exist, new
ones can be added but **it is not a goal to have an area for all issues.**

If the issue only applies to one platform, you can use a `platform/*` label.
These are only applied to single platform issues which are not on Linux.

For bugs you may want to add `panic`, `regression`, or `upstream` to provide
further detail.

Summary of Labels:

| Label | Description | Purpose |
| --- | ----------- | ---|
| `area/*` | These labels each corresponding to a plugin or group of plugins that can be added to identify the affected plugin or group of plugins | categorization |
| `breaking change` | Improvement to Telegraf that requires breaking changes to the plugin or agent; for minor/major releases | triage |
| `bug` | New issue for an existing component of Telegraf | triage |
| `cloud` | Issues or request around cloud environments | categorization |
| `dependencies` | Pull requests that update a dependency file | triage |
| `discussion` | Issues open for discussion | community/categorization |
| `documentation` | Issues related to Telegraf documentation and configuration descriptions | categorization |
| `error handling` | Issues related to error handling | categorization |
| `external plugin` | Plugins that would be ideal external plugin and expedite being able to use plugin w/ Telegraf | categorization |
| `good first issue` | This is a smaller issue suited for getting started in Telegraf, Golang, and contributing to OSS | community |
| `help wanted` | Request for community participation, code, contribution | community |
| `need more info` | Issue triaged but outstanding questions remain | community |
| `performance` | Issues or PRs that address performance issues | categorization|
| `platform/*` | Issues that only apply to one platform | categorization |
| `plugin/*` | Request for new plugins and issues/PRs that are related to plugins | categorization |
| `ready for final review` | Pull request has been reviewed and/or tested by multiple users and is ready for a final review | triage |
| `rfc` | Request for comment - larger topics of discussion that are looking for feedback | community |
| `support` |Telegraf questions, may be directed to community site or slack | triage |
| `upstream` | Bug or issues that rely on dependency fixes and we cannot fix independently | triage |
| `waiting for response` | Waiting for response from contributor | community/triage |
| `wip` | PR still Work In Progress, not ready for detailed review | triage |

Labels starting with `pm` are not applied by maintainers.

## Closing Issues

We close issues for the following reasons:

| Label | Reason |
| --- | ----------- |
| `closed/as-designed` | Labels to be used when closing an issue or PR with short description why it was closed |
| `closed/duplicate` | This issue or pull request already exists |
| `closed/external-candidate` | The feature request is best implemented by an external plugin |
| `closed/external-issue` | The feature request is best implemented by an external plugin |
| `closed/needs more info` | Did not receive the information we need within 3 months from last activity on issue |
| `closed/not-reproducible` | Given the information we have we can't reproduce the issue |
| `closed/out-of-scope` | The feature request is out of scope for Telegraf - highly unlikely to be worked on |
| `closed/question` | This issue is a support question, directed to community site or slack |
