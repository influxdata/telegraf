#!/usr/bin/env python
import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq

df = pandas.DataFrame({
    'value': [42],
    'timestamp': ["1710683608143228692"]
})

pyarrow.parquet.write_table(pyarrow.Table.from_pandas(df), "input.parquet")
