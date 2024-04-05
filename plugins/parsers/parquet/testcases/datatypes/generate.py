#!/usr/bin/env python
import pandas
import pyarrow
import pyarrow.parquet

df = pandas.DataFrame({
    'boolean': [True, False],
    'string': ["row1", "row2"],
    'int32': [-2147483648, 2147483647],
    'int64': [-9223372036854775808, 9223372036854775807],
    'float32': [1.000000001, 1.123456789],
    'float64': [64.00000000000000001, 65.12345678912121212],
    'byteArray': ["Short", "Much longer string here..."],
    'fixedLengthByteArray': ["STRING", "FOOBAR"],
    'timestamp': [
        "Sun, 17 Mar 2024 10:39:59 MST",
        "Sat, 27 Jun 1987 10:22:04 MST",
    ]
})

schema = pyarrow.schema([
    pyarrow.field('boolean', pyarrow.bool_()),
    pyarrow.field('string', pyarrow.string()),
    pyarrow.field('int32', pyarrow.int32()),
    pyarrow.field('int64', pyarrow.int64()),
    pyarrow.field('float32', pyarrow.float32()),
    pyarrow.field('float64', pyarrow.float64()),
    pyarrow.field('byteArray', pyarrow.binary()),
    pyarrow.field('fixedLengthByteArray', pyarrow.binary(6)),
    pyarrow.field('timestamp', pyarrow.binary())
])

pyarrow.parquet.write_table(pyarrow.Table.from_pandas(df, schema), "input.parquet")
