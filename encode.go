package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/runningwild/argus/qtree"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	// "image/png"
	"io"
	"io/ioutil"
	"math"
	"os"
	"runtime/pprof"
	"sort"
	"sync"
)

var inputArgus = flag.String("inargus", "", "If set skip encoding and use this file")
var cpuprof = flag.String("prof.cpu", "", "write cpu profile to file")

// File format
// Dims
// Kayframe (jpeg or png)
// for each frame:
// timestamp
// quad-tree representation of changed cells
// for each changed cell:
// a jpeg or png replacement

const maxPowerPerPixel = 35.0

func maxPowerForRegion(dx, dy int) float64 {
	region := float64(dx * dy)
	return maxPowerPerPixel * (math.Pow(region, 1.8))
}

// copies b onto a
func copyBlock(a, b *image.RGBA, x0, y0, x1, y1 int) {
	for y := y0; y < y1; y++ {
		start := a.PixOffset(x0, y)
		end := a.PixOffset(x1, y)
		copy(a.Pix[start:end], b.Pix[start:end])
	}
}

func power(a, b *image.RGBA, x0, y0 int) float64 {
	var p float64
	for y := y0; y < y0+8; y++ {
		for x := x0; x < x0+8; x++ {
			offset := a.PixOffset(x, y)
			rdiff := float64(a.Pix[offset+0]) - float64(b.Pix[offset+0])
			gdiff := float64(a.Pix[offset+1]) - float64(b.Pix[offset+1])
			bdiff := float64(a.Pix[offset+2]) - float64(b.Pix[offset+2])
			p += rdiff * rdiff
			p += gdiff * gdiff
			p += bdiff * bdiff
		}
	}
	return p
}

func doDiff(q *qtree.Tree, a, b *image.RGBA) {
	if !q.Bounds().Eq(a.Bounds()) || !a.Bounds().Eq(b.Bounds()) {
		panic("Cannot diff two images with different bounds.")
	}
	// If a cell needs to be replaced, then it should be removed from its parent.
	// If all of a cell's children need to be replaced then we just replace that cell as a whole.

	q.TraverseBottomUp(func(t *qtree.Tree) bool {
		t.Info = qtree.Info{}
		if t.Leaf() {
			t.Info.Power += power(a, b, t.Bounds().Min.X, t.Bounds().Min.Y)
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
		if t.Info.Power > maxPowerForRegion(t.Bounds().Dx(), t.Bounds().Dy()) {
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
	im      image.Image
	offsets []image.Point
}

func (sb *selectedBlocks) addBlock(x, y int) {
	sb.offsets = append(sb.offsets, image.Point{x, y})
}
func (sb *selectedBlocks) Bounds() image.Rectangle {
	return image.Rect(0, 0, 8*len(sb.offsets), 8)
}
func (sb *selectedBlocks) At(x, y int) color.Color {
	if x < 0 || x >= len(sb.offsets)*8 {
		return color.Black
	}
	offset := sb.offsets[x/8]
	return sb.im.At(x%8+offset.X, y+offset.Y)
}
func (sb *selectedBlocks) ColorModel() color.Model {
	return sb.im.ColorModel()
}

func loadImageFromFilenameOnto(filename string, dst *image.RGBA) error {
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

func readImage(r io.Reader) (*image.RGBA, error) {
	var length int32
	check(binary.Read(r, endian, &length))
	buf := make([]byte, int(length))
	_, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	rgba := image.NewRGBA(im.Bounds())
	draw.Draw(rgba, rgba.Bounds(), im, image.Point{}, draw.Over)
	return rgba, err
}

func writeImage(w io.WriteSeeker, im image.Image) error {
	// Placeholder for the length of the jpeg that we'll write in afterwards.
	err := binary.Write(w, endian, uint32(0))
	if err != nil {
		return err
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
	fmt.Printf("Write (%v) -> %d\n", im.Bounds(), endOfImage-startOfImage)
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

func decodeDiff(r *bytes.Buffer, frames chan<- image.Image, m *sync.Mutex) (err error) {
	defer close(frames)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Failure: %v", r)
		}
	}()
	refSrc, err := readImage(r)
	check(err)
	fmt.Printf("Decodingnggngn")
	ref := image.NewRGBA(refSrc.Bounds())
	refDebug := image.NewRGBA(refSrc.Bounds())
	draw.Draw(ref, ref.Bounds(), refSrc, image.Point{}, draw.Over)
	fmt.Printf("Loaded keyframe: %v\n", ref.Bounds())
	m.Lock()
	frames <- ref
	m.Lock()
	m.Unlock()
	q := qtree.MakeTree(ref.Bounds().Dx(), ref.Bounds().Dy())
	count := 0
	for {
		count++
		var offsets []image.Point
		fmt.Printf("Decoding frame %d\n", count)
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
			var diff image.Image
			diff, err := readImage(r)
			check(err)
			fmt.Printf("Offsets: %v\n", len(offsets))
			if diff.Bounds().Dx()/8 != len(offsets) {
				panic("balls")
			}
			for i, offset := range offsets {
				draw.Draw(ref, image.Rect(offset.X, offset.Y, offset.X+8, offset.Y+8), diff, image.Point{i * 8, 0}, draw.Over)
			}
			draw.Draw(refDebug, refDebug.Bounds(), ref, image.Point{}, draw.Over)
			for i, offset := range offsets {
				draw.Draw(refDebug, image.Rect(offset.X, offset.Y, offset.X+8, offset.Y+8), tintRed{diff}, image.Point{i * 8, 0}, draw.Over)
			}
		}
		m.Lock()
		frames <- ref
		m.Lock()
		m.Unlock()
	}
	return nil
}

// Format:
// a jpeg image
// then quad-tree representations:
//
type updateImage func(*image.RGBA) error

func encodeDiff(initialFrame *image.RGBA, updater updateImage, w io.WriteSeeker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Failure: %v", r)
		}
	}()
	ref := image.NewRGBA(initialFrame.Bounds())
	cur := image.NewRGBA(initialFrame.Bounds())
	draw.Draw(ref, ref.Bounds(), initialFrame, image.Point{}, draw.Over)
	draw.Draw(cur, cur.Bounds(), initialFrame, image.Point{}, draw.Over)
	q := qtree.MakeTree(ref.Bounds().Dx(), ref.Bounds().Dy())
	count := 0
	check(writeImage(w, ref))
	for {
		err := updater(cur)
		if err != nil {
			return nil
		}
		count++
		n, _ := w.Seek(0, 1)
		fmt.Printf("%d: %d\n", count, n)
		doDiff(q, ref, cur)
		q.TraverseTopDown(func(t *qtree.Tree) bool {
			if t.Info.Over {
				check(binary.Write(w, endian, byte(2)))
				return false
			}
			if t.Info.AboveOver {
				check(binary.Write(w, endian, byte(1)))
				return true
			}
			check(binary.Write(w, endian, byte(0)))
			return false
		})
		sb := selectedBlocks{im: cur}
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
						sb.addBlock(x, y)
						// ss := makeSubSection(im, image.Rect(x, y, x+8, y+8))
						// draw.Draw(ref, ss.Bounds().Add(ss.offset), ss, image.Point{}, draw.Over)
					}
				}
				return false
			}
			return t.Info.AboveOver
		})
		if sb.Bounds().Dx() > 0 {
			writeImage(w, &sb)
		}
		// out, _ := os.Create(fmt.Sprintf("ref-%02d.jpg", count))
		// count++
		// jpeg.Encode(out, ref, nil)
		// out.Close()
	}
	return
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
	if *inputArgus == "" {
		inputFilenames := flag.Args()
		if len(inputFilenames) < 2 {
			fmt.Printf("Must specify at least two files, you specified %v\n", inputFilenames)
			os.Exit(1)
		}
		sort.Strings(inputFilenames)

		*inputArgus = "diff.argus"
		argus, err := os.Create(*inputArgus)
		if err != nil {
			fmt.Printf("Failed to create output file: %v\n", err)
			os.Exit(1)
		}
		f, err := os.Open(inputFilenames[0])
		if err != nil {
			fmt.Printf("Unable to open file %q: %v\n", inputFilenames[0], err)
			os.Exit(1)
		}
		rawFrame, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			fmt.Printf("Unable to decode image %q: %v\n", inputFilenames[0], err)
			os.Exit(1)
		}
		initialImage := image.NewRGBA(rawFrame.Bounds())
		draw.Draw(initialImage, initialImage.Bounds(), rawFrame, image.Point{}, draw.Over)
		updater := func(im *image.RGBA) error {
			inputFilenames = inputFilenames[1:]
			if len(inputFilenames) == 0 {
				return fmt.Errorf("Ran out of images")
			}
			return loadImageFromFilenameOnto(inputFilenames[0], im)
		}
		err = encodeDiff(initialImage, updater, argus)
		if err != nil {
			fmt.Printf("Failed to encode argus: %v\n", err)
			os.Exit(1)
		}
		argus.Close()
	}
	return

	{
		fmt.Printf("Decoding...\n")
		data, err := ioutil.ReadFile(*inputArgus)
		if err != nil {
			panic(err)
		}
		buf := bytes.NewBuffer(data)
		frames := make(chan image.Image)
		var m sync.Mutex
		go func() {
			err := decodeDiff(buf, frames, &m)
			if err != nil {
				fmt.Sprintf("decode: %v", err)
			}
		}()
		count := 0
		for frame := range frames {
			count++
			fmt.Printf("Frame %d\n", count)
			// This is racy
			out, err := os.Create(fmt.Sprintf("ref-%04d.jpg", count))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			jpeg.Encode(out, frame, nil)
			out.Close()
			m.Unlock()
		}
	}
}
