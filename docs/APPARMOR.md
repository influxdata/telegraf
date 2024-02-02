# AppArmor

When running Telegraf under AppArmor users may see denial messages depending on
the Telegraf plugins used and the AppArmor profile applied. Telegraf does not
have control over the AppArmor profiles used. If users wish to address denials,
then they must understand the collections made by their choice of Telegraf
plugins, the denial messages, and the impact of changes to their AppArmor
profiles.

## Example Denial

For example, users might see denial messages such as:

```s
type=AVC msg=audit(1588901740.036:2457789): apparmor="DENIED" operation="ptrace" profile="docker-default" pid=9030 comm="telegraf" requested_mask="read" denied_mask="read" peer="unconfined"
```

In this case, Telegraf will also need the ability to ptrace(read). User's will
first need to analyze the denial message for the operation and requested mask.
Then consider if the required changes make sense. There may be additional
denials even after initial changes.

For more details around AppArmor settings and configuration, users can check out
the `man 5 apparmor.d` man page on their system or the [AppArmor wiki][wiki].

[wiki]: https://gitlab.com/apparmor/apparmor/-/wikis/home
