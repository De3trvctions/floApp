# floapp — NEM12 Meter Readings Importer

## Overview

- Streams NEM12 files and upserts interval consumption into `meter_readings` using Beego v2 ORM and QueryBuilder.
- Reads `NEM.txt` by default from the project root and performs batched, transactional inserts.

## Requirements

- Go 1.23+ (module uses `go 1.23`).
- MySQL reachable with credentials configured in `standard-library/db/db.go`.
- A `meter_readings` table with a unique key on `(nmi, timestamp)`.

## Database Setup

1. Configure DB connection in `standard-library/db/db.go`:
   - `dbUser`, `dbPass`, `dbHost`, `dbPort`, `dbName`.
2. Create table (MySQL):
   ```sql
   CREATE TABLE meter_readings (
     id BIGINT AUTO_INCREMENT PRIMARY KEY,
     nmi VARCHAR(10) NOT NULL,
     timestamp DATETIME NOT NULL,
     consumption DECIMAL(20,6) NOT NULL,
     UNIQUE KEY uniq_nmi_ts (nmi, timestamp)
   );
   ```
3. Ensure the schema matches your needs; unique key is required for upsert.

## Run

- From the project root:
  ```bash
  go run main.go
  ```
- Custom input path:
  ```bash
  go run main.go C:\full\path\to\NEM.txt
  ```
- Logs report counts when the import completes.

## What It Does

- Reads NEM12 records:
  - `200` record: captures `NMI` (field 2) and interval minutes (field 9).
  - `300` record: captures date (field 2) and interval values (fields 3–50).
- Computes timestamps as `date + i * intervalMinutes` starting from midnight.
- Batches rows and executes a single query per batch:
  - `INSERT INTO meter_readings (nmi, timestamp, consumption) VALUES ... ON DUPLICATE KEY UPDATE consumption=VALUES(consumption)`
- Uses a transaction:
  - Rolls back on error, commits when successful.

## Sample Data

- A sample `NEM.txt` is provided in the project root.
- Run with `go run main.go` to import it.

## Customization

- Batch size: change `GeneratorOptions{BatchSize: ...}` in `main.go`.
- DB credentials and database name: edit `standard-library/db/db.go`.
- Table name and field sizes: edit `models/meter_reading.go` and your DB schema.

## Troubleshooting

- “Cannot find path”:
  - Ensure you run from the project root or pass an absolute path to `NEM.txt`.
- DB connection errors:
  - Verify `db.go` credentials, MySQL is running, and database exists.
- Duplicate key errors:
  - Ensure the unique key on `(nmi, timestamp)` is present and correctly defined.

## Project Layout

- `main.go`: CLI entrypoint that initializes DB and runs the import.
- `standard-library/nem12/parser.go`: streaming NEM12 parser with batched upsert via QueryBuilder.
- `models/meter_reading.go`: Beego ORM model for `meter_readings`.
- `standard-library/db/db.go`: MySQL connection and Beego ORM initialization.
