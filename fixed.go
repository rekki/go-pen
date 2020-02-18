package pen

import (
	"encoding/binary"
	"os"
)

const FixedHeaderSize = 8

// Write at specific index
// format is:
//   8 byte LE checksum // go-metro(data)
//   data:
//   ...
//   ...
//   fixed size data
func FixedWriteAt(file *os.File, index uint64, encoded []byte) error {
	blobSize := FixedHeaderSize + len(encoded)
	blob := make([]byte, blobSize)
	copy(blob[FixedHeaderSize:], encoded)

	binary.LittleEndian.PutUint64(blob, Hash(encoded))
	_, err := file.WriteAt(blob, int64(index*uint64(blobSize)))
	if err != nil {
		return err
	}
	return nil
}

// Calculate the amount of objects based on the given fixed size
func FixedLen(file *os.File, fixedSize uint64) (uint64, error) {
	s, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return uint64(s.Size() / int64(fixedSize+FixedHeaderSize)), nil
}

// Read from specific index
func FixedReadAt(file *os.File, index uint64, into []byte) error {
	blockSize := len(into) + FixedHeaderSize
	block := make([]byte, blockSize)
	_, err := file.ReadAt(block, int64(index*uint64(blockSize)))
	if err != nil {
		return err
	}

	computedChecksumHeader := Hash(block[FixedHeaderSize:])
	checksumHeader := binary.LittleEndian.Uint64(block)
	if checksumHeader != computedChecksumHeader {
		return EBADSLT
	}
	copy(into, block[FixedHeaderSize:])
	return nil
}
