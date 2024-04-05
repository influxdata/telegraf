#!/usr/bin/env python
import pandas
import pyarrow
import pyarrow.parquet

df = pandas.DataFrame({
    'tag': ["row1", "row1", "row1", "row1", "row1", "row1", "row1"],
    'float_field': [64.0, 65.0, 66.0, 67.0, 68.0, 69.0, 70.0],
    'str_field': ["a", "b", "c", "d", "e", "f", "g"],
    'timestamp': [
        1710683695, 1710683695, 1710683695, 1710683695, 1710683695,
        1710683695, 1710683695,
    ]
})

pyarrow.parquet.write_table(pyarrow.Table.from_pandas(df), "input.parquet")
