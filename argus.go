package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
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
	var blocks []core.Block8RGB
	var ΔTables []ΔTable
	for i, block := range ref.Blocks() {
		blocks = append(blocks, block)
		ΔTables = append(ΔTables, ΔTable{ΔTableEntry{ms: uint32(now), block: int32(i)}})
	}
	return &epochWriter{
		f:         f,
		ref:       ref,
		maxPowerΔ: 10000,
		blockDx:   int32((ref.Bounds().Dx() + 7) / 8),
		blockDy:   int32((ref.Bounds().Dy() + 7) / 8),
		start:     now,
		end:       now,

		blocks:  blocks,
		ΔTables: ΔTables,
	}
}

func (e *epochWriter) scanPositions(bufferΔ, bufferBlock int) starts {
	var sizeHeader int64 = 32
	sizeΔTableIndexes := int64(12 * (len(e.ΔTables)))
	var sizeΔTables int64
	for _, ΔTable := range e.ΔTables {
		sizeΔTables += int64(8 * (len(ΔTable) + bufferΔ))
	}
	return starts{
		header:        0,
		ΔTableIndexes: sizeHeader,
		ΔTables:       sizeHeader + sizeΔTableIndexes,
		blocks:        sizeHeader + sizeΔTableIndexes + sizeΔTables,
	}
}

type starts struct {
	header        int64
	ΔTableIndexes int64
	ΔTables       int64
	blocks        int64
}

func (e *epochWriter) writeFile(f *os.File, bufferΔ, bufferBlock int) error {
	start := e.scanPositions(bufferΔ, bufferBlock)

	// Header - 32 bytes
	binary.Write(f, binary.LittleEndian, e.start)
	binary.Write(f, binary.LittleEndian, e.end)
	binary.Write(f, binary.LittleEndian, e.blockDx)
	binary.Write(f, binary.LittleEndian, e.blockDy)
	binary.Write(f, binary.LittleEndian, start.blocks)
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
			return err
		}
		for j := range e.ΔTables[i] {
			binary.Write(f, binary.LittleEndian, e.ΔTables[i][j].ms)
			binary.Write(f, binary.LittleEndian, e.ΔTables[i][j].block)
		}
	}
	{
		co, _ := f.Seek(0, io.SeekCurrent)
		if co+int64(bufferΔ*8) != start.blocks {
			panic(fmt.Sprintf("unexpected file offset, %d != %d", co, start.blocks))
		}
	}

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
		return err
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

	return nil
}

func (e *epochWriter) applyImage(im *rgb.Image, t time.Time, bufferΔ, bufferBlock int) error {
	// The following things all need to be updated:
	// End time
	// Indexes for every block that changed
	// Deltas for every block that changed
	// Blocks appended for all new blocks
	refBlocks := e.ref.Blocks()
	imBlocks := im.Blocks()

	// Update the end time
	e.end = t.UnixNano() / 1e6
	e.f.Seek(8, io.SeekStart)
	binary.Write(e.f, binary.LittleEndian, e.end)

	// Update the reference image and note each block that changed.
	var changed []int
	for i := range refBlocks {
		if core.Power(&refBlocks[i], &imBlocks[i]) > e.maxPowerΔ {
			fmt.Printf("Block %d changed \n", i)
			refBlocks[i] = imBlocks[i]
			changed = append(changed, i)
		}
	}

	start := e.scanPositions(bufferΔ, bufferBlock)

	// For each block that changed, update the index, the Δ table, and the blocks.
	for _, b := range changed {
		// Seek to this table's entry in the index, we'll need to read the offset into the Δ table
		// from here because we don't store that in memory (TODO: although we could - does it matter?).
		e.f.Seek(start.ΔTableIndexes+int64(12*b), io.SeekStart)
		var offset int64
		binary.Read(e.f, binary.LittleEndian, &offset)

		// Increment the length of deltas for this block in the index.
		entry := ΔTableEntry{ms: uint32(e.end - e.start), block: int32(len(e.blocks))}
		e.ΔTables[b] = append(e.ΔTables[b], entry)
		binary.Write(e.f, binary.LittleEndian, int32(len(e.ΔTables[b])))

		// Add this entry to the Δ table.
		e.f.Seek(offset+int64(8*len(e.blocks)), io.SeekStart)
		binary.Write(e.f, binary.LittleEndian, entry.ms)
		binary.Write(e.f, binary.LittleEndian, entry.block)

		// Seek to the last block previously recorded and find the offset and length, this will tell
		// us where we can put the next block.  We are guaranteed that this isn't the first block
		// because the initial write of the file must include every block.
		e.f.Seek(start.blocks+int64(len(e.blocks)-1), io.SeekStart)
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

	return nil
}

func main() {
	f, err := os.Create("foo")
	if err != nil {
		panic(err)
	}
	ref := rgb.Make(image.Rect(0, 0, 80, 80))
	im := rgb.Make(image.Rect(0, 0, 80, 80))
	im0 := rgb.Make(image.Rect(0, 0, 80, 80))
	now := time.Now()
	e := newEpoch(now, f, ref)
	e.writeFile(f, 10, 1000)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			im.Set(i, j, color.White)
		}
	}

	now.Add(time.Millisecond)
	e.applyImage(im, now, 10, 1000)

	now.Add(time.Millisecond)
	e.applyImage(im, now, 10, 1000)

	now.Add(time.Millisecond)
	e.applyImage(im, now, 10, 1000)

	now.Add(time.Millisecond)
	e.applyImage(im0, now, 10, 1000)
}

type epochWriter struct {
	f         *os.File
	ref       *rgb.Image
	maxPowerΔ uint64

	start, end       int64    //   16 bytes
	blockDx, blockDy int32    //    8 bytes
	ΔTables          []ΔTable //

	blocks []core.Block8RGB
}

// ΔTable contains all of the Δs for a block.  The entries are ordered by ms since the start of the epoch.
type ΔTable []ΔTableEntry

type ΔTableEntry struct {
	// TODO: This could just be a counter, and we could have a separate table for times
	ms    uint32 // Time in ms since the start of this epoch, good enough for 49 days.
	block int32  // Index into block table
}

type Reader struct {
}
