# Atlassian Jira Input Plugin
---

The Jira plugin gathers metrics of Atlassian Jira using [rest/api/latest](https://docs.atlassian.com/software/jira/docs/api/REST/7.6.1/) endpoint.

### Configuration:

This section contains the default TOML to configure the Jira-Plugin.
You can generate it using `telegraf --usage jira`.

```toml
# Works with multiple Atlassian Jira instances
[[inputs.jira]]
  # Multiple Hosts from which to read ticket stats
  hosts = ["http://jira:8080/"]

  # Give here all fields to be selected. Each field will be counted grouped by the hosts, tags and JQLs below
  fields = ["priority", "custom_field_1234"]

  # Create tags based on these fields values
  tag_fields = ["custom_field_666"]


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
```

### Authentication:

Atlassian announced that [BasicAuth will be deprecated](https://developer.atlassian.com/cloud/jira/platform/deprecation-notice-basic-auth-and-cookie-based-auth/). Therefore you should not use this anymore.

For BasicAuth you can use the `username` and `password` fields in the `inputs.jira.authentication` section. Just enter the plain values there.
If you're using the API-Token, what is also a Basic-Authentication in the background, use the `email` and `token` fields.


### Metrics:

The JQL-Filter-Names are the base for the field names. Based on the example Configuration it's `new, closed, total`. Each of these values is used as Prefix for each `fields` Field. Based on the `Value` of the response of the group, the fieldname is finally built: `JQL_FIELD_VALUE`.

Let's describe it more simple:

- Given is a JQL-Filter with the name `new` which is selecting all new Issues from today.
- Given is a JQL-Filter with the name `closed` which is selecting all today closed Issues.
- Given are the Fields `fields = [ "priority", "custom_field_1234" ]`
- Given are the Tagfields `tag_fields = [ "custom_field_666" ]`

The two given Filters are the base for the field names: `new_` and `closed_`.
Each field is then appended to the base fieldname: `new_priority_`, `closed_priority_` and `new_custom_field_1234_`, `closed_custom_field_1234_`
After the values of the fields are taken. In Priority we have values for `high`, `medium` and `low`. The customfield is for the customer, lets say it has `Company`, `ACME`.
These values are taken and appended ot the fields as well: `new_priority_high`, `new_priority_medium`, `new_priority_low`, `closed_priority_high`, `closed_priority_medium`, `closed_priority_low`.
And for the customfield as well: `new_custom_field_1234_Company`, `new_custom_field_1234_Company`, `closed_custom_field_1234_ACME`, `closed_custom_field_1234_ACME`

The Tagfields are used to first group the issues based on the value of the fields. Let's assume in the `custom_field_666` we have which Dev-Team is responsible for the bug: `DevTeam`, `ConfigTeam`.
Each issue is then first checked if it's for the `DevTeam` or the `ConfigTeam` and grouped by this value. This results in a clean output which has only the amount of issues for which the teams are responsible for and which are interreting for them.

In this Example where we have the Customer in the `custom_field_1234` field, it could be better to use that field also as a tag. Then we would have the data grouped by the team and customer.

For more Information about how to filter query issues from Atlassian Jira, see their (Jira Server REST APIs)[https://developer.atlassian.com/server/jira/platform/rest-apis/].


### Example output:

Based on the example Configuration.

**Tags:**

* server
* custom_field_666

**Output Fields:**

* new_priority_critical
* new_priority_high
* new_priority_low
* closed_priority_critical
* closed_priority_high
* closed_priority_low

**Output**

```
jira,server=https://jira:8080/,custom_field_666=DevTeam new_priority_low=55,new_priority_high=10,new_priority_critical=1,closed_priority_low=55,closed_priority_high=10,closed_priority_critical=1,total_priority_low=55,total_priority_high=10,total_priority_critical=1 1536707179000000000
jira,server=https://jira:8080/,custom_field_666=ConfigTeam new_priority_low=55,new_priority_high=10,new_priority_critical=1,closed_priority_low=55,closed_priority_high=10,closed_priority_critical=1,total_priority_low=55,total_priority_high=10,total_priority_critical=1 1536707179000000000
```
