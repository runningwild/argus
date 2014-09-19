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

const maxPowerPerPixel = 15.0

func maxPowerForRegion(dx, dy int) float64 {
	region := float64(dx * dy)
	return maxPowerPerPixel * (math.Pow(region, 1.8))
}

// copies b onto a
func copyBlock(a, b *image.YCbCr, x0, y0, x1, y1 int) {
	// fmt.Printf("Bounds: %v %v\n", a.Bounds(), b.Bounds())
	// fmt.Printf("Bounds: %d %d %d %d", x0, y0, x1, y1)
	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			yoff := a.YOffset(x, y)
			coff := a.COffset(x, y)
			a.Y[yoff] = b.Y[yoff]
			a.Cb[coff] = b.Cb[coff]
			a.Cr[coff] = b.Cr[coff]
		}
	}
}

func power(a, b *image.YCbCr, x, y int) float64 {
	// width := a.YStride / a.CStride
	// var power float64
	var p float64
	for j := 0; j < 8; j++ {
		for i := 0; i < 8; i++ {
			yoff := a.YOffset(x+i, y+j)
			coff := a.COffset(x+i, y+j)
			ydiff := float64(a.Y[yoff]) - float64(b.Y[yoff])
			cbdiff := float64(a.Cb[coff]) - float64(b.Cb[coff])
			crdiff := float64(a.Cr[coff]) - float64(b.Cr[coff])
			p += ydiff*ydiff + cbdiff*cbdiff + crdiff*crdiff
		}
	}
	// ar, ag, ab, _ := a.RGBA()
	// br, bg, bb, _ := b.RGBA()
	// ar = ar >> 8
	// ag = ag >> 8
	// ab = ab >> 8
	// br = br >> 8
	// bg = bg >> 8
	// bb = bb >> 8
	// p += float64((ar - br) * (ar - br))
	// p += float64((ag - bg) * (ag - bg))
	// p += float64((ab - bb) * (ab - bb))
	return p
}

func doDiff(q *qtree.Tree, a, b *image.YCbCr) {
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

func readImage(r io.Reader) (*image.YCbCr, error) {
	var length int32
	check(binary.Read(r, endian, &length))
	buf := make([]byte, int(length))
	_, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(bytes.NewBuffer(buf))
	raw, ok := im.(*image.YCbCr)
	if !ok {
		return nil, fmt.Errorf("Unexpected type: %T", err)
	}
	return raw, err
}

func writeImage(w io.Writer, im image.Image) error {
	buf := bytes.NewBuffer(nil)
	err := jpeg.Encode(buf, im, nil)
	if err != nil {
		return err
	}
	err = binary.Write(w, endian, int32(buf.Len()))
	if err != nil {
		return err
	}
	_, err = io.Copy(w, buf)
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
		// if r := recover(); r != nil {
		// 	err = fmt.Errorf("Failure: %v", r)
		// }
	}()
	refSrc, err := readImage(r)
	check(err)
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
func encodeDiff(ims <-chan *image.YCbCr, w *bytes.Buffer) (err error) {
	defer func() {
		// if r := recover(); r != nil {
		// 	err = fmt.Errorf("Failure: %v", r)
		// }
	}()
	var q *qtree.Tree
	var ref *image.YCbCr
	// count := 0
	im := <-ims
	count := 0
	check(writeImage(w, im))
	for im := range ims {
		count++
		fmt.Printf("%d: %d\n", count, w.Len())
		if ref == nil {
			ref = image.NewYCbCr(im.Bounds(), im.SubsampleRatio)
			copy(ref.Y, im.Y)
			copy(ref.Cb, im.Cb)
			copy(ref.Cr, im.Cr)
			q = qtree.MakeTree(im.Bounds().Dx(), im.Bounds().Dy())
			continue
		}
		doDiff(q, ref, im)
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
		sb := selectedBlocks{im: im}
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
				copyBlock(ref, im, t.Bounds().Min.X, t.Bounds().Min.Y, maxx, maxy)
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
			fmt.Printf("Write blocks: %v\n", sb.Bounds())
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

		ims := make(chan *image.YCbCr)
		go func() {
			for _, filename := range inputFilenames {
				data, err := ioutil.ReadFile(filename)
				if err != nil {
					fmt.Printf("Unable to read %q: %v\n", filename, err)
					os.Exit(1)
				}
				im, _, err := image.Decode(bytes.NewBuffer(data))
				if err != nil {
					fmt.Printf("Unable to decode %q: %v\n", filename, err)
					os.Exit(1)
				}
				raw, ok := im.(*image.YCbCr)
				if !ok {
					panic(fmt.Sprintf("Unable to something something %T", im))
				}
				ims <- raw
			}
			close(ims)
		}()
		var buf bytes.Buffer
		err := encodeDiff(ims, &buf)
		if err != nil {
			fmt.Printf("Failed to encode argus: %v\n", err)
			os.Exit(1)
		}
		*inputArgus = "diff.argus"
		argus, err := os.Create(*inputArgus)
		if err != nil {
			fmt.Printf("Failed to create output file: %v\n", err)
			os.Exit(1)
		}
		io.Copy(argus, &buf)
		argus.Close()
	}
	return

	{
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
