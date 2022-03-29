# Quantile Aggregator Plugin

The quantile aggregator plugin aggregates specified quantiles for each numeric field
per metric it sees and emits the quantiles every `period`.

## Configuration

```toml
# Keep the aggregate quantiles of each metric passing through.
[[aggregators.quantile]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Quantiles to output in the range [0,1]
  # quantiles = [0.25, 0.5, 0.75]

  ## Type of aggregation algorithm
  ## Supported are:
  ##  "t-digest" -- approximation using centroids, can cope with large number of samples
  ##  "exact R7" -- exact computation also used by Excel or NumPy (Hyndman & Fan 1996 R7)
  ##  "exact R8" -- exact computation (Hyndman & Fan 1996 R8)
  ## NOTE: Do not use "exact" algorithms with large number of samples
  ##       to not impair performance or memory consumption!
  # algorithm = "t-digest"

  ## Compression for approximation (t-digest). The value needs to be
  ## greater or equal to 1.0. Smaller values will result in more
  ## performance but less accuracy.
  # compression = 100.0
```

## Algorithm types

### t-digest

Proposed by [Dunning & Ertl (2019)][tdigest_paper] this type uses a
special data-structure to cluster data. These clusters are later used
to approximate the requested quantiles. The bounds of the approximation
can be controlled by the `compression` setting where smaller values
result in higher performance but less accuracy.

Due to its incremental nature, this algorithm can handle large
numbers of samples efficiently.  It is recommended for applications
where exact quantile calculation isn't required.

For implementation details see the underlying [golang library][tdigest_lib].

### exact R7 and R8

These algorithms compute quantiles as described in [Hyndman & Fan (1996)][hyndman_fan].
The R7 variant is used in Excel and NumPy.  The R8 variant is recommended
by Hyndman & Fan due to its independence of the underlying sample distribution.

These algorithms save all data for the aggregation `period`.  They require
a lot of memory when used with a large number of series or a
large number of samples. They are slower than the `t-digest`
algorithm and are recommended only to be used with a small number of samples and series.

## Benchmark (linux/amd64)

The benchmark was performed by adding 100 metrics with six numeric
(and two non-numeric) fields to the aggregator and the derive the aggregation
result.

| algorithm  | # quantiles   | avg. runtime  |
| :------------ | -------------:| -------------:|
| t-digest      |            3  |  376372 ns/op |
| exact R7      |            3  | 9782946 ns/op |
| exact R8      |            3  | 9158205 ns/op |
| t-digest      |          100  |  899204 ns/op |
| exact R7      |          100  | 7868816 ns/op |
| exact R8      |          100  | 8099612 ns/op |

## Measurements

Measurement names are passed trough this aggregator.

### Fields

For all numeric fields (int32/64, uint32/64 and float32/64) new *quantile*
fields are aggregated in the form `<fieldname>_<quantile*100>`. Other field
types (e.g. boolean, string) are ignored and dropped from the output.

For example passing in the following metric as *input*:

- somemetric
  - average_response_ms (float64)
  - minimum_response_ms (float64)
  - maximum_response_ms (float64)
  - status (string)
  - ok (boolean)

and the default setting for `quantiles` you get the following *output*

- somemetric
  - average_response_ms_025 (float64)
  - average_response_ms_050 (float64)
  - average_response_ms_075 (float64)
  - minimum_response_ms_025 (float64)
  - minimum_response_ms_050 (float64)
  - minimum_response_ms_075 (float64)
  - maximum_response_ms_025 (float64)
  - maximum_response_ms_050 (float64)
  - maximum_response_ms_075 (float64)

The `status` and `ok` fields are dropped because they are not numeric.  Note that the
number of resulting fields scales with the number of `quantiles` specified.

### Tags

Tags are passed through to the output by this aggregator.

### Example Output

```text
cpu,cpu=cpu-total,host=Hugin usage_user=10.814851731872487,usage_system=2.1679541490155687,usage_irq=1.046598554697342,usage_steal=0,usage_guest_nice=0,usage_idle=85.79616247197244,usage_nice=0,usage_iowait=0,usage_softirq=0.1744330924495688,usage_guest=0 1608288360000000000
cpu,cpu=cpu-total,host=Hugin usage_guest=0,usage_system=2.1601016518428664,usage_iowait=0.02541296060990694,usage_irq=1.0165184243964942,usage_softirq=0.1778907242693666,usage_steal=0,usage_guest_nice=0,usage_user=9.275730622616953,usage_idle=87.34434561626493,usage_nice=0 1608288370000000000
cpu,cpu=cpu-total,host=Hugin usage_idle=85.78199052131747,usage_nice=0,usage_irq=1.0476428036915637,usage_guest=0,usage_guest_nice=0,usage_system=1.995510102269591,usage_iowait=0,usage_softirq=0.1995510102269662,usage_steal=0,usage_user=10.975305562484735 1608288380000000000
cpu,cpu=cpu-total,host=Hugin usage_guest_nice_075=0,usage_user_050=10.814851731872487,usage_guest_075=0,usage_steal_025=0,usage_irq_025=1.031558489546918,usage_irq_075=1.0471206791944527,usage_iowait_025=0,usage_guest_050=0,usage_guest_nice_050=0,usage_nice_075=0,usage_iowait_050=0,usage_system_050=2.1601016518428664,usage_irq_050=1.046598554697342,usage_guest_nice_025=0,usage_idle_050=85.79616247197244,usage_softirq_075=0.1887208672481664,usage_steal_075=0,usage_system_025=2.0778058770562287,usage_system_075=2.1640279004292173,usage_softirq_050=0.1778907242693666,usage_nice_050=0,usage_iowait_075=0.01270648030495347,usage_user_075=10.895078647178611,usage_nice_025=0,usage_steal_050=0,usage_user_025=10.04529117724472,usage_idle_025=85.78907649664495,usage_idle_075=86.57025404411868,usage_softirq_025=0.1761619083594677,usage_guest_025=0 1608288390000000000
```

## References

- Dunning & Ertl: "Computing Extremely Accurate Quantiles Using t-Digests", arXiv:1902.04023 (2019)  [pdf][tdigest_paper]
- Hyndman & Fan: "Sample Quantiles in Statistical Packages", The American Statistician, vol. 50, pp. 361-365 (1996) [pdf][hyndman_fan]

[tdigest_paper]: https://arxiv.org/abs/1902.04023
[tdigest_lib]:   https://github.com/caio/go-tdigest
[hyndman_fan]:   http://www.maths.usyd.edu.au/u/UG/SM/STAT3022/r/current/Misc/Sample%20Quantiles%20in%20Statistical%20Packages.pdf
