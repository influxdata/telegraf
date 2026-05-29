# Prose Style Guide

This document describes the prose linting setup for Telegraf documentation.

## Overview

Telegraf uses [Vale](https://vale.sh/) for prose linting alongside existing
tools like markdownlint. While markdownlint checks Markdown formatting, Vale
checks prose style, grammar, and terminology consistency.

## Running Vale locally

### Installation

Install Vale using your package manager:

```shell
# macOS
brew install vale

# Windows (scoop)
scoop install vale

# Linux (snap)
snap install vale
```

Or download from [Vale releases](https://github.com/errata-ai/vale/releases).

### Syncing styles

After installation, download the required style packages:

```shell
vale sync
```

This downloads the Google style guide package specified in `.vale.ini`.

### Running the linter

Lint all Markdown files:

```shell
vale .
```

Lint specific files:

```shell
vale plugins/inputs/cpu/README.md
```

Lint only changed files in your branch:

```shell
vale $(git diff --name-only master -- '*.md')
```

## Style rules

Vale is configured in `.vale.ini` at the repository root. Custom Telegraf rules
are in `.vale/styles/Telegraf/`.

### Telegraf-specific rules

The following custom rules are enforced:

#### Latin abbreviations (`Telegraf.Latin`)

Avoid Latin abbreviations in documentation. Use plain English instead:

| Avoid | Use instead |
|-------|-------------|
| e.g.  | for example |
| i.e.  | that is     |
| etc.  | and so on   |
| viz.  | namely      |
| vs.   | versus      |

#### Grammar patterns (`Telegraf.Grammar`)

Avoid common grammar issues:

| Avoid | Use instead |
|-------|-------------|
| allows to | lets you |
| in order to | to |
| In case | If |

#### Terminology (`Telegraf.Terms`)

Use correct capitalization for product names:

| Avoid | Use instead |
|-------|-------------|
| influxdb | InfluxDB |
| telegraf | Telegraf |
| time-series | time series |

### Google style guide

Vale also applies rules from the
[Google Developer Documentation Style Guide](https://developers.google.com/style).
Some rules are relaxed for technical documentation (see `.vale.ini`).

## CI integration

Vale runs automatically on pull requests that modify Markdown files. The
workflow is defined in `.github/workflows/vale.yml`.

- Vale runs only on changed files
- Currently configured to report issues but not fail the build
- Results appear as PR check annotations

## Disabling rules

To disable a rule for a specific section, use Vale comments:

```markdown
<!-- vale Telegraf.Latin = NO -->
This section uses e.g. intentionally.
<!-- vale Telegraf.Latin = YES -->
```

To disable all rules for a section:

```markdown
<!-- vale off -->
This section is not linted.
<!-- vale on -->
```

## Updating rules

To modify or add rules:

1. Edit files in `.vale/styles/Telegraf/`
2. Test changes locally with `vale <file>`
3. Submit a pull request with the rule changes

See the [Vale documentation](https://vale.sh/docs/) for rule syntax.
