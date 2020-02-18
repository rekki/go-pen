package pen

import (
	"bytes"
	"encoding/binary"
	"io"
)

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
func WriteAtWriter64(file io.WriterAt, offset uint64, encoded []byte) error {
	blobSize := 16 + len(encoded)
	blob := make([]byte, blobSize)
	copy(blob[16:], encoded)
	binary.LittleEndian.PutUint32(blob[0:], uint32(len(encoded)))
	binary.LittleEndian.PutUint32(blob[4:], uint32(Hash(encoded)))
	copy(blob[8:], MAGIC)
	binary.LittleEndian.PutUint32(blob[12:], uint32(Hash(blob[:12])))

	_, err := file.WriteAt(blob, int64(offset))
	if err != nil {
		return err
	}
	return nil
}

func ReadFromReader64(reader io.ReaderAt, offset uint64, blockSize int) ([]byte, error) {
	block := make([]byte, blockSize)
	n, err := reader.ReadAt(block, int64(offset))

	// end of file, or not enough space to read whole block_size
	if n < 16 {
		return nil, err
	}
	if n != blockSize {
		block = block[:n]
	}

	header := block[:16]
	if !bytes.Equal(header[8:12], MAGIC) {
		return nil, EBADSLT
	}

	computedChecksumHeader := uint32(Hash(header[:12]))
	checksumHeader := binary.LittleEndian.Uint32(header[12:16])
	if checksumHeader != computedChecksumHeader {
		return nil, EBADSLT
	}

	metadataLen := binary.LittleEndian.Uint32(header)

	var readInto []byte
	if int(metadataLen) < len(block)-len(header) {
		readInto = block[len(header) : len(header)+int(metadataLen)]
	} else {
		readInto = make([]byte, metadataLen)
		_, err = reader.ReadAt(readInto, int64(offset)+int64(len(header)))
		if err != nil {
			return nil, err
		}
	}

	checksumHeaderData := binary.LittleEndian.Uint32(header[4:])
	computedChecksumData := uint32(Hash(readInto))

	if checksumHeaderData != computedChecksumData {
		return nil, EBADSLT
	}
	return readInto, nil
}
