# Columnar Storage
This demonstrates how columnar storage works. Each column is stored separately.
When saved to disk, this allows us to load the one single column alone for maybe aggregation, filtering etc.
This reduces I/O, memory usage and utilize SIMD to process the data in parallel

Whereas in row storage, the entire row has to be retrieved and processed. It's only good for transactional purposes where the row has to be processed as a unit.

We can compress the data using RLE and Dictionary compression since for each column, the data is of same datatype.

## Optimizations
```SQL
SELECT * FROM TAB WHERE a > 2 and b > 3;
```

* Read column `a`
* Read column `b`
* Compute bitmasks for a e.g., `a = [False, True, True, False, ...]`
* Compute bitmasks for b e.g., `b = [True, True, False, True, ...]`
* Compute final mask = a & b
* Filter the rows based on the final masks and only retrieve them


