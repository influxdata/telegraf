# GlusterFS Input Plugin

[GlusterFS](https://www.gluster.org) input plugin gathers metrics directly from any running GlusterFS volume. It can do so by using the gluster profiler.

### Configuration:

```toml
# SampleConfig
[[inputs.glusterfs]]
  volumes = ["volume-name"]
```

The configuration must include the volume(s) name(s) for this plugin to work.
You also need to activate the profiler on the volume(s) by running this command :
```
gluster volume profile volume-name start
```

### Measurements & Fields:

glusterfs_read is the number of bytes read from the brick
glusterfs_write is the number of bytes written to the brick

### Tags:

- All measurements have the following tags:
    - volume : the volume name
    - brick : the brick name
