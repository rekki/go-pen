package append

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var EBADSLT = errors.New("checksum mismatch")

type AppendReader struct {
	file *os.File
}

// Create New AppendReader (you just nice wrapper around ReadFromReader adn ScanFromReader)
// it is *safe* to use it concurrently
// Example usage
//	r, err := NewAppendReader(filename)
//	if err != nil {
//		panic(err)
//	}
//	// read specific offset
//	data, _, err := r.Read(docID)
//	if err != nil {
//		panic(err)
//	}
//	// scan from specific offset
//	err = r.Scan(0, func(offset uint32, data []byte) error {
//		log.Printf("%v",data)
//		return nil
//	})
//
func NewAppendReader(filename string) (*AppendReader, error) {
	fd, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	return &AppendReader{
		file: fd,
	}, nil
}

// Scan the open file, if the callback returns error this error is returned as the Scan error. just a wrapper around ScanFromReader.
func (ar *AppendReader) Scan(offset uint32, cb func(uint32, []byte) error) error {
	return ScanFromReader(ar.file, offset, cb)
}

// Read at specific offset (just wrapper around ReadFromReader)
func (ar *AppendReader) Read(offset uint32) ([]byte, uint32, error) {
	return ReadFromReader(ar.file, offset)
}

func (ar *AppendReader) Close() error {
	return ar.file.Close()
}

// Reads specific offset. returns data, nextOffset, error. You can
// ReadFromReader(nextOffset) if you want to read the next document, or
// use the Scan() helper
func ReadFromReader(reader io.ReaderAt, offset uint32) ([]byte, uint32, error) {
	header := make([]byte, 16)
	_, err := reader.ReadAt(header, int64(offset*PAD))
	if err != nil {
		return nil, 0, err
	}

	if !bytes.Equal(header[8:12], MAGIC) {
		return nil, 0, EBADSLT
	}

	computedChecksumHeader := uint32(Hash(header[:12]))
	checksumHeader := binary.LittleEndian.Uint32(header[12:])
	if checksumHeader != computedChecksumHeader {
		return nil, 0, EBADSLT
	}

	metadataLen := binary.LittleEndian.Uint32(header)
	nextOffset := (offset + ((uint32(len(header))+(uint32(metadataLen)))+PAD-1)/PAD)
	readInto := make([]byte, metadataLen)
	_, err = reader.ReadAt(readInto, int64(offset*PAD)+int64(len(header)))
	if err != nil {
		return nil, 0, err
	}
	checksumHeaderData := binary.LittleEndian.Uint32(header[4:])
	computedChecksumData := uint32(Hash(readInto))

	if checksumHeaderData != computedChecksumData {
		return nil, 0, EBADSLT
	}
	return readInto, nextOffset, nil
}

// Scan ReaderAt, if the callback returns error this error is returned as the Scan error
func ScanFromReader(reader io.ReaderAt, offset uint32, cb func(uint32, []byte) error) error {
	for {
		data, next, err := ReadFromReader(reader, offset)
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
		err = cb(next, data)
		if err != nil {
			return err
		}
		offset = next
	}
}
