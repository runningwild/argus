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
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
)

// File format
// Dims
// Kayframe (jpeg or png)
// for each frame:
// timestamp
// quad-tree representation of changed cells
// for each changed cell:
// a jpeg or png replacement

const maxPowerPerPixel = 5.0

func maxPowerForRegion(dx, dy int) float64 {
	region := float64(dx * dy)
	return maxPowerPerPixel * (math.Pow(region, 1.8))
}

func power(a, b color.Color) float64 {
	var power float64
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	ar = ar >> 8
	ag = ag >> 8
	ab = ab >> 8
	br = br >> 8
	bg = bg >> 8
	bb = bb >> 8
	power += float64((ar - br) * (ar - br))
	power += float64((ag - bg) * (ag - bg))
	power += float64((ab - bb) * (ab - bb))
	return power
}

func doDiff(q *qtree.Tree, a, b image.Image) {
	if !q.Bounds().Eq(a.Bounds()) || !a.Bounds().Eq(b.Bounds()) {
		panic("Cannot diff two images with different bounds.")
	}
	// If a cell needs to be replaced, then it should be removed from its parent.
	// If all of a cell's children need to be replaced then we just replace that cell as a whole.

	q.TraverseBottomUp(func(t *qtree.Tree) bool {
		t.Info = qtree.Info{}
		if t.Leaf() {
			for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y++ {
				for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x++ {
					t.Info.Power += power(a.At(x, y), b.At(x, y))
				}
			}
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

type imageInfo struct {
	im   image.Image
	data []byte
}

func encodeDiff(infos <-chan imageInfo, w *bytes.Buffer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Failure: %v", r)
		}
	}()
	var q *qtree.Tree
	var ref draw.Image
	count := 0
	for info := range infos {
		im := info.im
		fmt.Printf("%d\n", w.Len())
		if ref == nil {
			ref = image.NewRGBA(im.Bounds())
			draw.Draw(ref, ref.Bounds(), im, image.Point{}, draw.Over)
			q = qtree.MakeTree(im.Bounds().Dx(), im.Bounds().Dy())
			check(binary.Write(w, endian, int32(q.Bounds().Dx())))
			check(binary.Write(w, endian, int32(q.Bounds().Dy())))
			continue
		}
		start := w.Len()
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
		q.TraverseTopDown(func(t *qtree.Tree) bool {
			if t.Info.Over {
				ss := makeSubSection(im, t.Bounds())
				check(jpeg.Encode(w, ss, nil))
				draw.Draw(ref, ss.Bounds().Add(ss.offset), ss, image.Point{}, draw.Over)
				return false
			}
			return t.Info.AboveOver
		})
		if w.Len()-start > len(info.data) {
			fmt.Printf("Using full image\n")
			w.Truncate(start)
			_, err := io.Copy(w, bytes.NewBuffer(info.data))
			check(err)
		}
		out, _ := os.Create(fmt.Sprintf("ref-%02d.jpg", count))
		count++
		jpeg.Encode(out, ref, nil)
		out.Close()
	}
	return
}

func main() {
	flag.Parse()
	inputFilenames := flag.Args()
	if len(inputFilenames) < 2 {
		fmt.Printf("Must specify at least two files, you specified %v\n", inputFilenames)
		os.Exit(1)
	}
	sort.Strings(inputFilenames)

	ims := make(chan imageInfo)
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
			ims <- imageInfo{im, data}
		}
		close(ims)
	}()
	var buf bytes.Buffer
	err := encodeDiff(ims, &buf)
	if err != nil {
		fmt.Printf("Failed to encode argus: %v\n", err)
		os.Exit(1)
	}
	argus, err := os.Create("diff.argus")
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer argus.Close()
	io.Copy(argus, &buf)
}
