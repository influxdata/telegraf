# Goldilocks Cluster Input plugin 

This plugin gathers the statistic data from Goldilocks Cluster server. All metrics are configurable, and you can add/modify/remove metrics for Goldilocks statistics data via config file, without recompiling. 

## Prerequisites 
UnixODBC ( http://www.unixodbc.org ) is required. (above version 2.3.1 ) 

## Configuration of Goldilocks plugin 

The configuration of Goldilocks plugin is consist of two parts. One is for the connection informations to Goldilocks server and the other is for defining metrics. 

### Connection informations 

* goldilocks_odbc_driver_path : path to goldilocks odbc driver location. "?" means GOLDILOCKS_HOME enviroment variables. default("?/lib/libgoldilockscs-ul64.so")
* goldilocks_host  : host address  default ("127.0.0.1")
* goldilocks_port  : port number default ( 22581 )
* goldilocks_user  : user name default( "test" )
* goldilocks_password  : password default("test" )

### Metrics 

Metrics are array of inputs.goldilocks.elements sections. Each section is consist of followings. 

* series_name : series_name for storing data to influxdb 
* sql : sql text 
* tags : tags lists ( should be column name in result set )
* fields : fields lists ( should be column name in result set )
* pivot : if you want to transpose data rows to columns, then true 
* pivot_key : key for pivoting

### How to use

Modifying recreate.sh:54~

like 'rec_goldilocks sys gliese GOLDILOCKS MATCHING DUMMY DUMMY'
Add By node
