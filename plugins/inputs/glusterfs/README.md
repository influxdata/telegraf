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

Since telegraf doesn't run as root, you will need to configure sudo to work without
password for the gluster command. For example, run the command visudo then add :
```
telegraf ALL=(ALL) NOPASSWD:/usr/sbin/gluster
```

You might need to adjust that if your distribution has a different path.

### Measurements & Fields:

glusterfs -> read is the number of bytes read from the brick
glusterfs -> write is the number of bytes written to the brick

### Tags:

- All measurements have the following tags:
    - volume : the volume name
    - brick : the brick name
