Feedback from Aliaksey K., based on experiences such as erlang
XDCR-DCP use case...

- consider adding uuid to everywhere in Receiver interface -- anywhere
  that there's a sequence number?

- advanced option idea - how about an option to not magically heal if
  there's a vbucket state or cluster topology change?
  - some apps like xdcr might not want magically healing, and instead
    want vbucket stickiness to a node (like xdcr tries to maintain one
    connection per pair of local & remote nodes).  If something
    happens, it doesn't want a vbucket which moved to another source
    server to automatically be connected to.
  - one approach to do this today is the application can provide its
    own ConnectBucket() implementation in the BucketDataSourceOptions.
    if there's a connection attempt to a server the application
    doesn't want, then the application can reject the connection and
    do a BucketDataSource.Close().

- feedback on high level versus low level API mixing...
  - why not provide full bucket url?
  - or, why not accept already configured and auth'ed
    go-couchbase connection instance?
    - connection & auth approaches are varied and changing in future
    - if go-couchbase doesn't auto-heal, perhaps it should and not
      be a concern of cbdatasource?




