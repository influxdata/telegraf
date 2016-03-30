# Conntrack Plugin

Collects conntrack stats from the configured directories and files.

### Configuration:

```toml
 # Collects conntrack stats from the configured directories and files.
 [[inputs.conntrack]]
   ## The following defaults would work with multiple versions of contrack. Note the nf_ and ip_
   ## filename prefixes are mutually exclusive across conntrack versions, as are the directory locations.

   ## Superset of filenames to look for within the conntrack dirs. Missing files will be ignored.
   files = ["ip_conntrack_count","ip_conntrack_max","nf_conntrack_count","nf_conntrack_max"]

   ## Directories to search within for the conntrack files above. Missing directrories will be ignored.
   dirs = ["/proc/sys/net/ipv4/netfilter","/proc/sys/net/netfilter"]
```

### Measurements & Fields:

- conntrack
    - ip_conntrack_count (int, count): the number of entries in the conntrack table 
    - ip_conntrack_max (int, size): the max capacity of the conntrack table

### Tags:

This input does not use tags.

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter conntrack -test
conntrack,host=myhost ip_conntrack_count=2,ip_conntrack_max=262144 1461620427667995735
```
