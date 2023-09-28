# apparmor

When running Telegraf under apparmor users may see denial messages depending on
the Telegraf plugins used and the apparmor profile applied. Telegraf does not
have control over the apparmor profiles used. If users wish to address denials,
then they must understand the collections made by their choice of Telegraf
plugins, the denial messages, and the impact of changes to their apparmor
profiles.

## Example Denial

For example, users might see denial messages such as:

```s
type=AVC msg=audit(1588901740.036:2457789): apparmor="DENIED" operation="ptrace" profile="docker-default" pid=9030 comm="telegraf" requested_mask="read" denied_mask="read" peer="dovecot"
```

In this case, Telegraf will also need the ability to ptrace(read). User's will
first need to analyze the denial message for the operation and requested mask.
Then consider if the required changes make sense. There may be additional
denials even after initial changes.
