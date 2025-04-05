import csv

with open("us_datacenters.csv", "r", encoding="utf-8") as f:
    reader = csv.reader(f)
    for line_num, row in enumerate(reader, start=1):
        if len(row) != 3:
            print(f"Line {line_num} has {len(row)} fields: {row}")

