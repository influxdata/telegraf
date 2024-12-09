#!/usr/bin/env python
import pandas
import pyarrow
import pyarrow.parquet

df1 = pandas.DataFrame({
    'tag': ["row1", "row1", "row1", "row1", "row1", "row1", "row1"],
    'float_field': [64.0, 65.0, 66.0, 67.0, 68.0, 69.0, 70.0],
    'timestamp': [
        "1710683608143228692", "1710683608143228692", "1710683608143228692",
        "1710683608143228692", "1710683608143228692", "1710683608143228692",
        "1710683608143228692",
    ]
})

df2 = pandas.DataFrame({
    'tag': ["row1", "row1", "row1", "row1", "row1", "row1", "row1"],
    'float_field': [64.0, 65.0, 66.0, 67.0, 68.0, 69.0, 70.0],
    'timestamp': [
        "1710683608143228693", "1710683608143228693", "1710683608143228693",
        "1710683608143228693", "1710683608143228693", "1710683608143228693",
        "1710683608143228693",
    ]
})

df3 = pandas.DataFrame({
    'tag': ["row1", "row1", "row1", "row1", "row1", "row1", "row1"],
    'float_field': [64.0, 65.0, 66.0, 67.0, 68.0, 69.0, 70.0],
    'timestamp': [
        "1710683608143228694", "1710683608143228694", "1710683608143228694",
        "1710683608143228694", "1710683608143228694", "1710683608143228694",
        "1710683608143228694",
    ]
})

with pyarrow.parquet.ParquetWriter('input.parquet', pyarrow.Table.from_pandas(df1).schema) as writer:
    writer.write_table(pyarrow.Table.from_pandas(df1))
    writer.write_table(pyarrow.Table.from_pandas(df2))
    writer.write_table(pyarrow.Table.from_pandas(df3))
