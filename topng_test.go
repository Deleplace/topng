package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"
)

var (
	//go:embed testdata/urls.txt
	pictureURLsRaw string
	pictureURLs    = strings.Split(pictureURLsRaw, "\n")
)

var fleets = []int{1, 2, 3, 4, 6, 8, 10, 12, 16, 20, 24, 28, 32}

// BenchmarkDownloadConvertAndWriteWorkers:
// Download jpeg files from the internet, convert them to png, write png files
func BenchmarkDownloadConvertAndWriteWorkers(b *testing.B) {
	b.Logf("This machine has %d CPU cores", runtime.NumCPU())

	for _, workers := range fleets {
		benchname := fmt.Sprintf("%d_workers", workers)
		b.Run(benchname, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				count, totalSize := 0, int64(0)
				var muSize sync.Mutex

				g, _ := errgroup.WithContext(context.Background())
				g.SetLimit(workers)
				for _, url := range pictureURLs {
					if url == "" {
						continue // spurious newlines
					}
					url := url
					// artificially slow down the request
					// url += "&speed=100"
					count++
					g.Go(func() error {
						tmpfile, err := os.CreateTemp("", "DownloadConvertAndWriteWorkers")
						if err != nil {
							return err
						}
						jpegData, _, err := DownloadConvertToPNGAndWrite(context.Background(), url, tmpfile.Name())
						muSize.Lock()
						totalSize += int64(len(jpegData))
						muSize.Unlock()
						if err != nil {
							return err
						}
						//b.Logf("converted %q\n", url)

						// TODO check: tmpfile exists on fs, and is not empty
						return nil
					})
				}
				err := g.Wait()
				if err != nil {
					b.Fatalf("error while downloading and processing test files: %v", err)
				}
			}
		})
	}
}
