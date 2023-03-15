package main

import (
	"context"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

// BenchmarkReadConvert:
// Read jpeg files from disk, convert them to png in-memory
func BenchmarkReadConvert(b *testing.B) {
	for _, maxproc := range fleets {
		benchname := fmt.Sprintf("maxproc_%d", maxproc)
		b.Run(benchname, func(b *testing.B) {
			prevmaxprocs := runtime.GOMAXPROCS(maxproc)
			defer runtime.GOMAXPROCS(prevmaxprocs)
			var wg sync.WaitGroup
			err := filepath.Walk("./testdata/", func(path string, info fs.FileInfo, err error) error {
				//b.Log(path)
				if err != nil {
					b.Errorf("failure accessing path %q: %v\n", path, err)
					return err
				}
				if ext := strings.ToLower(filepath.Ext(path)); ext != ".jpg" && ext != ".jpeg" {
					return nil // skip
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					b.Logf("visiting JPEG file: %q\n", path)
					input, err := os.ReadFile(path)
					if err != nil {
						b.Error(err)
						return
					}
					output, err := ConvertToPNG(context.Background(), input)
					if err != nil {
						b.Error(err)
						return
					}
					if len(output) == 0 {
						b.Errorf("Empty PNG output when converting %q!", path)
					}
					b.Logf("converted %q\n", path)
				}()
				return nil
			})
			wg.Wait()
			if err != nil {
				b.Errorf("error while processing test files: %v", err)
			}
		})
	}
}

// BenchmarkReadConvertWorkers:
// Read jpeg files from disk, convert them to png in-memory
func BenchmarkReadConvertWorkers(b *testing.B) {
	for _, workers := range fleets {
		benchname := fmt.Sprintf("%d_workers", workers)
		b.Run(benchname, func(b *testing.B) {
			g, _ := errgroup.WithContext(context.Background())
			g.SetLimit(workers)
			err := filepath.Walk("./testdata/", func(path string, info fs.FileInfo, err error) error {
				//b.Log(path)
				if err != nil {
					b.Errorf("failure accessing path %q: %v\n", path, err)
					return err
				}
				if ext := strings.ToLower(filepath.Ext(path)); ext != ".jpg" && ext != ".jpeg" {
					return nil // skip
				}
				g.Go(func() error {
					//b.Logf("visiting JPEG file: %q\n", path)
					input, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					output, err := ConvertToPNG(context.Background(), input)
					if err != nil {
						return err
					}
					if len(output) == 0 {
						return fmt.Errorf("Empty PNG output when converting %q!", path)
					}
					//b.Logf("converted %q\n", path)
					return nil
				})
				return nil
			})
			if err != nil {
				b.Errorf("error walking: %v", err)
			}
			err = g.Wait()
			if err != nil {
				b.Errorf("error while processing test files: %v", err)
			}
		})
	}
}

// BenchmarkReadConvertAndWrite:
// Read jpeg files from disk, convert them to png, write png files
func BenchmarkReadConvertAndWrite(b *testing.B) {
	for _, maxproc := range fleets {
		benchname := fmt.Sprintf("maxproc_%d", maxproc)
		b.Run(benchname, func(b *testing.B) {
			prevmaxprocs := runtime.GOMAXPROCS(maxproc)
			defer runtime.GOMAXPROCS(prevmaxprocs)
			var wg sync.WaitGroup
			err := filepath.Walk("./testdata/", func(path string, info fs.FileInfo, err error) error {
				//b.Log(path)
				if err != nil {
					b.Errorf("failure accessing path %q: %v\n", path, err)
					return err
				}
				if ext := strings.ToLower(filepath.Ext(path)); ext != ".jpg" && ext != ".jpeg" {
					return nil // skip
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					//b.Logf("visiting JPEG file: %q\n", path)
					input, err := os.ReadFile(path)
					if err != nil {
						b.Error(err)
						return
					}
					tmpfile, err := os.CreateTemp("", "ReadConvertAndWrite")
					if err != nil {
						b.Errorf("creating a temp file: %v", err)
						return
					}
					defer os.Remove(tmpfile.Name())
					_, err = ConvertToPNGAndWrite(context.Background(), input, tmpfile.Name())
					if err != nil {
						b.Error(err)
						return
					}
					//b.Logf("converted %q\n", path)

					// TODO check: tmpfile exists on fs, and is not empty
				}()
				return nil
			})
			wg.Wait()
			if err != nil {
				b.Errorf("error while processing test files: %v", err)
			}
		})
	}
}

// BenchmarkReadConvertAndWriteWorkers:
// Read jpeg files from disk, convert them to png, write png files
func BenchmarkReadConvertAndWriteWorkers(b *testing.B) {
	for _, workers := range fleets {
		benchname := fmt.Sprintf("%d_workers", workers)
		b.Run(benchname, func(b *testing.B) {
			g, _ := errgroup.WithContext(context.Background())
			g.SetLimit(workers)
			err := filepath.Walk("./testdata/", func(path string, info fs.FileInfo, err error) error {
				//b.Log(path)
				if err != nil {
					b.Errorf("failure accessing path %q: %v\n", path, err)
					return err
				}
				if ext := strings.ToLower(filepath.Ext(path)); ext != ".jpg" && ext != ".jpeg" {
					return nil // skip
				}
				g.Go(func() error {
					//b.Logf("visiting JPEG file: %q\n", path)
					input, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					tmpfile, err := os.CreateTemp("", "ReadConvertAndWriteWorkers")
					if err != nil {
						return err
					}
					defer os.Remove(tmpfile.Name())
					_, err = ConvertToPNGAndWrite(context.Background(), input, tmpfile.Name())
					if err != nil {
						return err
					}
					//b.Logf("converted %q\n", path)

					// TODO check: tmpfile exists on fs, and is not empty
					return nil
				})
				return nil
			})
			if err != nil {
				b.Errorf("error walking: %v", err)
			}
			err = g.Wait()
			if err != nil {
				b.Errorf("error while processing test files: %v", err)
			}
		})
	}
}

// BenchmarkDownloadConvertAndWriteWorkers:
// Download jpeg files from the internet, convert them to png, write png files
func BenchmarkDownloadConvertAndWriteWorkers(b *testing.B) {
	fleets = []int{28, 32, 48, 64, 96, 128}
	expectedCount, expectedSize := countJpegFiles(b)

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
				/*
					if count != expectedCount {
						// We want the same number of local test files as downloaded test files
						b.Errorf("Expected %d downloads, got %d", expectedCount, count)
					}
					if totalSize != expectedSize {
						// We want the same amount of data in local test files as in downloaded test files
						b.Errorf("Expected %d files having total size %d , got %d", expectedCount, expectedSize, totalSize)
					}
				*/
				_, _ = expectedCount, expectedSize
			}
		})
		//time.Sleep(400 * time.Millisecond)
	}
}

func countJpegFiles(t testing.TB) (n int, totalSize int64) {
	path := "./testdata/"
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			t.Errorf("failure accessing path %q: %v\n", path, err)
			return err
		}
		if ext := strings.ToLower(filepath.Ext(path)); ext != ".jpg" && ext != ".jpeg" {
			return nil // skip
		}
		n++
		totalSize += info.Size()
		return nil
	})
	if err != nil {
		t.Errorf("failure walking path %q: %v\n", path, err)
		return
	}
	return n, totalSize
}
