# HTCondor Input Plugin

HTCondor is a specialized workload management system for compute-intensive jobs. Like other full-featured batch systems, HTCondor provides a job queueing mechanism, scheduling policy, priority scheme, resource monitoring, and resource management. Users submit their serial or parallel jobs to HTCondor, HTCondor places them into a queue, chooses when and where to run the jobs based upon a policy, carefully monitors their progress, and ultimately informs the user upon completion.

Reference: https://research.cs.wisc.edu/htcondor/description.html

### Configuration:

```toml
# Gather outputs from condor_q command
[[inputs.htcondor]]
  # no configuration
```

### Example Output:

```
> htcondor,host=127.0.0.1 completed=2i,held=6i,idle=4i,jobs=1i,removed=3i,running=5i,suspended=7i 1555274373000000000
```
