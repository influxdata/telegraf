# Pull Requests

## Before Review

Ensure that the CLA is signed (the `telegraf-tiger` bot performs this check).  The
only exemption would be non-copyrightable changes such as fixing a typo.

Check that all tests are passing.  Due to intermittent errors in the CI tests
it may be required to check the cause of test failures and restart failed
tests and/or create new issues to fix intermittent test failures.

Ensure that PR is opened against the master branch as all changes are merged
to master initially.  It is possible to change the branch a pull request is
opened against but it often results in many conflicts, change it before
reviewing and then if needed ask the contributor to rebase.

Ensure there are no merge conflicts.  If there are conflicts, ask the
contributor to merge or rebase.

## Review

[Review the pull request](https://github.com/influxdata/telegraf/blob/master/docs/developers/REVIEWS.md).

## Merge

Determine what release the change will be applied to.  New features should
be added only to master, and will be released in the next minor version (1.x).
Bug fixes can be backported to the current release branch to go out with the
next patch release (1.7.x) unless the bug is too risky to backport or there is
an easy workaround.  Set the correct milestone on the pull request and any
associated issue.

All pull requests are merged using the "Squash and Merge" strategy on Github.
This method is used because many pull requests do not have a clean change
history and this method allows us to normalize commit messages as well as
simplifies backporting.

### Rewriting the commit message

After selecting "Squash and Merge" you may need to rewrite the commit message.
Usually the body of the commit messages should be cleared as well, unless it
is well written and applies to the entire changeset.

- Use imperative present tense for the first line of the message:
  - Use "Add tests for" (instead of "I added tests for" or "Adding tests for")
- The default merge commit messages include the PR number at the end of the
commit message, keep this in the final message.
- If applicable mention the plugin in the message.

**Example Enhancement:**

> Add user tag to procstat input (#4386)

**Example Bug Fix:**

> Fix output format of printer processor (#4417)

## After Merge

[Update the Changelog](https://github.com/influxdata/telegraf/blob/master/docs/maintainers/CHANGELOG.md).

If required, backport the patch and the changelog update to the current
release branch.  Usually this can be done by cherry picking the commits:

```shell
git cherry-pick -x aaaaaaaa bbbbbbbb
```

Backporting changes to the changelog often pulls in unwanted changes.  After
cherry picking commits, double check that the only the expected lines are
modified and if needed clean up the changelog and amend the change.  Push the
new master and release branch to Github.
