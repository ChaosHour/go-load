package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ChaosHour/go-load/config"
	"github.com/ChaosHour/go-load/importer"
)

func main() {
	start := time.Now()
	cfg := config.NewConfig()

	flag.StringVar(&cfg.IniFile, "ini-file", "", "Path to INI configuration file")
	flag.StringVar(&cfg.Host, "host", cfg.Host, "MySQL host")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "MySQL port")
	flag.StringVar(&cfg.User, "user", cfg.User, "MySQL username")
	flag.StringVar(&cfg.Password, "password", cfg.Password, "MySQL password")
	flag.StringVar(&cfg.File, "file", "", "SQL file to import")
	flag.StringVar(&cfg.Directory, "directory", "", "Directory containing SQL files")
	flag.StringVar(&cfg.Pattern, "pattern", cfg.Pattern, "File pattern for SQL files")
	flag.IntVar(&cfg.Workers, "workers", cfg.Workers, "Number of parallel workers")
	flag.IntVar(&cfg.ChunkSize, "chunk-size", cfg.ChunkSize, "Size of query chunks")
	flag.StringVar(&cfg.Database, "database", "", "Database name")
	flag.IntVar(&cfg.ChannelBufferSize, "channel-buffer-size", cfg.ChannelBufferSize, "Size of channel buffer")
	flag.Parse()

	// Load INI file if specified
	if err := cfg.LoadIniFile(); err != nil {
		log.Fatal(err)
	}

	if cfg.File == "" && cfg.Directory == "" {
		fmt.Println("Please specify either -file or -directory flag")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	imp, err := importer.NewImporter(dsn, cfg.Workers, cfg.ChunkSize, cfg.ChannelBufferSize)
	if err != nil {
		log.Fatal(err)
	}
	defer imp.Close()

	fmt.Printf("Starting import with %d workers\n", cfg.Workers)

	if cfg.Directory != "" {
		err = imp.ImportDirectory(cfg.Directory, cfg.Pattern)
	} else {
		err = imp.Import(cfg.File)
	}

	if err != nil {
		log.Fatal(err)
	}

	elapsed := time.Since(start)
	fmt.Printf("\nImport completed successfully. Execution time: %s\n", elapsed)
}
