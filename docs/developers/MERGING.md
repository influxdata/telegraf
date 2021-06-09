# Merging pull requests

When a pull request is ready to be merged and meets all the review criteria, then you can click `squash and merge`.

## Follow up steps after merging a `bug` pull request

If the pull request fixes a bug in Telegraf (and have the `bug` label) then you must prepare it for the next patch release. A pull request can be considered a bug fix for multiple reasons, such as fixing a security issue or a fix to prevent a plugin from throwing an unexpected error. Anything introducing new features or behaviour, such as an entire new plugin should not be considered a bug and will be released in the next feature release. Any script, makefile, or ci change is also not considered a bug, even if it is techinically a bug it is not an issue with the final resulting Telegraf binary then it can just stay in the master branch.

### 1. Update `CHANGELOG.md` in `master` branch

Add a line describing the merged pull request in the `CHANGELOG.md` in the `master` branch. If there isn't already a new header for the next patch release that this bug will be included in then create a new one. Then under `Bugfixes` add a line following this format (this is an example taken from the changelog):

`- [#9182](https://github.com/influxdata/telegraf/pull/9182) Update pgx to v4`

If you need more examples, look at previous releases and follow the same format. Be sure to spend time on creating a good description that will help users understand what changed.

### 2. Cherry pick the merged pull request into the latest release branch

The release branches follow the naming convention `release-x.xx` where the `x` are replaced by the release number (the latest branch being the one with the highest number). The git commands you need to run are as follows:

1. `git checkout #{release_branch} origin/#{release_branch}`
2. `git cherry-pick -x #{sha}`

### 3. Cherry-pick the updated `CHANGELOG.md` into the latest release branch

Use the same commands as step 1, but be sure to check if there will be another patch release in the google calander. If there isn't going to be another patch release you can skip this step.

### 4. Add a comment to the merged pull request stating you have added it to the release branch

suggested text: "Thank you for your contribution! This change will be included in the next minor release vx.xx.x"
