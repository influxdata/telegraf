# Metric Format Changes

When making changes to an existing input plugin, care must be taken not to change the metric format in ways that will cause trouble for existing users.  This document helps developers understand how to make metric format changes safely.

## Changes can cause incompatibilities

If the metric format changes, data collected in the new format can be incompatible with data in the old format.  Database queries designed around the old format may not work with the new format.  This can cause application failures.

Some metric format changes don't cause incompatibilities.  Also, some unsafe changes are necessary.  How do you know what changes are safe and what to do if your change isn't safe?

## Guidelines

The main guideline is just to keep compatibility in mind when making changes.  Often developers are focused on making a change that fixes their particular problem and they forget that many people use the existing code and will upgrade.  When you're coding, keep existing users and applications in mind.

### Renaming, removing, reusing

Database queries refer to the metric and its tags and fields by name.  Any Telegraf code change that changes those names has the potential to break an existing query.  Similarly, removing tags or fields can break queries.

Changing the meaning of an existing tag value or field value or reusing an existing one in a new way isn't safe.  Although queries that use these tags/field may not break, they will not work as they did before the change.

Adding a field doesn't break existing queries.  Queries that select all fields and/or tags (like "select * from") will return an extra series but this is often useful.

### Performance and storage

Time series databases can store large amounts of data but many of them don't perform well on high cardinality data.  If a metric format change includes a new tag that holds high cardinality data, database performance could be reduced enough to cause existing applications not to work as they previously did.  Metric format changes that dramatically increase the number of tags or fields of a metric can increase database storage requirements unexpectedly.  Both of these types of changes are unsafe.

### Make unsafe changes opt-in

If your change has the potential to seriously affect existing users, the change must be opt-in.  To do this, add a plugin configuration setting that lets the user select the metric format.  Make the setting's default value select the old metric format.  When new users add the plugin they can choose the new format and get its benefits.  When existing users upgrade, their config files won't have the new setting so the default will ensure that there is no change.

When adding a setting, avoid using a boolean and consider instead a string or int for future flexibility.  A boolean can only handle two formats but a string can handle many.  For example, compare use_new_format=true and features=["enable_foo_fields"]; the latter is much easier to extend and still very descriptive.

If you want to encourage existing users to use the new format you can log a warning once on startup when the old format is selected.  The warning should tell users in a gentle way that they can upgrade to a better metric format.  If it doesn't make sense to maintain multiple metric formats forever, you can change the default on a major release or even remove the old format completely.  See [[Deprecation]] for details.

### Utility

Changes should be useful to many or most users.  A change that is only useful for a small number of users may not accepted, even if it's off by default.

## Summary table

|         | delete | rename | add |
| ------- | ------ | ------ | --- |
| metric  | unsafe | unsafe | safe |
| tag     | unsafe | unsafe | be careful with cardinality |
| field   | unsafe | unsafe | ok as long as it's useful for existing users and is worth the added space |

## References

InfluxDB Documentation: "Schema and data layout"
