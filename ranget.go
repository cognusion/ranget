package main

import (
	"context"
	"net/http"

	"github.com/cheggaaa/pb/v3"
	"github.com/cognusion/go-humanity"
	"github.com/cognusion/go-rangetripper/v2"
	"github.com/spf13/pflag"

	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func main() {

	var (
		urlString   string
		outFile     string
		chunkString string
		chunkSize   int64
		max         int
		debug       bool
		trash       bool
		timeout     time.Duration

		progressChan chan int64
		doneChan     = make(chan any)
		bar          *pb.ProgressBar
		barTmpl      = `{{ counters . }} {{ bar . }} {{ percent . }} {{ rtime . }} {{ speed . "%s/s"}}`

		timingOut = log.New(io.Discard, "[TIMING]", 0)
		debugOut  = log.New(io.Discard, "[DEBUG] ", 0)
	)

	pflag.StringVar(&urlString, "url", "", "What to fetch")
	pflag.StringVar(&outFile, "out", "./afile", "Where to write it it")
	pflag.StringVar(&chunkString, "size", "10MB", "Size of chunks to download (whole-numbers with suffixes of B,KB,MB,GB,PB)")
	pflag.IntVar(&max, "max", 10, "Maximum number of simultaneous downloaders")
	pflag.BoolVar(&debug, "debug", false, "Enable debugging output (disables progress bar)")
	pflag.BoolVar(&trash, "trash", false, "Delete the file after downloading (e.g. if benchmarking, etc.)")
	pflag.DurationVar(&timeout, "timeout", 1*time.Minute, "Set a general timeout for the download")
	pflag.Parse()

	if urlString == "" {
		fmt.Println("Please at least set --url")
		pflag.Usage()
		os.Exit(1)
	}

	var cerr error
	chunkSize, cerr = humanity.StringAsBytes(chunkString)
	if cerr != nil {
		fmt.Printf("Please use wholenumbers with suffixes of B,KB,MB,GB,PB")
		pflag.Usage()
		os.Exit(1)
	}

	if debug {
		timingOut = log.New(os.Stdout, "[TIMING] ", 0)
		debugOut = log.New(os.Stdout, "[DEBUG] ", 0)
	}

	client := new(http.Client)                                                                // make a new Client
	rtclient := rangetripper.NewRetryClientWithExponentialBackoff(10, 1*time.Second, timeout) // make a new Client
	client.Timeout = timeout

	rt, nerr := rangetripper.NewWithLoggers(10, timingOut, debugOut)
	if nerr != nil {
		panic(nerr)
	}
	rt.SetClient(rtclient)
	rt.SetChunkSize(chunkSize)
	rt.SetMax(max)

	if trash {
		defer os.Remove(outFile) // clean up after ourselves
	}

	if !debug {
		// Not debugging
		progressChan = make(chan int64)

		defer close(doneChan)

		go func(done chan any, progress <-chan int64) {

			contentLength := <-progress // first item is the contentLength
			bar = pb.ProgressBarTemplate(barTmpl).Start64(contentLength)
			bar.Set(pb.Bytes, true)
			defer bar.Finish()

			for {
				select {
				case b := <-progress:
					bar.Add64(b)
				case <-done:
					return
				}
			}
		}(doneChan, progressChan)

	}
	client.Transport = rt // Use the RangeTripper as the Transport

	debugOut.Printf("GETting %s!\n", urlString)
	ctx := rangetripper.WithOutfile(context.Background(), outFile)
	ctx = rangetripper.WithProgressChan(ctx, progressChan)
	if req, err := http.NewRequestWithContext(ctx, "GET", urlString, nil); err == nil {
		if _, derr := client.Do(req); derr != nil {
			panic(derr)
		}
	} else {
		panic(err)
	}

	if !debug {
		// Let the pb drain out
		<-time.After(200 * time.Millisecond)
	}
}
