## 🚨 AI Policy & Operations

Operational constraints and procedures extracted from CONTRIBUTING.md.

### AI Policy

We currently cannot accept AI-generated code contributions. Code contributed must be your own per the CLA.

### Creating a Pull Request

Follow these steps when creating a pull request:

1. **Sign the CLA** - Sign the Contributor License Agreement
2. **Open an issue** - Discuss changes you would like to make (not strictly required but helps reduce rework)
3. **Make changes** - Use the guidelines in CONTRIBUTING.md for plugins:
   - Input Plugins
   - Processor Plugins
   - Aggregator Plugins
   - Output Plugins
4. **Add tests and documentation** - Ensure you have added proper unit tests and documentation
5. **Open pull request** - Submit your pull request for review
6. **Follow naming conventions** - The pull request title must follow [conventional commit format][semcommit]

> **Note:** If you have a pull request with only one commit, that commit needs to follow the conventional commit format. GitHub will use the pull request title if there are multiple commits, but will use the commit message if there is only one commit.

[semcommit]: https://www.conventionalcommits.org/en/v1.0.0/#summary

### Adding Dependencies

When adding a new dependency to the project:

1. Use `go get github.com/[dependency]/[new-package]` to add the dependency
2. Ensure the dependency is properly documented
3. Update relevant documentation about the new dependency

### Support and Questions

For general support questions such as "How do I use the MongoDB plugin?", please refer to:

- [Community Slack][slack]
- [Community Page][forum]

Do not file these as GitHub issues. Code contributed must be your own per the CLA.

[slack]: https://influxdata.com/slack
[forum]: https://community.influxdata.com/