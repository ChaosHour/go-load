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
