package main

import (
	"flag"
	"fmt"
	"github.com/runningwild/argus/qtree"
	"github.com/runningwild/argus/rgb"
	"image"
	"os"
	"runtime/pprof"
)

func main() {
	flag.Parse()
	f, err := os.Create("cpu.prof")
	if err != nil {
		fmt.Printf("Failed to open profile: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Setup
	r := rgb.Make(image.Rect(0, 0, 640, 480))
	t := qtree.MakeTree(r.Bounds().Dx(), r.Bounds().Dy(), 0, 10)

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	for i := 0; i < 10000; i++ {
		t.SetToImage(r)
	}
}
