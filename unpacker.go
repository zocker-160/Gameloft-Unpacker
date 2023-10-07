package main

// Warning: terrible code ahead! Was just hacked together in a few minutes.

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz/lzma"
)


type FileContainer struct {
	Header Header
	Offsets []uint32
	Entries []*Entry
}

type Header struct {
	Unk1 uint16
	Unk2 uint16 // could also be uint32 who the fuck knows
	Unk3 uint16
	NumEntries uint16
}

type Entry struct {
	Magic byte
	Data []byte
}

func parseEntry(f io.Reader, size int) (*Entry, error) {
	var magic byte
	var data = make([]byte, size-1)

	err := binary.Read(f, binary.LittleEndian, &magic)
	if err != nil {
		return nil, fmt.Errorf("failed to read entry magic")
	}
	
	//fmt.Println("magic:", magic, size)

	/*
	if magic != 0x80 {
		return nil, fmt.Errorf("unexpected magic: %v", magic)
	}
	*/

	r, err := lzma.NewReader(f)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, data)
	if err != nil {
		return nil, fmt.Errorf("failed to read entry data")
	}

	return &Entry{
		Magic: magic,
		Data: data,
	}, nil
}

const helpMessage = `
Usage:
    unpacker[.exe] <file>

NOTE: on Windows you can just drag and drop the file onto the .exe
`

func main() {
	fmt.Println("Unpacker by zocker_160")

	if len(os.Args) != 2 {
		fmt.Print(helpMessage)

		fmt.Println("Press any key to exit")
		os.Stdin.Read(make([]byte, 1))
		return
	}

	var input = os.Args[1]

	file, err := os.Open(input)
	if err != nil {
		handleError(err)
	}
	defer file.Close()


	var header = new(Header)
	err = binary.Read(file, binary.LittleEndian, header)
	if err != nil {
		handleError(err)
	}

	fmt.Println("numEntries:", header.NumEntries)

	
	// last entry is the file size, aka the offset of the file end
	var offsets = make([]uint32, header.NumEntries + 1)
	err = binary.Read(file, binary.LittleEndian, offsets)
	if err != nil {
		handleError(err)
	}

	var entries = make([]*Entry, header.NumEntries)
	for i := range entries {
		pos, err := file.Seek(int64(offsets[i]), io.SeekStart)
		if err != nil {
			handleError(err)
		}

		_ = pos
		//fmt.Println("pos:", pos)

		fe, err := parseEntry(file, int(offsets[i+1]-offsets[i]))
		if err != nil {
			handleError(err)
		}

		entries[i] = fe

		drawProgressbar("Decompressing data", i+1, int(header.NumEntries))
	}

	// write files

	var output = strings.ReplaceAll(strings.ReplaceAll(input, "\\", "/"), "/", "") + "_out"
	fmt.Println("output directory:", output)

	err = os.MkdirAll(output, os.FileMode(0775))
	if err != nil {
		handleError(err)
	}

	for i, e := range entries {
		// bad attempt to detect file formats
		var formatMagic [4]byte 
		copy(formatMagic[:], e.Data[:4])

		var ext string

		if formatMagic == [4]byte{'M', 'T', 'h', 'd'} {
			ext = "midi"
		} else if formatMagic == [4]byte{'O', 'g', 'g', 'S'} {
			ext = "ogg"
		} else if formatMagic == [4]byte{'R', 'I', 'F', 'F'} {
			ext = "wav"
		} else {
			ext = "unkn"
		}

		outFile := filepath.Join(output, fmt.Sprintf("file_%d.%s", i+1, ext))
		ofile, err := os.Create(outFile)
		if err != nil {
			handleError(err)
		}

		_, err = ofile.Write(e.Data)
		if err != nil {
			handleError(err)
		}

		ofile.Close()

		drawProgressbar("Saving files", i+1, int(header.NumEntries))
	}

	fmt.Println("unpacking done!")
}

// helper functions

// some helper functions

func handleError(err error) {
	fmt.Println("\nERROR:", err.Error())

	fmt.Println("press any key to exit")
	os.Stdin.Read(make([]byte, 1))
}

func drawProgressbar(descr string, start, end int) {
	progress := int( float32(start) / float32(end) * 100 )
	pHalf := progress / 2

	bar := strings.Repeat("#", pHalf) + strings.Repeat(".", 50-pHalf)

	fmt.Printf("%s: (%v%%) [%s]", descr, progress, bar)
	if progress != 100 {
		fmt.Print("\r")
	} else {
		fmt.Print("\n")
	}

	//time.Sleep(50 * time.Millisecond)
}
