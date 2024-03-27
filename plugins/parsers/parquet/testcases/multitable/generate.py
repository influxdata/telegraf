#!/usr/bin/env python
import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq

df = pandas.DataFrame({
    'tag': ["row1", "row1", "row1", "row1", "row1", "row1", "row1"],
    'float_field': [64.0, 65.0, 66.0, 67.0, 68.0, 69.0, 70.0],
    'str_field': ["a", "b", "c", "d", "e", "f", "g"],
    'timestamp': [
        1710683608143228692, 1710683608143228692, 1710683608143228692,
        1710683608143228692, 1710683608143228692, 1710683608143228692,
        1710683608143228692
    ]
})

pyarrow.parquet.write_table(pyarrow.Table.from_pandas(df), "input.parquet")
