#!/usr/bin/env python
import pandas
import pyarrow
import pyarrow.parquet

df = pandas.DataFrame({
    'value': [42],
    'timestamp': ["1710683608143228692"]
})

pyarrow.parquet.write_table(pyarrow.Table.from_pandas(df), "input.parquet")
