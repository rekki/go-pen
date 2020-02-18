package pen

import (
	"encoding/binary"
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
func (ow *OffsetWriter) ReadOrDefault(def int64) int64 {
	storedOffset := make([]byte, 8)
	var offset int64
	err := FixedReadAt(ow.fd, 0, storedOffset)
	if err != nil {
		offset = def
	} else {
		offset = int64(binary.LittleEndian.Uint64(storedOffset))
	}
	return offset
}

// Write the offsed and its checksum
func (ow *OffsetWriter) SetOffset(offset int64) error {
	storedOffset := make([]byte, 8)
	binary.LittleEndian.PutUint64(storedOffset, uint64(offset))
	return FixedWriteAt(ow.fd, 0, storedOffset)
}
