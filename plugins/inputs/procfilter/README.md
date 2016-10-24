# Telegraf plugin: procfilter

## Description

The procfilter plugin monitors system resource usage by process or groups of processes.

The description of what you want to measure is done using a simple language (in the script= or script_file= configuration option)

A correct script is composed of a sequence of:

* Comments:  
A comment starts with a # and stops at the end of line.
eg: `# This is a nice comment.`  

* Filters:  
A filter describes rules to select processes. (see below for the list of filters)  
eg: `joe \<\- user('joe')`   
This filter will select all processes belonging to the joe user.  

* Measurements:  
A measurement has a name, a tag() and/or field() declaration and an input filter (situated after the <-)  
eg: `m1 = tag(cmd) field(cpu,rss) <- joe`  
This measurement named m1 will emit cpu and rss field values tagged with the command name for all processes selected by the joe filter.  



### Example script:

```
[[inputs.procfilter]]
  script = """
    root <- user(0)
    hogs <- top(rss,3)
    apache <- user('apache')
    tomcat <- children(cmd('tomcat'))
/var/run/my.pid')  

    hogs = tag(cmd_pid) field(rss,args) <- hogs
    wl_root = field(cpu,rss) <- pack(root)
    wl_http = field(cpu,rss,swap,process_nb) <- pack(apache,tomcat)
"""
```

## Filters

To simplify and clarify scripts, filters can be nammed.  
eg: `f1 <- all`  
Declares a filter nammed 'f1' that will select all processes (effectively being an alias to the predefined 'all' filter).

The order of declaration is meaningful and you cannot do forward references to yet undeclared filters.

#### Available filters:

* All  
all or all()  
This filter selects all processes present on the server at sampling time.
It is optional as a parameter in expressions requiring at least one input filter.  
eg: `top(cpu,4,all())`, `top(cpu,4,all)` and `top(cpu,4)` are identical.  


* User  
user(number)
Select processes belonging to user with UID {number}.
user('name')
Select processes belonging to user named {name}.  
eg: `root <- user(0)`  
Declares a filter named 'root' containing all processes with UID 0.  
eg: `students <- user('^s[0-9]{8}'r)`  
Declares a filter nammed students containing all processes of all users named s[0-9]{8}  
Please read the section about regular expressions for more information about the 'r' suffix for strings.  


* Group  
group(number)  
Select processes belonging to group with GID {number}.  
group('name')  
Select processes belonging to group named {name}.  
eg: `apache <- group('apache')`  
Declares a filter named 'apache' containing all processes of group 'apache'.   


* PID  
pid(number)  
Select process with that PID.  
pid('file')
Select processes with the PID stored in the {file}.  
eg: `pid('/var/run/my.pid')`    
Select the process with the PID stored in /var/run/my.pid.



* Chilldren  
children(i1[,i2,...][,depth])  
Select all processes descending from one of the processes in {in}. You can specify an optional {depth} to cut the descent in the processes tree.  
eg: `children(pid('/var/run/my.pid'))`  
Select processes in the process tree rooted on PID found in file my.pid.  


* Cmd or command  
cmd('name')  
Select processes whith a basename equal to {name}.  
eg: `cmd('bash')`  
Select all 'bash' processes.   
eg: `cmd('sh$'r)`  
Select all processes with a name ending with 'sh' (bash,ksh,zsh,sh,...)  


* Path  
path('my_path')  
Select processes with a dirname matching {my_path}. This is the basename of the command.  
eg: `path('/opt/oracle/bin')`  
Select all processes with executable files residing in '/opt/oracle/bin'.


* Args  
args('my_arg')  
Select processes with at least one argument matching {my_arg}.  
Note that the my_arg is matched against one argument at a time. Thus you do not have to bother with argument ordering or separators.  
eg: `args('my_SID'r)`  
Select all processes with 'my_SID' in one of their arguments.  


* Cmdline  
cmdline('my_cmdline')  
Select processes with a command line matching 'my_cmdline'.  
The commandline is one string with path exe and arguments concatenated (eg: "/usr/bin/scp -R remote:/ .")  
eg: `cmdline("^/home/joe/crack -all"r)`  
Select all processes with a command line starting with '/home/joe/crack -all'  

* Top  
top(criteria,number,input)  
Select the {number} biggest for {criteria} processes from {input} filter.  
See the Criteria section for a list of known criteria.  
eg: `top(cpu,2,user("joe"))`  
Select the two most cpu consumming processes for the joe user.  


* Exceed  
exceed(criteria,value,input)  
Select processes from {input} filter that exceed {value} for {criteria}.  
eg: `exceed(vsz,5000000,group('tomcat'))`  
Select processes of group 'tomcat' with a virtual memory size greater than 5G.  


* Or, union  
or(i1,i2[,in])  
Select all processes in input filters {i*}. Note that there must be at least two arguments.  
eg:` union(apache,tomcat)`  
eg: `or(user('fu'),user('bar'))`  

* Filters  
filters('name')  
Select content of all filtes matching {name}.  
eg: `filters('http'r)`  
Select processes in all filters named '.*httpd.*'.  



## Set filters

These filters use set algrebra. They implicitly unpack their inputs. You may need to use pack or packby to (re)aggregate their resulting set of processes.  

* And, intersection  
and(i1,i2[,in])  
Select processes present in all input filters {i*}. Note that there must be at least two arguments.  
eg: `and(user('root'),group('root'))`  
Select processes beloinging both to user 'root' and group 'root'.  


* Not, complement  
not(input)  
Select all processes that are not in {input} filter.  
eg: `not(user('root'))`  
Select all processes not belonging to 'root' user.  

* Xor, difference  
xor(i1,i2[,in])  
Select processes present in one and only one of the input filters {i*}. This is the synthetic difference of set of processes {i\*}.  Note that there must be at least two arguments.  



## Aggregation

You can aggregate metrics for a group of processes with pack(). 
This is very handy to define workloads that will output synthetic measurements.  
eg: `wl_httpd = field(cpu) <- pack(user('apache'),user('tomcat'))`  
Will output the measurement wl_httpd with only one field per sample: the sum of all CPU usage for processes belonging to user apache or tomcat.  

  
* Pack
pack(i1[,i2,...])  
All processes selected by fitlers {i*} are aggregated as a single pseudo process. When a measurement uses such a packed set of processes, fields values are the sum of all processes values.  
eg: `m1 = field(cpu) <- pack(user(0))`  
The m1 measurement will contain only one line per sample with the sum of cpu usage for all root processes.  


* Unpack  
unpack(input)  
If {input} contains aggregated processes, unpack them as individual processes.  


* Packby, by  
packby(criteria,i1[,i2,...])  
Pack processes according to {criteria} values (similar to a SQL group by).  
The subset of criteria available for groupby is: user,group,cmd  
eg: `packby(user)`  
Build aggregates of processes by owner (user).  



## Criteria

Top, exceed, packby, ... use a criteria to specify what metric to use.
eg: `top(vsz,1)`  
'vsz' is the virtual memory size of a process and top will select the biggest process according to this criteria.  

* CPU  
CPU usage. Unit is implicity %.  

* RSS  
Resident memory size. In normal (non swap) situation this indicates how much RAM the process uses.  

* VSZ  
Virtual memory size.  

* Swap  
Swap used.  

* FD_nb  
Number of open files (file descriptors).  

* Thread_nb  
Number of threads.  

* Process_nb  
Number of processes. Obviously one unless applied on a pack() filter.  



## Regular expressions

When the 'r' suffix is added to a string the string content becomes a regular expression.  
eg:  
```
cmd("apache") will select processes named "apache".
cmd("apache"r) will select processes with names mathings the regexp "apache". This means that my_apache or apache2 will match.
cmd("^apache$"r) will match only processes named apache (like the simpler "apache" version)
```  
You can invert the match by appending ! to the r.  
eg: `user('^s[0-9]{8}$'r!)`  
Select processes from all users not a student. You could also use not() to the same effect but this syntax is faster at runtime.  

The regular expression syntax used in procfilter is the same as in golang, python or perl. (see: https://github.com/google/re2/wiki/Syntax)



#### Notes on general syntax

String delimiters " and ' can be used indifferently.  
eg: `"joe"` and `'joe'`.  

Identifiers are case insensitive.  
eg: `Top(Apache)` and `top(apache)`  

Identifiers are made of letters, digits '.' or '_'.


## Tags

A measurement may have a tags() declaration.  
This will specify what values will be output as tags for this measurement.  
eg: `m1 = tag(cmd,user) <- cmd('sh$'r)`  
This measurement 'm1' will emit tags with the shell name and user name for all processes with command name ending with 'sh'.  
In line protocole it will look something like:
```m1,cmd="nosh",user="bart"
m1,cmd="bash",user="maggy"
```  
Note that tag and tags are synonyms.  

Identifiers known by the tag instruction:  
user
uid
group
gid
cmd
exe
path
pid
cmd_pid  

Note that pid and cmd_pid could be considered harmfull for your influxdb performance until the cardinality issues are less problematic. (as of influxDB 1.0 this is still an issue)  



## Fields

A measurement may have a fields() declaration.  
This will specify what values will be output as fields for this measurement.  
eg: `m1 = tag(user) field(cpu,rss) <- cmd('eclipse')`  
This measurement 'm1' will emit values with the cpu usage and resident memory size for all 'eclipse' processes.  
In line protocole it will look something like:
```
m1,user="bart" cpu=0.1,rss=1i
m1,user="maggy" cpu=99.9,rss=91823981i
```  
Note that field and fields are synonyms.  

Identifiers known by the field instruction:  
user
uid
group
gid
cmd
exe
path
cmd\_line
pid
cmd\_pid
rss
vsz
thead\_nb
process\_nb
fd\_nb


## More examples

`by.user = tag(user) fields(cpu,rss,vsz,swap,process\_nb,thread\_nb,fd\_nb) <- packby(user)`

This script will generate a measurement named by.user with one line per user with the user name as tag and field values for cpu,vsz,rss,....

`top.cpu.by.user = tag(user) fields(cpu) <- top(cpu,5,packby(user))`

This script will output metrics for the five users consuming the most CPU.

```
tomcat <- user(tomcat)
apache <- user(apache)
root   <- and(user(root),not(cmd(emacs)))
wl_http = fields(cpu,rss,swap) <- pack(tomcat,apache)
wl_root = fields(cpu,rss,swap) <- pack(root)
```  
This script declares three filters to collect all processes from tomcat,apache and root (except emacs).
Then it outputs the measurement wl_http with the cpu,rss and swap fields aggregated for apache and tomcat.
Another measurement will be the cpu,rss and swap aggregagted for all root processes.  

Feel free to send me your own scripts to be used as examples.  


## Grafana

In the github directory for this plugin you will find a complete dashboard ready for import in grafana and the corresponding telegraf configuration (grafana.json, procfilter.conf). Please use and abuse it and report your improvments.  

