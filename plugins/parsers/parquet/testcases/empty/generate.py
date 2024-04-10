#!/usr/bin/env python
import pandas
import pyarrow
import pyarrow.parquet

pyarrow.parquet.write_table(pyarrow.Table.from_pandas(pandas.DataFrame()), "input.parquet")
