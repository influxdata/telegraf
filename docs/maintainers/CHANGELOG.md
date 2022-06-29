# Changelog

The changelog contains the list of changes by version in addition to release
notes.  The file is updated immediately after adding a change that impacts
users.  Changes that don't effect the functionality of Telegraf, such as
refactoring code, are not included.

The changelog entries are added by a maintainer after merging a pull request.
We experimented with requiring the pull request contributor to add the entry,
which had a nice side-effect of reducing the number of changelog only commits
in the log history, however this had several drawbacks:

- The entry often needed reworded.
- Entries frequently caused merge conflicts.
- Required contributor to know which version a change was accepted into.
- Merge conflicts made it more time consuming to backport changes.

Changes are added only to the first version a change is added in.  For
example, a change backported to 1.7.2 would only appear under 1.7.2 and not in
1.8.0.  This may become confusing if we begin supporting more than one
previous version but works well for now.

## Updating

If the change resulted in deprecation, mention the deprecation in the Release
Notes section of the version.  In general all changes that require or
recommend the user to perform an action when upgrading should be mentioned in
the release notes.

If a new plugin has been added, include it in a section based on the type of
the plugin.

All user facing changes, including those already mentioned in the release
notes or new plugin sections, should be added to either the Features or
Bugfixes section.

Features should generally link to the pull request since this describes the
actual implementation.  Bug fixes should link to the issue instead of the pull
request since this describes the problem, if a bug has been fixed but does not
have an issue then it is okay to link to the pull request.

It is usually okay to just use the shortlog commit message, but if needed
it can differ or be further clarified in the changelog.
