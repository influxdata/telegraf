# GitHub Input Plugin

The [GitHub](https://www.github.com) input plugin gathers statistics from GitHub repositories.

### Configuration:

```toml
[[inputs.github]]
    repositories = [
      "owner/repository",
    ]
```

### Metrics:

For more details about collected metrics reference the [HAProxy CSV format
documentation](https://cbonte.github.io/haproxy-dconv/1.8/management.html#9.1).

- github
  - tags:
    - `full_name` - The full name of the repository, including owner/organization
    - `name` - The repository name
    - `owner` - The owner of the repository
    - `language` - The primary language of the repository
  - fields:
    - `stars` (int)
    - `forks` (int)
    - `open_issues` (int)
    - `size` (int)

### Example Output:

```
github,full_name=influxdata/telegraf,language=Go,license=MIT\ License,name=telegraf,owner=influxdata stars=6401i,forks=2421i,open_issues=722i,size=22611i 1552651811000000000
```
