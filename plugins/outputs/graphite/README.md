# Graphite Output Plugin

This plugin writes to [Graphite](http://graphite.readthedocs.org/en/latest/index.html) via raw TCP.

Parameters:

    Servers []string
    Prefix  string
    Timeout int

* `servers`: List of strings, ["mygraphiteserver:2003"].
* `prefix`: String use to prefix all sent metrics.
* `timeout`: Connection timeout in second.
