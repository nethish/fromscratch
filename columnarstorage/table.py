from compression import RLECompressor
from compression import DictionaryCompressor


class Column:
    def __init__(self, name, dtype, compression=None):
        self.name = name
        self.dtype = dtype
        self.data = []
        self.compression = compression
        self._compressed_data = None  # For storing compressed form

    def append(self, value):
        if not isinstance(value, self.dtype):
            raise TypeError(f"Value {value} does not match type {self.dtype}")
        self.data.append(value)

    def compress(self):
        if self.compression == "rle":
            self._compressed_data = RLECompressor.compress(self.data)
        elif self.compression == "dict":
            self._compressed_data = DictionaryCompressor.compress(self.data)
        else:
            self._compressed_data = None

    def decompress(self):
        if self.compression == "rle":
            self.data = RLECompressor.decompress(self._compressed_data)
        elif self.compression == "dict":
            compressed, dictionary = self._compressed_data
            self.data = DictionaryCompressor.decompress(compressed, dictionary)

    def __getitem__(self, idx):
        return self.data[idx]


class ColumnarTable:
    def __init__(self, schema):
        """
        schema: list of (name, type, compression)
        """
        self.columns = {}
        for name, dtype, *rest in schema:
            compression = rest[0] if rest else None
            self.columns[name] = Column(name, dtype, compression)
        self.num_rows = 0

    def insert(self, row):
        if len(row) != len(self.columns):
            raise ValueError("Row does not match schema")
        for (name, _), value in zip(self.columns.items(), row):
            self.columns[name].append(value)
        self.num_rows += 1

    def compress_all(self):
        for column in self.columns.values():
            column.compress()

    def decompress_all(self):
        for column in self.columns.values():
            column.decompress()

    def select_column(self, column_name):
        return self.columns[column_name].data

    def select_where(self, column_name, condition_fn):
        col = self.columns[column_name]
        matching_indices = [i for i, val in enumerate(col.data) if condition_fn(val)]

        result = {}
        for name, column in self.columns.items():
            result[name] = [column.data[i] for i in matching_indices]
        return result

    def query(self, select=None, where=None, agg=None):
        """
        select: list of column names to return
        where: function that takes a row dict and returns True/False
        agg: dict of column -> aggregation function name ("sum", "avg", etc.)
        """
        select = select or list(self.columns.keys())
        rows = []

        # Materialize data (for now; later we can stream/compress)
        data = {name: col.data for name, col in self.columns.items()}

        for i in range(self.num_rows):
            row = {name: data[name][i] for name in data}
            if where is None or where(row):
                rows.append(row)

        if agg:
            result = {}
            for col, fn in agg.items():
                values = [row[col] for row in rows]
                if fn == "sum":
                    result[col] = sum(values)
                elif fn == "avg":
                    result[col] = sum(values) / len(values)
                elif fn == "min":
                    result[col] = min(values)
                elif fn == "max":
                    result[col] = max(values)
                elif fn == "count":
                    result[col] = len(values)
                else:
                    raise ValueError(f"Unknown aggregation {fn}")
            return result

        return [{col: row[col] for col in select} for row in rows]
