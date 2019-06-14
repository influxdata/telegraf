# Date Processor Plugin

The `date` processor adds the months and years as tags to your data.

Provides the ability to group by months or years. 

A few example usecases include: 
1) consumption data for utilities on per month basis 
2) bandwith capacity per month
3) compare energy production or sales on a yearly or monthly basis 


### Configuration:

```toml
[[processors.date]]
  ##Specify the date tags to add rename operation.
  tagKey = "month"
  dateFormat = "Jan"
```

### Tags:

Tags are applied by this processor. 

### Example processing:

```
- throughput, hostname=example.com lower=10i,upper=1000i,mean=500i 1502489900000000000
+ throughput,host=backend.example.com,month=Mar min=10i,max=1000i,mean=500i 1502489900000000000
```
