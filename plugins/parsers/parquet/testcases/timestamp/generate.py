#!/usr/bin/env python
import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq

df = pandas.DataFrame({
    'value': [1.1, 2.2, 3.3],
    'timestamp': [
        "2024-03-15T14:05:06+00:00", "2024-03-16T14:05:06+00:00",
        "2024-03-17T14:05:06+00:00",
    ]
})

pyarrow.parquet.write_table(pyarrow.Table.from_pandas(df), "input.parquet")
