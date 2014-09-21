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
)

var inputArgus = flag.String("inargus", "", "If set skip encoding and use this file")
var cpuprof = flag.String("prof.cpu", "", "write cpu profile to file")
var maxPowerPerPixel = flag.Float64("ppp", 200.0, "Maximum power-per-pixel")
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

func maxPowerForRegion(dx, dy int) float64 {
	region := float64(dx * dy)
	return *maxPowerPerPixel * (math.Pow(region, 1.4))
}

// copies b onto a
func copyBlock(a, b *image.RGBA, x0, y0, x1, y1 int) {
	for y := y0; y < y1; y++ {
		start := a.PixOffset(x0, y)
		end := a.PixOffset(x1, y)
		copy(a.Pix[start:end], b.Pix[start:end])
	}
}

func power(a, b *image.RGBA, x0, y0 int) (pow float64, over bool) {
	var p float64
	for y := y0; y < y0+8; y++ {
		for x := x0; x < x0+8; x++ {
			offset := a.PixOffset(x, y)
			rdiff := float64(a.Pix[offset+0]) - float64(b.Pix[offset+0])
			gdiff := float64(a.Pix[offset+1]) - float64(b.Pix[offset+1])
			bdiff := float64(a.Pix[offset+2]) - float64(b.Pix[offset+2])
			add := rdiff*rdiff + gdiff*gdiff + bdiff*bdiff
			if add > *maxPowerPerPixel {
				// return add, true
			}
			p += add
		}
	}
	return p, false
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
			power, over := power(a, b, t.Bounds().Min.X, t.Bounds().Min.Y)
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
	blocks *image.RGBA
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

func (sb *selectedBlocks) addBlock(src *image.RGBA, x, y int) {
	if sb.blocks == nil {
		sb.blocks = image.NewRGBA(image.Rect(0, 0, 8, 0))
	}
	blocky := sb.blocks.Rect.Dy()
	sb.blocks.Rect = image.Rect(0, 0, 8, blocky+8)
	if cap(sb.blocks.Pix) >= len(sb.blocks.Pix)+64*4 {
		sb.blocks.Pix = sb.blocks.Pix[0 : len(sb.blocks.Pix)+64*4]
	} else {
		pix := make([]byte, len(sb.blocks.Pix)*2+64*4)
		copy(pix, sb.blocks.Pix)
		sb.blocks.Pix = pix[0 : len(sb.blocks.Pix)+64*4]
	}
	for i := 0; i < 8; i++ {
		dstOffset := sb.blocks.PixOffset(0, blocky+i)
		srcOffset := src.PixOffset(x, y+i)
		if srcOffset >= len(src.Pix) {
			return
		}
		copy(sb.blocks.Pix[dstOffset:dstOffset+32], src.Pix[srcOffset:])
	}
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
	if length == 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0)), nil
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
	updater(ref)
	q := qtree.MakeTree(ref.Bounds().Dx(), ref.Bounds().Dy())
	count := 0
	var momentFramesRemaining int32 = 0
	var momentBlocks *image.RGBA
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
		}
		updater(ref)
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
		// if r := recover(); r != nil {
		// 	err = fmt.Errorf("Failure: %v", r)
		// }
	}()
	ref := image.NewRGBA(initialFrame.Bounds())
	cur := image.NewRGBA(initialFrame.Bounds())
	draw.Draw(ref, ref.Bounds(), initialFrame, image.Point{}, draw.Over)
	draw.Draw(cur, cur.Bounds(), initialFrame, image.Point{}, draw.Over)
	q := qtree.MakeTree(ref.Bounds().Dx(), ref.Bounds().Dy())
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

type rawRGB struct {
	dx, dy int
	data   []byte
}

func (r *rawRGB) At(x, y int) color.Color {
	if !(image.Point{x, y}).In(r.Bounds()) {
		return color.Black
	}
	index := x + y*r.dx
	return color.RGBA{r.data[index], r.data[index+1], r.data[index+2], 255}
}
func (r *rawRGB) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}).In(r.Bounds()) {
		return
	}
	cr, cg, cb, _ := c.RGBA()
	index := (x + y*r.dx) * 3
	r.data[index+0] = (byte)(cr >> 8)
	r.data[index+1] = (byte)(cg >> 8)
	r.data[index+2] = (byte)(cb >> 8)
}
func (r *rawRGB) Bounds() image.Rectangle {
	return image.Rect(0, 0, r.dx, r.dy)
}
func (r *rawRGB) ColorModel() color.Model {
	return color.RGBAModel
}

func loadUncompressedRGB(filename string, dx, dy int) (*rawRGB, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to read %q: %v", filename, err)
	}
	if len(data) != dx*dy*3 {
		return nil, fmt.Errorf("Unexpected file length")
	}
	return &rawRGB{dx: dx, dy: dy, data: data}, nil
}

func loadImage(filename string, guessDx, guessDy int) (*rawRGB, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return loadUncompressedRGB(filename, guessDx, guessDy)
	}
	raw := &rawRGB{
		dx:   im.Bounds().Dx(),
		dy:   im.Bounds().Dy(),
		data: make([]byte, im.Bounds().Dx()*im.Bounds().Dy()*3),
	}
	draw.Draw(raw, raw.Bounds(), im, image.Point{}, draw.Over)
	return raw, nil
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
		rawFrame, err := loadImage(inputFilenames[0], 640, 480)
		if err != nil {
			fmt.Printf("Unable to open file %q: %v\n", inputFilenames[0], err)
			os.Exit(1)
		}
		initialImage := image.NewRGBA(rawFrame.Bounds())
		draw.Draw(initialImage, rawFrame.Bounds(), rawFrame, image.Point{}, draw.Over)
		updater := func(im *image.RGBA) error {
			inputFilenames = inputFilenames[1:]
			if len(inputFilenames) == 0 {
				return fmt.Errorf("Ran out of images")
			}
			rawFrame, err := loadImage(inputFilenames[0], 640, 480)
			if err != nil {
				return err
			}
			for i := 0; i < len(im.Pix)/4; i++ {
				im.Pix[i*4+0] = rawFrame.data[i*3+0]
				im.Pix[i*4+1] = rawFrame.data[i*3+1]
				im.Pix[i*4+2] = rawFrame.data[i*3+2]
			}
			return nil
		}
		err = encodeDiff(initialImage, updater, argus)
		if err != nil {
			fmt.Printf("Failed to encode argus: %v\n", err)
			os.Exit(1)
		}
		argus.Close()
	}

	{
		fmt.Printf("Decoding...\n")
		f, err := os.Open(*inputArgus)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		count := 0
		update := func(frame *image.RGBA) error {
			count++
			out, err := os.Create(fmt.Sprintf("ref-%04d.jpg", count))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return nil
			}
			jpeg.Encode(out, frame, nil)
			return nil
		}

		err = decodeDiff(f, update)
		if err != nil {
			fmt.Printf("Error on decoding: %v\n", err)
		}
	}
}
