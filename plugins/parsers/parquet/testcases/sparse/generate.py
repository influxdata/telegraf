#!/usr/bin/env python
import pandas
import pyarrow
import pyarrow.parquet

df = pandas.DataFrame({
    'tag': ["row1", "row2", "row3", "row4", "row5", "row6", "multi_field"],
    'float_field': [64.0, 65.0, None, None, None, None, None],
    'int_field': [None, None, 65, None, None, None, None],
    'uint_field': [None, None, None, 5, None, None, None],
    'bool_field': [None, None, None, None, True, None, False],
    'str_field': [None, None, None, None, None, "blargh", "blargh"],
    'timestamp': [
        "2024-03-01T17:10:32", "2024-03-02T17:10:32", "2024-03-03T17:10:32",
        "2024-03-04T17:10:32", "2024-03-05T17:10:32", "2024-03-06T17:10:32",
        "2024-03-07T17:10:32",
    ]
})

pyarrow.parquet.write_table(pyarrow.Table.from_pandas(df), "input.parquet")
