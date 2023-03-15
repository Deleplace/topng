package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 2 {
		usage()
	}
	imageURL := flag.Arg(0)
	destPath := flag.Arg(1)
	ctx := context.Background()
	log.Println("Downloading", imageURL)

	_, _, err := DownloadConvertToPNGAndWrite(ctx, imageURL, destPath)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("Written", destPath)
}

func usage() {
	log.Printf("Usage: %s <source image url> <local output filename>\n", os.Args[0])
	os.Exit(1)
}

func ConvertToPNG(ctx context.Context, jpegInput []byte) (pngOutput []byte, err error) {
	if ctx.Err() != nil {
		return nil, errContextCanceled
	}

	r := bytes.NewBuffer(jpegInput)
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	var outBuffer bytes.Buffer
	err = png.Encode(&outBuffer, img)
	return outBuffer.Bytes(), err
}

func ConvertToPNGAndWrite(ctx context.Context, jpegInput []byte, outputPath string) (pngBytes []byte, err error) {
	if ctx.Err() != nil {
		return nil, errContextCanceled
	}

	outBytes, err := ConvertToPNG(ctx, jpegInput)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(outputPath, outBytes, 0777)
	return outBytes, err
}

func DownloadConvertToPNGAndWrite(ctx context.Context, imageURL string, outputPath string) (jpegBytes, pngBytes []byte, err error) {
	if ctx.Err() != nil {
		return nil, nil, errContextCanceled
	}

	code, jpegInput, _, err := download(imageURL)
	if err != nil {
		return nil, nil, fmt.Errorf("downloading %q: %v", imageURL, err)
	}
	if code <= 199 || code >= 300 {
		return nil, nil, fmt.Errorf("downloading %q: unexpected response status %d", imageURL, code)
	}

	pngOutput, err := ConvertToPNGAndWrite(ctx, jpegInput, outputPath)
	return jpegInput, pngOutput, err
}

var errContextCanceled = errors.New("context cancelled")

func download(u string) (statusCode int, payload []byte, d time.Duration, err error) {
	//fmt.Println("Downloading", u)
	start := time.Now()
	resp, err := http.Get(u) // TODO use a Context?
	d = time.Since(start)
	if err != nil {
		return 0, nil, d, err
	}
	code := resp.StatusCode
	if code < 200 || code >= 300 {
		return code, nil, d, nil
	}
	payload, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	return code, payload, d, err
}
