# go-load

A CLI tool to import SQL files created by [go-dump](https://github.com/ChaosHour/go-dump).

## Installation

```bash
go install github.com/yourusername/go-load/cmd/load@latest
```

## Usage

Using command line flags:
```bash
go-load -host localhost -user root -password secret -port 3306 -file dump.sql
```

Using INI file:
```bash
go-load --ini-file ./my-conf.ini -file dump.sql
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
- `-file`: SQL file to import
- `-ini-file`: Path to INI configuration file
