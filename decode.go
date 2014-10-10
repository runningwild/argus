package main

import (
	"encoding/binary"
	"fmt"
	"github.com/runningwild/argus/qtree"
	"github.com/runningwild/argus/rgb"
	"image"
	"image/draw"
	"io"
)

func decodeDiff(r io.ReadSeeker, updater updateImage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Failure: %v", r)
		}
	}()
	refSrc, err := readImage(r)
	check(err)
	ref := rgb.Make(refSrc.Bounds())
	refDebug := rgb.Make(refSrc.Bounds())
	draw.Draw(ref, ref.Bounds(), refSrc, image.Point{}, draw.Over)
	fmt.Printf("Loaded keyframe: %v\n", ref.Bounds())
	updater(ref)
	var backBufferSize uint8
	binary.Read(r, endian, &backBufferSize)
	q := qtree.MakeTree(ref.Bounds().Dx(), ref.Bounds().Dy(), *maxPowerPerPixel, backBufferSize)
	count := 0
	var momentFramesRemaining int32 = 0
	var momentBlocks *rgb.Image
	momentBlockCount := 0
	// var momentEnd int64 = -1
	for {
		count++
		var offsets []image.Point
		fmt.Printf("Decoding frame %d\n", count)

		if momentFramesRemaining == 0 {
			check(binary.Read(r, endian, &momentFramesRemaining))
			var err error
			momentBlocks, err = readImage(r)
			check(err)
			momentBlockCount = 0
			fmt.Printf("Starting moment: %d frames, %d blocks.\n", momentFramesRemaining, momentBlocks.Bounds().Dy()/8)
		}
		momentFramesRemaining--

		q.TraverseTopDown(func(t *qtree.Tree) bool {
			var b byte
			check(binary.Read(r, endian, &b))
			if b == 0 {
				return false
			}
			if b == 1 {
				return true
			}
			if b != 2 {
				panic(fmt.Sprintf("Got %d, expected 2.", b))
			}
			for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x += 8 {
				for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y += 8 {
					offsets = append(offsets, image.Point{x, y})
				}
			}
			return false
		})
		if len(offsets) > 0 {
			fmt.Printf("Offsets: %v\n", len(offsets))
			for i, offset := range offsets {

				draw.Draw(ref, image.Rect(offset.X, offset.Y, offset.X+8, offset.Y+8), momentBlocks, image.Point{0, (momentBlockCount + i) * 8}, draw.Over)
			}
			draw.Draw(refDebug, refDebug.Bounds(), ref, image.Point{}, draw.Over)
			for i, offset := range offsets {
				draw.Draw(refDebug, image.Rect(offset.X, offset.Y, offset.X+8, offset.Y+8), tintRed{momentBlocks}, image.Point{0, (momentBlockCount + i) * 8}, draw.Over)
			}
			momentBlockCount += len(offsets)
			updater(ref)
		}
	}
	return nil
}
