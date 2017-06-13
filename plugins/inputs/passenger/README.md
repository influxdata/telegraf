# Telegraf plugin: passenger

Get phusion passenger stat using their command line utility
`passenger-status`

# Measurements

Meta:

- tags:

  * name
  * passenger_version
  * pid
  * code_revision

Measurement names:

- passenger:

  * Tags: `passenger_version`
  * Fields:

    - process_count
		- max
		- capacity_used
		- get_wait_list_size

- passenger_supergroup:

    * Tags: `name`
    * Fields:

      - get_wait_list_size
      - capacity_used

- passenger_group:

  * Tags:

    - name
    - app_root
    - app_type

  * Fields:

    - get_wait_list_size
    - capacity_used
    - processes_being_spawned

- passenger_process:

  * Tags:

    - group_name
    - app_root
    - supergroup_name
    - pid
    - code_revision
    - life_status
    - process_group_id

  * Field:

    - concurrency
    - sessions
    - busyness
    - processed
    - spawner_creation_time
    - spawn_start_time
    - spawn_end_time
    - last_used
    - uptime
    - cpu
    - rss
    - pss
    - private_dirty
    - swap
    - real_memory
    - vmsize

# Example output

Using this configuration:

```
[[inputs.passenger]]
  # Path of passenger-status.
  #
  # Plugin gather metric via parsing XML output of passenger-status
  # More information about the tool:
  #   https://www.phusionpassenger.com/library/admin/apache/overall_status_report.html
  #
  #
  # If no path is specified, then the plugin simply execute passenger-status
  # hopefully it can be found in your PATH
  command = "passenger-status -v --show=xml"
```

When run with:

```
./telegraf --config telegraf.conf --input-filter passenger --test
```

It produces:

```
> passenger,passenger_version=5.0.17 capacity_used=23i,get_wait_list_size=0i,max=23i,process_count=23i 1452984112799414257
> passenger_supergroup,name=/var/app/current/public capacity_used=23i,get_wait_list_size=0i 1452984112799496977
> passenger_group,app_root=/var/app/current,app_type=rack,name=/var/app/current/public capacity_used=23i,get_wait_list_size=0i,processes_being_spawned=0i 1452984112799527021
> passenger_process,app_root=/var/app/current,code_revision=899ac7f,group_name=/var/app/current/public,life_status=ALIVE,pid=11553,process_group_id=13608,supergroup_name=/var/app/current/public busyness=0i,concurrency=1i,cpu=58i,last_used=1452747071764940i,private_dirty=314900i,processed=951i,pss=319391i,real_memory=314900i,rss=418548i,sessions=0i,spawn_end_time=1452746845013365i,spawn_start_time=1452746844946982i,spawner_creation_time=1452746835922747i,swap=0i,uptime=226i,vmsize=1563580i 1452984112799571490
> passenger_process,app_root=/var/app/current,code_revision=899ac7f,group_name=/var/app/current/public,life_status=ALIVE,pid=11563,process_group_id=13608,supergroup_name=/var/app/current/public busyness=2147483647i,concurrency=1i,cpu=47i,last_used=1452747071709179i,private_dirty=309240i,processed=756i,pss=314036i,real_memory=309240i,rss=418296i,sessions=1i,spawn_end_time=1452746845172460i,spawn_start_time=1452746845136882i,spawner_creation_time=1452746835922747i,swap=0i,uptime=226i,vmsize=1563608i 1452984112799638581
```

# Note

You have to ensure that you can run the `passenger-status` command under
telegraf user. Depend on how you install and configure passenger, this
maybe an issue for you. If you are using passenger standlone, or compile
yourself, it is straight forward. However, if you are using gem and
`rvm`, it maybe harder to get this right.

Such as with `rvm`, you can use this command:

```
~/.rvm/bin/rvm default do passenger-status -v --show=xml
```

You can use `&` and `;` in the shell command to run comlicated shell command
in order to get the passenger-status such as load the rvm shell, source the
path
```
command = "source .rvm/scripts/rvm && passenger-status -v --show=xml"
```

Anyway, just ensure that you can run the command under `telegraf` user, and it
has to produce XML output.
