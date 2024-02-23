# Telegraf Specification Overview

## Objective

Define and layout the Telegraf specification process.

## Overview

The general goal of a spec is to detail the work that needs to get accomplished
for a new feature. A developer should be able to pick up a spec and have a
decent understanding of the objective, the steps required, and most of the
general design decisions.

The specs can then live in the Telegraf repository to share and involve the
community in the process of planning larger changes or new features. The specs
also serve as a public historical record for changes.

## Process

The general workflow is for a user to put up a PR with a spec outlining the
task, have any discussion in the PR, reach consensus, and ultimately commit
the finished spec to the repo.

While researching a new feature may involve an investment of time, writing the
spec should be relatively quick. It should not take hours of time.

## Spec naming

Please name the actual file prefixed with `tsd` and the next available
number, for example:

* tsd-001-agent-write-ahead-log.md
* tsd-002-inputs-apache-increase-timeout.md
* tsd-003-serializers-parquet.md

All lower-case and separated by hyphens.

## What belongs in a spec

A spec should involve the creation of a markdown file with at least an objective
and overview:

* Objective (required) - One sentence headline
* Overview (required) - Explain the reasoning for the new feature and any
  historical information. Answer the why this is needed.

Please feel free to make a copy the template.md and start with that.

The user is free to add additional sections or parts in order to express and
convey a new feature. For example this might include:

* Keywords - Help identify what the spec is about
* Is/Is-not - Explicitly state what this change includes and does not include
* Prior Art - Point at existing or previous PRs, issues, or other works that
  demonstrate the feature or need for it.
* Open Questions - Section with open questions that can get captured in
  updates to the PR

## Changing existing specs

Small changes which are non-substantive, like grammar or formatting are gladly
accepted.

After a feature is complete it may make sense to come back and update a spec
based on the final result.

Other changes that make substantive changes are entirely up to the maintainers
whether the edits to an existing RFC will be accepted. In general, finished
specs should be considered complete and done, however, priorities, details, or
other situations may evolve over time and as such introduce the need to make
updates.
