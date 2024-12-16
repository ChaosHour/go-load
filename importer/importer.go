package importer

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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
}

func NewImporter(dsn string, workers, chunkSize, channelBufferSize int) (*Importer, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return &Importer{
		db:                db,
		workers:           workers,
		chunkSize:         chunkSize,
		channelBufferSize: channelBufferSize,
	}, nil
}

func (i *Importer) Close() error {
	if i.db != nil {
		return i.db.Close()
	}
	return nil
}

type queryJob struct {
	query string
	err   error
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

	queries := strings.Split(string(content), ";")
	queries = filterEmptyQueries(queries)

	bar := progressbar.NewOptions64(
		int64(len(queries)),
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

	jobs := make(chan string, i.channelBufferSize)
	results := make(chan queryJob, i.channelBufferSize)
	var wg sync.WaitGroup

	// Start status reporter
	stopStatus := make(chan bool)
	go i.reportStatus(stopStatus)

	// Start workers
	for w := 0; w < i.workers; w++ {
		wg.Add(1)
		go i.worker(jobs, results, &wg)
	}

	// Send jobs to workers
	go func() {
		for _, query := range queries {
			jobs <- query
		}
		close(jobs)
	}()

	// Wait for results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	for result := range results {
		if result.err != nil {
			return result.err
		}
		bar.Add(1)
	}

	fmt.Println() // Add newline after progress bar
	stopStatus <- true
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

	// Process schema files first
	for _, file := range files {
		if !file.IsSchema {
			continue
		}
		if err := i.Import(file.Path); err != nil {
			return err
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

func (i *Importer) worker(jobs <-chan string, results chan<- queryJob, wg *sync.WaitGroup) {
	defer wg.Done()
	for query := range jobs {
		_, err := i.db.Exec(query)
		results <- queryJob{query: query, err: err}
	}
}

func filterEmptyQueries(queries []string) []string {
	filtered := make([]string, 0, len(queries))
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query != "" {
			filtered = append(filtered, query)
		}
	}
	return filtered
}

// Simplify reportStatus to just handle cancellation
func (i *Importer) reportStatus(stop chan bool) {
	<-stop
}
