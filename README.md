# go-load

A CLI tool to import SQL files created by [go-dump](https://github.com/ChaosHour/go-dump).

## Installation

```bash
go install github.com/ChaosHour/go-load/cmd/load@latest
```

## Usage

Using command line flags:

```bash
go-load -host 192.168.50.50 -user root -password s3cr3t -directory ./backup4 -database sakila -workers 8
```

Using INI file:

```bash
go-load -ini-file ./my-conf.ini -directory ./backup4 -database sakila -workers 8
```

### INI File Format

```ini
[go-load]
mysql-user = root
mysql-password = s3cr3t
mysql-host = 192.168.50.50
```

### Flags

- `-host`: MySQL host (default: localhost)
- `-port`: MySQL port (default: 3306)
- `-user`: MySQL username
- `-password`: MySQL password
- `-database`: Database name
- `-file`: Single SQL file to import
- `-directory`: Directory containing SQL files
- `-workers`: Number of parallel workers (default: 4)
- `-chunk-size`: Size of query chunks (default: 50000)
- `-channel-buffer-size`: Size of channel buffer (default: 2000)
- `-ini-file`: Path to INI configuration file
- `-pattern`: SQL file pattern (default: "*-thread*.sql")

## Testing

```bash

mysql --defaults-group-suffix=_primary1 -Bse "SELECT table_name, table_rows FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'sakila'" | cat -n
     1 actor 200
     2 actor_info NULL
     3 address 603
     4 category 16
     5 city 600
     6 country 109
     7 customer 599
     8 customer_list NULL
     9 film 1000
    10 film_actor 5462
    11 film_category 1000
    12 film_list NULL
    13 film_text 1000
    14 inventory 4581
    15 language 6
    16 nicer_but_slower_film_list NULL
    17 payment 16086
    18 rental 16005
    19 sales_by_film_category NULL
    20 sales_by_store NULL
    21 staff 2
    22 staff_list NULL
    23 store 2
    24 store_no_pk 2



Drop the customer table:
mysql --defaults-group-suffix=_primary1 -e "set foreign_key_checks=0; drop table if exists sakila.customer; set foreign_key_checks=1"



mysql --defaults-group-suffix=_primary1 -Bse "SELECT table_name, table_rows FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'sakila'" | cat -n
     1 actor 200
     2 actor_info NULL
     3 address 603
     4 category 16
     5 city 600
     6 country 109
     7 customer_list NULL
     8 film 1000
     9 film_actor 5462
    10 film_category 1000
    11 film_list NULL
    12 film_text 1000
    13 inventory 4581
    14 language 6
    15 nicer_but_slower_film_list NULL
    16 payment 16086
    17 rental 16005
    18 sales_by_film_category NULL
    19 sales_by_store NULL
    20 staff 2
    21 staff_list NULL
    22 store 2
    23 store_no_pk 2



Load the customer data into the table:

./go-load ./cmd/load/ --ini-file './my-conf.ini' --directory ../go-dump/backup4 --database sakila --workers 8
Starting import with 8 workers
Found 2 files to process
sakila.customer-definition.sql 100% [==============================]         
Importing data file: sakila.customer-thread7.sql
sakila.customer-thread7.sql 100% [==============================]         

Import completed successfully. Execution time: 66.668335ms


mysql --defaults-group-suffix=_primary1 -Bse "SELECT table_name, table_rows FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'sakila'" | cat -n
     1 actor 200
     2 actor_info NULL
     3 address 603
     4 category 16
     5 city 600
     6 country 109
     7 customer 599
     8 customer_list NULL
     9 film 1000
    10 film_actor 5462
    11 film_category 1000
    12 film_list NULL
    13 film_text 1000
    14 inventory 4581
    15 language 6
    16 nicer_but_slower_film_list NULL
    17 payment 16086
    18 rental 16005
    19 sales_by_film_category NULL
    20 sales_by_store NULL
    21 staff 2
    22 staff_list NULL
    23 store 2
    24 store_no_pk 2
```
