# Template Processor

The `template` processor applies a go template to tag, field, measurement and time values to create a new tag.

Golang [Template Documentation]

### Configuration

```toml
  # Concatenate two tags to create a new tag
  [[processors.template]]
     ## Tag to create
     tag = "topic"
     ## Template to create tag
     # Note: Single quotes (') are used, so the double quotes (") don't need escaping (\")
     template = '{{ .Tag "hostname" }}.{{ .Tag "level" }}'
```

### Example

```diff
- cpu,level=debug,hostname=localhost value=42i
+ cpu,level=debug,hostname=localhost,topic=localhost.debug value=42i
```

[Template Documentation]:https://golang.org/pkg/text/template/