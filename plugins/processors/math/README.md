# Math Processor Plugin

The math processor plugin applies the function to all or selected fields.

### Configuration:

```toml
## Example config that processe all fields of the metric.
# [[processor.math]]
#   ## Math function
# 	function = "abs"
#   ## The name of metric.
#   measurement_name = "cpu"

## Example config that processe only specific fields of the metric.
# [[processor.math]]
#	## Math function
#	function = "abs"
#   ## The name of metric.
#   measurement_name = "diskio"
#   ## The concrete fields of metric
#   fields = ["io_time", "read_time", "write_time"]
```

### Measurements & Fields:

- measurement1
    - field1_functionName


### Available functions:

| function in config|  golang       |
| ----------------- |:-------------:|
| "abs"             | math.Abs      |    
| "acos"            | math.Acos     |  
| "acosh"           | math.Acosh    |   
| "asin"            | math.Asin     |  
| "asinh"           | math.Asinh    |   
| "atan"            | math.Atan     |  
| "atanh"           | math.Atanh    |   
| "cbrt"            | math.Cbrt     |  
| "ceil"            | math.Ceil     |  
| "cos"             | math.Cos      | 
| "cosh"            | math.Cosh     |  
| "erf"             | math.Erf      | 
| "erfc"            | math.Erfc     |  
| "exp"             | math.Exp      | 
| "exp2"            | math.Exp2     |  
| "expm1"           | math.Expm1    |   
| "floor"           | math.Floor    |   
| "gamma"           | math.Gamma    |   
| "j0"              | math.J0       |
| "j1"              | math.J1       |
| "log"             | math.Log      | 
| "log10"           | math.Log10    |   
| "log1p"           | math.Log1p    |   
| "log2"            | math.Log2     |  
| "logb"            | math.Logb     |  
| "sin"             | math.Sin      | 
| "sinh"            | math.Sinh     |  
| "sqrt"            | math.Sqrt     |  
| "tan"             | math.Tan      | 
| "tanh"            | math.Tanh     |  
| "trunc"           | math.Trunc    |   
| "y0"              | math.Y0       |
| "y1"              | math.Y1       |



### Tags:

No tags are applied by this processor.

### Example Output:

```
$ telegraf --config telegraf.conf --quiet
cpu,cpu=cpu4 usage_user=2.429149797571139,usage_user_ceil=3 1508512030000000000


```