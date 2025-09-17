package importer

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/schollz/progressbar/v3"
)

type Importer struct {
	db                *sql.DB
	workers           int
	chunkSize         int
	channelBufferSize int
	skipSchema        bool
}

func NewImporter(dsn string, workers, chunkSize, channelBufferSize int, skipSchema bool) (*Importer, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return &Importer{
		db:                db,
		workers:           workers,
		chunkSize:         chunkSize,
		channelBufferSize: channelBufferSize,
		skipSchema:        skipSchema,
	}, nil
}

func (i *Importer) Close() error {
	if i.db != nil {
		return i.db.Close()
	}
	return nil
}

type SQLFile struct {
	Path     string
	IsSchema bool
}

func (i *Importer) Import(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Remove USE statements and chunk comments that may cause issues
	contentStr := string(content)
	useRe := regexp.MustCompile(`USE ` + "`" + `[^` + "`" + `]*` + "`" + `\n`)
	contentStr = useRe.ReplaceAllString(contentStr, "")
	chunkRe := regexp.MustCompile(`-- Chunk \d+ - from \d+ to \d+\n`)
	contentStr = chunkRe.ReplaceAllString(contentStr, "")
	// Normalize line endings to Unix style
	contentStr = strings.ReplaceAll(contentStr, "\r\n", "\n")

	bar := progressbar.NewOptions64(
		int64(1),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription(filepath.Base(filename)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: "-",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	_, err = i.db.Exec(contentStr)
	if err != nil {
		return err
	}
	bar.Add(1)

	fmt.Println() // Add newline after progress bar
	return nil
}

func (i *Importer) ImportDirectory(directory, pattern string) error {
	files, err := i.findSQLFiles(directory, pattern)
	if err != nil {
		return err
	}

	// Sort files to ensure schema files are processed first
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsSchema != files[j].IsSchema {
			return files[i].IsSchema
		}
		return files[i].Path < files[j].Path
	})

	fmt.Printf("Found %d files to process\n", len(files))

	// Process schema files first (if not skipping)
	if !i.skipSchema {
		for _, file := range files {
			if !file.IsSchema {
				continue
			}
			if err := i.Import(file.Path); err != nil {
				return err
			}
		}
	}

	// Process data files in parallel
	var wg sync.WaitGroup
	errors := make(chan error, len(files))
	semaphore := make(chan struct{}, i.workers)

	for _, file := range files {
		if file.IsSchema {
			continue
		}
		wg.Add(1)
		go func(f SQLFile) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("Importing data file: %s\n", filepath.Base(f.Path))
			if err := i.Import(f.Path); err != nil {
				errors <- fmt.Errorf("error importing %s: %v", f.Path, err)
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(errors)
	}()

	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Importer) findSQLFiles(directory, pattern string) ([]SQLFile, error) {
	var files []SQLFile
	entries, err := filepath.Glob(filepath.Join(directory, pattern))
	if err != nil {
		return nil, err
	}

	// Find any definition files regardless of pattern
	definitionFiles, err := filepath.Glob(filepath.Join(directory, "*-definition.sql"))
	if err != nil {
		return nil, err
	}

	// Add definition files
	for _, entry := range definitionFiles {
		files = append(files, SQLFile{
			Path:     entry,
			IsSchema: true,
		})
	}

	// Add data files
	for _, entry := range entries {
		if !strings.Contains(entry, "-definition") {
			files = append(files, SQLFile{
				Path:     entry,
				IsSchema: false,
			})
		}
	}

	return files, nil
}
