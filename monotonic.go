package pen

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// completely thread unsafe
// lock accordingly
type Monotonic struct {
	indexFD           *os.File
	dataFD            *os.File
	current           uint64
	currentDataOffset uint64
}

func NewMonotonic(fn string) (*Monotonic, error) {
	dataFD, err := os.OpenFile(fmt.Sprintf("%s.data", fn), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	indexFD, err := os.OpenFile(fmt.Sprintf("%s.index", fn), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return NewMonotonicFromFile(indexFD, dataFD)

}

func NewMonotonicFromFile(indexFD, dataFD *os.File) (*Monotonic, error) {
	currentDataOffset, err := dataFD.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}

	current, err := FixedLen(indexFD, 8)
	if err != nil {
		return nil, err
	}

	return &Monotonic{indexFD: indexFD, dataFD: dataFD, currentDataOffset: uint64(currentDataOffset), current: current}, nil
}

func (m *Monotonic) AppendAt(index uint64, b []byte) error {
	// append in data
	actualSize := len(b) + 16

	currentDataOffset := m.currentDataOffset
	m.currentDataOffset += uint64(actualSize)

	if index > m.current {
		m.current = index
	}

	err := WriteAtWriter64(m.dataFD, currentDataOffset, b)
	if err != nil {
		return err
	}

	o := make([]byte, 8)
	binary.LittleEndian.PutUint64(o, currentDataOffset)

	err = FixedWriteAt(m.indexFD, index, o)
	if err != nil {
		return err
	}
	return nil
}

func (m *Monotonic) Append(b []byte) (uint64, error) {
	current := m.current
	m.current++
	err := m.AppendAt(current, b)
	if err != nil {
		return 0, err
	}
	return current, nil
}

func (m *Monotonic) Last() ([]byte, error) {
	if m.current == 0 {
		return nil, io.EOF
	}
	return m.Read(m.current - 1)
}

func (m *Monotonic) MustLast() []byte {
	b, err := m.Read(m.current - 1)
	if err != nil {
		panic(err)
	}
	return b
}

func (m *Monotonic) MustAppend(b []byte) uint64 {
	d, err := m.Append(b)
	if err != nil {
		panic(err)
	}
	return d
}

func (m *Monotonic) MustRead(id uint64) []byte {
	d, err := m.Read(id)
	if err != nil {
		panic(err)
	}
	return d
}

// This function is super racy, if you are going to use it protect the whole Monotonic object with a lock
// it truncates two files the requested id
func (m *Monotonic) TruncateAt(id uint64) error {
	o := make([]byte, 8)
	err := FixedReadAt(m.indexFD, id, o)
	if err != nil {
		return err
	}
	dataOffset := binary.LittleEndian.Uint64(o)
	data, err := ReadFromReader64(m.dataFD, dataOffset, 16)
	if err != nil {
		return err
	}

	m.current = id
	m.currentDataOffset = dataOffset + 16 + uint64(len(data))

	err = m.indexFD.Truncate(int64(id) * int64(8+8)) // 8 header 8 data
	if err != nil {
		return err
	}
	// now it is safe to trucate the data

	err = m.dataFD.Truncate(int64(dataOffset + 16 + uint64(len(data))))
	if err != nil {
		return err
	}

	return nil
}

func (m *Monotonic) Read(id uint64) ([]byte, error) {
	o := make([]byte, 8)
	err := FixedReadAt(m.indexFD, id, o)
	if err != nil {
		return nil, err
	}
	off := binary.LittleEndian.Uint64(o)
	data, err := ReadFromReader64(m.dataFD, off, 16)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Monotonic) Count() uint64 {
	return m.current
}

func (m *Monotonic) Sync() error {
	err1 := m.dataFD.Sync()
	err2 := m.indexFD.Sync()

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil

}
func (m *Monotonic) Close() error {
	err1 := m.dataFD.Close()
	err2 := m.indexFD.Close()

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}
