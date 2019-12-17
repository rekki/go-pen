package append

import (
	"encoding/binary"
	"fmt"
	"os"
)

type OffsetWriter struct {
	fd *os.File
}

// Creates new OffsetWriter, used to store one integer in a file with a checksum.
// Usually used for when you want to store your position in a log file that you want to scan from certain offset.
// example:
//	ow, err := NewOffsetWriter(filename)
//	if err != nil {
//		panic(err)
//	}
//	offset, err := ow.ReadOrDefault(0) // whatever the stored value is or 0
//	if err == nil {
//		panic(err)
//	}
func NewOffsetWriter(fn string) (*OffsetWriter, error) {
	state, err := os.OpenFile(fn, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	return &OffsetWriter{
		fd: state,
	}, nil
}

func (ow *OffsetWriter) Close() error {
	return ow.fd.Close()
}

// Read the offset or if empty/corrupt/not existant return default value
func (ow *OffsetWriter) ReadOrDefault(def int64) (int64, error) {
	storedOffset := make([]byte, 16)
	var offset int64
	_, err := ow.fd.ReadAt(storedOffset, 0)
	if err != nil {
		/*
			 FIXME:
				if err != io.EOF {
					ow.l.WithError(err).Warnf("failed to read bytes, setting offset to %d, err: %s", def, err)
				}
		*/
		offset = def
	} else {
		offset = int64(binary.LittleEndian.Uint64(storedOffset))
		sum := binary.LittleEndian.Uint64(storedOffset[8:])
		expected := Hash(storedOffset[:8])
		if expected != sum {
			return 0, fmt.Errorf("bad checksum expected: %d, got %d, bytes: %v", expected, sum, storedOffset)
		}
	}
	return offset, nil
}

// Write the offsed and its checksum
// format is:
//   8 byte LE offset
//   8 byte LE checksum // go-metro(offset bytes)
//
func (ow *OffsetWriter) SetOffset(offset int64) error {
	storedOffset := make([]byte, 16)
	binary.LittleEndian.PutUint64(storedOffset[0:], uint64(offset))
	offsetBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(offsetBytes, uint64(offset))
	binary.LittleEndian.PutUint64(storedOffset[8:], Hash(offsetBytes))
	_, err := ow.fd.WriteAt(storedOffset, 0)
	return err
}
