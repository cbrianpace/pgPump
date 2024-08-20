<div>
  <h1 style="font-size: 70px;text-align: center">pgPump</h1>
  <h2 style="text-align: center">Data Export and Import</h2>
</div>
<hr>

[![License](...)](LICENSE)

# Export and Import

pgPump is a golang application designed to provide an efficient and flexible solution for coping data between schemas, databases, and tables.  For use in situations where you are moving data, including:

- **Schema Refresh:**  Exporting and importing data between Postgres environments is typically done with pg_dump/pg_dumpall.  These tools do a fantastic job, but have many challenges when you need to move data between schemas or only copy data to/from certain columns.  pgPump provides flexibility to copy data in parallel without limiting flexibility.

- **Table Recovery:** pgPump can efficiently export data from a recovered database and import it into a different schema/database.

At a higher level, pgPump copies data from the source (list of tables, single table, limited columns) and dumps the data to an external file in a binary or csv format.  On import, the created dump files are loaded back into the database with flexibility of schema and column mappings.  Both the exports and imports can be parallelized to speed up the movement of larger amounts of data.

# Installation

## Requirements

Before initiating the build and installation process, ensure the following prerequisites are met:

1. Go version 1.22 or higher.
2. Postgres version 15 or higher (to use for the pgCompare Data Compare repository).

## Compile
Once the prerequisites are met, begin by forking the repository and cloning it to your host machine:

```shell
YOUR_GITHUB_UN="<your GitHub username>"
git clone --depth 1 "git@github.com:${YOUR_GITHUB_UN}/pgPump.git"
cd pgPump
```

Optionally compile the source:

```shell
go build cmd/pgpump/main.go
```

# Getting Started

## Perform Data Export

To perform the export, simply executed the compiled version of the code or using `go run`:

### Compiled

```shell
pgpump export --user=postgres --password=welcome1 --host=127.0.0.1 --database=mlb --schema=mlb  --dir=/app/temp/data
```

### Using `go run`

```shell
go run cmd/pgpump/main.go export --user=postgres --password=welcome1 --host=127.0.0.1 --database=mlb --schema=mlb  --dir=/app/temp/data
```

In the above example, the minimal required parameters are presented.  If the `--password` is not provided, the application will pull the password from the PGPASSWORD environment variable.

If export is not limited to a specific table (using the `--table` argument), all tables in the specified schema will be exported.  Many tables can be exported in parallel by providing the `--parallel=n` argument where `n` is the desired number of threads to perform the exports.  Additional arguments to select a specific table (`--table=mytable`) and only export specific columns (`--columns=cola,colb`) can be provided to limit what is exproted.

The format of the output defaults to binary.  This can be changed to CSV by specifying the format (`--format=csv`).

## Perform Data Import

To perform the import, simply executed the compiled version of the code or using `go run`:

### Compiled

```shell
pgpump import --user=postgres --password=welcome1 --host=127.0.0.1 --database=mlb --schema=mlb  --dir=/app/temp/data
```

### Using `go run`

```shell
go run cmd/pgpump/main.go import --user=postgres --password=welcome1 --host=127.0.0.1 --database=mlb --schema=mlb  --dir=/app/temp/data
```

In the above example, the minimal required parameters are presented.  If the `--password` is not provided, the application will pull the password from the PGPASSWORD environment variable.  The connection information and the schema should point to the desired target for the imported data.

If export is not limited to a specific table (using the `--table` argument), all dump files in the specified directory (`--dir`) will be imported into the  specified schema.  Many tables can be imported in parallel by providing the `--parallel=n` argument where `n` is the desired number of threads to perform the imports.  Additional arguments to select a specific table (`--table=mytable`) and only export specific columns (`--columns=cola,colb`) can be provided to limit what is exproted.  Note that the columns argument must match the values used in the export.

# Reference

## Usage

pgpump export|import --user=<dbuser> [--password=<dbpassword>] --host=<postgres host> [--port=<postgres port>] --database=<dbname> --schema=<database schema> --dir=<dump directory> [--parallel=n] [--table=<table name>] [--columns="cola,colb"] [--format=binary|csv] [--help] [--version]

**Action**

- export = Export the data to files in the specified directory
- import = Import data in files located in the specified directory. 

**Arguments**

--columns="cola,colb"     List of columns to be exported/imported.  Must be used in conjunction with --table and contain a comman seperated list with no spaces.
--database=<dbname>       Postgres database to connect to.
--dir=<dump directory>    Directory where dump files will be written/read from. Default is '.'.
--format=binary|csv       Format of dump file.  Default is binary.
--help                    Shows the help output.
--host=<postgres host>    Name or IP address of the Postgrest host.  Default is 'localhost'
--parallel=n              Number of processes to use for the import/export.  Default is 1.
--port=<postgres port>    Portgrest port.  Default is 5432
--password=<password>     Postgres password for the specified user.  Can also be passed using PGPASSWORD environment variable.
--schema=<schema>         Schema for tables to be exported/imported.  Default is 'public'
--table=<table name>      Name of a specific table to export/import.
--user=<dbuser>           Postgres database user. Default is 'postgres'
--version                 Displays the version of pgPump.