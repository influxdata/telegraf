# Dependency license verification tool

This tool allows the verification of information in
`docs/LICENSE_OF_DEPENDENCIES.md` against the linked license
information. To do so, the license reported by the user is
checked against the license classification of the downloaded
license file for each dependency.

## Building

```shell
make build_tools
```

## Running

The simplest way to run the verification tool is to execute

```shell
telegraf$ ./tools/license_checker/license_checker
```

using the current directory as telegraf's root directory and verifies
all licenses. Only errors will be reported by default.

There are multiple options you can use to customize the verification.
Take a look at

```shell
telegraf$ ./tools/license_checker/license_checker --help
```

to get an overview.

As the verification tool downloads each license file linked in the
dependency license document, you should be careful on not exceeding
the access limits of e.g. GitHub by running the tool too frequent.

Some packages change the license for newer versions. As we always
link to the latest license text the classification might not match
the actual license of our used dependency. Furthermore, some license
text might be wrongly classified, or not classified at all. In these
cases, you can use a _whitelist_ to explicitly state the license
SPDX classifier for those packages.
See the [whitelist section](#whitelist) for more details.

The recommended use in telegraf is to run

```shell
telegraf$ ./tools/license_checker/license_checker \
              -whitelist ./tools/license_checker/data/whitelist
```

using the code-versioned whitelist. This command will report all
non-matching entries with an `ERR:` prefix.

## Whitelist

Whitelist entries contain explicit license information for
a set of packages to use instead of classification. Each entry
in the whitelist is a line of the form

```text
[comparison operator]<package name>[@vX.Y.Z] <license SPDX>
```

where the _comparison operator_ is one of `>`, `>=`, `=`, `<=` or `<`
and the _license SPDX_ is a [SPDX license identifier][spdx].
In case no package version is specified, the entry matches all versions
of the library. Furthermore, the comparison operator can be omitted
which is equivalent to an exact match (`=`).

The entries are processed in order until the first match is found.

Here is an example of a whitelist. Assume that you have library
`github.com/foo/bar` which started out with the `MIT` license
until version 1.0.0 where it changed to `EFL-1.0` until it again
changed to `EFL-2.0` starting __after__ version 2.3.0. In this case
the whitelist should look like this

```text
<github.com/foo/bar@v1.0.0 MIT
<=github.com/foo/bar@v2.3.0 EFL-1.0
github.com/foo/bar EFL-2.0
```

All versions below 1.0.0 are matched by the first line and are thus
classified as `MIT`. The second line matches everything that is
above 1.0.0 (thus not matched by the first line) until (and including)
2.3.0. The last line with catch everything that was passing the first
two lines i.e. everything after 2.3.0.

[spdx]: https://spdx.org/licenses/
