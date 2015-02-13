package main

import (
	"flag"
	"fmt"
	"github.com/runningwild/argus/rgb"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func compare(a, b *rgb.Image) {

}

var dir = flag.String("dir", ".", "Directory of images")

type job struct {
	path string
	n    int
	err  error
	im   image.Image
}

func decoderInternal(in <-chan job, out chan<- job) {
	for j := range in {
		f, err := os.Open(j.path)
		if err != nil {
			j.err = err
			out <- j
			continue
		}
		j.im, _, j.err = image.Decode(f)
		f.Close()
		out <- j
	}
}

func startDecoders(n int) (chan<- string, <-chan image.Image) {
	in := make(chan string)
	out := make(chan image.Image)
	jobIn := make(chan job)
	jobOut := make(chan job)

	go func() {
		defer close(jobIn)
		count := 0
		for p := range in {
			jobIn <- job{path: p, n: count}
			count++
		}
	}()
	go func() {
		defer close(out)
		next := 0
		incoming := make(map[int]job)
		for j := range jobOut {
			incoming[j.n] = j
			for j, ok := incoming[next]; ok; j, ok = incoming[next] {
				next++
				if j.err == nil {
					out <- j.im
				}
			}
		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			decoderInternal(jobIn, jobOut)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(jobOut)
	}()
	return in, out
}

func main() {
	flag.Parse()
	in, out := startDecoders(4)
	start := time.Now()
	go func() {
		filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			in <- path
			return nil
		})
		close(in)
	}()
	count := 0
	for _ = range out {
		count++
	}
	end := time.Now()
	fmt.Printf("Processed %d in %v, %v average.\n", count, end.Sub(start), end.Sub(start)/time.Duration(count))
}
