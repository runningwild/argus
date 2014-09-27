package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/runningwild/argus/qtree"
	"github.com/runningwild/argus/rgb"
	"github.com/runningwild/argus/utils"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sort"
	"time"
)

var inputDir = flag.String("dir", "", "Directory to deposit frames into.")
var inputArgus = flag.String("input", "", ".argus file to decode from.")
var cmd = flag.String("cmd", "", "encode or decode")
var cpuprof = flag.String("prof.cpu", "", "write cpu profile to file")
var maxPowerPerPixel = flag.Uint64("ppp", 200, "Maximum power-per-pixel")
var maxFramesPerMoment = flag.Int("fpm", 100, "Maximum frames per moment")
var maxBlocksPerMoment = flag.Int("bpm", 2000, "Maximum blocks per moment")

// File format
// Dims
// Kayframe (jpeg or png)
// for each frame:
// timestamp
// quad-tree representation of changed cells
// for each changed cell:
// a jpeg or png replacement

// copies a blcok from b onto a.  Assumes that a and b are the same dimensions.
func copyBlock(a, b *rgb.Image, x0, y0, x1, y1 int) {
	for y := y0; y < y1; y++ {
		start := a.PixOffset(x0, y)
		end := a.PixOffset(x1, y)
		copy(a.Pix[start:end], b.Pix[start:end])
	}
}

func doDiff(q *qtree.Tree, a, b *rgb.Image) {
	if !q.Bounds().Eq(a.Bounds()) || !a.Bounds().Eq(b.Bounds()) {
		panic("Cannot diff two images with different bounds.")
	}
	// If a cell needs to be replaced, then it should be removed from its parent.
	// If all of a cell's children need to be replaced then we just replace that cell as a whole.

	q.TraverseBottomUp(func(t *qtree.Tree) bool {
		t.Info = qtree.Info{}
		if t.Leaf() {
			power, over := utils.Power(a, b, t.Bounds().Min.X, t.Bounds().Min.Y, *maxPowerPerPixel)
			t.Info.Power += power
			t.Info.Over = over
		} else {
			under := 0
			for i := 0; i < t.NumChildren(); i++ {
				child := t.Child(i)
				if child.Info.Over {
					under++
				}
				if child.Info.Over || child.Info.AboveOver {
					t.Info.AboveOver = true
				}
				t.Info.Power += child.Info.Power
			}
			if under == t.NumChildren() {
				t.Info.Over = true
			}
		}
		if t.Info.Power > t.MaxPower() {
			t.Info.Over = true
		}
		if t.Info.Over {
			t.Info.Power = 0
		}
		return true
	})
}

var endian = binary.LittleEndian

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type subSection struct {
	im     image.Image
	bounds image.Rectangle
	offset image.Point
}

func makeSubSection(im image.Image, region image.Rectangle) *subSection {
	return &subSection{
		im:     im,
		bounds: region.Sub(region.Min),
		offset: region.Min,
	}
}
func (ss *subSection) Bounds() image.Rectangle {
	return ss.bounds
}
func (ss *subSection) At(x, y int) color.Color {
	return ss.im.At(x+ss.offset.X, y+ss.offset.Y)
}
func (ss *subSection) ColorModel() color.Model {
	return ss.im.ColorModel()
}

type selectedBlocks struct {
	blocks *rgb.Image
}

func (sb *selectedBlocks) At(x, y int) color.Color {
	if sb.blocks == nil {
		return color.Black
	}
	return sb.blocks.At(x, y)
}

func (sb *selectedBlocks) Bounds() image.Rectangle {
	if sb.blocks == nil {
		return image.Rect(0, 0, 0, 0)
	}
	return sb.blocks.Bounds()
}

func (sb *selectedBlocks) ColorModel() color.Model {
	return color.RGBAModel
}

func (sb *selectedBlocks) clear() {
	if sb.blocks == nil {
		return
	}
	sb.blocks.Rect = image.Rect(0, 0, 8, 0)
	sb.blocks.Pix = sb.blocks.Pix[0:0]
}

func (sb *selectedBlocks) addBlock(src *rgb.Image, x, y int) {
	if sb.blocks == nil {
		sb.blocks = rgb.Make(image.Rect(0, 0, 8, 0))
	}
	blocky := sb.blocks.Rect.Dy()
	sb.blocks.Rect = image.Rect(0, 0, 8, blocky+8)
	if cap(sb.blocks.Pix) >= len(sb.blocks.Pix)+64*3 {
		sb.blocks.Pix = sb.blocks.Pix[0 : len(sb.blocks.Pix)+64*3]
	} else {
		pix := make([]byte, len(sb.blocks.Pix)*2+64*3)
		copy(pix, sb.blocks.Pix)
		sb.blocks.Pix = pix[0 : len(sb.blocks.Pix)+64*3]
	}
	for i := 0; i < 8; i++ {
		dstOffset := sb.blocks.PixOffset(0, blocky+i)
		srcOffset := src.PixOffset(x, y+i)
		if srcOffset >= len(src.Pix) {
			return
		}
		copy(sb.blocks.Pix[dstOffset:dstOffset+24], src.Pix[srcOffset:])
	}
}

func readImage(r io.Reader) (*rgb.Image, error) {
	var length int32
	check(binary.Read(r, endian, &length))
	if length == 0 {
		return rgb.Make(image.Rect(0, 0, 0, 0)), nil
	}
	buf := make([]byte, int(length))
	_, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	rgba := rgb.Make(im.Bounds())
	draw.Draw(rgba, rgba.Bounds(), im, image.Point{}, draw.Over)
	return rgba, err
}

func writeImage(w io.WriteSeeker, im image.Image) error {
	// Placeholder for the length of the jpeg that we'll write in afterwards.
	err := binary.Write(w, endian, uint32(0))
	if err != nil {
		return err
	}

	// In the event that we tried to encode an empty image, just return, the length was already
	// recorded as zero so we'll notice that there's nothing to do when we decode it.
	if im.Bounds().Dx()*im.Bounds().Dy() == 0 {
		return nil
	}

	// Seek to get the current offset
	startOfImage, err := w.Seek(0, 1)
	if err != nil {
		return err
	}
	err = jpeg.Encode(w, im, nil)
	if err != nil {
		return err
	}
	endOfImage, err := w.Seek(0, 1)
	if err != nil {
		return err
	}
	_, err = w.Seek(startOfImage-4, 0)
	if err != nil {
		return err
	}
	err = binary.Write(w, endian, uint32(endOfImage-startOfImage))
	if err != nil {
		return err
	}
	_, err = w.Seek(endOfImage, 0)
	return err
}

type tintRed struct {
	image.Image
}

func (tr tintRed) At(x, y int) color.Color {
	r, g, b, a := tr.Image.At(x, y).RGBA()
	r = 32565
	return color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(a)}
}

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
	q := qtree.MakeTree(ref.Bounds().Dx(), ref.Bounds().Dy(), *maxPowerPerPixel)
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

type updateImage func(*rgb.Image) error

func encodeDiff(initialFrame *rgb.Image, updater updateImage, w io.WriteSeeker) (err error) {
	defer func() {
		// if r := recover(); r != nil {
		// 	err = fmt.Errorf("Failure: %v", r)
		// }
	}()
	ref := rgb.Make(initialFrame.Bounds())
	cur := rgb.Make(initialFrame.Bounds())
	copy(ref.Pix, initialFrame.Pix)
	copy(cur.Pix, initialFrame.Pix)
	q := qtree.MakeTree(ref.Bounds().Dx(), ref.Bounds().Dy(), *maxPowerPerPixel)
	qbuf := bytes.NewBuffer(nil)
	count := -1
	check(writeImage(w, ref))
	momentStartFrame := 0
	var sb selectedBlocks
	for {
		count++
		err := updater(cur)
		done := err != nil
		if done ||
			count-momentStartFrame > *maxFramesPerMoment ||
			sb.Bounds().Dy()/8 > *maxBlocksPerMoment {
			// Write the number of frames in this moment
			// then the blocks
			// then the qtree
			fmt.Printf("Moment: %d frames, %d blocks.\n", count-momentStartFrame, sb.Bounds().Dy()/8)
			check(binary.Write(w, endian, uint32(count-momentStartFrame)))
			momentStartFrame = count
			check(writeImage(w, &sb))
			sb.clear()
			_, err = io.Copy(w, qbuf)
			check(err)
			qbuf.Truncate(0)
		}
		if done {
			return nil
		}
		doDiff(q, ref, cur)
		q.TraverseTopDown(func(t *qtree.Tree) bool {
			if t.Info.Over {
				check(binary.Write(qbuf, endian, byte(2)))
				return false
			}
			if t.Info.AboveOver {
				check(binary.Write(qbuf, endian, byte(1)))
				return true
			}
			check(binary.Write(qbuf, endian, byte(0)))
			return false
		})
		q.TraverseTopDown(func(t *qtree.Tree) bool {
			if t.Info.Over {
				maxx := t.Bounds().Max.X
				if t.Bounds().Max.X > ref.Bounds().Max.X {
					maxx = ref.Bounds().Max.X
				}
				maxy := t.Bounds().Max.Y
				if t.Bounds().Max.Y > ref.Bounds().Max.Y {
					maxy = ref.Bounds().Max.Y
				}
				copyBlock(ref, cur, t.Bounds().Min.X, t.Bounds().Min.Y, maxx, maxy)
				for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x += 8 {
					for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y += 8 {
						sb.addBlock(cur, x, y)
					}
				}
				return false
			}
			return t.Info.AboveOver
		})
	}
	return
}

func loadUncompressedRGBOnto(filename string, dx, dy int, im *rgb.Image) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Unable to read %q: %v", filename, err)
	}
	if len(data) != im.Bounds().Dx()*im.Bounds().Dy()*3 {
		return fmt.Errorf("Unexpected file length")
	}
	im.Pix = data
	return nil
}

func loadImage(filename string, guessDx, guessDy int) (*rgb.Image, error) {
	raw := rgb.Make(image.Rect(0, 0, guessDx, guessDy))
	err := loadUncompressedRGBOnto(filename, guessDx, guessDy, raw)
	if err == nil {
		return raw, nil
	}

	// The rest of this is just a convenience if we want to run this on jpegs
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return nil, err
	}
	draw.Draw(raw, raw.Bounds(), im, image.Point{}, draw.Over)
	return raw, nil
}

func loadImageFromFilenameOnto(filename string, dst *rgb.Image) error {
	err := loadUncompressedRGBOnto(filename, dst.Bounds().Dx(), dst.Bounds().Dy(), dst)
	if err == nil {
		return nil
	}

	// The rest of this is just a convenience if we want to run this on jpegs
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	im, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return err
	}
	draw.Draw(dst, dst.Bounds(), im, image.Point{}, draw.Over)
	return nil
}

type fileInfo struct {
	name string
	data []byte
}

func consumeFiles(dir string) (<-chan fileInfo, <-chan error) {
	files := make(chan fileInfo)
	errors := make(chan error)
	go func() {
		defer close(files)
		defer close(errors)
		for {
			d, err := os.Open(dir)
			if err != nil {
				errors <- err
				return
			}
			names, err := d.Readdirnames(-1)
			d.Close()
			if err != nil && err != io.EOF {
				errors <- err
				return
			}
			if len(names) <= 1 {
				fmt.Printf("No files available...\n")
				time.Sleep(time.Second)
				continue
			}
			sort.Strings(names)
			filename := filepath.Join(dir, names[0])
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				errors <- err
				return
			}
			if len(data) != 640*480*3 {
				fmt.Printf("Data length was %d, trying jpeg...\n", len(data))
				im, _, err := image.Decode(bytes.NewBuffer(data))
				if err != nil {
					time.Sleep(time.Second)
					continue
				}
				frame := rgb.Make(image.Rect(0, 0, 640, 480))
				draw.Draw(frame, frame.Bounds(), im, image.Point{}, draw.Over)
			}
			files <- fileInfo{name: filename, data: data}
			err = os.Remove(filename)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
	return files, errors
}

func encodeCmd() {
	files, errs := consumeFiles(*inputDir)
	updater := func(frame *rgb.Image) error {
		select {
		case file := <-files:
			fmt.Printf("%s\n", file.name)
			frame.Pix = file.data
		case err := <-errs:
			return err
		}
		return nil
	}
	argus, err := os.Create("diff.argus")
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	var rawFrame *rgb.Image
	select {
	case file := <-files:
		rawFrame = rgb.MakeWithData(image.Rect(0, 0, 640, 480), file.data)
	case err := <-errs:
		fmt.Printf("Failed to load the first image: %v\n", err)
		os.Exit(1)
	}
	err = encodeDiff(rawFrame, updater, argus)
	if err != nil {
		fmt.Printf("Failed to encode argus: %v\n", err)
		os.Exit(1)
	}
	argus.Close()
}

func decodeCmd() {
	f, err := os.Open(*inputArgus)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	count := 0
	update := func(frame *rgb.Image) error {
		count++
		out, err := os.Create(fmt.Sprintf("ref-%05d.jpg", count))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return nil
		}
		jpeg.Encode(out, frame, nil)
		out.Close()
		return nil
	}

	err = decodeDiff(f, update)
	if err != nil {
		fmt.Printf("Error on decoding: %v\n", err)
	}
}

func main() {
	flag.Parse()
	if *cpuprof != "" {
		f, err := os.Create(*cpuprof)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	validCmds := map[string]bool{"encode": true, "decode": true}
	if !validCmds[*cmd] {
		fmt.Printf("Must specify a valid cmd with --cmd: 'encode' or 'decode'.\n")
		os.Exit(1)
	}

	switch *cmd {
	case "encode":
		if *inputDir == "" {
			fmt.Printf("Must specify input directory with --dir.\n")
			os.Exit(1)
		}
		encodeCmd()
	case "decode":
		decodeCmd()
	}

}
