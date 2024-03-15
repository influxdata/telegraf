# Empty

The test file was created via the following:

```python
#!/usr/bin/env python
import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq

pyarrow.parquet.write_table(pyarrow.Table.from_pandas(pandas.DataFrame()), "input.parquet")
```
