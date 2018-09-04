Two modes. Create and bulkGet mode

1. To create documents

./perf -set=true 

Default setting will create 2 million documents of 1024 bytes with 10 threads
Following command will create 5 million documents of 512 bytes with 20 threads

./perf -set=true -size=512 -threads=20 -documents=5000000

2. BulkGet performance

./perf 

Default setting will fetch 2 million documents in 10 threads with size of each bulkGet(quantum) 1024
To fetch 5 million documents in 20 threads with a quantum of 2048 

./perf -documents=5000000 -threads=20 -quantum=2048

Other options
--------------

-serverURL. Default http://localhost:9000
-bucketName. Default : default
-poolName. Default: default
