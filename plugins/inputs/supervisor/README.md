# Supervisor Input Plugin

The plugin gathers metrics of supervisord managed processes

### Configuration:

```toml
# Description
[[inputs.supervisor]]
  # host = "http://localhost:9001/RPC2"  # default
```

### Measurements & Fields:

Structure based on http://supervisord.org/api.html#supervisor.rpcinterface.SupervisorNamespaceRPCInterface.getProcessInfo

- supervisor
    - Name          (string)
    - Group         (string)
    - Description   (string)
    - Start         (int, UNIX timestamp)
    - Stop          (int, UNIX timestamp)
    - Now           (int, UNIX timestamp)
    - State         (int, http://supervisord.org/subprocess.html#process-states)
    - Statename     (string, http://supervisord.org/subprocess.html#process-states)
    - StdoutLogfile (string)
    - StderrLogfile (string)
    - SpawnErr      (string)
    - ExitStatus    (int, 0 if running)
    - Pid           (int, UNIX process ID)

### Tags:

- All measurements have the following tags:
    - server (supervisor xmlrpc url)
    - process (name of the program managed by supervisor)

### Sample Queries:

```
SELECT "Name", "Statename", "Pid" FROM "supervisor" GROUP BY "process"
```

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter supervisor -test
supervisor,server=http://127.0.0.1:9001/RPC2,process=process-1,Start=1484735889i,Stop=0i,State=20i,Statename="RUNNING",StdoutLogfile="/tmp/process-1-stdout---supervisor-GLbfBU.log",SpawnErr="",Group="process-1",Description="pid 9, uptime 0:08:30",StderrLogfile="/tmp/process-1-stderr---supervisor-KxEKGw.log",ExitStatus=0i,Pid=9i,Name="process-1",Now=1484736399i 1484736400000000000
```
