package append

import (
	"encoding/binary"
	"os"
	"sync/atomic"
)

// the offsets are 32 bit, but usually you want to store more than 4gb of data
// so we just pad things to minimum 64 byte chunks
var PAD = uint32(64)

var MAGIC = []byte{0xb, 0xe, 0xe, 0xf} // change it if you wish

type AppendWriter struct {
	file   *os.File
	offset uint32
}

// Creates new writer and seeks to the end
// The writer is *safe* to be used concurrently, because it uses bump pointer like allocation of the offset.
// example usage:
//
//	w, err := NewAppendWriter(filename)
//	if err != nil {
//		panic(err)
//	}
//
//	docID, err := w.Append([]byte("hello world"))
//	if err != nil {
//		panic(err)
//	}
//
//
//	r, err := NewAppendReader(filename)
//	if err != nil {
//		panic(err)
//	}
//	data, _, err := r.Read(docID)
//	if err != nil {
//		panic(err)
//	}
//	log.Printf("%s",string(data))
//
func NewAppendWriter(filename string) (*AppendWriter, error) {
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	off, err := fd.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}

	return &AppendWriter{
		file:   fd,
		offset: uint32((off + int64(PAD) - 1) / int64(PAD)),
	}, nil
}

func (fw *AppendWriter) Close() error {
	return fw.file.Close()
}

func (fw *AppendWriter) Sync() error {
	return fw.file.Sync()
}

// Append bytes to the end of file
// format is:
//   16 byte header
//   XX variable length data
//
//   header:
//      4 bytes LE len(data) [1] // LE = Little Endian
//      4 bytes LE HASH(data)[2] // go-metro
//      4 bytes MAGIC        [3] // 0xbeef
//      4 bytes LE HASH(1 2 3)   // hash of the first 12 bytes
//   data:
//      ..
//      ..
// Then the blob(header + data) is padded to PAD size using ((uint32(blobSize) + PAD - 1) / PAD).
//
// it returns the addressable offset that you can use ReadFromReader() on
func (fw *AppendWriter) Append(encoded []byte) (uint32, error) {
	blobSize := 16 + len(encoded)
	blob := make([]byte, blobSize)
	copy(blob[16:], encoded)
	binary.LittleEndian.PutUint32(blob[0:], uint32(len(encoded)))
	binary.LittleEndian.PutUint32(blob[4:], uint32(Hash(encoded)))
	copy(blob[8:], MAGIC)
	binary.LittleEndian.PutUint32(blob[12:], uint32(Hash(blob[:12])))

	padded := ((uint32(blobSize) + PAD - 1) / PAD)

	current := atomic.AddUint32(&fw.offset, padded)
	current -= uint32(padded)

	_, err := fw.file.WriteAt(blob, int64(current*PAD))
	if err != nil {
		return 0, err
	}
	return uint32(current), nil
}
