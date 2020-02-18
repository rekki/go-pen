package pen

import (
	"errors"
	"io"
	"os"
)

var EBADSLT = errors.New("checksum mismatch")
var EINVAL = errors.New("invalid argument")

type Reader struct {
	file      *os.File
	blockSize int
}

// Create New AppendReader (you just nice wrapper around ReadFromReader adn ScanFromReader)
// it is *safe* to use it concurrently
// Example usage
//	r, err := NewReader(filename, 4096)
//	if err != nil {
//		panic(err)
//	}
//	// read specific offset
//	data, _, err := r.Read(docID)
//	if err != nil {
//		panic(err)
//	}
//	// scan from specific offset
//	err = r.Scan(0, func(data []byte, offset, next uint32) error {
//		log.Printf("%v",data)
//		return nil
//	})
//
// each Read requires 2 syscalls, one to read the header and one to read the data (since the length of the data is in the header).
// You can reduce that to 1 syscall if your data fits within 1 block, do not set blockSize < 16 because this is the header length.
// blockSize 0 means 16
func NewReader(filename string, blockSize int) (*Reader, error) {
	if blockSize == 0 {
		blockSize = 16
	}
	if blockSize < 16 {
		return nil, EINVAL
	}

	fd, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	return NewReaderFromFile(fd, blockSize)
}

func NewReaderFromFile(fd *os.File, blockSize int) (*Reader, error) {
	if blockSize == 0 {
		blockSize = 16
	}
	if blockSize < 16 {
		return nil, EINVAL
	}

	return &Reader{
		file:      fd,
		blockSize: blockSize,
	}, nil
}

// Scan the open file, if the callback returns error this error is returned as the Scan error. just a wrapper around ScanFromReader.
func (ar *Reader) Scan(offset uint32, cb func([]byte, uint32, uint32) error) error {
	return ScanFromReader(ar.file, offset, ar.blockSize, cb)
}

// Read at specific offset (just wrapper around ReadFromReader), returns the data, next readable offset and error
func (ar *Reader) Read(offset uint32) ([]byte, uint32, error) {
	return ReadFromReader(ar.file, offset, ar.blockSize)
}

func (ar *Reader) Close() error {
	return ar.file.Close()
}

// Reads specific offset. returns data, nextOffset, error. You can
// ReadFromReader(nextOffset) if you want to read the next document, or
// use the Scan() helper
func ReadFromReader(reader io.ReaderAt, offset uint32, blockSize int) ([]byte, uint32, error) {
	b, err := ReadFromReader64(reader, uint64(offset*PAD), blockSize)
	if err != nil {
		return nil, 0, err
	}
	nextOffset := (offset + ((uint32(16+len(b)))+PAD-1)/PAD)
	return b, uint32(nextOffset), nil
}

// Scan ReaderAt, if the callback returns error this error is returned as the Scan error
func ScanFromReader(reader io.ReaderAt, offset uint32, blockSize int, cb func([]byte, uint32, uint32) error) error {
	for {
		data, next, err := ReadFromReader(reader, offset, blockSize)
		if err == io.EOF {
			return nil
		}
		if err == EBADSLT {
			// assume corrupted file, so just skip until we find next valid entry
			offset++
			continue
		}
		if err != nil {
			return err
		}
		err = cb(data, offset, next)
		if err != nil {
			return err
		}
		offset = next
	}
}
