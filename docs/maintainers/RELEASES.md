# Releases

## Release Branch

On master, update `etc/telegraf.conf` and commit:

```sh
./telegraf config > etc/telegraf.conf
```

Create the new release branch:

```sh
git checkout -b release-1.15
```

Push the changes:

```sh
git push origin release-1.15 master
```

Update next version strings on master:

```sh
git checkout master
echo 1.16.0 > build_version.txt
```

## Release Candidate

Release candidates are created only for new minor releases (ex: 1.15.0).   Tags
are created but some of the other tasks, such as adding a changelog entry are
skipped.  Packages are added to the github release page and posted to
community but are not posted to package repos or docker hub.

```sh
git checkout release-1.15
git commit --allow-empty -m "Telegraf 1.15.0-rc1"
git tag -s v1.15.0-rc1 -m "Telegraf 1.15.0-rc1"
git push origin release-1.15 v1.15.0-rc1
```

## Release

On master, set the release date in the changelog and cherry-pick the change
back:

```sh
git checkout master
vi CHANGELOG.md
git commit -m "Set 1.8.0 release date"
git checkout release-1.8
git cherry-pick -x <rev>
```

Double check that the changelog was applied as desired, or fix it up and
amend the change before pushing.

Tag the release:

```sh
git checkout release-1.8
# This just improves the `git show 1.8.0` output
git commit --allow-empty -m "Telegraf 1.8.0"
git tag -s v1.8.0 -m "Telegraf 1.8.0"
```

Check that the version was set correctly, the tag can always be altered if a
mistake is made but only before you push it to Github:

```sh
make
./telegraf --version
Telegraf v1.8.0 (git: release-1.8 aaaaaaaa)
```

When you push a branch with a tag to Github, CircleCI will be triggered to
build the packages.

```sh
git push origin master release-1.8 v1.8.0
```

Set the release notes on Github.

Update webpage download links.

Update apt and yum repositories hosted at repos.influxdata.com.

Update the package signatures on S3, these are used primarily by the docker images.

Update docker image [influxdata/influxdata-docker](https://github.com/influxdata/influxdata-docker):

```sh
cd influxdata-docker
git co master
git pull
git co -b telegraf-1.8.0
telegraf/1.8/Dockerfile
telegraf/1.8/alpine/Dockerfile
git commit -am "telegraf 1.8.0"
```

Official company post to RSS/community.

Update documentation on docs.influxdata.com
