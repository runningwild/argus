package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/runningwild/argus/core"
	"github.com/runningwild/argus/rgb"
)

const BasicFormat = "2006-01-02_15-04-05"

type Writer struct {
	timeFormat   string
	ref          rgb.Image
	maxPowerΔ    uint64
	timeFrontier time.Time
	dx, dy       int
	f            *os.File
}

func NewWriter(dx, dy int, dir, format string) (*Writer, error) {
	// start := time.Now()
	// f, err := os.Create(filepath.Join(dir, start.Format("2006-01-02_15-04-05")))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create epoch file: %v", err)
	// }
	return &Writer{}, nil

}

func (w *Writer) newEpoch() error {
	w.flushEpoch()
	// start := time.Now()
	// f, err := os.Create(filepath.Join(dir, start.Format(w.timeFormat)))
	// if err != nil {
	// return nil, fmt.Errorf("failed to create epoch file: %v", err)
	// }
	return nil
}

func (w *Writer) flushEpoch() error {
	return nil
}

func newEpoch(t time.Time, f *os.File, ref *rgb.Image) *epochWriter {
	now := t.UnixNano() / 1e6
	var blocks []*core.Block8RGB
	var ΔTables []ΔTable
	for i, block := range ref.Blocks() {
		blocks = append(blocks, block)
		ΔTables = append(ΔTables, ΔTable{ΔTableEntry{Ms: 0, Block: int32(i)}})
	}
	return &epochWriter{
		f:         f,
		ref:       ref,
		maxPowerΔ: 30000,
		header: &epochHeader{
			Start:   now,
			End:     now,
			BlockDx: int32((ref.Bounds().Dx() + 7) / 8),
			BlockDy: int32((ref.Bounds().Dy() + 7) / 8),
		},

		blocks:  blocks,
		ΔTables: ΔTables,
	}
}

func (e *epochWriter) scanPositions(bufferΔ, bufferBlock int) starts {
	sizeΔTableIndexes := int64(12 * (len(e.ΔTables)))
	var sizeΔTables int64
	for _, ΔTable := range e.ΔTables {
		sizeΔTables += int64(8 * (len(ΔTable) + bufferΔ))
	}
	return starts{
		header:        0,
		ΔTableIndexes: sizeEpochHeader,
		ΔTables:       sizeEpochHeader + sizeΔTableIndexes,
		blocks:        sizeEpochHeader + sizeΔTableIndexes + sizeΔTables,
	}
}

type starts struct {
	header        int64
	ΔTableIndexes int64
	ΔTables       int64
	blocks        int64
}

func (e *epochWriter) writeFile(f *os.File, bufferΔ, bufferBlock int) (*epochHeader, error) {
	start := e.scanPositions(bufferΔ, bufferBlock)

	// Header - 32 bytes
	header := e.header
	header.BlockOffset = start.blocks
	binary.Write(f, binary.LittleEndian, header)
	{
		co, _ := f.Seek(0, io.SeekCurrent)
		if co != start.ΔTableIndexes {
			panic(fmt.Sprintf("unexpected file offset, %d != %d", co, start.ΔTableIndexes))
		}
	}

	// Δ Table Indexes - N * 12 bytes
	var offsets []int64
	offset := start.ΔTables
	for i := range e.ΔTables {
		binary.Write(f, binary.LittleEndian, offset)
		binary.Write(f, binary.LittleEndian, int32(len(e.ΔTables[i])))
		offsets = append(offsets, offset)
		offset += 8 * int64(len(e.ΔTables[i])+bufferΔ)
	}
	{
		co, _ := f.Seek(0, io.SeekCurrent)
		if co != start.ΔTables {
			panic(fmt.Sprintf("unexpected file offset, %d != %d", co, start.ΔTables))
		}
	}

	// Δ Tables
	for i := range e.ΔTables {
		if _, err := f.Seek(offsets[i], io.SeekStart); err != nil {
			return nil, err
		}
		for j := range e.ΔTables[i] {
			binary.Write(f, binary.LittleEndian, e.ΔTables[i][j])
		}
	}
	{
		co, _ := f.Seek(0, io.SeekCurrent)
		if co+int64(bufferΔ*8) != start.blocks {
			panic(fmt.Sprintf("unexpected file offset, %d != %d", co, start.blocks))
		}
	}
	f.Seek(start.blocks, io.SeekStart)

	// Blocks table

	offset = start.blocks + 8*int64(len(e.blocks)+bufferBlock)
	offsets = offsets[0:0]
	for i := range e.blocks {
		binary.Write(f, binary.LittleEndian, int32(offset))
		binary.Write(f, binary.LittleEndian, int32(len(e.blocks[i])))
		offsets = append(offsets, offset)
		offset += int64(len(e.blocks[i]))
	}

	if _, err := f.Seek(offsets[0], io.SeekStart); err != nil {
		return nil, err
	}
	for i := range e.blocks {
		{
			co, _ := f.Seek(0, io.SeekCurrent)
			if co != offsets[i] {
				panic(fmt.Sprintf("unexpected file offset, %d != %d", co, offsets[i]))
			}
		}
		binary.Write(f, binary.LittleEndian, e.blocks[i])
	}

	return header, nil
}

func (e *epochWriter) applyImage(im *rgb.Image, t time.Time, bufferΔ, bufferBlock int) (int, error) {
	refBlocks := e.ref.Blocks()
	imBlocks := im.Blocks()

	// Update the end time
	e.header.End = t.UnixNano() / 1e6
	e.f.Seek(8, io.SeekStart)
	binary.Write(e.f, binary.LittleEndian, e.header.End)

	// Update the reference image and note each block that changed.
	var changed []int
	for i := range refBlocks {
		if core.Power(refBlocks[i], imBlocks[i]) > e.maxPowerΔ {
			refBlocks[i] = imBlocks[i]
			changed = append(changed, i)
		}
	}

	// For each block that changed, update the index, the Δ table, and the blocks.
	for _, b := range changed {
		// Seek to this table's entry in the index, we'll need to read the offset into the Δ table
		// from here because we don't store that in memory (TODO: although we could - does it matter?).
		e.f.Seek(sizeEpochHeader+int64(12*b), io.SeekStart)
		var offset int64
		binary.Read(e.f, binary.LittleEndian, &offset)

		// Check previous blocks to see if there are any we can reuse.
		blockIndex := int32(len(e.blocks))
		table := e.ΔTables[b]
		for i := len(table) - 1; i >= 0; i-- {
			block := table[i]
			// fmt.Printf("%d %d\n", len(e.blocks), block.Block)
			if power := core.Power(e.blocks[block.Block], refBlocks[b]); power < e.maxPowerΔ {
				fmt.Printf("Reusing block %d in block %d", block.Block, b)
				blockIndex = block.Block
				break
			}
		}

		// Increment the length of deltas for this block in the index.
		entryNum := int32(len(e.ΔTables[b]))
		entry := ΔTableEntry{Ms: uint32(e.header.End - e.header.Start), Block: blockIndex}
		e.ΔTables[b] = append(e.ΔTables[b], entry)
		binary.Write(e.f, binary.LittleEndian, int32(len(e.ΔTables[b])))

		// Add this entry to the Δ table.
		e.f.Seek(offset+int64(8*entryNum), io.SeekStart)
		fmt.Printf("Writing entry %v at %d\n", entry, offset+int64(8*entryNum))
		binary.Write(e.f, binary.LittleEndian, entry)

		if blockIndex == int32(len(e.blocks)) {
			// Seek to the last block previously recorded and find the offset and length, this will tell
			// us where we can put the next block.  We are guaranteed that this isn't the first block
			// because the initial write of the file must include every block.
			e.f.Seek(e.header.BlockOffset+8*int64(len(e.blocks)-1), io.SeekStart)
			var off32 int32
			var length int32
			binary.Read(e.f, binary.LittleEndian, &off32)
			binary.Read(e.f, binary.LittleEndian, &length)
			binary.Write(e.f, binary.LittleEndian, off32+length)
			binary.Write(e.f, binary.LittleEndian, int32(len(refBlocks[b])))

			// Now actually write the block itself.
			e.f.Seek(int64(off32+length), io.SeekStart)
			binary.Write(e.f, binary.LittleEndian, refBlocks[b])

			e.blocks = append(e.blocks, refBlocks[b])
		}
	}

	return len(changed), nil
}

// dumpAll extracts all images and dumps them into a directory.
func dumpAll(f *os.File, target string, blocks map[int]bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	dumpAllPanicy(f, target, blocks)
	return
}

func check(errs ...interface{}) {
	if err := errs[len(errs)-1]; err != nil {
		panic(err)
	}
}

func dumpAllPanicy(f *os.File, target string, blocks map[int]bool) {
	var header epochHeader
	check(binary.Read(f, binary.LittleEndian, &header))
	fmt.Printf("Start/End: %v/%v\n", time.Unix(0, header.Start*1e6), time.Unix(0, header.End*1e6))
	fmt.Printf("Dx/Dy: %d/%d\n", header.BlockDx, header.BlockDy)

	// Build the index.
	var index []ΔTableIndex
	numBlocks := int(header.BlockDx * header.BlockDy)
	for i := 0; i < numBlocks; i++ {
		var x ΔTableIndex
		check(binary.Read(f, binary.LittleEndian, &x))
		index = append(index, x)
	}

	// Get the Δs for each block.  Along the way also track which timestamps are mentioned.
	var tables []ΔTable
	ts := make(map[uint32]bool)
	for _, x := range index {
		fmt.Printf("Index %v\n", x)
		var table ΔTable
		check(f.Seek(x.Offset, io.SeekStart))
		for i := 0; i < int(x.Length); i++ {
			var entry ΔTableEntry
			co, _ := f.Seek(0, io.SeekCurrent)
			binary.Read(f, binary.LittleEndian, &entry)
			fmt.Printf("Read entry %v at %d\n", entry, co)
			table = append(table, entry)
			ts[entry.Ms] = true
		}
		tables = append(tables, table)
	}

	// Get an ordered list of timestamps.
	var tsOrder []uint32
	for t := range ts {
		tsOrder = append(tsOrder, t)
	}
	sort.Sort(tsSlice(tsOrder))
	fmt.Printf("TS: %v\n", tsOrder)

	getBlock := func(n int) (*core.Block8RGB, error) {
		check(f.Seek(header.BlockOffset+int64(n*8), io.SeekStart))
		var offset, length int32
		check(binary.Read(f, binary.LittleEndian, &offset))
		check(binary.Read(f, binary.LittleEndian, &length))
		check(f.Seek(int64(offset), io.SeekStart))
		var b core.Block8RGB
		check(binary.Read(f, binary.LittleEndian, &b))
		return &b, nil
	}

	// Start by creating the reference image.
	im := rgb.Make(image.Rect(0, 0, 8*int(header.BlockDx), 8*int(header.BlockDy)))
	for _, table := range tables {
		b, _ := getBlock(int(table[0].Block))
		im.SetBlock(int(table[0].Block), b)
	}

	out, err := os.Create(filepath.Join(target, "ref.png"))
	if err != nil {
		panic(fmt.Errorf("failed to create output file: %v", err))
	}
	defer out.Close()
	if err := png.Encode(out, im); err != nil {
		panic(fmt.Errorf("failed to write image: %v", err))
	}

	// Generate an image for each timestep.
	changedBlocks := make(map[int]bool)
	for _, t := range tsOrder[1:] {
		fmt.Printf("Timestamp %d\n", t)
		relevant := false
		for b := range tables {
			for len(tables[b]) > 1 && tables[b][1].Ms <= t {
				tables[b] = tables[b][1:]
				changedBlocks[b] = true
				relevant = relevant || (blocks == nil || blocks[b])
			}
		}
		if len(changedBlocks) == 0 || !relevant {
			continue
		}
		for b := range changedBlocks {
			block, err := getBlock(int(tables[b][0].Block))
			if err != nil {
				panic(err)
			}
			im.SetBlock(b, block)
			fmt.Printf("Modified block %d to %d\n", b, tables[b][0])
		}
		out, err := os.Create(filepath.Join(target, fmt.Sprintf("t-%07d.png", t)))
		if err != nil {
			panic(fmt.Errorf("failed to create output file: %v", err))
		}
		if err := png.Encode(out, im); err != nil {
			out.Close()
			panic(fmt.Errorf("failed to write image: %v", err))
		}
		out.Close()
		changedBlocks = make(map[int]bool)
	}
}

type tsSlice []uint32

func (t tsSlice) Len() int           { return len(t) }
func (t tsSlice) Less(i, j int) bool { return t[i] < t[j] }
func (t tsSlice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

const sizeEpochHeader = 32

type epochHeader struct {
	Start, End       int64
	BlockDx, BlockDy int32
	BlockOffset      int64
}

var (
	readFile = flag.Bool("read", false, "whether to read")
)

func main() {
	flag.Parse()
	if *readFile {
		f, err := os.Open("foo")
		if err != nil {
			panic(err)
		}
		blocks := make(map[int]bool)
		for i := 0; i < 40; i++ {
			blocks[i] = true
		}
		blocks = nil
		if err := dumpAll(f, "output", blocks); err != nil {
			panic(err)
		}
		return
	}
	f, err := os.Create("foo")
	if err != nil {
		panic(err)
	}

	frames := make(chan *rgb.Image)
	go func() {
		defer close(frames)
		filepath.Walk("input", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".png") {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					fmt.Printf("failed to load %q: %v", path, err)
					return nil
				}
				frame, _, err := image.Decode(bytes.NewBuffer(data))
				if err != nil {
					fmt.Printf("failed to decode %q: %v", path, err)
					return nil
				}
				im := rgb.Make(frame.Bounds())
				draw.Draw(im, frame.Bounds(), frame, image.Point{}, draw.Over)
				frames <- im
			}
			return nil
		})
	}()

	frame := <-frames
	ref := rgb.Make(frame.Bounds())
	now := time.Now()
	e := newEpoch(now, f, ref)
	h, err := e.writeFile(f, 3000, 1000000)
	if err != nil {
		panic(err)
	}
	e.header = h

	for frame := range frames {
		now = now.Add(time.Millisecond)
		n, _ := e.applyImage(frame, now, 3000, 1000000)
		fmt.Printf("%d blocks changed\n", n)
	}
	{
		comp, err := os.Create("compressed.argus")
		if err != nil {
			panic(err)
		}
		defer comp.Close()
		e.writeFile(comp, 0, 0)
	}
}

type epochWriter struct {
	f         *os.File
	ref       *rgb.Image
	maxPowerΔ uint64

	header  *epochHeader
	ΔTables []ΔTable

	blocks []*core.Block8RGB
}

// ΔTable contains all of the Δs for a block.  The entries are ordered by ms since the start of the epoch.
type ΔTable []ΔTableEntry

type ΔTableIndex struct {
	Offset int64 // Offset from the start of the file of Δ table for this block.
	Length int32 // Number of Δs this block has.
}

type ΔTableEntry struct {
	// TODO: This could just be a counter, and we could have a separate table for times
	Ms    uint32 // Time in ms since the start of this epoch, good enough for 49 days.
	Block int32  // Index into block table
}

type Reader struct {
}
