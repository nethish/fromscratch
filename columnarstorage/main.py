from table import ColumnarTable

schema = [("id", int, None), ("name", str, "dict"), ("age", int, "rle")]

table = ColumnarTable(schema)
table.insert((1, "Alice", 30))
table.insert((2, "Alice", 30))
table.insert((3, "Bob", 35))
table.insert((4, "Bob", 35))

table.compress_all()
print(
    "Compressed:",
    table.columns["name"]._compressed_data,
    table.columns["age"]._compressed_data,
)

table.decompress_all()
print("Decompressed age column:", table.select_column("age"))


# SELECT name, age WHERE salary > 50000
results = table.query(select=["name", "age"], where=lambda row: row["age"] > 30)
print(results)

# SELECT AVG(age) WHERE name = 'Alice'
avg_age = table.query(where=lambda row: row["name"] == "Alice", agg={"age": "avg"})
print(avg_age)

# SELECT COUNT(*) WHERE age > 30
count = table.query(
    where=lambda row: row["age"] > 30,
    agg={"id": "count"},  # any column works for count
)
print(count)
